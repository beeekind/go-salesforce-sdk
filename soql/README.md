## SOQL (Salesforce Object Query Language)

Utilitzes method chaining and the builder design pattern to elegantly build SOQL queries.

In this context SOQL is a read-only API for querying data from the salesforce REST API. It does **not** support
parameterized inputs. There are **SQL Injection Vulnerabilities** if you allow unsanitized input from the web. 

One method of SQL Injection for example would be to unescape the single quotes which encapsulate strings in 
SOQL and inserting additional fields within a query. This is difficult but especially because this code is 
open sourced it is still a threat.

Sanitize your inputs. Even 

```golang
builder := soql.
    Select("Id", "Name").
    From("Lead").
    Where(soql.And{
        soql.Eq{"FirstName": "Benjamin"},
        // salesforce datetime's use a custom format which types.Datetime accomodates 
        soql.Gt{"CreatedDate": types.NewDatetime(time.Now().Add(time.Hour).String())},
    }).
    Limit(1)
```

SOQL is intended to be used directly with the requests package.

```golang
type entityQuery struct {
    types.QueryResponse
    Records []*entitydefinitions.EntityDefinition `json:"records"`
}

var response entityQuery 
_, err := requests. 
    Sender(salesforce.DefaultClient). 
    URL("tooling/query"). 
    SQLizer(soql.
        Select("QualifiedApiName"). 
        From("EntityDefinition").
        Limit(10)).
    JSON(&response)
```

It has no issues querying parent records or child records (via subqueries). However you should
use generated types with a recursion level of 1 so that they may be unmarshalled automatically.

```bash
go-salesforce-sdk generate Lead ./ leads 1
```

```golang
type leadQuery struct {
    types.QueryResponse
    Records []*leads.Lead `json:"records"`
}

var response2 leadQuery
subquery := soql.Select("Id", "Body").From("Attachments")
_, err = requests.
    Sender(salesforce.DefaultClient).
    URL("query").
    SQLizer(
        soql.
            Select("Id", "Name").
            Column(soql.SubQuery(subquery)).
            From("Lead").
            Limit(10),
    ).
    JSON(&response2)
```