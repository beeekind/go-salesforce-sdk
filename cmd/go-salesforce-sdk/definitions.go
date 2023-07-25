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

// ObjectsDefinition ...
type ObjectsDefinition struct {
	Client         *client.Client
	Objects        []*codegen.Struct
	ObjectNames    []string
	OutputPath     string
	PackageName    string
	RecursionLevel int
	Relations      codegen.Structs
}

// Options ...
func (o *ObjectsDefinition) Options() ([]codegen.Option, error) {
	if len(o.ObjectNames) == 0 {
		return nil, errors.New("empty ObjectsDefinition.ObjectsNames")
	}

	if o.OutputPath == "" {
		return nil, errors.New("missing output path")
	}

	if o.PackageName == "" {
		return nil, errors.New("missing output package name")
	}

	entities := make(codegen.Structs, 0)

	for _, on := range o.ObjectNames {
		ens, err := defineEntity(on, o.RecursionLevel)
		if err != nil {
			return nil, err
		}

		entities = append(entities, ens...)
	}

	seenObjs := make(map[string]struct{})

outer:
	for _, entity := range entities {
		for _, on := range o.ObjectNames {
			if on != entity.Name {
				continue
			}

			if _, ok := seenObjs[on]; ok {
				continue outer
			}

			o.Objects = append(o.Objects, entity)
			continue outer
		}

		seenObjs[entity.Name] = struct{}{}
		o.Relations = append(o.Relations, entity)
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
