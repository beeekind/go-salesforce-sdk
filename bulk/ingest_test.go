// +build integration

package bulk_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/beeekind/go-salesforce-sdk"
	"github.com/beeekind/go-salesforce-sdk/bulk"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/stretchr/testify/require"
)

var req = requests.Sender(salesforce.DefaultClient)

//go:embed bulk_example.csv
var exampleCSV []byte

func TestIngestUsage(t *testing.T) {
	// creates a job that is intended to insert Account records
	job, err := bulk.CreateJob(req, &bulk.CreateJobRequest{
		Object:    "Account",
		Operation: bulk.OperationInsert,
	})

	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	// upload csv data containing account records
	if _, err := bulk.UploadJob(req, job.ID, bytes.NewBuffer(exampleCSV)); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	// tell the job to process the uploaded csv data
	if _, err := bulk.UpdateJob(req, job.ID, &bulk.UpdateJobRequest{
		State: bulk.JobStateUploadComplete,
	}); err != nil {
		t.Log(err.Error())
		t.FailNow() 
	}

	// wait for the job to execute, this is a race condition
	time.Sleep(time.Second * 10)

	// retrieve failed inserts
	csvReader, err := bulk.GetUnprocessedJobs(req, job.ID)
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
			require.Nil(t, err)
		}

		fmt.Printf("%v\n", record)
	}

	// delete the underlying job
	if err := bulk.DeleteJob(req, job.ID); err != nil {
		t.Log(err.Error())
		t.FailNow() 
	}
}