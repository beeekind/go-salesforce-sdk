package requests_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/b3ntly/salesforce/codegen"
	"github.com/b3ntly/salesforce/requests"
	"github.com/stretchr/testify/require"
)

var docTemplate = "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_%s.htm"
var testObjects = []string{
	"AcceptedEventRelation",
	"ActivityHistory",
	"LeadCleanInfo",
	"ProcessInstance",
	"Lead",
	"Opportunity",
	"Account",
	"Contact",
	"Event",
}

func TestParseWebApp(t *testing.T) {
	t.Skip()
	documentationURL := "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_%s.htm"
	doc, err := requests.ParseWebApp(strings.ToLower(fmt.Sprintf(documentationURL, "lead")))
	require.Nil(t, err)
	objectName := doc.Find("span#topic-title.ph").Text()
	require.Equal(t, objectName, "Lead")
}

func TestGetWebApp(t *testing.T) {
	for _, objectName := range testObjects {
		URL := fmt.Sprintf(docTemplate, strings.ToLower(objectName))
		doc, err := requests.ParseWebApp(URL)
		require.Nil(t, err)
		require.NotNil(t, doc)

		entity, err := codegen.FromHTML(doc)
		if errors.Is(err, codegen.ErrObjectDocumentationNotFound) {
			fmt.Println(err.Error())
		}
		require.Nil(t, err, URL)
		require.NotNil(t, entity)
	}
}
