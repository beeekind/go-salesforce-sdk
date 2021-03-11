// +build broken

package tree_test

import (
	"github.com/beeekind/go-salesforce-sdk/types"
)

var (
	lastName          = "smith"
	title             = types.NewString("Senator")
	email             = types.NewString("smith@example.com")
	phone             = types.NewString("6508675309")
	website           = types.NewString("google.com")
	numberOfEmployees = types.NewInt(100)
	industry          = types.NewString("bar")
)

/**

var req = requests.Base.Sender(client.Must())

var parseNodeTests = map[interface{}]*tree.Node{
	&accounts.Account{
		Attributes: &types.Attributes{
			Type:        "Account",
			ReferenceID: "random",
		},
		Name:              "loo",
		Phone:             phone,
		Website:           website,
		NumberOfEmployees: numberOfEmployees,
		Industry:          industry,
		Contacts: struct {
			Done      bool                "json:\"done\""
			Count     int                 "json:\"count\""
			TotalSize int                 "json:\"totalSize\""
			Records   []*accounts.Contact "json:\"records\""
		}{
			Records: []*accounts.Contact{{
				LastName: lastName,
				Title:    title,
				Email:    email,
			}},
		},
	}: {
		Attributes: &types.Attributes{
			Type:        "Account",
			ReferenceID: "random",
		},
		Fields: map[string]interface{}{
			"Name":              "loo",
			"Phone":             phone,
			"Website":           website,
			"NumberOfEmployees": numberOfEmployees,
			"Industry":          industry,
		},
		Children: map[string]*tree.Request{
			"Contacts": {[]*tree.Node{
				{
					Attributes: &types.Attributes{
						Type:        "Contact",
						ReferenceID: "random",
					},
					Fields: map[string]interface{}{
						"LastName": lastName,
						"Title":    title,
						"Email":    email,
					},
					Children: map[string]*tree.Request{},
				},
			}},
		},
	},
}

func TestParseNode(t *testing.T) {
	for in, expected := range parseNodeTests {
		actual, err := tree.ParseNode(in)
		if err != nil {
			println(err.Error())
			t.FailNow()
		}

		require.Equal(t, len(expected.Fields), len(actual.Fields), "len(fields)")
		require.Equal(t, len(expected.Children), len(actual.Children), "len(children)")
		require.Equal(t, expected, actual)
	}
}

func TestCreate(t *testing.T) {
	node, err := tree.ParseNode(&accounts.Account{
		Name:              "loo",
		Phone:             phone,
		Website:           website,
		NumberOfEmployees: numberOfEmployees,
		Industry:          industry,
		Contacts: struct {
			Done      bool                "json:\"done\""
			Count     int                 "json:\"count\""
			TotalSize int                 "json:\"totalSize\""
			Records   []*accounts.Contact "json:\"records\""
		}{
			Records: []*accounts.Contact{{
				LastName: lastName,
				Title:    title,
				Email:    email,
			}},
		},
	})

	if err != nil {
		println(err.Error())
		t.FailNow()
	}

	response, err := tree.Create(req, "Account", node)
	if err != nil {
		println(err.Error())
		t.FailNow()
	}

	contents, _ := json.MarshalIndent(response, "\t", "")
	println(string(contents))
}
*/
