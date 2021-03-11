## Bulk API

Allows data to be uploaded/downloaded in text/csv format. This is a preferred method for uploading large quantities of records because it has a much higher throughput then other insert methods via the REST API. 

Salesforce also has a GUI for Bulk operations which includes field mapping and other features, so consider using that for non-programmatic cases. 

## Basic Usage 

The Bulk API abstracts operations as Jobs which are created, prepared, started, and deleted. 

```golang 
var req = requests.Sender(salesforce.DefaultClient)

// creates a job that is intended to insert Account records
job, err := bulk.CreateJob(req, &bulk.CreateJobRequest{
    Object: "Account",
    Operation: bulk.OperationInsert,
})

// upload csv data containing account records 
csvData := []byte(`some,csv,data\n1,2,3`)
statusCode, err := bulk.UploadJob(req, job.ID, bytes.Newbuffer(csvData)))

// tell the job to process the uploaded csv data 
job, err := bulk.UpdateJob(req, job.ID, &bulk.UpdateJobRequest{
    State: bulk.JobStateUploadComplete,
})

time.Sleep(time.Minute)

// retrieve failed inserts 
csvReader, err := bulk.GetUnprocessedJobs(req, job.ID)
```

## Bulk Query 

The bulk-query API allows you to query resources in text/csv format. I'm not a huge fan of this portion of the API because text/csv encoding can lose type precision as data types are marshalled in and out of text format. Creating an ETL pipeline or similar work based on the input/output of massive CSV files is not a great thing. BUT if you have to do it here you go.

```golang
// create a job that will execute the query and store its results 
job, err := bulk.CreateQuery(req, bulk.DelimiterComma, bulk.LineEndingLF, soql. 
    Select("Id", "Name", "CreatedBy.Name"). 
    From("Account"), 
)

time.Sleep(time.Second * 2)

// retrieve the results 
nextLocator, csvReader, err := bulk.GetQueryResults(reqs, job.ID, "", 0)
for {
    record, err := csvReader.Read() 
    if err != nil && err == io.EOF {
        break 
    }

    if err != nil {
        fmt.Println(err.Error())
        return 
    }
}

```
