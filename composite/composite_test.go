package composite_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/beeekind/go-salesforce-sdk/client"
	"github.com/beeekind/go-salesforce-sdk/composite"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/stretchr/testify/require"
)

var sender = client.Must()
var req = requests.Base.Sender(sender)

type input struct {
	composite.Builder
}

type output struct {
	resultContainsErrors bool
	requestBody          []byte
}

var builderTests = map[*input]*output{
	{composite.
		Client(sender).
		AllOrNone(true).
		CollateSubrequests(true).
		Post("Lead", "ref", nil, map[string]interface{}{
			"LastName": "Richards",
			"Company":  "ACME inc",
		}).
		Patch("Lead/@{ref.id}", "ref2", nil, map[string]interface{}{
			"FirstName": "fred",
			"LastName":  "johnson",
		}).
		Get("Lead/@{ref.id}", "ref3", nil, nil).
		Delete("Lead/@{ref.id}", "ref4", nil, nil),
	}: {false, []byte(`{
		"allOrNone":true,
		"collateSubrequests":true,
		"compositeRequest":[
			{
				"method":"POST",
				"url":"/services/data/v51.0/sobjects/Lead",
				"referenceId":"ref",
				"body":{
					"Company":"ACME inc",
					"LastName":"Richards"
				}
			},
			{
				"method":"PATCH",
				"url":"/services/data/v51.0/sobjects/Lead/@{ref.id}",
				"referenceId":"ref2",
				"body":{
					"FirstName":"fred",
					"LastName":"johnson"
				}
			},
			{
				"method":"GET",
				"url":"/services/data/v51.0/sobjects/Lead/@{ref.id}",
				"referenceId":"ref3"
			},
			{
				"method":"DELETE",
				"url":"/services/data/v51.0/sobjects/Lead/@{ref.id}",
				"referenceId":"ref4"
			}]
		}`)},
	{composite.
		Client(sender).
		AllOrNone(true).
		CollateSubrequests(false).
		Post("Lead", "ref", nil, map[string]interface{}{
			"LastName": "Richards",
			"Company":  "ACME inc",
		}).
		Patch("Lead/@{ref1.id}", "ref2", nil, map[string]interface{}{
			"FirstName": "fred",
			"LastName":  "johnson",
		}).
		Get("Lead/@{ref1.id}", "ref3", nil, nil).
		Delete("Lead/@{ref1.id}", "ref4", nil, nil)}: {true, []byte(`{
			"allOrNone":true,
			"collateSubrequests":false,
			"compositeRequest":[{
				"method":"POST",
				"url":"/services/data/v51.0/sobjects/Lead",
				"referenceId":"ref",
				"body":{
					"Company":"ACME inc",
					"LastName":"Richards"
				}
			},{
				"method":"PATCH",
				"url":"/services/data/v51.0/sobjects/Lead/@{ref1.id}",
				"referenceId":"ref2",
				"body":{
					"FirstName":"fred",
					"LastName":"johnson"
				}
			},{
				"method":"GET",
				"url":"/services/data/v51.0/sobjects/Lead/@{ref1.id}",
				"referenceId":"ref3"
			},{
				"method":"DELETE",
				"url":"/services/data/v51.0/sobjects/Lead/@{ref1.id}",
				"referenceId":"ref4"
			}]
		}`)},
}

func TestJSONPayload(t *testing.T) {
	for in, out := range builderTests {
		t.Run("", func(t *testing.T) {
			req, err := in.Request()
			require.Nil(t, err)
			actual, err := ioutil.ReadAll(req.Body)
			require.Nil(t, err)
			outStr := strings.ReplaceAll(string(out.requestBody), "\n", "")
			outStr = strings.ReplaceAll(outStr, "\t", "")
			require.Equal(t, outStr, string(actual))
		})
	}
}

func TestUnmarshalTypes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	for in, out := range builderTests {
		t.Run("", func(t *testing.T) {
			response, err := in.Send()

			if out.resultContainsErrors {
				require.NotNil(t, err)
			} else {
				if err != nil {
					println(err.Error())
					require.Nil(t, err)
				}

				t.Logf("%v successful composite actions\n", len(response.Items))
			}
		})
	}
}
