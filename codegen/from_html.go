package codegen

// html.go converts html tables documenting a salesforce object to a codegen.Struct representing that same object

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	// ErrObjectDocumentationNotFound ...
	ErrObjectDocumentationNotFound = errors.New("documentation not found")
	// ErrEmptyDocumentation ...
	ErrEmptyDocumentation = errors.New("empty documentation table: likely a chromium rendering issue")
	// ErrInternalProperty ...
	ErrInternalProperty = errors.New("internal salesforce property do not use")
	// ErrPossiblyNotEnabled ...
	ErrPossiblyNotEnabled = errors.New("property is disabled by default")
	// ErrPropertyNameNotFound ...
	ErrPropertyNameNotFound = errors.New("property name not found")
	// ErrTableStructureUnknown ...
	ErrTableStructureUnknown = errors.New("property row table structure unknown")
	// ErrEmptyRow ...
	ErrEmptyRow = errors.New("property row is empty")
)

// FromHTML converts the an HTML table representing a Salesforce Object to a codegen.Struct
func FromHTML(doc *goquery.Document) (*Struct, error) {
	entity := &Struct{}

	errMsgHidden := doc.Find(".error-container.ng-hide h2").Text()
	errMsg := doc.Find(".error-container h2").Text()
	if errMsgHidden == "" && strings.Contains(errMsg, "Sorry, the document you are looking for doesn't exist or could not be retrieved.") {
		return nil, ErrObjectDocumentationNotFound
	}

	// some documentation pages are empty and thats ok, some documentation pages have multiple tables and thats ok too
	tableSelection, err := parseTable(doc.Selection)
	if err != nil {
		return nil, ErrEmptyDocumentation
	}

	entity.Name, err = prepareStructName(stripNewLinesAndTabs(doc.Find("h1.helpHead1").Text()), singular)
	if err != nil {
		return nil, ErrInvalidPropertyName
	}

	//
	entity.DocumentationURL, _ = doc.Find(`link[rel="canonical"]`).Attr("href")

	//
	entity.Documentation = enforceLineLimit(stripNewLinesAndTabs(doc.Find(".shortdesc").Text()), 90)
	if entity.Name == "" || entity.Documentation == "" {
		return nil, errors.New("cannot identify structName or documentation")
	}

	tableSelection.Children().Each(func(idx int, s *goquery.Selection) {
		if err != nil {
			return
		}

		var prop *Property
		prop, err = parseRow(s)
		// skip ErrInternalProperty as a property that should not be processed
		if errors.Is(err, ErrInternalProperty) || errors.Is(err, ErrEmptyRow) {
			err = nil
			return
		}

		if errors.Is(err, ErrPossiblyNotEnabled) || errors.Is(err, ErrEmptyRow) {
			err = nil
			return
		}

		if err != nil {
			return
		}

		entity.Properties = append(entity.Properties, prop)
	})

	if err != nil {
		return nil, fmt.Errorf("parsing documentation property: %w", err)
	}

	return entity, nil
}

// we collect information on object properties via a table in the documentation page,
// however some pages have multiple tables and the table is not always the same
//
// so we return the table with the greatest amount of text as the presumed desired table
//
// further we ensure any returned table has two columns as indicated by the number of th elements
// directly descendant of thead
func parseTable(s *goquery.Selection) (*goquery.Selection, error) {
	var currentTable *goquery.Selection
	s.Find("table").Each(func(idx int, s2 *goquery.Selection) {
		if currentTable == nil {
			currentTable = s2.Find("tbody").First()
		}

		if len(s2.Text()) > len(currentTable.Text()) {
			if len(s2.Find("thead > tr").Children().Nodes) == 2 {
				currentTable = s2.Find("tbody").First()
			}
		}
	})

	if currentTable == nil {
		return nil, ErrEmptyDocumentation
	}

	return currentTable, nil
}

func parseRow(s *goquery.Selection) (*Property, error) {
	prop := &Property{}

	// 0) retrieve the property name
	// * usually the name can be retried in tr > td.first() > span.keyword.parname
	// * sometimes the span is missing and must be retrieved via td.entry
	name := stripNewLinesAndTabs(s.Find(".keyword.parmname").First().Text())
	if name == "" {
		name = stripNewLinesAndTabs(s.Find("td.entry").First().Text())
		if name == "" {
			return nil, ErrEmptyRow
		}
	}

	propertyNameSingular, err := prepareStructName(name, neither)
	if err != nil {
		return nil, err
	}

	prop.Name = propertyNameSingular

	// 1) all property details are contained in a table column (element=td) but the structure may vary between properties
	// * most properties are structured as td > dl > dt, dd, dt, dd, dt, dd
	// * some properties are structured as td > dl > dt, dt, dt, dd, dt, dd
	// * some properties are "internal" and do not contain a dl element with any property details, we skip these properties entirely without error
	section := s.Find(".dl.detailList")
	if len(section.Children().Nodes) <= 3 && (strings.Contains(s.Find("td").Last().Text(), "For internal use only.") ||
		strings.Contains(s.Find("td").Last().Text(), "eserved for future use.")) {
		return nil, ErrInternalProperty
	}

	if strings.Contains(s.Text(), "This field is only available to organizations"){
		return nil, ErrPossiblyNotEnabled
	}

	if strings.Contains(s.Text(), "This field is available if you enabled Salesforce to Salesforce"){
		return nil, ErrPossiblyNotEnabled
	}

	if len(section.Children().Nodes) < 4 {
		return nil, fmt.Errorf("field %s: %w", prop.Name, ErrTableStructureUnknown)
	}

	// [0]: type label
	// [1]: type value
	// [2]: property label OR MAYBE type value
	// [3]: property value
	// ... successive nodes are optional and may not be present for some properties
	// [4]: description label
	// [5]: description value
	// [5:]: description++
	var description string
	var properties string
	section.Children().Each(func(idx int, s2 *goquery.Selection) {
		if err != nil {
			return 
		}

		switch x := idx; {
		case x == 1:
			prop.Type, err = prepareDatatype(prop.Name, stripNewLinesAndTabs(s2.Text()))

		case x == 3:
			// in this context properties are a comma separated list of attributes i.e. "Create, Filter, Group, Sort, " that
			// we append to the end of the description
			properties = stripNewLinesAndTabs(s2.Text())
		case x == 5:
			description = stripNewLinesAndTabs(s2.Text())
		case x > 5:
			description += "\n\t// " + stripNewLinesAndTabs(s2.Text()) + "\n"
		}
	})

	if err != nil {
		return nil, err
	}

	prop.Documentation = enforceLineLimit(description+"\n\t//\n\t// Properties:"+properties, 90)
	prop.Tag = jsonTag(name)
	return prop, nil
}
