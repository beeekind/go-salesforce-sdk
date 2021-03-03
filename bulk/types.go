package bulk

import (
	"net/http"

	"github.com/b3ntly/salesforce/types"
)

type client interface {
	Do(*http.Request) (*http.Response, error)
	URL(string) string
}

type delimiter string

const (
	// DelimiterBackQuote ...
	DelimiterBackQuote delimiter = "BACKQUOTE"
	// DelimeterCaret ...
	DelimeterCaret delimiter = "CARET"
	// DelimiterComma ...
	DelimiterComma delimiter = "COMMA"
	// DelimiterPipe ...
	DelimiterPipe delimiter = "PIPE"
	// DelimiterSemiColon ...
	DelimiterSemiColon delimiter = "SEMICOLON"
	// DelimiterTab ...
	DelimiterTab delimiter = "TAB"
)

type contentType string

const (
	// ContentTypeCSV ...
	ContentTypeCSV contentType = "CSV"
)

type lineEnding string

const (
	// LineEndingLF linefeed character
	LineEndingLF lineEnding = "LF"
	// LineEndingCRLF carriage return character followed by linefeed character
	LineEndingCRLF lineEnding = "CRLF"
)

type operation string

const (
	// OperationInsert ...
	OperationInsert operation = "insert"
	// OperationDelete ...
	OperationDelete operation = "delete"
	// OperationHardDelete ...
	OperationHardDelete operation = "hardDelete"
	// OperationUpdate ...
	OperationUpdate operation = "update"
	// OperationUpsert ...
	OperationUpsert operation = "upsert"
)

// CreateJobRequest ...
type CreateJobRequest struct {
	// AssignmentRuleID The ID of an assignment rule to run for a Case or a Lead.
	// The assignment rule can be active or inactive.
	// The ID can be retrieved by using the Lightning Platform SOAP API
	// or the Lightning Platform REST API to query the AssignmentRule object.
	// This property is available in API version 49.0 and later.
	AssignmentRuleID string `json:"assignmentRuleId,omitempty"`
	// ColumnDelimiter The column delimiter used for CSV job data. The default value is COMMA.
	ColumnDelimiter delimiter `json:"columnDelimiter,omitempty"`
	// ContentType The content type for the job. The only valid value (and the default) is CSV.
	ContentType contentType `json:"contentType,omitempty"`
	// ExternalIDFieldName The external ID field in the object being updated.
	// Only needed for upsert operations. Field values must also exist in CSV job data.
	ExternalIDFieldName string `json:"externalIdFieldName,omitempty"`
	// LineEnding The line ending used for CSV job data, marking the end of a data row. The default is LF.
	LineEnding lineEnding `json:"lineEnding,omitempty"`
	// Object The object type for the data being processed.
	// Use only a single object type per job.
	// I.e. "Contact"
	Object string `json:"object"`
	// Operation The processing operation for the job.
	Operation operation `json:"operation"`
}

type jobType string

const (
	// JobTypeObject ...
	JobTypeObject jobType = "BigObjectIngest"
	// JobTypeClassic ...
	JobTypeClassic jobType = "Classic"
	// JobTypeIngest ...
	JobTypeIngest jobType = "V2Ingest"
	// JobTypeQuery ... 
	JobTypeQuery jobType = "V2Query"
)

type jobState string

const (
	// JobStateOpen — The job has been created, and data can be added to the job.
	JobStateOpen jobState = "Open"
	// JobStateUploadComplete — No new data can be added to this job. You can’t edit or save a closed job.
	JobStateUploadComplete jobState = "UploadComplete"
	// JobStateAborted - The job has been aborted. You can abort a job if you created it or if you have the “Manage Data Integrations” permission.
	JobStateAborted jobState = "Aborted"
	// JobStateComplete — The job was processed by Salesforce.
	JobStateComplete jobState = "JobComplete"
	// JobStateFailed — Some records in the job failed. Job data that was successfully processed isn’t rolled back.
	JobStateFailed jobState = "Failed"
)

// UpdateJobRequest Closes or aborts a job. If you close a job, Salesforce queues the job and
// uploaded data for processing, and you can’t add any additional job data.
// If you abort a job, the job does not get queued or processed.
type UpdateJobRequest struct {
	State jobState `json:"state"`
}

// GetJobsRequest ...
type GetJobsRequest struct {
	// IsPKChunkingEnabled If set to true, filters jobs with PK chunking enabled.
	IsPKChunkingEnabled bool `json:"isPkChunkingEnabled"`
	// Filters jobs based on job type.
	JobType jobType `json:"jobType"`
	// QueryLocator Use queryLocator with a locator value to get a specific set of job results.
	// Get All Jobs returns up to 1000 result rows per request, along with a nextRecordsUrl value that contains the locator value used to get the next set of results.
	QueryLocator string `json:"queryLocator"`
}

// GetJobsResponse ...
type GetJobsResponse struct {
	// Indicates whether there are more jobs to get. If false, use the nextRecordsUrl value to retrieve the next group of jobs.
	Done bool `json:"done"`
	// NextRecordsURL A URL that contains a query locator used to get the next set of results in a subsequent request if done isn’t true.
	NextRecordsURL string `json:"NextRecordsUrl"`
	// Records Contains information for each retrieved job.
	Records []*JobInfo `json:"records"`
}

