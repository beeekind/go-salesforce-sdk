package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/codegen"
	"github.com/beeekind/go-salesforce-sdk/metadata"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/beeekind/go-salesforce-sdk/types"
)

var objectDocumentationTmpl = "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_%s.htm"
var toolingDocumentationTmpl = "https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_%s.htm"
var invalidCommandText = "expected command 'ls' or 'generate {objectName}' run command with --help for more info"

var helpText = `
Welcome to the go-salesforce-sdk CLI!

This CLI depends on the default authentication methods used by Salesforce.DefaultClient. 
To function properly set the following environment variables:

For the JWT flow (recommended):

https://mannharleen.github.io/2020-03-03-salesforce-jwt/

SALESFORCE_SDK_CLIENT_ID 
SALESFORCE_SDK_USERNAME
SALESFORCE_SDK_PEM_PATH

For the Password Flow:

SALESFORCE_SDK_CLIENT_ID
SALESFORCE_SDK_CLIENT_SECRET
SALESFORCE_SDK_USERNAME
SALESFORCE_SDK_PASSWORD
SALESFORCE_SDK_SECURITY_TOKEN

There are currently two commands:

---
ls 
---

list salesforce objects for your org. Takes no arguments.

---
generate {objectName|string} {outputPath|string} {packageName|string} {relationshipLevel|int}
---

generate type definitions for the given salesforce object i.e. {Lead, User, Account}. 
relationshipLevel denotes how many relations should be generated. Because salesforce objects
are so deeply connected we suggest 0 or 1 for this value.

Examples:

go-salesforce-sdk ls 
go-salesforce-sdk generate Lead ./ leads 0
go-salesforce-sdk generate Account /path/to/desired/output accounts 1
`

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(invalidCommandText)
	}

	for _, arg := range os.Args {
		switch arg {
		case "help", "--help", "-help":
			helpCommand()
			return
		}
	}

	switch os.Args[1] {
	case "ls":
		lsCommand()
	case "generate":
		generateCommand()
	default:
		fmt.Println(invalidCommandText)
	}
}

func lsCommand() {
	results, err := salesforce.SObjects()
	if err != nil {
		fmt.Printf("Error retrieving salesforce objects: %s\n", err.Error())
		return
	}

	if len(results.Sobjects) == 0 {
		fmt.Println("No Salesforce objects returned. Has salesforce.DefaultClient authenticated properly?")
		return
	}

	for _, sobject := range results.Sobjects {
		fmt.Println(sobject.Name)
	}
}

func generateCommand() {
	if len(os.Args) < 6 {
		fmt.Printf("Too few arguments to 'generate' command expected 6 got %v - use --help for more", len(os.Args))
		return
	}

	objectName := os.Args[2]
	outputPath := os.Args[3]
	outputPackageName := os.Args[4]
	recursionLevel, err := strconv.Atoi(os.Args[5])
	if err != nil {
		fmt.Println("Generate expects its final argument to be an integer indicating how many levels of relations should be generated")
		return
	}

	definer := &ObjectDefinition{
		Client:         salesforce.DefaultClient,
		ObjectName:     objectName,
		OutputPath:     outputPath,
		PackageName:    outputPackageName,
		RecursionLevel: recursionLevel,
	}
	seed, err := codegen.From(definer)

	panicIfErr(err)

	err = codegen.Generate(seed)
	panicIfErr(err)
}

func helpCommand() {
	fmt.Println(helpText)
}

