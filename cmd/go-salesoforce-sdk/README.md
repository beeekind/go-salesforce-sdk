# go-salesforce-sdk CLI

This command line tool is used to executing the ls and generate commands with the go-salesforce-sdk.

ls will list available Salesforce objects in your organization that can be generated.

generate will generate a golang package containing one or more type definitions for your object.

# Usage 

```bash
git clone https://github.com/beeekind/go-salesforce-sdk
go install go-salesforce-sdk/cmd/go-salesforce-sdk

go-salesforce-sdk ls

go-salesforce-sdk generate Lead ./ leads 0 
go-salesforce-sdk generate Account ./ accounts 0 
go-salesforce-sdk generate Contact ./ contacts 1
```

# Arguments passed to the generate command 

args[0] the name of the installed program which defaults to the directory found in go-salesforce-sdk/cmd
args[1] the initial command {ls|generate}
args[2] the object name case-sensitive
args[3] the output directory
args[4] the output package name 
args[5] the recursionLevel - indicating how many levels of relations we should generate types for (useful when doing subqueries on generated types)