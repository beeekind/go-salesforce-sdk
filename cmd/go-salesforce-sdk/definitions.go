package main

import (
	"errors"
	"os"
	"text/template"

	"github.com/beeekind/go-salesforce-sdk/client"
	"github.com/beeekind/go-salesforce-sdk/codegen"
)

// definition.go contains utility functions for generating salesforce type
// definitions

var gopath = os.Getenv("GOPATH")

// ObjectDefinition ...
type ObjectDefinition struct {
	Client         *client.Client
	Object         *codegen.Struct
	ObjectName     string
	OutputPath     string
	PackageName    string
	RecursionLevel int
	Relations      codegen.Structs
}

// Options ...
func (o *ObjectDefinition) Options() ([]codegen.Option, error) {
	if o.ObjectName == "" {
		return nil, errors.New("missing ObjectDefinition.ObjectName")
	}

	if o.OutputPath == "" {
		return nil, errors.New("missing output path")
	}

	if o.PackageName == "" {
		return nil, errors.New("missing output package name")
	}

	entities, err := defineEntity(o.ObjectName, o.RecursionLevel)
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		if entity.Name == o.ObjectName {
			o.Object = entity
		} else {
			o.Relations = append(o.Relations, entity)
		}
	}

	return []codegen.Option{
		codegen.WithPackageName(o.PackageName),
		codegen.WithOutputDirectory(o.OutputPath),
		codegen.WithTemplateMap(map[string]*template.Template{
			"objects.go":   template.Must(template.New("objects.gohtml").Funcs(codegen.DefaultFuncMap).ParseFiles(gopath + "/src/github.com/beeekind/go-salesforce-sdk/templates/objects.gohtml")),
			"relations.go": template.Must(template.New("objects.relations.gohtml").Funcs(codegen.DefaultFuncMap).ParseFiles(gopath + "/src/github.com/beeekind/go-salesforce-sdk/templates/objects.relations.gohtml")),
			// "api.go": template.Must(template.New("objects.api.gohtml").Funcs(codegen.DefaultFuncMap).ParseFiles(GOPATH + "/src/github.com/beeekind/go-salesforce-sdk/templates/objects.api.gohtml")),
		}),
		codegen.WithData(o),
	}, nil
}
