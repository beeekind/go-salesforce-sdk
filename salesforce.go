package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/b3ntly/ratelimit"
	"github.com/b3ntly/ratelimit/memory"
	"github.com/b3ntly/salesforce/client"
	"github.com/b3ntly/salesforce/codegen"
	"github.com/b3ntly/salesforce/composite"
	"github.com/b3ntly/salesforce/internal/async"
	"github.com/b3ntly/salesforce/metadata"
	"github.com/b3ntly/salesforce/requests"
	"github.com/b3ntly/salesforce/soql"
	"github.com/b3ntly/salesforce/types"
)

// "If the data can't be found here, check out what's behind API endpoint number 5"

// DefaultClient ...
var DefaultClient = client.Must(
	client.WithLoginFailover(
		client.WithPasswordBearer(
			os.Getenv("SALESFORCE_SDK_CLIENT_ID"),
			os.Getenv("SALESFORCE_SDK_CLIENT_SECRET"),
			os.Getenv("SALESFORCE_SDK_USERNAME"),
			os.Getenv("SALESFORCE_SDK_PASSWORD"),
			os.Getenv("SALESFORCE_SDK_SECURITY_TOKEN"),
		),
		client.WithJWTBearer(
			os.Getenv("SALESFORCE_SDK_CLIENT_ID"),
			os.Getenv("SALESFORCE_SDK_USERNAME"),
			"../private.pem",
		),
	),
	client.WithPool(async.New(100, nil, ratelimit.New(5, time.Second*1, 10, memory.New()))),
	client.WithLimiter(ratelimit.New(5, time.Second, 5, memory.New())),
)

/**
[{"message":"The users password has expired, you must call SetPassword before attempting any other API operations","errorCode":"INVALID_OPERATION_WITH_EXPIRED_PASSWORD"}] SELECT QualifiedApiName, Description FROM EntityDefinition WHERE QualifiedApiName IN ('Account')
*/

// Versions returns data on available API versions for the Salesforce REST API
//
// This method is used during authentication so that we may default to the latest
// REST API endpoint. 
func Versions()([]*client.APIVersion, error){
	return DefaultClient.APIVersions()
}

// Services are api endpoints for RESTful operations within the Salesforce API
func Services() (services map[string]string, err error) {
	_, err = requests.
		Sender(DefaultClient).
		URL("").
		JSON(&services)

	if err != nil {
		return nil, err
	}

	// map values like "/services/data/v49.0/tooling" => "tooling"
	for serviceName, serviceURL := range services {
		parts := strings.Split(serviceURL, "/")
		services[serviceName] = strings.Split(serviceURL, "/")[len(parts)-1]
	}

	return services, nil
}

// Types returns a golang type definition(s) for the JSON response of an endpoint
func Types(structName string, endpoint string) (codegen.Structs, error) {
	response, err := requests.
		Sender(DefaultClient).
		URL(endpoint).
		Response()

	if err != nil {
		return nil, err
	}

	contents, err := requests.ReadAndCloseResponse(response)
	// we can still derive types from an error response for an endpoint, so as long as
	// there is something in response.Body we will ignore this error. For example, disabled API features
	// return errors and thats ok. 
	//
	// parameterizedSearch 400 | serviceTemplates 401 | payments 404 |
	// compactLayouts 400 | smartdatadiscovery 403 | prechatForms 405 |
	// jsonxform 500
	if err != nil {
		if contents == nil || len(contents) == 0 {
			return nil, fmt.Errorf("retrieving types for %s: %w", endpoint, err)
		}
	}

	repr, err := codegen.FromJSON(structName, "", contents)
	if err != nil {
		return nil, err
	}

	return repr, nil
}

