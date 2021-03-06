// +build integration

package salesforce_test

import (
	"testing"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/examples/leads"
	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/stretchr/testify/require"
)

var commonObjects = []string{
	"RelationshipDomain",
	"EntityDefinition",
	"FieldDefinition",
	"Lead",
	"Contact",
	"User",
	"Account",
	"Asset",
	"Calendar",
	"Event",
	//"Goal",
	"Order",
	"Partner",
	"Report",
	"Vote",
	"Task",
	"Solution",
	"Refund",
	"Profile",
	"OrderStatus",
	"Opportunity",
	"Balloon__c",
}

var metadataObjects = []string{
	"EntityDefinition",
	"Lead",
	//"FieldDefinition",
	//"RelationshipInfo",
	//"RelationshipDomain",
}

func TestVersions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	versions, err := salesforce.Versions()
	require.Nil(t, err)
	require.Greater(t, len(versions), 0)
}

func TestServices(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	services, err := salesforce.Services()
	require.Nil(t, err)
	require.Greater(t, len(services), 0)
}

func TestTypes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	results, err := salesforce.Types("Describe", "/sobjects/EntityDefinition/describe")
	require.Nil(t, err)
	require.Greater(t, len(results), 0)
}

func TestSObjects(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	results, err := salesforce.SObjects()
	require.Nil(t, err)
	require.NotNil(t, results)
	require.Greater(t, len(results.Sobjects), 0)
}

func TestDescribe(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	results, err := salesforce.Describe("Lead")
	require.Nil(t, err)
	require.NotNil(t, results)
	require.Greater(t, len(results.Fields), 0)
}

func TestDownloadFile(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	testFile := "0684x000000KdUVAA0"
	contents, err := salesforce.DownloadFile(testFile)
	require.Nil(t, err)
	require.Greater(t, len(contents), 0)
}

func TestCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	count, err := salesforce.Count("Lead")
	require.Nil(t, err)
	require.Greater(t, count, 0)
}

func TestFind(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	q, err := soql.
		Select("Name", "CreatedDate").
		From("Lead").
		Where(soql.Eq{"FirstName": "testtest"}).
		Limit(1).
		ToSQL()

	require.Nil(t, err)

	var leads []*leads.Lead
	require.Nil(t, salesforce.Find(q, &leads))
	require.Greater(t, len(leads), 0)
}

func TestFindByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	q, err := soql.Select("Id").From("Lead").Where(soql.Eq{"FirstName": "testtest"}).Limit(1).ToSQL()
	require.Nil(t, err)

	var items []*leads.Lead
	require.Nil(t, salesforce.Find(q, &items))
	require.Greater(t, len(items), 0)

	var lead *leads.Lead
	require.Nil(t, salesforce.FindByID("Lead", items[0].ID, []string{"Id", "Name"}, &lead))
	require.Nil(t, err)
}

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ID, err := salesforce.Create("Lead", map[string]interface{}{
		"FirstName":   "testtest",
		"LastName":    "go-salesforce-sdk",
		"Description": "This is a lead for Benjamin",
		"Company":     "ACME inc",
	})
	require.Nil(t, err)

	var lead *leads.Lead
	require.Nil(t, salesforce.FindByID("Lead", ID, []string{"Id", "Name"}, &lead))
	t.Log(ID)
}

func TestUpdateByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var items []*leads.Lead
	q, err := soql.
		Select("Id").
		From("Lead").
		Where(soql.Eq{"FirstName": "testtest"}).
		Limit(1).
		ToSQL()

	require.Nil(t, salesforce.Find(q, &items))
	require.Nil(t, err)
	require.Greater(t, len(items), 0)

	require.Nil(t, salesforce.UpdateByID("Lead", items[0].ID, map[string]interface{}{
		"Title": "qa",
	}))
}

func TestDeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var items []*leads.Lead
	q, err := soql.Select("Id").From("Lead").Where(soql.Eq{"LastName": "go-salesforce-sdk"}).Limit(1).ToSQL()
	require.Nil(t, err)

	require.Nil(t, salesforce.Find(q, &items))
	require.Nil(t, err)
	require.Greater(t, len(items), 0)
	require.Nil(t, salesforce.DeleteByID("Lead", items[0].ID))
}
