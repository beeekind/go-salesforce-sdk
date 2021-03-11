package soql

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/lann/builder"
)

// Builder is used to compose highly dynamic and reusable SOQL queries
type Builder builder.Builder

// selectData is a data structure holding the individual components of a SOQL query
type selectData struct {
	Prefixes     []SQLizer
	Options      []string
	Columns      []SQLizer
	From         SQLizer
	WhereParts   []SQLizer
	GroupBys     []string
	HavingParts  []SQLizer
	OrderByParts []SQLizer
	Limit        string
	Offset       string
	Suffixes     []SQLizer
}

// Fragment ...
type Fragment string

// ToSQL ...
func (f *Fragment) ToSQL() (sql string, err error) {
	if f == nil {
		return "", errors.New("soql.Fragment method ToSQL() called with nil receiver")
	}
	return string(*f), nil
}

// baseSQLizer is a SQLizer intended for multiple properties of selectData
type baseSQLizer struct {
	predicate interface{}
}

// whereSQLizer is a SQLizer intended for a WHERE clause
type whereSQLizer baseSQLizer

// ToSQL marshalls the selectData builder into an SOQL string
func (b Builder) ToSQL() (string, error) {
	data := builder.GetStruct(b).(selectData)
	return data.toSQL()
}

// MustSQL calls ToSQL and panics instead of returning an error
func (b Builder) MustSQL() string {
	sql, err := b.ToSQL()
	if err != nil {
		panic(err)
	}

	return sql
}

// toSQL composes the properties of selectData into a SOQL query string
func (d *selectData) toSQL() (string, error) {
	if len(d.Columns) == 0 && len(d.Prefixes) == 0 && len(d.Suffixes) == 0 {
		return "", fmt.Errorf("select statements must have at least one column")
	}

	sql := &bytes.Buffer{}

	if len(d.Prefixes) > 0 {
		if err := appendToSQL(d.Prefixes, sql, " "); err != nil {
			return "", fmt.Errorf("%w", err)
		}

		sql.WriteString(" ")
	}

	if len(d.Columns) > 0 {
		sql.WriteString("SELECT ")
	}

	if len(d.Options) > 0 {
		sql.WriteString(strings.Join(d.Options, " "))
		sql.WriteString(" ")
	}

	if len(d.Columns) > 0 {
		if err := appendToSQL(d.Columns, sql, ", "); err != nil {
			return "", fmt.Errorf("%w", err)
		}
	}

	if d.From != nil {
		sql.WriteString(" FROM ")
		if err := appendToSQL([]SQLizer{d.From}, sql, ""); err != nil {
			return "", fmt.Errorf("%w", err)
		}
	}

	if len(d.WhereParts) > 0 {
		sql.WriteString(" WHERE ")
		if err := appendToSQL(d.WhereParts, sql, " AND "); err != nil {
			return "", fmt.Errorf("%w", err)
		}
	}

	if len(d.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(d.GroupBys, ", "))
	}

	if len(d.HavingParts) > 0 {
		sql.WriteString(" HAVING ")
		if err := appendToSQL(d.HavingParts, sql, " AND "); err != nil {
			return "", fmt.Errorf("%w", err)
		}
	}

	if len(d.OrderByParts) > 0 {
		sql.WriteString(" ORDER BY ")
		if err := appendToSQL(d.OrderByParts, sql, ", "); err != nil {
			return "", fmt.Errorf("%w", err)
		}
	}

	if len(d.Limit) > 0 {
		sql.WriteString(" LIMIT ")
		sql.WriteString(d.Limit)
	}

	if len(d.Offset) > 0 {
		sql.WriteString(" OFFSET ")
		sql.WriteString(d.Offset)
	}

	if len(d.Suffixes) > 0 {
		sql.WriteString(" ")

		if err := appendToSQL(d.Suffixes, sql, " "); err != nil {
			return "", fmt.Errorf("%w", err)
		}
	}

	return sql.String(), nil
}

// Prefix appends the given SQL to the beginning of a query
func (b Builder) Prefix(sql string) Builder {
	return builder.Append(b, "Prefixes", Expr(sql)).(Builder)
}

// Options adds an additional clause between SELECT and COLUMNS,
// not sure if this has any application within SOQL but I'll keep it
// in case it proves useful one day (I don't think its unreasonable to
// expect Salesforce to port more features from Postgres/MariaDB/etc
// over time)
func (b Builder) Options(options ...string) Builder {
	return builder.Extend(b, "Options", options).(Builder)
}

// Columns adds result columns to the query.
func (b Builder) Columns(columns ...string) Builder {
	parts := make([]interface{}, 0, len(columns))
	for _, str := range columns {
		parts = append(parts, newBaseSQLizer(str))
	}

	return builder.Extend(b, "Columns", parts).(Builder)
}

