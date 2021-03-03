package codegen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
)

var pluralizer = pluralize.NewClient()

// edgecaseDataTypes return hardcoded types for specific properties which are problematic to parse
var edgecaseDatatypes = map[string]string{
	"locale":                      "string",
	"Permissions<PermissionName>": "string",
}

var dataTypes = map[string]string{
	"null": "interface{}",
	// string-ish types
	"interface{}":                           "interface{}",
	"[]string":                              "[]string",
	"junctionidlist":                        "[]string",
	"string":                                "string",
	"bool":                                  "bool",
	"float64":                               "float64",
	"id":                                    "string",
	"lookup":                                "string",
	"reference":                             "string",
	"text":                                  "string",
	"textarea":                              "string",
	"combobox":                              "string",
	"complexvalue":                          "string",
	"apexclassmetadata":                     "interface{}",
	"url":                                   "string",
	"email":                                 "string",
	"roll-up summary (sum invoice line)":    "string",
	"phone":                                 "string",
	"picklist":                              "string",
	"manageablestate enumerated list":       "string",
	"actionemailsendertype enumerated list": "string",
	"encryptedstring":                       "string",
	"mns:workflowoutboundmessage":           "string",
	"mns:workflowtask":                      "string",
	"mns:workflowrule":                      "string",
	"mns:workflowfieldupdate":               "string",
	"mns:usercriteria":                      "string",
	"mns:workskillrouting":                  "string",
	"mns:timesheettemplate":                 "string",
	"mns:compactlayout":                     "string",
	"mns: compactlayout":                    "string",
	"mns:customapplication":                 "string",
	"mns:workflowalert":                     "string",
	"mns:embeddedserviceconfig":             "string",
	"mns:embeddedservicefieldservice":       "string",
	"mns: customobject":                     "string",
	"mns:eventdelivery":                     "string",
	"mns:eventdescription":                  "string",
	"mns:eventsubscription":                 "string",
	"mns:embeddedserviceliveagent":          "string",
	"mns: flow":                             "string",
	"https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_embeddedservicebranding.htm": "string",
	"restricted picklist": "string",
	"multipicklist":       "string",
	// boolean-ish types
	"boolean":  "bool",
	"checkbox": "bool",
	// numeric-ish types
	"int":      "int",
	"currency": "float64",
	"number":   "int",
	"double":   "float64",
	"long":     "int64",
	"int64":    "int64",
	"integer":  "int",
	"decimal":  "float64",
	"percent":  "float64",
	// complex types
	"address":                           "types.Address",
	"date":                              "types.Date",
	"datetime":                          "types.Datetime",
	"date/time":                         "types.Datetime",
	"object":                            "json.RawMessage",
	"queryresult":                       "types.QueryResult",
	"anytype":                           "json.RawMessage",
	"types.datetime":                    "types.Datetime",
	"types.address":                     "types.Address",
	"types.date":                        "types.Date",
	"time.time":                         "time.Time",
	"time":                              "types.Datetime",
	"base64":                            "string",
	"json.rawmessage":                   "json.RawMessage",
	"any":                               "interface{}",
	"types.queryresult":                 "types.QueryResult",
	"symboltable":                       "json.RawMessage",
	"apexcomponentmetadata":             "json.RawMessage",
	"entitydefinition":                  "json.RawMessage",
	"fielddefinition":                   "json.RawMessage",
	"apexresult":                        "json.RawMessage",
	"heapdump":                          "json.RawMessage",
	"soqlresult":                        "json.RawMessage",
	"apexpagemetadata":                  "json.RawMessage",
	"executeanonymousresult":            "json.RawMessage",
	"apextriggermetadata":               "json.RawMessage",
	"brandingset":                       "json.RawMessage",
	"compactlayoutinfo":                 "json.RawMessage",
	"deploydetails":                     "json.RawMessage",
	"flexipage":                         "json.RawMessage",
	"customfieldmetadata":               "json.RawMessage",
	"customfield":                       "json.RawMessage",
	"customtabmetadata":                 "json.RawMessage",
	"customlabel":                       "json.RawMessage",
	"embeddedserviceconfig":             "json.RawMessage",
	"embeddedserviceflowconfig":         "json.RawMessage",
	"embeddedservicemenusettings":       "json.RawMessage",
	"publisher":                         "json.RawMessage",
	"relationshipreferenceto":           "json.RawMessage",
	"flexipagemetadata":                 "json.RawMessage",
	"flowdefinition":                    "json.RawMessage",
	"flow":                              "json.RawMessage",
	"customvalue[]":                     "json.RawMessage",
	"array of typeextent":               "json.RawMessage",
	"mns:inboundnetworkconnection":      "json.RawMessage",
	"mns:keywordlist":                   "json.RawMessage",
	"datacategorygroupreference":        "json.RawMessage",
	"mns:layout":                        "json.RawMessage",
	"mns:lightningcomponentbundle":      "json.RawMessage",
	"user":                              "json.RawMessage",
	"location":                          "json.RawMessage",
	"lookupfilter":                      "json.RawMessage",
	"mns: managedcontenttype":           "json.RawMessage",
	"mns:moderationrule":                "json.RawMessage",
	"operationparameters":               "json.RawMessage",
	"mns:outboundnetworkconnection":     "json.RawMessage",
	"subscriberpackageinstallerrors":    "json.RawMessage",
	"msn:pathassistant":                 "json.RawMessage",
	"querylocator":                      "json.RawMessage",
	"mns: recommendationstrategy":       "json.RawMessage",
	"mns:recordactiondeploymentchannel": "json.RawMessage",
	"relationshipinfo":                  "json.RawMessage",
	"queryresultmetadata":               "json.RawMessage",
	"searchlayoutbuttonsdisplayed":      "json.RawMessage",
	"standardvalue[]":                   "json.RawMessage",
	"subscriberpackagecsptrustedsites":  "json.RawMessage",
	"properties":                        "json.RawMessage",
	"array of constructor":              "json.RawMessage",
	"authprovider":                      "json.RawMessage",
	"validationrule metadata":           "json.RawMessage",
	"any type":                          "json.RawMessage",
	"recordtypessupported":              "json.RawMessage",
	"datatype":                          "json.RawMessage",
	"mns:flowdefinition":                "json.RawMessage",
	"ratelimittimeperiod (enumeration of type string)": "json.RawMessage",
	"subscriberpackageprofilemappings":                 "json.RawMessage",
	"sobject":                                          "json.RawMessage",
	"mns:recordactiondeploymentcontext":                "json.RawMessage",
	"array of mapvalue":                                "json.RawMessage",
	"searchlayoutfieldsdisplayed":                      "json.RawMessage",
	"subscriberpackagedependencies":                    "json.RawMessage",
	"array of externalreference":                       "json.RawMessage",
	"userentityaccess":                                 "json.RawMessage",
	"moderationruletype (enumeration of type string)":  "json.RawMessage",
	"mns:recordactiondeployment":                       "json.RawMessage",
	"raw":                                              "json.RawMessage",
	"array of symboltable":                             "json.RawMessage",
	"mns:recordactionrecommendation":                   "json.RawMessage",
	"subscriberpackageprofiles":                        "json.RawMessage",
	"array of string":                                  "json.RawMessage",
	"mns:recordactionselectableitem":                   "json.RawMessage",
	"subscriberpackageremotesitesettings":              "json.RawMessage",
	"array of method":                                  "json.RawMessage",
	"array of visibilitysymbol":                        "json.RawMessage",
	"array of symbol":                                  "json.RawMessage",
	//
	"[]*field": "[]*Field",
	"[]int64":  "[]int64",
	"*tar":     "*Tar",
	"*zar":     "*Zar",
	"types.AlmostBool": "types.AlmostBool",
}

