package client_test

import (
	"os"
	"testing"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/examples/leads"
	"github.com/beeekind/go-salesforce-sdk/client"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/beeekind/go-salesforce-sdk/types"
	"github.com/stretchr/testify/require"
)

var c = client.Must()

func TestNew(t *testing.T) {
	client, err := client.New(client.WithLoginFailover(
		client.WithPasswordBearer(
			os.Getenv("SALESFORCE_SDK_CLIENT_ID"),
			os.Getenv("SALESFORCE_SDK_CLIENT_SECRET"),
			os.Getenv("SALESFORCE_SDK_USERNAME"),
			os.Getenv("SALESFORCE_SDK_PASSWORD"),
			os.Getenv("SALESFORCE_SDK_SECURITY_TOKEN"),
		),
		client.WithJWTBearer(
			os.Getenv("SALESFORCE_SDK_CLIENT_ID"),
			os.Getenv("SALESFORCE_SDK_USERNAME"),
			"../private.pem",
		),
	))

	require.Nil(t, err)
	require.NotNil(t, client)
}

func TestNewWithJWT(t *testing.T) {
	t.Skip() 
	client, err := client.New(client.WithJWTBearer(
		os.Getenv("SALESFORCE_SDK_CLIENT_ID"),
		os.Getenv("SALESFORCE_SDK_USERNAME"),
		"../private.pem",
	))

	require.Nil(t, err)
	require.NotNil(t, client)
}

func TestQueryMore(t *testing.T) {
	type response struct {
		types.QueryResponse
		Records []*leads.Lead `json:"records"`
	}

	var results response
	_, err := requests.
		Sender(salesforce.DefaultClient).
		URL("query").
		SQLizer(soql.Select("Id", "Name").From("Lead")).
		JSON(&results)

	require.Nil(t, err)
	require.Greater(t, len(results.Records), 0)
	t.Log(len(results.Records))
}

type clientURLInput string

type clientURLOutput string

var clientURLTests = map[clientURLInput]clientURLOutput{
	"https://some.fully.qualified.url/bar": "https://some.fully.qualified.url/bar",
	"path/segment":                         "https://placeholder-dev-ed.my.salesforce.com/services/data/v51.0/path/segment",
	"/path/segment":                        "https://placeholder-dev-ed.my.salesforce.com/services/data/v51.0/path/segment",
	"/query?q=select Name, Id from foo":    "https://placeholder-dev-ed.my.salesforce.com/services/data/v51.0/query?q=select Name, Id from foo",
}

func TestClientURL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	for in, out := range clientURLTests {
		t.Run(string(in), func(t *testing.T) {
			result := c.URL(string(in))
			require.Equal(t, string(out), result, string(in))
		})
	}
}
