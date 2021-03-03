package bulk_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"testing"

	"github.com/b3ntly/salesforce/bulk"
	"github.com/b3ntly/salesforce/client"
	"github.com/b3ntly/salesforce/requests"
	"github.com/stretchr/testify/require"
)

var req = requests.Base.Sender(client.Must())

//go:embed bulk_example.csv
var exampleCSV []byte

func TestCreateJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	job, err := bulk.CreateJob(req, &bulk.CreateJobRequest{
		Object:    "Account",
		Operation: bulk.OperationInsert,
	})

	require.Nil(t, err)
	t.Log(job.ID)
	t.Log(job.State)
	t.Log(job.ContentURL)
}

func TestUploadJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	statusCode, err := bulk.UploadJob(req, "7504x000002FEUzAAO", bytes.NewBuffer(exampleCSV))
	require.Nil(t, err)
	t.Log(statusCode)
}

func TestUpdateJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	job, err := bulk.UpdateJob(req, "7504x000002FEUzAAO", &bulk.UpdateJobRequest{
		State: bulk.JobStateUploadComplete,
	})

	require.Nil(t, err)
	t.Log(job.State)
}

func TestGetJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	response, err := bulk.GetJobs(req, false, bulk.JobTypeIngest, "")
	require.Nil(t, err)

	for _, j := range response.Records {
		t.Log(j.ID, j.State, j.SystemModstamp.Value.String())
	}
}

func TestGetJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	job, err := bulk.GetJob(req, "7504x000002FEUzAAO")
	require.Nil(t, err)

	t.Log(job.ID, job.State, job.ApexProcessingTime, job.APIActiveProcessingTime, job.NumRecordsProcessed, job.NumRecordsFailed, job.Retries)
}

func TestGetSuccessfulRecords(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	reader, err := bulk.GetSuccessfulRecords(req, "7504x000002FEUzAAO")
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
}

func TestGetFailedRecords(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	reader, err := bulk.GetFailedRecords(req, "7504x000002FEUzAAO")
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
}

func TestGetUnprocessedRecords(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	reader, err := bulk.GetUnprocessedJobs(req, "7504x000002FEUzAAO")
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
}

func TestDeleteJob(t *testing.T) {
	t.Skip()
	result, err := bulk.GetJobs(req, false, "", "")
	require.Nil(t, err)

	success := false
	if len(result.Records) > 0 {
		for _, job := range result.Records {
			if job.State == string(bulk.JobStateComplete) {
				err := bulk.DeleteJob(req, job.ID)
				require.Nil(t, err)
				success = true
				break
			}
		}
	}

	if !success {
		t.Log("no deletablle record found")
	}
}
