package soql 

import "github.com/lann/builder"

// SB is a parent builder for other builders, e.g. SelectBuilder.
var sb = Builder(builder.EmptyBuilder)

// Empty is a empty builder useful for 
var Empty = sb 

// SQLizer defines a component which can be converted to SQL. It is the base building block 
// of this package.
type SQLizer interface {
	ToSQL()(string, error)
}

// Select returns a SelectBuilder with the given columns 
func Select(columns ...string) Builder {
	return sb.Columns(columns...)
}

// String returns a SelectBuilder composed as a singular string. This builder should not be extended
// as the query is stored and prepended as a prefix to form the entire query. Its used to satisy the
// requests.SQLizer signature when a string is the only input.
func String(query string) Builder {
	return sb.Prefix(query)
}

// init is necessary for preparing data structures used by builder.Builder
func init() {
	builder.Register(Builder{}, selectData{})
}
