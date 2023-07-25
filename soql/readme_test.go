package soql_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/examples/entitydefinitions"
	"github.com/beeekind/go-salesforce-sdk/examples/leads"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/beeekind/go-salesforce-sdk/types"
)

func TestReadMeExamples(t *testing.T) {
	_, err := soql.
		Select("Id", "Name").
		From("Lead").
		Where(soql.And{
			soql.Eq{"FirstName": "Benjamin"},
			// salesforce datetime's use a custom format which types.Datetime accomodates
			soql.Gt{"CreatedDate": types.NewDatetime(time.Now().Add(time.Hour)).String()},
		}).
		Limit(1).
		ToSQL()

	require.Nil(t, err)

	type entityQuery struct {
		types.QueryResponse
		Records []*entitydefinitions.EntityDefinition `json:"records"`
	}

	var response entityQuery
	_, err = requests.
		Sender(salesforce.DefaultClient()).
		URL("tooling/query").
		SQLizer(soql.
			Select("QualifiedApiName").
			From("EntityDefinition").
			Limit(10)).
		JSON(&response)

	require.Nil(t, err)

	type leadQuery struct {
		types.QueryResponse
		Records []*leads.Lead `json:"records"`
	}

	var response2 leadQuery
	subquery := soql.Select("Id").From("Attachments")
	_, err = requests.
		Sender(salesforce.DefaultClient()).
		URL("query").
		SQLizer(
			soql.
				Select("Id", "Name").
				Column(soql.SubQuery(subquery)).
				From("Lead").
				Limit(10),
		).
		JSON(&response2)

	require.Nil(t, err)
}
