package bulk

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/b3ntly/salesforce/requests"
)

const ingestEndpoint = "jobs/ingest"

// CreateJob ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/create_job.htm
func CreateJob(builder requests.Builder, req *CreateJobRequest) (job *JobInfo, err error) {
	contents, err := builder.
		Method(http.MethodPost).
		URL(ingestEndpoint).
		Header("Content-Type", "application/json").
		Marshal(req).
		JSON(&job)

	if err != nil {
		println(string(contents))
	}

	return job, err
}

// UploadJob ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/upload_job_data.htm#upload_job_data
func UploadJob(builder requests.Builder, jobID string, body io.Reader) (statusCode int, err error) {
	response, err := builder.
		Method(http.MethodPut).
		URL(fmt.Sprintf("%s/%s/batches", ingestEndpoint, jobID)).
		Body(body).
		Header("Content-Type", "text/csv").
		Response()

	if response == nil {
		return 0, err
	}

	return response.StatusCode, err
}

// UpdateJob ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/close_job.htm
func UpdateJob(builder requests.Builder, jobID string, update *UpdateJobRequest) (job *JobInfo, err error) {
	contents, err := builder.
		Method(http.MethodPatch).
		URL(fmt.Sprintf("%s/%s", ingestEndpoint, jobID)).
		Header("Content-Type", "application/json").
		Marshal(update).
		JSON(&job)

	if err != nil {
		println(string(contents))
	}

	return job, err
}

// DeleteJob ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/delete_job.htm
func DeleteJob(builder requests.Builder, jobID string) error {
	_, err := builder.
		Method(http.MethodDelete).
		URL(fmt.Sprintf("%s/%s", ingestEndpoint, jobID)).
		Response()

	return err
}

// GetJobs ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/get_all_jobs.htm
func GetJobs(builder requests.Builder, isPkChunkingEnabled bool, jobType jobType, queryLocator string) (jobs *GetJobsResponse, err error) {
	builder = builder.
		Method(http.MethodGet).
		URL(ingestEndpoint).
		Param("isPkChunkingEnabled", strconv.FormatBool(isPkChunkingEnabled))

	if jobType != "" {
		builder = builder.Param("jobType", string(jobType))
	}

	if queryLocator != "" {
		builder = builder.Param("queryLocator", queryLocator)
	}

	contents, err := builder.JSON(&jobs)

	if err != nil {
		println(string(contents))
		return nil, err
	}

	return jobs, nil
}

// GetJob ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/get_job_info.htm
func GetJob(builder requests.Builder, jobID string) (job *GetJobInfoResponse, err error) {
	_, err = builder.
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/%s", ingestEndpoint, jobID)).
		JSON(&job)

	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetSuccessfulRecords ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/get_job_successful_results.htm
func GetSuccessfulRecords(builder requests.Builder, jobID string) (*csv.Reader, error) {
	response, err := builder.
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/%s/successfulResults", ingestEndpoint, jobID)).
		Response()

	if err != nil {
		return nil, err
	}

	contents, err := requests.ReadAndCloseResponse(response)
	if err != nil {
		return nil, err 
	}

	return csv.NewReader(bytes.NewBuffer(contents)), nil
}

// GetFailedRecords ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/get_job_failed_results.htm
func GetFailedRecords(builder requests.Builder, jobID string) (*csv.Reader, error) {
	response, err := builder.
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/%s/failedResults", ingestEndpoint, jobID)).
		Response()

	if err != nil {
		return nil, err
	}

	contents, err := requests.ReadAndCloseResponse(response)
	if err != nil {
		return nil, err 
	}

	return csv.NewReader(bytes.NewBuffer(contents)), nil
}

// GetUnprocessedJobs ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/get_job_unprocessed_results.htm
func GetUnprocessedJobs(builder requests.Builder, jobID string) (*csv.Reader, error) {
	response, err := builder.
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/%s/unprocessedrecords", ingestEndpoint, jobID)).
		Response()

	if err != nil {
		return nil, err
	}

	contents, err := requests.ReadAndCloseResponse(response)
	if err != nil {
		return nil, err 
	}

	return csv.NewReader(bytes.NewBuffer(contents)), nil
}

