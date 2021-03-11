package bulk

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
)

const queryEndpoint = "jobs/query"

// CreateQuery ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/query_create_job.htm
func CreateQuery(builder requests.Builder, delimiter delimiter, lineEnding lineEnding, q soql.Builder) (job *JobInfo, err error) {
	sql, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	_, err = builder.
		Method(http.MethodPost).
		URL(queryEndpoint).
		Header("Content-Type", "application/json").
		Marshal(&CreateQueryJobRequest{
			Operation:       operation(QueryOperation),
			ContentType:     ContentTypeCSV,
			ColumnDelimiter: delimiter,
			LineEnding:      lineEnding,
			Query:           sql,
		}).
		JSON(&job)

	return job, err
}

// GetQueries ... 
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/query_get_all_jobs.htm
func GetQueries(builder requests.Builder, isPkChunkingEnabled bool, jobType jobType, queryLocator string)(jobs *GetJobsResponse, err error){
	builder = builder.
		Method(http.MethodGet).
		URL(queryEndpoint).
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

// GetQuery ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/query_get_one_job.htm
func GetQuery(builder requests.Builder, jobID string) (job *GetJobInfoResponse, err error) {
	_, err = builder.
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/%s", queryEndpoint, jobID)).
		JSON(&job)

	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetQueryResults ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/query_get_job_results.htm
func GetQueryResults(builder requests.Builder, jobID string, locator string, maxRecords int) (nextLocator string, reader *csv.Reader, err error) {
	builder = builder.
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/%s/results", queryEndpoint, jobID))

	if locator != "" {
		builder = builder.Param("locator", locator)
	}

	if maxRecords > 0 {
		builder = builder.Param("maxRecords", strconv.Itoa(maxRecords))
	}

	response, err := builder.Response()
	if err != nil {
		return "", nil, err
	}

	contents, err := requests.ReadAndCloseResponse(response)
	if err != nil {
		return "", nil, err
	}

	return response.Header.Get("Sforce-Locator"), csv.NewReader(bytes.NewBuffer(contents)), nil
}

// UpdateQuery ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/query_abort_job.htm
func UpdateQuery(builder requests.Builder, jobID string, update *UpdateJobRequest) (job *JobInfo, err error) {
	contents, err := builder.
		Method(http.MethodPatch).
		URL(fmt.Sprintf("%s/%s", queryEndpoint, jobID)).
		Header("Content-Type", "application/json").
		Marshal(update).
		JSON(&job)

	if err != nil {
		println(string(contents))
	}

	return job, err
}

// DeleteQuery ...
// https://developer.salesforce.com/docs/atlas.en-us.api_bulk_v2.meta/api_bulk_v2/query_delete_job.htm
func DeleteQuery(builder requests.Builder, jobID string) error {
	_, err := builder.
		Method(http.MethodDelete).
		URL(fmt.Sprintf("%s/%s", queryEndpoint, jobID)).
		Response()

	return err
}