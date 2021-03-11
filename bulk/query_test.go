// +build integration

package bulk_test

import (
	"io"
	"testing"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/bulk"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
)

var reqs = requests.Sender(salesforce.DefaultClient)

func TestQueryUsage(t *testing.T) {
	job, err := bulk.CreateQuery(reqs, bulk.DelimiterComma, bulk.LineEndingLF, soql.
		Select("Id", "Name", "CreatedBy.Name").
		From("Account"),
	)

	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	_, csvReader, err := bulk.GetQueryResults(reqs, job.ID, "", 0)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	for {
		record, err := csvReader.Read()
		if err != nil && err == io.EOF {
			break
		}

		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}

		t.Logf("%v\n", record)
	}
}
