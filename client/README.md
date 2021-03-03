# Client 

Wraps http.Client and session state, provides configuration and ratelimiting logic.

## Authentication 

Making authenticated requests to the Salesforce REST API requires appending an Authorization header
with a valid oauth2 access token. Of the [available methods](https://help.salesforce.com/articleView?id=sf.remoteaccess_oauth_flows.htm&type=5) for generating an access code we support the [OAuth 2.0 JWT Bearer Flow for Server-to-Server Integration](https://help.salesforce.com/articleView?id=remoteaccess_oauth_jwt_flow.htm&type=5) and the [OAuth 2.0 Username-Password Flow for Special Scenarios](https://help.salesforce.com/articleView?id=remoteaccess_oauth_username_password_flow.htm&type=5).

## Usage 

Both the client.Client object and its underlying http.Client use the functional-option pattern for configuration.

```go
client, err := client.New(
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
    client.WithLimiter(ratelimit.New(5, time.Second, 5, memory.New())),
)

```

The above code will attempt to login first with the password bearer flow, then the JWT bearer flow. Then it sets the client.Limiter interface to use a ratelimiting library such as github.com/beeekind/ratelimit.

We formally recommend the JWT based flow, even if many production systems default to the password bearer flow.

For a full list of available options see [options.go](https://github.com/beeekind/go-salesforce-sdk/blob/main/client/options.go). Also review the variable defaultOptions in client.go .

---

Clients are intended to fulfill the requests.Sender interface.

```go
var result metadata.Describe 
contents, err := requests. 
    Sender(client). 
    URL("sobjects/Lead/describe").
    JSON(&result)
```

This provides a fluent API that can be heavily customized.

```go
var result metadata.Describe 
contents, err := requests. 
    Sender(client). 
    URL("sobjects/Lead/describe"). 
    Header("foo", "bar"). 
    Ctx(context.WithTimeout(context.Background(), time.Minute)). 
    JSON(&result)
```

The client.QueryMore method is an important method for retrieving paginated records from salesforce.

```go
var results types.QueryResponse
response, err := requests. 
    Sender(client). 
    URL("query"). 
    SQLizer(soql.Select("Id", "Name").From("Lead")). 
    Response() 

if err := requests.Unmarshal(response, &results); err != nil {
    // ...
}

```

Finally all methods may be used in conjunction with generated types.

```go 
type response struct {
    types.QueryResponse
    Records []*leads.Lead `json:"records"`
}

var results response
_, err := requests. 
    Sender(client). 
    URL("query"). 
    SQLizer(soql.Select("Id", "Name").From("Lead")). 
    JSON(&results) 
```