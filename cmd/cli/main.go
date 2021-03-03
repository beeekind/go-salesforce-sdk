package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/b3ntly/salesforce/client"
	"github.com/b3ntly/salesforce/codegen"
	"github.com/b3ntly/salesforce/metadata"
	"github.com/b3ntly/salesforce/requests"
	"github.com/b3ntly/salesforce/soql"
	"github.com/b3ntly/salesforce/types"
)

var (
	id            = os.Getenv("SALESFORCE_SDK_CLIENT_ID")
	secret        = os.Getenv("SALESFORCE_SDK_CLIENT_SECRET")
	username      = os.Getenv("SALESFORCE_SDK_USERNAME")
	password      = os.Getenv("SALESFORCE_SDK_PASSWORD")
	securityToken = os.Getenv("SALESFORCE_SDK_SECURITY_TOKEN")
)

var sender = client.Must()
var req = requests.Base.Sender(sender)
var objectDocumentationTmpl = "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_%s.htm"
var toolingDocumentationTmpl = "https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_%s.htm"

// warm the cache

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	println("hydrating...")
	panicIfErr(hydrate())
	println("hydrated...")

	definer := &ObjectDefinition{Client: sender}
	seed, err := codegen.From(definer)

	panicIfErr(err)

	err = codegen.Generate(seed)
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
	_, err := req.
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
		println(6, entity.Name, err.Error())
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
	_, err := req.
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

func hydrate() error {
	fp := "../../chromedp/all"
	files, err := ioutil.ReadDir(fp)
	if err != nil {
		return err
	}

	for _, f := range files {
		contents, err := os.ReadFile(filepath.Join(fp, f.Name()))
		if err != nil {
			println(1, f.Name())
			return err
		}

		reader, err := gzip.NewReader(bytes.NewBuffer(contents))
		if err != nil {
			println(2, f.Name())
			return err
		}

		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			println(3, f.Name())
			return err
		}

		referenceStruct, err := codegen.FromHTML(doc)
		if err != nil {
			println(4, f.Name())
			return err
		}

		parts := strings.Split(f.Name(), ".")
		objectName := parts[0]
		referenceCache.set(objectName, codegen.Structs{referenceStruct})
	}

	return nil
}