var nillableDataTypes = map[string]string{
	"int": "types.NullableInt",
	"bool": "types.NullableBool",
	"string": "types.NullableString",
	"float64": "types.NullableFloat64",
}

// prepareDatatype checks for property.Name specific overrides before calling convertType to
// produce a datatype for the given property
func prepareDatatype(propertyName, rawType string) (string, error) {
	if rawType == "" {
		potentialType, exists := edgecaseDatatypes[strings.ToLower(propertyName)]
		if exists {
			return potentialType, nil
		}
	}

	return convertType(rawType)
}

// convertType attempts to convert a string representation of a salesforce datatype
// to a string representation of a golang datatype. Salesforce datatypes may be retrieved from 3
// locations:
//
// 1) the DataType field of a FieldDefinition from the tooling/query API
// 2) the HTML table within the object and tooling reference documentation pages
// 3) the HTML table within the field reference documentation page
//
// importantly these 3 sources use different string representations for the same type. for example,
// the object reference docs may call a foreign-key-like datatype a "reference" and the field reference docs may
// call it an "id". For that reason this method is prone to faults and should error hard, forcing users
// to ensure proper usage and outputs.
//
// Also there is at least one instance of a typo that suggests much of the reference documentation types
// was handwritten or adlibbed.
func convertType(str string) (string, error) {
	if str == "" {
		return "", fmt.Errorf("convertingType '%s' (empty string)", str)
	}

	dataType, exists := dataTypes[strings.ToLower(strings.TrimSpace(str))]
	if exists {
		return dataType, nil
	}

	return str, fmt.Errorf("convertingType '%s'", strings.ToLower(str))
}

