package soql_test

import (
	"errors"
	"testing"

	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/stretchr/testify/require"
)

type exprInput struct {
	expr soql.SQLizer
}

type exprOutput struct {
	expectedSQL string
	expectedErr error
	description string
}

var eqTests = map[*exprInput]*exprOutput{
	{soql.Eq{"ID": "foobar"}}:                       {"ID = 'foobar'", nil, "where clause with string input"},
	{soql.Eq{"Amount": 100}}:                        {"Amount = 100", nil, "where clause with int32 input"},
	{soql.Eq{"ID": "foobar", "Amount": 100}}:        {"Amount = 100 AND ID = 'foobar'", nil, "where clause with string & int32 input"},
	{soql.Eq{}}:                                     {"", nil, "empty where clause"},
	{soql.Eq{"Things": []int{1, 2, 3}}}:             {"Things IN (1, 2, 3)", nil, "where clause with slice input"},
	{soql.Eq{"Things": []interface{}{1, "two", 3}}}: {"Things IN (1, 'two', 3)", nil, "where clause with generic slice input"},
	{soql.Eq{"bar": nil}}:                           {"bar IS null", nil, "where with null value"},
}

func TestEq(t *testing.T) {
	for in, out := range eqTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var notEqTests = map[*exprInput]*exprOutput{
	{soql.NotEq{"ID": "foobar"}}:                       {"ID != 'foobar'", nil, "where clause with string input"},
	{soql.NotEq{"Amount": 100}}:                        {"Amount != 100", nil, "where clause with int32 input"},
	{soql.NotEq{"ID": "foobar", "Amount": 100}}:        {"Amount != 100 AND ID != 'foobar'", nil, "where clause with string & int32 input"},
	{soql.NotEq{}}:                                     {"", nil, "empty where clause"},
	{soql.NotEq{"Things": []int{1, 2, 3}}}:             {"Things NOT IN (1, 2, 3)", nil, "where clause with slice input"},
	{soql.NotEq{"Things": []interface{}{1, "two", 3}}}: {"Things NOT IN (1, 'two', 3)", nil, "where clause with generic slice input"},
	{soql.NotEq{"bar": nil}}:                           {"bar IS NOT null", nil, "where with null value"},
}

func TestNotEq(t *testing.T) {
	for in, out := range notEqTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var ltTests = map[*exprInput]*exprOutput{
	{soql.Lt{"a": 1}}:           {"a < 1", nil, "lt clause with int32 value"},
	{soql.Lt{"c": "a"}}:         {"c < 'a'", nil, "lt clause with string value"},
	{soql.Lt{"c": "a", "a": 1}}: {"a < 1 AND c < 'a'", nil, "lt clause with string and int32 value"},
	{soql.Lt{}}:                 {"", nil, "empty lt clause"},
	{soql.Lt{"a": nil}}:         {"", errors.New("cannot use null with Lt or Gt operators"), "nil lt value"},
}

func TestLt(t *testing.T) {
	for in, out := range ltTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var lteTests = map[*exprInput]*exprOutput{
	{soql.LtOrEq{"a": 1}}:           {"a <= 1", nil, "lte clause with int32 value"},
	{soql.LtOrEq{"c": "a"}}:         {"c <= 'a'", nil, "lte clause with string value"},
	{soql.LtOrEq{"c": "a", "a": 1}}: {"a <= 1 AND c <= 'a'", nil, "lte clause with string and int32 value"},
	{soql.LtOrEq{}}:                 {"", nil, "empty lte clause"},
	{soql.LtOrEq{"a": nil}}:         {"", errors.New("cannot use null with Lt or Gt operators"), "nil lt value"},
}

func TestLtEqual(t *testing.T) {
	for in, out := range lteTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var gtTests = map[*exprInput]*exprOutput{
	{soql.Gt{"a": 1}}:           {"a > 1", nil, "gt clause with int32 value"},
	{soql.Gt{"c": "a"}}:         {"c > 'a'", nil, "gt clause with string value"},
	{soql.Gt{"c": "a", "a": 1}}: {"a > 1 AND c > 'a'", nil, "gt clause with string and int32 value"},
	{soql.Gt{}}:                 {"", nil, "empty gt clause"},
	{soql.Gt{"a": nil}}:         {"", errors.New("cannot use null with Lt or Gt operators"), "nil lt value"},
}

func TestGt(t *testing.T) {
	for in, out := range gtTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var gteTests = map[*exprInput]*exprOutput{
	{soql.GtOrEq{"a": 1}}:           {"a >= 1", nil, "gte clause with int32 value"},
	{soql.GtOrEq{"c": "a"}}:         {"c >= 'a'", nil, "gte clause with string value"},
	{soql.GtOrEq{"c": "a", "a": 1}}: {"a >= 1 AND c >= 'a'", nil, "gte clause with string and int32 value"},
	{soql.GtOrEq{}}:                 {"", nil, "empty gt clause"},
	{soql.GtOrEq{"a": nil}}:         {"", errors.New("cannot use null with Lt or Gt operators"), "nil lt value"},
}

func TestGte(t *testing.T) {
	for in, out := range gteTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var likeTests = map[*exprInput]*exprOutput{
	{soql.Like{"a": "foo"}}:             {"a LIKE 'foo'", nil, "like clause with basic string"},
	{soql.Like{"b": "zoo", "a": "foo"}}: {"a LIKE 'foo' AND b LIKE 'zoo'", nil, "like clause with basic string"},
	{soql.Like{"a": "%foo%"}}:           {"a LIKE '%foo%'", nil, "like clause with catch-all"},
	{soql.Like{"a": ""}}:                {"a LIKE ''", nil, "like clause with empty string value"},
	{soql.Like{}}:                       {"", nil, "empty like clause"},
}

func TestLike(t *testing.T) {
	for in, out := range likeTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var notLikeTests = map[*exprInput]*exprOutput{
	{soql.NotLike{"a": "foo"}}:   {"a NOT LIKE 'foo'", nil, "not like clause with basic string"},
	{soql.NotLike{"a": "%foo%"}}: {"a NOT LIKE '%foo%'", nil, "not like clause with catch-all"},
	{soql.NotLike{"a": ""}}:      {"a NOT LIKE ''", nil, "not like clause with empty string value"},
	{soql.NotLike{}}:             {"", nil, "empty like clause"},
}

func TestNotLike(t *testing.T) {
	for in, out := range notLikeTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var conjugationTests = map[*exprInput]*exprOutput{
	{soql.And{
		soql.Eq{"a": 1},
		soql.Lt{"c": 0},
		soql.And{
			soql.NotEq{"b": 2},
			soql.GtOrEq{"d": "foo"},
		},
	}}: {"(a = 1 AND c < 0 AND (b != 2 AND d >= 'foo'))", nil, "nested conjugation with And{}"},
	{soql.Or{
		soql.Eq{"a": 1},
		soql.Lt{"c": 0},
		soql.Or{
			soql.NotEq{"b": 2},
			soql.GtOrEq{"d": "foo"},
		},
	}}: {"(a = 1 OR c < 0 OR (b != 2 OR d >= 'foo'))", nil, "nested conjugation with Or{}"},
}

func TestConjugations(t *testing.T) {
	for in, out := range conjugationTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}

var subqueryTests = map[*exprInput]*exprOutput{
	{soql.Select("three").Column(soql.SubQuery(soql.Select("one").From("two"))).From("four")}: {
		"SELECT three, (SELECT one FROM two) FROM four", nil, "basic subquery",
	},
}

func TestSubQuery(t *testing.T) {
	for in, out := range subqueryTests {
		t.Run(out.description, func(t *testing.T) {
			sql, err := in.expr.ToSQL()
			require.Equal(t, out.expectedErr, err)
			require.Equal(t, out.expectedSQL, sql)
		})
	}
}
