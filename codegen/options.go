package codegen

import (
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

// DefaultFuncMap ...
var DefaultFuncMap = template.FuncMap{
	"ToLower":          strings.ToLower,
	"ToPlural":         pluralizer.Plural,
	"ToCamelCase":      toCamelCase,
	"ToLowerCamelCase": toLowerCamelCase,
	"ToFieldName":      toFieldName,
	"Type":             convertType,
	"Nullable":         isNillable,
}

// Option is a functional option used for dynamically configuring an instance of Seed
type Option func(*Seed)

var defaultOptions = []Option{
	WithFormatting(&imports.Options{
		TabWidth:  8,
		TabIndent: true,
		Comments:  true,
		Fragment:  true,
	}),
	WithFuncMap(DefaultFuncMap),
}

// WithOutputDirectory ...
func WithOutputDirectory(outputDirectory string) Option {
	return func(seed *Seed) {
		seed.OutputDirectory = outputDirectory
	}
}

// WithPackageName ...
func WithPackageName(packageName string) Option {
	return func(seed *Seed) {
		seed.PackageName = packageName
	}
}

// WithPackageImports ...
func WithPackageImports(packageImports []string) Option {
	return func(seed *Seed) {
		seed.PackageImports = packageImports
	}
}

// WithDefinition ...
func WithDefinition(definer Definer) Option {
	return func(seed *Seed) {
		seed.Definition = definer
	}
}

// WithDistinctSubpackages ...
func WithDistinctSubpackages(distinctSubpackages []string) Option {
	return func(seed *Seed) {
		seed.DistinctSubpackages = distinctSubpackages
	}
}

// WithTemplateMap ...
func WithTemplateMap(templateMap map[string]*template.Template) Option {
	return func(seed *Seed) {
		seed.TemplateMap = templateMap
	}
}

// WithData ...
func WithData(data interface{}) Option {
	return func(seed *Seed) {
		seed.Data = data
	}
}

// WithDistinctData ...
func WithDistinctData(distinctData map[string]interface{}) Option {
	return func(seed *Seed) {
		seed.DistinctData = distinctData
	}
}

// WithFormatting ...
func WithFormatting(opts *imports.Options) Option {
	return func(seed *Seed) {
		seed.FormattingOptions = opts
	}
}

// WithFuncMap ...
func WithFuncMap(funcMap template.FuncMap) Option {
	return func(seed *Seed) {
		seed.FuncMap = funcMap
	}
}
