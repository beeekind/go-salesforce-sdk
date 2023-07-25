package codegen

import (
	"fmt"
	"reflect"
	"sort"
)

// structs.go contains utilities for working with []Struct

// Structs allows us to define utility methods on a group of Struct objects
type Structs []*Struct

// Valid ...
func (s Structs) Valid() bool {
	for _, elem := range s {
		if ok := elem.Valid(); !ok {
			return false
		}
	}
	return true
}

// RemoveRelations returns a new set of structs with all polymorphic structs
// and all properties representing relationships removed
func (s Structs) RemoveRelations() Structs {
	var s2 Structs
	for i := 0; i < len(s); i++ {
		if s[i].IsPolymorphicModel {
			continue
		}

		s2 = append(s2, s[i].RemoveRelations())
	}

	return s2
}

// Dedupe ...
func (s Structs) Dedupe(overrideNulls bool) (results Structs) {
	return s

	set := make(map[string]*Struct, len(s))
	for _, entity := range s {
		// not all Structs are equivalent and we want to preserve the one with the greatest number of properties
		if previousObject, exists := set[entity.Name]; exists {
			if len(previousObject.Properties) > len(entity.ParentName) {
				if overrideNulls {
					previousObject.Properties = previousObject.Properties.Merge(
						previousObject.Properties.Merge(
							entity.Properties,
							MergePropertyDocumentation,
							false,
						),
						MergeOverrideNulls,
						false,
					)
				}

				set[entity.Name] = previousObject
				continue
			}

			if overrideNulls {
				entity.Properties = entity.Properties.Merge(
					entity.Properties.Merge(
						previousObject.Properties,
						MergePropertyDocumentation,
						false,
					),
					MergeOverrideNulls,
					false,
				)
			}
		}

		set[entity.Name] = entity
	}

	for _, v := range set {
		item := v
		results = append(results, item)
	}

	return results
}

// Sort ...
func (s Structs) Sort() Structs {
	for i := 0; i < len(s); i++ {
		sort.Slice(s[i].Properties, func(j, k int) bool {
			return s[i].Properties[j].Name < s[i].Properties[k].Name
		})
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Name < s[j].Name
	})

	return s
}

// DocSize ...
func (s Structs) DocSize() int {
	size := 0

	for _, item := range s {
		size += len(item.Documentation)

		for _, p := range item.Properties {
			size += len(p.Documentation)
		}
	}

	return size
}

// ConvertNillable ...
func (s Structs) ConvertNillable() Structs {
	for i := 0; i < len(s); i++ {
		s[i].Properties = s[i].Properties.ConvertNillable()
	}

	return s
}

// Merge merges two []*Struct objects via the function fn.
//
// This allows for merge behavior to change based on the source of the given []*Struct.
//
// For example code derived from the reference documentation may only wish to contribute its
// struct.Documentation property where as code derived from the tooling/query api should contriubte its struct.Type .
func (s Structs) Merge(s2 Structs, fn func(old, new Struct) Struct, includeDistinct bool) Structs {
	var final []*Struct

	set := make(map[string]*Struct)

	for i := 0; i < len(s); i++ {
		if _, exists := set[s[i].Name]; exists {
			item := fn(*set[s[i].Name], *s[i])
			set[s[i].Name] = &item
			continue
		}

		set[s[i].Name] = s[i]
	}

	for i := 0; i < len(s2); i++ {
		if _, exists := set[s2[i].Name]; !exists {
			if includeDistinct {
				set[s2[i].Name] = s2[i]
			}

			continue
		}

		item := fn(*set[s2[i].Name], *s2[i])
		set[s2[i].Name] = &item
	}

	for _, v := range set {
		final = append(final, v)
	}

	return final
}

// Properties ...
type Properties []*Property

// ConvertNillable ...
func (props Properties) ConvertNillable() (results Properties) {
	for i := 0; i < len(props); i++ {
		final := props[i]
		// alternative means of testing nillable strings.Contains(final.Documentation, "Nillable")
		if final.IsNillable {
			nillableDataType, _ := convertNillableType(final.Type)
			final.Type = nillableDataType
		}

		results = append(results, final)
	}

	return results
}

