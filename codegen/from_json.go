package codegen

// json.go converts JSON to condegen.Struct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// FromJSON converts a JSON payload to its requisite Struct components. Nested objects will be separated into
// discrete Struct objects rather then nesting them like Matt Holt's JSON2go website. StructDocumentation
// refers to Struct.Documentation and will append comments above the Struct.
func FromJSON(structName string, structDocumentation string, JSON []byte) (results Structs, err error) {
	// 0) use of the reflect package can incur panics, this defer/recover statement captures that
	// and returns a more usable error message
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("recovered while converting JSON2go: %s", x)
			case error:
				err = fmt.Errorf("recovered while converting JSON2go: %w", err)
			default:
				// Fallback err (per specs, error strings should be lowercase w/o punctuation
				err = errors.New("unknown panic while converting JSON2go")
			}
		}
	}()

	// 1) unmarshal the JSON payload into a map[string]interface{}
	raw := map[string]interface{}{}
	if err := json.Unmarshal(JSON, &raw); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON to map[string]interface{}: %w", err)
	}

	// 2) iterate over each key in raw and inspect its type
	var properties Properties
	idx := 0
	for k, v := range raw {
		propertyName, err := prepareStructName(k, neither)
		if err != nil {
			return nil, err
		}

		prop := &Property{
			Name: propertyName,
			Tag:  jsonTag(k),
		}

		// it wouldn't be an enterprise rest api if it didn't have absurd edge cases
		if structName == "Field" && k == "defaultValue" {
			prop.Type = "types.AlmostBool"
			properties = append(properties, prop)
			idx++
			continue
		}

		reflectType := reflect.TypeOf(v)

		// 3) if the value is null we may return immediately
		if reflectType == nil {
			prop.Type, err = convertType("null")
			if err != nil {
				return nil, err
			}
			properties = append(properties, prop)
			idx++
			continue
		}

		// 4) inspect the type of the JSON value to derive its go type
		// if type is map then it is a nested object and should be handled as such
		// if type is slice then check if it is a primitive type of a slice of nested objects
		switch reflectType.Kind() {
		case reflect.Int:
			prop.Type, err = convertType("int")
		case reflect.Int64:
			prop.Type, err = convertType("int64")
		case reflect.Bool:
			prop.Type, err = convertType("bool")
		case reflect.String:
			prop.Type, err = convertType("string")

		case reflect.Float64:
			strRepr := fmt.Sprintf("%v", v)
			if strings.Contains(strRepr, ".") {
				prop.Type = "float64"
			} else {
				prop.Type = "int64"
			}

		case reflect.Map:
			// the name of a map becomes
			propertyName, err := prepareStructName(k, singular)
			if err != nil {
				return nil, err
			}

			prop.Name = propertyName
			prop.Type = fmt.Sprintf("*%s", propertyName)
			contents, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("marshalling nested map while converting JSON2Go: %w", err)
			}

			nestedStructs, err := FromJSON(propertyName, "", contents)
			if err != nil {
				return nil, fmt.Errorf("converting nested map while converting JSON2Go: %w", err)
			}

			// annotate nested objects with the name of their parent object
			for i := 0; i < len(nestedStructs); i++ {
				nestedStructs[i].ParentName = structName
			}

			results = append(results, nestedStructs...)

		case reflect.Slice:
			propertyNameSingular, err := prepareStructName(k, singular)
			if err != nil {
				return nil, err
			}

			prop.Type = fmt.Sprintf("[]*%s", propertyNameSingular)

			elements, ok := v.([]interface{})
			if !ok {
				return nil, fmt.Errorf("casting reflect.Slice to []interface{}")
			}

			// JSON represented as an empty slice "[]"
			if len(elements) == 0 {
				prop.Type, err = convertType("[]json.RawMessage")
				properties = append(properties, prop)
				idx++

				if structName == "Field" && prop.Name == "ReferenceTo" {
					// println(0, prop.Type)
				}
				continue
			}

			// identify if elements is a slice of objects
			_, ok = elements[0].(map[string]interface{})
			// if not we can immediately parse it into a slice of primitive value
			if !ok {
				// then its a slice of primitive types
				switch item := elements[0].(type) {
				case int:
					prop.Type = "[]int"
				case int32:
					prop.Type = "[]int32"
				case int64:
					prop.Type = "[]int64"
				case float64:
					if strings.Contains(fmt.Sprintf("%v", v), ".") {
						prop.Type = "[]float64"
					} else {
						prop.Type = "[]int64"
					}

				case string:
					prop.Type = "[]string"

				default:
					t := reflect.TypeOf(elements[0])
					return nil, fmt.Errorf("converting slice of primitive type: %s: %v", t.Kind().String(), item)
				}

				properties = append(properties, prop)
				idx++
				continue
			}

			// sometimes an array of objects will have overlapping types,
			// meaning that elements[0] may have a key of foo with value null
			// where as elements[1] may have a key of foo with value of "zar"
			// therefore we want to scan multiple entries in the array to get a more
			// comprehensive understanding of their underlying type
			var allSubstructs Structs
			for _, entry := range elements {
				obj := entry.(map[string]interface{})
				contents, err := json.Marshal(obj)
				if err != nil {
					return nil, fmt.Errorf("10 %s %s %w", structName, k, err)
				}

				propertyNameSingular, err := prepareStructName(k, singular)
				if err != nil {
					return nil, fmt.Errorf("20 %s %s %w", structName, k, err)
				}

				subStructs, err := FromJSON(propertyNameSingular, "", contents)
				for i := 0; i < len(subStructs); i++ {
					subStructs[i].ParentName = structName
				}

				allSubstructs = append(allSubstructs, subStructs...)
			}

			//println("900 deduping", k, structName)
			allSubstructs = allSubstructs.Dedupe(true)
			results = append(results, allSubstructs...)

		default:
			return nil, fmt.Errorf("unrecognized type converting JSON2go: %s", reflectType.Kind().String())
		}

		properties = append(properties, prop)
		idx++
	}

	if err != nil {
		return nil, err
	}

	results = append(results, &Struct{
		Name:          structName,
		Documentation: structDocumentation,
		Properties:    properties,
	})

	return results, nil
}