func defineEntity(objectName string, recursionLevel int) (codegen.Structs, error) {
	println("defining entity", objectName)

	// describe root object and its list of named dependencies
	structs, err := describe(objectName, recursionLevel == 0)
	if err != nil {
		return nil, err
	}

	//
	root := structs[0]

	// if entity is entitydefinition append the description field
	// this is an edgecase
	if objectName == "EntityDefinition" {
		root.Properties = append(root.Properties, &codegen.Property{
			Name: "Description",
			Type: "string",
			Tag:  codegen.Tag{"json": []string{"Description"}},
		})
	}

	// retrieve the corresponding reference documentation for root object and its dependencies
	if err := reference(structs[0]); err != nil {
		return nil, err
	}

	// retrieve the corresponsing tooling query descriptions for root object and its dependencies
	if err := description(structs[0]); err != nil {
		return nil, err
	}

	//
	if recursionLevel == 0 {
		return structs, nil
	}

	// describe dependant objects
	nextRecursionLevel := recursionLevel - 1
	for _, dependant := range structs[0].Dependencies {
		relatedStructs, err := defineEntity(dependant, nextRecursionLevel)
		if err != nil {
			return nil, err
		}

		structs = append(structs, relatedStructs...)
	}

	// merge results
	structs = structs.Dedupe(true).Sort().ConvertNillable()

	return structs, nil
}

func describe(objectName string, ignoreRelations bool) (codegen.Structs, error) {
	if structs, exists := describeCache.get(objectName); exists {
		return structs, nil
	}

	var describe metadata.Describe
	uri := fmt.Sprintf("%s/%s/%s", "sobjects", objectName, "describe")
	_, err := requests.
		Sender(salesforce.DefaultClient).
		URL(uri).
		JSON(&describe)

	if err != nil {
		return nil, err
	}

	structs, err := codegen.FromDescribe(&describe, ignoreRelations)
	if err != nil {
		return nil, err
	}

	describeCache.set(objectName, structs)
	return structs, nil
}

func reference(entity *codegen.Struct) error {
	if structs, exists := referenceCache.get(entity.Name); exists {
		final := codegen.MergeDocumentation(*entity, *structs[0])
		entity.Documentation = final.Documentation
		entity.Properties = final.Properties
		return nil
	}

	objectURL := fmt.Sprintf(objectDocumentationTmpl, strings.ToLower(entity.Name))
	toolingURL := fmt.Sprintf(toolingDocumentationTmpl, strings.ToLower(entity.Name))

	doc, err := requests.ParseWebApp(objectURL)
	if err != nil {
		doc, err = requests.ParseWebApp(toolingURL)
		if err != nil {
			println(5, entity.Name, err.Error())
			return nil
		}
		err = nil
	}

	documentationStruct, err := codegen.FromHTML(doc)
	if err != nil {
		println(entity.Name, err.Error())
		return nil
	}

	final := codegen.MergeDocumentation(*entity, *documentationStruct)
	entity.Documentation = final.Documentation
	entity.Properties = final.Properties
	referenceCache.set(entity.Name, codegen.Structs{entity})
	return nil
}

func description(entity *codegen.Struct) error {
	if structs, exists := descriptionCache.get(entity.Name); exists {
		final := codegen.MergeDocumentation(*entity, codegen.Struct{Name: entity.Name, Documentation: structs[0].Documentation})
		entity.Documentation = final.Documentation
	}

	type entities struct {
		types.QueryResponse
		Records []*metadata.EntityDefinition `json:"records"`
	}

	builder := soql.
		Select("Description").
		From("EntityDefinition").
		Where(soql.Eq{"QualifiedApiName": entity.Name})

	var result entities
	_, err := requests.
		Sender(salesforce.DefaultClient).
		URL(fmt.Sprintf("%s/%s", "tooling", "query")).
		SQLizer(builder).
		JSON(&result)

	if err != nil {
		return err
	}

	if len(result.Records) != 1 {
		return fmt.Errorf("tooling query response returned %v records when %v was expected", len(result.Records), 1)
	}

	final := codegen.MergeDocumentation(*entity, codegen.Struct{Name: entity.Name, Documentation: result.Records[0].Description})
	entity.Documentation = final.Documentation
	descriptionCache.set(entity.Name, codegen.Structs{{Name: entity.Name, Documentation: result.Records[0].Description}})
	return nil
}