// Column adds a result column to the query.
func (b Builder) Column(column interface{}) Builder {
	return builder.Append(b, "Columns", newBaseSQLizer(column)).(Builder)
}

// SetColumns is like Columns but it replaces existing columns rather then extending them
func (b Builder) SetColumns(columns ...string) Builder {
	parts := make([]SQLizer, 0, len(columns))
	for _, str := range columns {
		parts = append(parts, newBaseSQLizer(str))
	}

	return builder.Set(b, "Columns", parts).(Builder)
}

// From sets the FROM clause of the query.
func (b Builder) From(from string) Builder {
	return builder.Set(b, "From", newBaseSQLizer(from)).(Builder)
}

// FromSelect sets a subquery into the FROM clause of the query.
func (b Builder) FromSelect(from Builder, alias string) Builder {
	// Prevent misnumbered parameters in nested selects (#183).
	return builder.Set(b, "From", Alias(from, alias)).(Builder)
}

// Where builds a where clause using nested SQLizers. Each SQLizer is responsible
// for serializing itself and any nested SQLizer, so all the SelectBuilder needs to do
// is maintain and apply a list of top-level SQLizers. See expressions.go for how the SQLizers
// work in practice.
//
// The following types are accepted as a predicate (as seen in newWhereSQLizer()):
//
// switch predicate.(type) {
// case nil:
//	   // noop
//	   return "", nil
// case SQLizer:
//     return pred.ToSQL()
// case map[string]interface{}:
//	   return Eq(pred).ToSQL()
// case string:
// 	   return pred, nil
func (b Builder) Where(predicate interface{}) Builder {
	if predicate == nil || predicate == "" {
		return b
	}

	return builder.Append(b, "WhereParts", newWhereSQLizer(predicate)).(Builder)
}

// GroupBy appends group by claus(es) to selectData
func (b Builder) GroupBy(groupBys ...string) Builder {
	return builder.Extend(b, "GroupBys", groupBys).(Builder)
}

// Having appends a having clause to selectData
func (b Builder) Having(pred interface{}) Builder {
	return builder.Append(b, "HavingParts", newWhereSQLizer(pred)).(Builder)
}

// OrderByClause appends an OrderBy clause to selectData
func (b Builder) OrderByClause(pred interface{}) Builder {
	return builder.Append(b, "OrderByParts", newBaseSQLizer(pred)).(Builder)
}

// OrderBy appends order by clauses to selectData
func (b Builder) OrderBy(orderBys ...string) Builder {
	for _, orderBy := range orderBys {
		b = b.OrderByClause(orderBy)
	}

	return b
}

// Limit appends a limit to the query
func (b Builder) Limit(limit int) Builder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(Builder)
}

// RemoveLimit removes a limit from a query
func (b Builder) RemoveLimit() Builder {
	return builder.Delete(b, "Limit").(Builder)
}

// Offset appends an offset to the query
func (b Builder) Offset(offset int) Builder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(Builder)
}

// RemoveOffset removes an offset from the queery
func (b Builder) RemoveOffset() Builder {
	return builder.Delete(b, "Offset").(Builder)
}

// Suffix adds an expression to the end of the query
func (b Builder) Suffix(sql string) Builder {
	return builder.Append(b, "Suffixes", Expr(sql)).(Builder)
}

// newBaseSQLizer allows multiple types for a predicate argument to be
// converted into sql.
func newBaseSQLizer(predicate interface{}) SQLizer {
	return baseSQLizer{predicate}
}

// ToSql ...
func (part baseSQLizer) ToSQL() (sql string, err error) {
	switch predicate := part.predicate.(type) {
	case nil:
		// no-op
	case SQLizer:
		sql, err = predicate.ToSQL()
	case string:
		sql = predicate
	default:
		err = fmt.Errorf("expected string or Sqlizer, not %T", predicate)
	}
	return sql, err
}

// appendToSQL
func appendToSQL(parts []SQLizer, w io.Writer, sep string) error {
	for index, p := range parts {
		partSQL, err := p.ToSQL()
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if len(partSQL) == 0 {
			continue
		}

		// append a separator after every character
		if index > 0 {
			_, err := io.WriteString(w, sep)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		_, err = io.WriteString(w, partSQL)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// newWhereSQLizer allows multiple types for a predicate argument to be
// converted into sql.
func newWhereSQLizer(predicate interface{}) SQLizer {
	return &whereSQLizer{predicate}
}

// ToSQL
func (part *whereSQLizer) ToSQL() (sql string, err error) {
	switch pred := part.predicate.(type) {
	case nil:
		// noop
		return "", nil
	case SQLizer:
		return pred.ToSQL()
	case map[string]interface{}:
		return Eq(pred).ToSQL()
	case string:
		return pred, nil
	default:
		return "", fmt.Errorf("expected string-keyed map, SQLizer, or string, not %T", pred)
	}
}
