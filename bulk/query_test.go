package bulk_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/b3ntly/salesforce/bulk"
	"github.com/b3ntly/salesforce/client"
	"github.com/b3ntly/salesforce/requests"
	"github.com/b3ntly/salesforce/soql"
	"github.com/stretchr/testify/require"
)

var reqs = requests.Base.Sender(client.Must())

func TestCreateQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	job, err := bulk.CreateQuery(reqs, bulk.DelimiterComma, bulk.LineEndingLF, soql.
		Select("Id", "Name", "CreatedBy.Name").
		From("Account"),
	)

	require.Nil(t, err)
	t.Log(2, job.ID, job.ContentURL)
}

func TestQueryResults(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	nextLocator, reader, err := bulk.GetQueryResults(reqs, "7504x000002FEcPAAW", "", 0)
	require.Nil(t, err)

	for {

		record, err := reader.Read()
		if err != nil && err == io.EOF {
			break
		}

		if err != nil {
			require.Nil(t, err)
		}

		fmt.Printf("%v\n", record)
	}

	t.Log(nextLocator)
}

func TestUpdateQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	job, err := bulk.UpdateQuery(reqs, "7504x000002FEcPAAW", &bulk.UpdateJobRequest{
		State: "Aborted",
	})

	require.Nil(t, err)
	t.Log(job.ID, job.State)
}

func TestDeleteQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	err := bulk.DeleteQuery(reqs, "7504x000002FEcPAAW")
	require.Nil(t, err)
}

func TestGetQueries(t *testing.T) {
	result, err := bulk.GetQueries(reqs, true, bulk.JobTypeQuery, "")
	require.Nil(t, err)

	for _, r := range result.Records {
		t.Log(r.ID, r.State, r.SystemModstamp.Value.String())
	}

	t.Log(len(result.Records))
}