// convertNullableType returns a custom type extending convertType
func convertNillableType(str string) (string, error) {
	dataType, err := convertType(str)
	if err != nil {
		return "", err
	}

	nillableDataType, exists := nillableDataTypes[dataType]
	if exists {
		return nillableDataType, nil 
	}

	return str, fmt.Errorf("convertingNillableType '%s'", dataType)
}

// commonInitialisms is a list of common initialisms.
// This list should contain a subset of `golang.org/x/lint/golint`.
var commonInitialisms = [][]string{
	{"ACL", "Acl"},
	{"API", "Api"},
	{"ASCII", "Ascii"},
	{"CPU", "Cpu"},
	{"CSS", "Css"},
	{"DNS", "Dns"},
	{"EOF", "Eof"},
	{"GUID", "Guid"},
	{"HTML", "Html"},
	{"HTTP", "Http"},
	{"HTTPS", "Https"},
	{"ID", "Id"},
	{"IP", "Ip"},
	{"JSON", "Json"},
	{"LHS", "Lhs"},
	{"QPS", "Qps"},
	{"RAM", "Ram"},
	{"RHS", "Rhs"},
	{"RPC", "Rpc"},
	{"SLA", "Sla"},
	{"SMTP", "Smtp"},
	{"SObject", "Sobject"},
	{"SObjects", "Sobjects"},
	{"SQL", "Sql"},
	{"SSH", "Ssh"},
	{"TCP", "Tcp"},
	{"TLS", "Tls"},
	{"TTL", "Ttl"},
	{"UDP", "Udp"},
	{"UI", "Ui"},
	{"UID", "Uid"},
	{"UUID", "Uuid"},
	{"URI", "Uri"},
	{"URL", "Url"},
	{"UTF8", "Utf8"},
	{"VM", "Vm"},
	{"XML", "Xml"},
	{"XMPP", "Xmpp"},
	{"XSRF", "Xsrf"},
	{"XSS", "Xss"},
}

// ConvertInitialisms returns a string converted to Go case.
func convertInitialisms(s string) string {
	for i := 0; i < len(commonInitialisms); i++ {
		s = strings.ReplaceAll(s, commonInitialisms[i][1], commonInitialisms[i][0])
	}
	return s
}

// RevertInitialisms returns a string converted from Go case to normal case.
func revertInitialisms(s string) string {
	for i := 0; i < len(commonInitialisms); i++ {
		s = strings.ReplaceAll(s, commonInitialisms[i][0], commonInitialisms[i][1])
	}
	return s
}

var keywords = map[string]string{
	"break":       "salesforce_break",
	"default":     "salesforce_default",
	"func":        "salesforce_func",
	"interface":   "salesforce_interface",
	"select":      "salesforce_select",
	"case":        "salesforce_case",
	"defer":       "salesforce_defer",
	"go":          "salesforce_go",
	"map":         "salesforce_map",
	"struct":      "salesforce_struct",
	"chan":        "salesforce_chan",
	"else":        "salesforce_else",
	"goto":        "salesforce_goto",
	"package":     "salesforce_package",
	"switch":      "salesforce_switch",
	"const":       "salesforce_const",
	"fallthrough": "salesforce_fallthrough",
	"if":          "salesforce_if",
	"range":       "salesforce_range",
	"type":        "salesforce_type",
	"continue":    "salesforce_continue",
	"for":         "salesforce_for",
	"import":      "salesforce_import",
	"return":      "salesforce_return",
	"var":         "salesforce_var",
}

// stripReservedKeywords ensures no Golang keywords will be used in a package name
func stripReservedKeywords(str string) string {
	if sanitizedValue, exists := keywords[str]; exists {
		return sanitizedValue
	}

	return str
}

