package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/codegen"
	"github.com/beeekind/go-salesforce-sdk/metadata"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/beeekind/go-salesforce-sdk/types"
)

const (
	objectDocumentationTmpl  = "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_%s.htm"
	toolingDocumentationTmpl = "https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_%s.htm"
	invalidCommandText       = "expected command 'ls' or 'generate', run command with --help for more info"

	helpTextF = `
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
generate {objectName|strings} --workdir={outputPath|string} --package={packageName|string} --depth={relationshipLevel|int}

[Flags]:
%s
---

generate type definitions for the given salesforce object i.e. {Lead, User, Account}. 
relationshipLevel denotes how many relations should be generated. Because salesforce objects
are so deeply connected we suggest 0 or 1 for this value.

Examples:

go-salesforce-sdk ls
go-salesforce-sdk generate 
go-salesforce-sdk generate Lead --workdir=./ --package=leads --depth=0
go-salesforce-sdk generate Account --workdir=/path/to/desired/output --package=accounts --depth=1
`
)

type Config struct {
	UseReferenceDocs bool
	WorkDir          string
	Package          string
	Depth            uint

	ObjectNames []string
}

var (
	// init flagset for this cmd
	fs = flag.NewFlagSet("go-salesforce-sdk", flag.ContinueOnError)

	// runtime config, currently only used by "generate" command.
	conf Config
)

func helpCommand() {
	// create temp buffer
	buff := bytes.NewBuffer(nil)
	// set output to our buffer
	fs.SetOutput(buff)
	// print defaults to buffer
	fs.PrintDefaults()
	// print help text with output from buffer
	fmt.Printf(helpTextF, buff.String())
}

func init() {
	// override default help text printer
	fs.Usage = helpCommand

	// define flag vars

	fs.BoolVar(
		&conf.UseReferenceDocs,
		"use-reference-docs",
		false,
		"If provided, attempts to further annotate each generated model from the reference documentation HTML.",
	)
	fs.StringVar(
		&conf.WorkDir,
		"workdir",
		".",
		"Root working directory.  Package directories will be created under this path.",
	)
	fs.StringVar(
		&conf.Package,
		"package",
		"sfmodels",
		"Name of package to define generated models within.  Directory of same name will be created under --workdir.",
	)
	fs.UintVar(
		&conf.Depth,
		"depth",
		0,
		"Depth of relationship level to generate models for.",
	)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// parse flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		panicIfErr(err)
	}

	// ensure at least 1 arg is provided
	if fs.NArg() < 1 {
		fmt.Println(invalidCommandText)
		os.Exit(1)
	}

	// parse arg
	switch fs.Arg(0) {
	case "help":
		helpCommand()
	case "ls":
		lsCommand()
	case "generate":
		generateCommand()

	default:
		fmt.Println(invalidCommandText)
		os.Exit(1)
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

	// parse any / all object names
	for i := 1; i < fs.NArg(); i++ {
		str := strings.TrimSpace(fs.Arg(i))
		if len(str) == 0 {
			continue
		}
		conf.ObjectNames = append(conf.ObjectNames, str)
	}

	// ensure we have at least one object to fetch
	if len(conf.ObjectNames) == 0 {
		panic("must provide at least one object name to parse")
	}

	seeds := make([]*codegen.Seed, 0)

	// iterate through object names, generating code.
	for _, objName := range conf.ObjectNames {
		definer := &ObjectDefinition{
			Client:         salesforce.DefaultClient(),
			ObjectName:     objName,
			OutputPath:     conf.WorkDir,
			PackageName:    conf.Package,
			RecursionLevel: int(conf.Depth),
		}

		seed, err := codegen.From(definer)
		panicIfErr(err)

		seeds = append(seeds, seed)
	}

	err := codegen.Generate(seeds...)
	panicIfErr(err)
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

	// if configured to use references, retrieve the corresponding reference documentation for root object and
	// its dependencies
	if conf.UseReferenceDocs {
		if err := reference(structs[0]); err != nil {
			return nil, err
		}
	}

	// retrieve the corresponsing tooling query descriptions for root object and its dependencies
	if err := description(structs[0]); err != nil {
		return nil, err
	}

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
	var describe metadata.Describe
	uri := fmt.Sprintf("%s/%s/%s", "sobjects", objectName, "describe")
	_, err := requests.
		Sender(salesforce.DefaultClient()).
		URL(uri).
		JSON(&describe)

	if err != nil {
		return nil, err
	}

	structs, err := codegen.FromDescribe(&describe, ignoreRelations)
	if err != nil {
		return nil, err
	}

	return structs, nil
}

func reference(entity *codegen.Struct) error {
	objectURL := fmt.Sprintf(objectDocumentationTmpl, strings.ToLower(entity.Name))
	toolingURL := fmt.Sprintf(toolingDocumentationTmpl, strings.ToLower(entity.Name))

	doc, err := ParseWebApp(objectURL)
	if err != nil {
		doc, err = ParseWebApp(toolingURL)
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
	return nil
}

func description(entity *codegen.Struct) error {
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
		Sender(salesforce.DefaultClient()).
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
	return nil
}
