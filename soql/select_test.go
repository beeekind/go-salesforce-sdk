package soql_test

import (
	"testing"

	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/stretchr/testify/require"
)

type selectInput struct {
	builder soql.Builder
}

type selectOutput struct {
	expectedSQL string
	expectedErr error
	description string
}

var selectBuilderTests = map[*selectInput]*selectOutput{
	{soql.
		Select("a", "b", "c").
		From("Lead").
		Prefix("With prefix As foo").
		Columns("Name", "Number", "Id").
		Where(soql.Eq{"a": 1}).
		Where(soql.Or{soql.Eq{"b": "two"}, soql.Expr("a < 2")}).
		GroupBy("l").
		Having("m = n").
		OrderBy("Name ASC").
		Limit(1200).
		Offset(200).
		Suffix("RETURNING Id"),
	}: {"With prefix As foo SELECT a, b, c, Name, Number, Id FROM Lead WHERE a = 1 AND (b = 'two' OR a < 2) GROUP BY l HAVING m = n ORDER BY Name ASC LIMIT 1200 OFFSET 200 RETURNING Id", nil, "an exhaustive select query"},
	{soql.
		Select("a", "b", "c").
		From("Account").
		Where(soql.Eq{
			"foo": "bar' AND zar = 'lar",
		}),
	}: {"SELECT a, b, c FROM Account WHERE foo = 'bar' AND zar = 'lar'", nil, "sql injection in where clause is escaped"},
}

func TestSelectBuilder(t *testing.T) {
	for in, out := range selectBuilderTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.builder.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}
