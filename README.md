# Go Salesforce SDK

go-salesforce-sdk is an unofficial SDK for the Salesforce REST API. 

Checkout our [release notes](https://github.com/beeekind/go-salesforce-sdk/releases) for information about the latest bug fixes, updates, and features added to the sdk.

Jump To:

* [Features](https://github.com/beeekind/go-salesforce-sdk#Features)
* [Examples](https://github.com/beeekind/go-salesforce-sdk#Examples)
* [Packages](https://github.com/beeekind/go-salesforce-sdk#Packages)
* [Contribute](https://github.com/beeekind/go-salesforce-sdk#Contribute)
* [Credits](https://github.com/beeekind/go-salesforce-sdk#Credits)

# Features 
--- 

- [x] Generate type definitions 
    - [x] Standard Objects 
    - [x] Tooling Objects 
- [x] Authentication Mechanisms
    - [x] JWT flow (recommended)
    - [x] Password flow
- [x] Concurrent Processing 
    - [x] Pre-compute paginated resources for retrieving all paginated records quickly
- [x] HTTP Client Wrapper
    - [x] HttpTransport customization
    - [x] Ratelimiting 
- [x] Querybuilder (based on [squirrel](https://github.com/Masterminds/squirrel))
    - [x] Select
    - [x] Where
        - [x] Equality | Inequality 
        - [x] Subqueries 
        - [x] Like | NotLike
        - [x] GT | LT | GTE | LTE 
        - [x] Conjugations (And | Or)
    - [x] Group By
    - [x] Order By(s)
    - [x] Limit
    - [x] Offset 
    - [x] Prefixes 
    - [x] Suffixes 
- [x] Request Builder 
    - [x] URL composition 
    - [x] Method
    - [x] URL parameters 
    - [x] Headers 
    - [x] SOQL embeding 
    - [x] Build as http.Response
    - [x] Unmarshal into struct 
    - [x] application/x-www-form-urlencoded submissions
- [x] Custom Types
    - [x] Nullable (bool | string | int | float)
    - [x] Date / Datetime 
    - [x] Address 
    - [x] [AlmostBool](https://github.com/beeekind/go-salesforce-sdk/blob/892727d16ecf24f6cadd0a287bc06f890d47657f/types/absurd.go#L16)
- [x] Metadata Response Types 
    - [x] /describe 
    - [x] /describe/{objectName}
    - [x] Limits 
    - [x] Query 
    - [x] Tooling/query
- [x] Bulk API v2
    - [x] Ingest 
    - [x] Query 
- [x] Composite 
    - [x] Create 
    - [x] Read 
    - [x] Update 
    - [x] Delete 
- [x] Tree 
    - [x] ParseNode(typeDefinition)
    - [x] Recursive object nesting 
- [x] Execute Anonymous Apex
    - [x] SingleEmailMessage 

- ### And much more...

# Examples 

This SDK consists of a high level and a low level API. The high level API can be found in the root package while all other packages should be considered low level.

We recommend you actually use the low level API as it is much more configurable and still simple to use.

Use these examples, all *_test.go files, the root package, and the godoc, as documentation for using this SDK.

# Packages 

| Package            | Link | Description                                               | 
| ------------------ | ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| go-salesforce-sdk  | [Link](https://github.com/beeekind/go-salesforce-sdk)              | Root package with high level API methods. Other packages should be considered the low-level API |
| apex               | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/apex)      | Demonstrates using the Execute Anonymous Apex endpoint to send an email                        | 
| bulk               | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/bulk)      | Methods for bulk uploading and retrieving objects as text/csv                                  | 
| client             | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/client)    | Wraps http.Client and provides authentication, ratelimiting, and http.Transport customization  | 
| composite          | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/composite) | Provides Create, Read, Update, and Delete, operations with the Composite API  | 
| internal           | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/internal)  | Internal use only.  | 
| metadata           | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/metadata)  | Content Cell  | 
| requests           | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/requests)  | HTTP request building using the builder design pattern  | 
| soql               | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/soql)      | SOQL (Salesforce Object Query Language) building using the builder design pattern  |
| templates          | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/templates) | Templates for generating Type definitions, Response types, Apex code, and other artifacts  | 
| tree               | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/tree)      | Tree API operations for saving nested objects based on their relations. Uses generated types.  |
| types              | [Link](https://github.com/beeekind/go-salesforce-sdk/tree/main/types)     | Type definitions for Salesforce specific types like Date and Datetime  | 

# Contribute



# Credits 