func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := initCase
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if capNext {
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		} else if i == 0 {
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		}
		if vIsCap || vIsLow {
			n.WriteByte(v)
			capNext = false
		} else if vIsNum := v >= '0' && v <= '9'; vIsNum {
			n.WriteByte(v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}

// ToCamelCase converts a string to CamelCase
func toCamelCase(s string) string {
	return toCamelInitCase(s, true)
}

// ToLowerCamelCase converts a string to lowerCamelCase
func toLowerCamelCase(s string) string {
	return toCamelInitCase(s, false)
}

func toFieldName(str string) string {
	return convertInitialisms(toCamelCase(str))
}

// enforceLineLimit adds line breaks to a docstring that exceeds lineLimit
func enforceLineLimit(str string, lineLimit int) string {
	cummulativeLength := 0
	var parts []string
	for _, elem := range strings.Split(str, " ") {
		cummulativeLength += len(elem)
		if cummulativeLength > lineLimit {
			cummulativeLength = 0
			parts = append(parts, "\n\t//")
		}

		parts = append(parts, elem)
	}

	return strings.Join(parts, " ")
}

// stripNewLinesAndTabs ...
func stripNewLinesAndTabs(str string) string {
	str = strings.ReplaceAll(str, "\n", " ")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, string('\u200B'), "")
	str = strings.ReplaceAll(str, string('\u200D'), "")
	str = strings.ReplaceAll(str, string('\u2014'), "")
	return strings.Join(strings.Fields(str), " ")
}

type field struct {
	Createable bool
	IsNillable bool
}

// RequiredFields ...
func requiredFields(fields []*field) (results []*field) {
	for i := 0; i < len(fields); i++ {
		if fieldIsRequired(fields[i]) {
			results = append(results, fields[i])
		}
	}

	return results
}

func fieldIsRequired(f *field) bool {
	// f.Required is not an actual field of Salesforce but a field this library sets on the
	// SObjectDescribeField struct so that we may explicitly assert a field is required as a result
	// of finding the string "required" within Salesforce's Object Reference documentation:
	// i.e. : https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_lead.htm
	return f.Createable && !f.IsNillable // && !f.DefaultedOnCreate)
}

// agreement https://en.wikipedia.org/wiki/Agreement_(linguistics)
type agreement int

const (
	neither agreement = iota
	singular
	plural
)

// RelationshipName ...
func RelationshipName(name string) (string, error) {
	return prepareStructName(name, plural)
}

var (
	// ErrInvalidPropertyName ...
	ErrInvalidPropertyName = errors.New("property name contains invalid characters")
	// ErrPropertyKeyBlank ...
	ErrPropertyKeyBlank = errors.New("property name must not be empty string")

	protectedChars = map[rune]struct{}{
		' ':  {},
		',':  {},
		'\n': {},
		'-':  {},
		'.':  {},
		// U+200B is a 0-width space character that should be thrown into Mordor where it was probably forged
		'\u200B': {},
		'\u200D': {},
		'\u2014': {},
		'\u003C': {},
		'>':      {},
	}
)

// prepareStructName ...
func prepareStructName(propertyKey string, pluralization agreement) (string, error) {
	// 0) check for empty string
	if propertyKey == "" {
		return "", ErrPropertyKeyBlank
	}

	// 1a) check for a common typo i.e. "Postal Code" instead of "PostalCode"
	parts := strings.Split(propertyKey, " ")
	if len(parts) == 2 {
		// edgecase for i.e. "Community (Zone)"
		if strings.Contains(parts[1], "(") {
			propertyKey = parts[0]
		} else {
			propertyKey = strings.Join(parts, "")
		}

	}

	// 1b) check for restricted chars
	var b strings.Builder
	for _, char := range propertyKey {
		if _, exists := protectedChars[char]; !exists {
			b.Write([]byte(string(char)))
		}
	}

	propertyKey = b.String()

	// 1b) check for restricted chars
	for _, char := range propertyKey {
		if _, exists := protectedChars[char]; exists {
			return "", fmt.Errorf("invalid character '%U': %w", char, ErrInvalidPropertyName)
		}
	}

	// 2) ensure proper pluralization
	switch pluralization {
	case singular:
		propertyKey = pluralizer.Singular(propertyKey)
	case plural:
		propertyKey = pluralizer.Plural(propertyKey)
	}

	// 3) remove initialisms
	propertyKey = convertInitialisms(propertyKey)

	// 4) titlecase
	return strings.Title(propertyKey), nil
}

func isNillable(p Property) bool {
	return strings.Contains(p.Documentation, "Nillable")
}

func dedupe(elems []string, ignoreElem string) []string {
	m := make(map[string]struct{})
	var final []string 
	for _, elem := range elems {
		if elem == ignoreElem {
			continue 
		}
		
		if _, exists := m[elem]; exists {
			continue 
		} 
		m[elem] = struct{}{} 
		final = append(final, elem)
	}

	return final 
}