// Merge ...
func (props Properties) Merge(p2 Properties, fn func(old, new Property) Property, includeDistinct bool) (final Properties) {
	set := make(map[string]*Property)

	for _, prop := range props {
		if _, exists := set[prop.Name]; exists {
			item := fn(*set[prop.Name], *prop)
			set[prop.Name] = &item
			continue
		}

		set[prop.Name] = prop
	}

	for _, prop := range p2 {
		if _, exists := set[prop.Name]; !exists {
			if includeDistinct {
				set[prop.Name] = prop
			}

			continue
		}

		item := fn(*set[prop.Name], *prop)
		set[prop.Name] = &item
	}

	for _, v := range set {
		final = append(final, v)
	}

	return final
}

// MergeAll ...
func MergeAll(old, new Struct) Struct {
	final := old

	if new.DocumentationURL != "" {
		final.DocumentationURL = new.DocumentationURL
	}

	if new.Documentation != "" {
		if old.Documentation != "" {
			final.Documentation = fmt.Sprintf("%s\n//%s", old.Documentation, new.Documentation)
		} else {
			final.Documentation = new.Documentation
		}
	}

	final.Properties = final.Properties.Merge(new.Properties, MergeProperty, true)
	return final
}

// MergeDocumentation merges only the .Documentation property of two structs
func MergeDocumentation(old, new Struct) Struct {
	if new.Documentation == "" && new.DocumentationURL == "" {
		return old
	}

	final := old

	if new.DocumentationURL != "" {
		final.DocumentationURL = new.DocumentationURL
	}

	if new.Documentation != "" {
		if old.Documentation != "" {
			final.Documentation = fmt.Sprintf("%s\n//%s", old.Documentation, new.Documentation)
		} else {
			final.Documentation = new.Documentation
		}
	}

	final.Properties = old.Properties.Merge(new.Properties, MergePropertyDocumentation, false)
	return final
}

// MergeProperty returns a new Property which overrides
// fields of oldProperty if conditions are met:
// 1) newProperty.{{Field}} is not a zero value
// 2) newProperty.{{Field}} != oldProperty.{{Field}}
// 3) the case of merging a property.Tag is harder as a user may pass in
//    a variety of values with an unknown intent for how they should be merged,
//    for example a k,v of "": []string{} or a Tag{}. In these cases we will not
//    override existing values with a zero value "" or []string{}
// 4) the documentation property is concatenated if the old.Documentation != ""
func MergeProperty(old, new Property) Property {
	final := old

	if new.Documentation != "" {
		if old.Documentation == "" {
			final.Documentation = new.Documentation
		} else {
			final.Documentation = fmt.Sprintf("%s\n\t%s", old.Documentation, new.Documentation)
		}
	}

	if new.Type != "" && new.Type != old.Type {
		final.Type = new.Type
	}

	//
	if len(new.Tag) != 0 && !reflect.DeepEqual(new.Tag, old.Tag) {
		finalTag := old.Tag
		for k, v := range new.Tag {
			if k != "" && len(v) != 0 {
				finalTag[k] = v
			}
		}
		final.Tag = finalTag
	}

	return final
}

// MergePropertyDocumentation merges only the Documentation field of
// two Propeprty objects returning the remaining properties of the initial
// Property argument unchanged
func MergePropertyDocumentation(old, new Property) Property {
	final := old
	if new.Documentation != "" {
		if old.Documentation == "" {
			final.Documentation = new.Documentation
		} else {
			final.Documentation = fmt.Sprintf("%s\n\t%s", old.Documentation, new.Documentation)
		}
	}

	return final
}

// MergeOverrideNulls ...
func MergeOverrideNulls(old, new Property) Property {
	final := old

	if old.Type == "interface{}" || old.Type == "[]json.RawMessage" || old.Type == "[]interface{}" {
		if new.Type != "interface{}" && new.Type != "[]json.RawMessage" && new.Type != "[]interface{}" {
			final.Type = new.Type
		}
	}

	return final
}