// JobInfo ...
type JobInfo struct {
	// APIVersion The API version that the job was created in.
	APIVersion float64 `json:"apiVersion"`
	// AssignmentRuleID The ID of the assignment rule. This property is only shown if an assignment rule is specified when the job is created.
	AssigmentRuleID string `json:"assignmentRuleId"`
	// ColumnDelimiter The column delimiter used for CSV job data.
	ColumnDelimiter delimiter `json:"columnDelimiter"`
	// ConcurrencyMode For future use. How the request was processed.
	// Currently only parallel mode is supported. (When other modes are added, the mode will be chosen automatically by the API and will not be user configurable.)
	ConcurrencyMode string `json:"concurrencyMode"`
	// ContentType The format of the data being processed. Only CSV is supported.
	ContentType contentType `json:"contentType"`
	// ContentURL The URL to use for Upload Job Data requests for this job. Only valid if the job is in Open state.
	ContentURL string `json:"contentUrl"`
	// CreatedByID The ID of the user who created the job.
	CreatedByID string `json:"createdById"`
	// CreatedDate The date and time in the UTC time zone when the job was created.
	CreatedDate types.Datetime `json:"CreatedDate"`
	// ExternalIDFieldName The name of the external ID field for an upsert.
	ExternalIDFieldName string `json:"externalIdFieldName"`
	// ID Unique ID for this job
	ID string `json:"id"`
	// JobType The job’s type.
	JobType string `json:"jobType"`
	// LineEnding The line ending used for CSV job data.
	LineEnding lineEnding `json:"lineEnding"`
	// Object The object type for the data being processed.
	Object string `json:"object"`
	// Operation The processing operation for the job.
	Operation operation `json:"Operation"`
	// State The current state of processing for the job.
	State string `json:"state"`
	// SystemModstamp Date and time in the UTC time zone when the job finished.
	SystemModstamp types.Datetime `json:"systemModstamp"`
}

// GetJobInfoRequest ...
type GetJobInfoRequest struct {
	// JobID Required ...
	JobID string `json:"jobId"`
}

// GetJobInfoResponse ...
type GetJobInfoResponse struct {
	JobInfo
	// The number of milliseconds taken to process triggers and other processes related to the job data.
	// This doesn't include the time used for processing asynchronous and batch Apex operations.
	// If there are no triggers, the value is 0.
	ApexProcessingTime int64 `json:"apexProcessingTime"`
	// APIActiveProcessingTime The number of milliseconds taken to actively process the job and includes apexProcessingTime,
	// but doesn't include the time the job waited in the queue to be processed or the time required for serialization and deserialization.
	APIActiveProcessingTime int64 `json:"apiActiveProcessingTime"`
	// NumRecordsFailed The number of records that were not processed successfully in this job.
	// This property is of type int in API version 46.0 and earlier.
	NumRecordsFailed int `json:"numRecordsFailed"`
	// NumRecordsProcessed The number of records already processed.
	// This property is of type int in API version 46.0 and earlier.
	NumRecordsProcessed int `json:"numRecordsProcessed"`
	// Retries The number of times that Salesforce attempted to save the results of an operation. The repeated attempts are due to a problem, such as a lock contention.
	Retries int `json:"retries"`
	TotalProcessingTime int `json:"totalProcessingTime"`
}

type queryOperation string

const (
	// QueryOperation returns data that has not been deleted or archived. For more information, see query() in the SOAP API Developer Guide.
	QueryOperation queryOperation = "query"
	// QueryOperationAll returns records that have been deleted because of a merge or delete, and returns information about archived Task and Event records.
	QueryOperationAll queryOperation = "queryAll"
)

// CreateQueryJobRequest ...
type CreateQueryJobRequest struct {
	// ColumnDelimiter the column delimiter used for CSV job data
	ColumnDelimiter delimiter `json:"columnDelimiter"`
	// ContentType the format to be used for the results.
	// Currently the only supported value is CSV (comma-separated variables).
	// Defaults to CSV.
	ContentType contentType `json:"contentType"`
	// LineEnding the line ending used for CSV job data, marking the end of a data row. The default is LF.
	LineEnding lineEnding `json:"lineEnding"`
	// Operation - the type of query.
	Operation operation `json:"operation"`
	// Query the SOQL query to be performed.
	Query string `json:"query"`
}

// GetQueryJobRequest ...
type GetQueryJobRequest struct {
	// QueryJobID the ID of the query job
	QueryJobID string `json:"queryJobId"`
	// Locator A string that identifies a specific set of query results. Providing a value for this parameter returns only that set of results.
	// Omitting this parameter returns the first set of results.
	// You can find the locator string for the next set of results in the response of each request. See Example and Rules and Guidelines.
	// As long as the associated job exists, the locator string for a set of results does not change. You can use the locator to retrieve a set of results multiple times.
	Locator string `json:"locator"`
	// MaxRecords The maximum number of records to retrieve per set of results for the query. The request is still subject to the size limits.
	// If you are working with a very large number of query results, you may experience a timeout before receiving all the data from Salesforce.
	// To prevent a timeout, specify the maximum number of records your client is expecting to receive in the maxRecords parameter. This splits the results into smaller sets with this value as the maximum size.
	// If you don’t provide a value for this parameter, the server uses a default value based on the service.
	MaxRecords int `json:"maxRecords"`
}

type abortedState string

const (
	// AbortedState ...
	AbortedState = "Aborted"
)

// AbortQueryJobRequest ...
type AbortQueryJobRequest struct {
	// State must be "Aborted"
	State abortedState `json:"state"`
}

// GetQueryJobsResponse ...
type GetQueryJobsResponse struct {
	// Done ...
	Done bool `json:"done"`
	// NextRecordsURL ...
	NextRecordsURL string `json:"mextRecordsUrl"`
	// Records ...
	Records []*JobInfo `json:"records"`
}
