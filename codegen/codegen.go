package codegen

// package codegen contains primitives for generating go source code using a variety of data sources

import (
	"bytes"
	"errors"
	"fmt"
	"go/scanner"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

// Definer defines options for a seed, for example
// to query additional data from a rest API before rendering
// a seed.
type Definer interface {
	Options() ([]Option, error)
}

// Seed contains the structures needed to generate go source code 
type Seed struct {
	// OutputDirectory ...
	OutputDirectory string
	// PackageName ...
	PackageName string
	// SubPackageName string
	SubPackageName string
	// PackageImports ...
	PackageImports []string
	// Definition ...
	Definition Definer
	// DistinctSubpackages indicates each Objects should produce
	// a subpackage in a directory that matches the pluralized lowercase
	// name of the distinct package
	DistinctSubpackages []string
	// Options ...
	Options []Option
	// Templates is map linking the absolute file path of a go template with an output filename
	// for its generated result
	// k:outputFileName: template
	TemplateMap map[string]*template.Template
	// Data is the data passed to a go template, configured via options
	Data interface{}
	// Distinct data is data but for DistinctSubpackages
	DistinctData map[string]interface{}
	//
	FormattingOptions *imports.Options
	//
	FuncMap template.FuncMap
}

// Source is the output of a generated Seed
// k:absoluteFilePath v:fileContents
type Source map[string][]byte

// From ...
func From(definer Definer) (*Seed, error) {
	options, err := definer.Options()
	if err != nil {
		return nil, fmt.Errorf("defining options: %w", err)
	}

	return &Seed{Options: options}, nil
}

// Generate ...
func Generate(seeds ...*Seed) error {
	// Render() each seed
	for i := 0; i < len(seeds); i++ {
		seed := seeds[i]

		// 1) render the seed to a final []byte representing rendered source code
		src, err := Render(seed)
		if err != nil {
			return fmt.Errorf("Generate(): %w", err)
		}

		// 2) writeFile() each derived file
		for outputPath, outputContents := range src {
			if err := writeFile(outputPath, 0755, outputContents); err != nil {
				return fmt.Errorf("Generate(): %w", err)
			}
		}
	}

	return nil
}

// Render executes all templates in seed.TemplateMap.
//
// If DistinctObjects is not nil or empty then each template will be executed
// for each distinctObject and use DistinctData[distinctObject] for the template execution
// instead of seed.Data. This allows us to generate subpackages for each distinct object.
func Render(seed *Seed) (Source, error) {
	// 0) apply default options
	for i := 0; i < len(defaultOptions); i++ {
		defaultOptions[i](seed)
	}

	// 1) apply each option for the given seed
	for i := 0; i < len(seed.Options); i++ {
		seed.Options[i](seed)
	}

	// 2) parse templatefiles for seed
	idx := 0
	tmplPaths := make([]string, len(seed.TemplateMap))
	for absoluteTemplatePath := range seed.TemplateMap {
		tmplPaths[idx] = absoluteTemplatePath
		idx++
	}

	source := Source{}

	// 3) iterate through templates and prepare to execute each template to a buffer
	for outputFileName, templ := range seed.TemplateMap {
		var buffer bytes.Buffer 

		// 4) if we're not rendering distinct packages, execute the given template with seed.Data
		if len(seed.DistinctSubpackages) == 0 {
			if err := templ.ExecuteTemplate(&buffer, templ.Name(), seed); err != nil {
				return nil, fmt.Errorf("Render(): failed to execute template %s: %w", templ.Name(), err)
			}

			// 5) process the generated file using goimports to format and add package import declarations
			absoluteOutputFilepath, formattedCode, err := postProcess(seed.FormattingOptions, buffer.Bytes(), seed.OutputDirectory, seed.PackageName, outputFileName)
			if err != nil {
				return nil, fmt.Errorf("Render(): %w", err)
			}

			// 6) append result to sourcemap
			source[absoluteOutputFilepath] = formattedCode
			continue
		}

		// 7) else we are rendering distinct subpackages, execute the given template for each subpackage and use distinctData instead of data
		for _, objectName := range seed.DistinctSubpackages {
			var buffer bytes.Buffer 

			//
			seed.Data = seed.DistinctData[objectName]

			//
			seed.SubPackageName = strings.ToLower(pluralizer.Plural(objectName))

			if err := templ.ExecuteTemplate(&buffer, templ.Name(), seed); err != nil {
				return nil, fmt.Errorf("Render(): failed to execute template %s: %w", templ.Name(), err)
			}

			absoluteOutputFilepath, formattedCode, err := postProcess(seed.FormattingOptions, buffer.Bytes(), seed.OutputDirectory, seed.PackageName, seed.SubPackageName, outputFileName)
			if err != nil {
				return nil, fmt.Errorf("Render(): %w", err)
			}

			source[absoluteOutputFilepath] = formattedCode
		}
	}

	return source, nil
}

func getCodeTemplates(funcMap template.FuncMap, filepaths ...string) (*template.Template, error) {
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(filepaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return tmpl, nil
}

// writeFile is a utility method for creating a file on a fully qualified path
// writeFile will make any subdirectories not already present on the absolutePath
// writeFile will overwrite existing files on the given absolute path
func writeFile(absolutePath string, perm os.FileMode, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(absolutePath), perm); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	file, err := os.OpenFile(absolutePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}

	defer file.Close()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("file.Write: %w", err)
	}

	return nil
}

// postProcess does extra work to the result of an executed template, mainly
// applying goimports to the generated code
func postProcess(formattingOptions *imports.Options, src []byte, paths ...string) (absoluteOutputFilepath string, formattedCode []byte, err error) {
	absoluteOutputFilePath := path.Join(paths...)
	formattedCode, err = imports.Process(absoluteOutputFilePath, src, formattingOptions)
	if err != nil {
		var errList scanner.ErrorList
		if errors.As(err, &errList) {
			var builder strings.Builder
			for idx, e := range errList {
				builder.WriteString(fmt.Sprintf("[%v]: %s\n", idx, e.Msg))
			}
			writeFile("./output.go", 0770, src)
			return "", nil, fmt.Errorf("postProcess(): failed to apply goimports for file %s: %w", absoluteOutputFilePath, errors.New(builder.String()))
		}

		return "", nil, fmt.Errorf("postProcess(): failed to apply goimports for file %s: %w", absoluteOutputFilePath, err)
	}

	// writeFile("./output.go", 0770, formattedCode)
	return absoluteOutputFilePath, formattedCode, nil
}