// SObjects returns the result of a request to the /sobjects endpoint
func SObjects() (results *metadata.Sobjects, err error) {
	_, err = requests.
		Sender(DefaultClient).
		URL("sobjects").
		JSON(&results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Describe returns the description of a given Salesforce Object
func Describe(objectName string) (describe *metadata.Describe, err error) {
	_, err = requests.
		Sender(DefaultClient).
		URL(fmt.Sprintf("%s/%s/%s", "sobjects", objectName, "describe")).
		JSON(&describe)

	if err != nil {
		return nil, err
	}

	return describe, nil
}

// AllEntities ...
// "If the data can't be found here, check out what's behind API endpoint number 5"
// ~~ B. Bonnette

// DownloadFile returns the given file via its ContentVersion ID
func DownloadFile(contentVersionID string) ([]byte, error) {
	response, err := requests.
		Sender(DefaultClient).
		URL(fmt.Sprintf("sobjects/ContentVersion/%s/VersionData", contentVersionID)).
		Method(http.MethodGet).
		Response()

	if err != nil {
		return nil, err
	}

	return requests.ReadAndCloseResponse(response)
}

// Attachment returns the given Attachment by ID
func Attachment(ID string) ([]byte, error) {
	response, err := requests.
		Sender(DefaultClient).
		URL(fmt.Sprintf("sobjects/Attachment/%s/body", ID)).
		Method(http.MethodGet).
		Response()

	if err != nil {
		return nil, err
	}

	return requests.ReadAndCloseResponse(response)
}

// Document returns the given Document by ID
func Document(req requests.Builder, ID string) ([]byte, error) {
	response, err := req.
		Sender(DefaultClient).
		URL(fmt.Sprintf("sobjects/Document/%s/body", ID)).
		Method(http.MethodGet).
		Response()

	if err != nil {
		return nil, err
	}

	return requests.ReadAndCloseResponse(response)
}

// Count for the given objectName "Lead" "Account" or "User"
func Count(objectName string) (int, error) {
	var response types.QueryResponse
	contents, err := requests.
		Sender(DefaultClient).
		URL("query").
		SQLizer(soql.Select("count()").From(objectName)).
		JSON(&response)

	if err != nil {
		println(string(contents))
		println(err.Error())
		return 0, err
	}

	return response.TotalSize, nil
}

// Find returns all paginated resources for a given query. If there
// are many results and/or many fields this method will take longer 
// to execute and use more of your org's API limit. 
//
// The parameter dst should be a pointer value to a slice of types matching
// the expected query records.
func Find(query string, dst interface{}) error {
	return requests.
		Sender(DefaultClient).
		URL("query").
		QueryMore(soql.String(query), dst, false)
}

// FindAll is akin to the queryAll resource which returns
// soft deleted resources in addition to regular records. 
//
// The parameter dst should be a pointer value to a slice of types matching
// the expected query records.
func FindAll(query string, dst interface{}) error {
	return requests.
		Sender(DefaultClient).
		URL("queryAll").
		QueryMore(soql.String(query), dst, true)
}

// FindByID returns a single result filtered by Id.
//
// The parameter dst should be a pointer to a type matching the 
// expected query record. 
func FindByID(objectName string, objectID string, fields []string, dst interface{}) error {
	var response types.QueryParts
	_, err := requests.
		Sender(DefaultClient).
		URL(metadata.QueryEndpoint).
		SQLizer(
			soql.Select(fields...).
				From(objectName).
				Where(soql.Eq{"Id": objectID}),
		).
		JSON(&response)

	if err != nil {
		return err
	}

	if len(response.Records) != 1 {
		return fmt.Errorf("resource.FindByID query did not return any records")
	}

	if err := json.Unmarshal(response.Records[0], dst); err != nil {
		return err
	}

	return nil
}

// Create created the given objectName
//
// This SDK goes out of its way to not be an ORM which is why the method
// signature doesnt use a generated type like those derived from the codegen
// package. 
func Create(objectName string, fields map[string]interface{}) (ID string, err error) {
	var response composite.Output
	contents, err := requests.
		Sender(DefaultClient).
		URL(fmt.Sprintf("%s/%s", metadata.SobjectsEndpoint, objectName)).
		Method(http.MethodPost).
		Header("Content-Type", "application/json").
		Marshal(fields).
		JSON(&response)

	if err != nil {
		var errors composite.Errors
		if err = json.Unmarshal(contents, &errors); err != nil {
			return "", err
		}

		return "", errors
	}

	//if len(response.Errors) > 0 {
	//	return "", response.Errors
	//}

	return response.ID, nil
}

// UpdateByID updates the given objectName with the given ID. 
//
// Salesforce update responses return empty response bodies and statusCode 
// 204 upon success. 
func UpdateByID(objectName string, ID string, fields map[string]interface{}) error {
	var result composite.Error

	_, err := requests.
		Sender(DefaultClient).
		URL(fmt.Sprintf("%s/%s/%s", metadata.SobjectsEndpoint, objectName, ID)).
		Method(http.MethodPatch).
		Header("Content-Type", "application/json").
		Marshal(fields).
		JSON(&result)

	if err != nil && !errors.Is(err, requests.ErrUnmarshalEmpty) {
		return fmt.Errorf("deleting object %s: %s: %w", objectName, ID, err)
	}

	return nil
}

// DeleteByID deletes the given objectName with the given ID
func DeleteByID(objectName string, ID string) error {
	var result composite.Error
	_, err := requests.
		Sender(DefaultClient).
		URL(fmt.Sprintf("%s/%s/%s", metadata.SobjectsEndpoint, objectName, ID)).
		Method(http.MethodDelete).
		JSON(&result)

	if err != nil && !errors.Is(err, requests.ErrUnmarshalEmpty) {
		return fmt.Errorf("crud.DeleteByID %s: %w", ID, err)
	}

	return nil
}
