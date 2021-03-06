package soql

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type expr struct {
	sql string
}

// Expr represents a sql expression that is already fully formed 
func Expr(sql string) SQLizer {
	return expr{sql}
}

// ToSQL ... 
func (e expr) ToSQL() (sql string, err error) {
	return e.sql, nil
}

type aliasExpr struct {
	expr  SQLizer
	alias string
}

// Alias produces an SQL alias expression
func Alias(expr SQLizer, alias string) SQLizer {
	return aliasExpr{expr, alias}
}

// ToSQL ... 
func (e aliasExpr) ToSQL() (sql string, err error) {
	sql, err = e.expr.ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return fmt.Sprintf("(%s) As %s", sql, e.alias), nil
}

type subquery struct {
	expr SQLizer 
}

// SubQuery ... 
func SubQuery(expr SQLizer) SQLizer {
	return subquery{expr}
}

func (sq subquery) ToSQL()(sql string, err error){
	sql, err = sq.expr.ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return fmt.Sprintf("(%s)", sql), nil
}

// Eq produces a complex clause for use by Where/Having methods. It may contain nested values
// and lists which are converted into IN queries. 
type Eq map[string]interface{}

// toSql converts a map[string]interface{} into a set of sql expressions. It is re-used by
// NotEq by inverting the comparison operators via the usueNotOperator argument. 
//
// Note that we escapeSingleQuote() properties and values where applicable. This will use the SOQL
// escape character '\' when a single quote that is not at the beginning or end of a string is identified. 
// it is not a silver bullet to SQL injection within a Where clause and you should still validate your
// user input based on your own threat model. 
func (eq Eq) toSQL(useNotOperator bool) (sql string, err error) {
	if len(eq) == 0 {
		return "", nil
	}

	var (
		exprs       []string
		equalOpr    = "="
		inOpr       = "IN"
		nullOpr     = "IS"
		inEmptyExpr = "()"
	)

	if useNotOperator {
		equalOpr = "!="
		inOpr = "NOT IN"
		nullOpr = "IS NOT"
		inEmptyExpr = "()"
	}

	// sort the keys so we can produce an indempotent query that can be tested
	sortedKeys := getSortedKeys(eq)
	for _, key := range sortedKeys {
		var expr string
		val := eq[key]

		// 1: inspect the type of val:interface{}
		r := reflect.ValueOf(val)
		if r.Kind() == reflect.Ptr {
			if r.IsNil() {
				val = nil
			} else {
				val = r.Elem().Interface()
			}
		}

		// 2: if val is nil return a "key is/is not NULL" clause
		if val == nil {
			expr = fmt.Sprintf("%s %s null", key, nullOpr)
			exprs = append(exprs, expr)
			continue 
		} 
		
		// 3: if val is list prepare an "in/not in ()" clause"
		if isListType(val) {
			valVal := reflect.ValueOf(val)
			if valVal.Len() == 0 {
				// append empty array notation
				expr = inEmptyExpr
			} else {
				var items []string
				for i := 0; i < r.Len(); i++ {
					v := reflect.ValueOf(r.Index(i).Interface())
					if v.Kind() == reflect.String {
						items = append(items, fmt.Sprintf("'%s'", r.Index(i).Interface().(string)))
					} else {
						items = append(items, fmt.Sprintf("%v", r.Index(i)))
					}
				}
				// append list notation
				expr = fmt.Sprintf("%s %s (%s)", key, inOpr, strings.Join(items, ", "))
			}

			exprs = append(exprs, expr)
			continue 
		} 
		
		// 4: if val is a string escape any single quotes not at the beginning or end of the string 
		if r.Kind() == reflect.String {
			value := reflect.ValueOf(val).String()
			expr = fmt.Sprintf("%s %s '%s'", key, equalOpr, value)
			exprs = append(exprs, expr)
			continue 	
		} 

		// 5: else prepare an "key =/!= val" claue
		expr = fmt.Sprintf("%s %s %v", key, equalOpr, val)
		exprs = append(exprs, expr)
	}

	// join the WHERE clause together using AND
	sql = strings.Join(exprs, " AND ")
	return sql, nil
}

// ToSQL ...
func (eq Eq) ToSQL() (string, error) {
	return eq.toSQL(false)
}

// NotEq is the inverse of Eq. See the documentation for Eq.
type NotEq Eq

// ToSQL reuses Eq.toSQL() to convert NotEq to an sql string 
func (neq NotEq) ToSQL() (sql string, err error) {
	return Eq(neq).toSQL(true)
}

// Like prepares a LIKE clause. 
type Like map[string]string

// toSQL ...
func (lk Like) toSQL(opr string) (sql string, err error) {
	var exprs []string
	sortedKeys := getSortedStringKeys(lk)
	for _, key := range sortedKeys {
		expr := fmt.Sprintf("%s %s '%s'", key, opr, lk[key])
		exprs = append(exprs, expr)
	}

	sql = strings.Join(exprs, " AND ")
	return sql, nil
}

// ToSQL ...
func (lk Like) ToSQL() (sql string, err error) {
	return lk.toSQL("LIKE")
}

// NotLike is the inverse of Like.
type NotLike Like

// ToSQL ...
func (nlk NotLike) ToSQL() (sql string, err error) {
	return Like(nlk).toSQL("NOT LIKE")
}

// Lt prepares a Less than clause. k < v
type Lt map[string]interface{}

// toSQL is a reusable method for Lt, Gt, LtOrEq, GtOrEq
func (lt Lt) toSQL(opposite bool, orEqual bool) (sql string, err error) {
	var exprs []string
	opr := "<"

	if opposite {
		opr = ">"
	}

	if orEqual {
		opr = fmt.Sprintf("%s%s", opr, "=")
	}

	sortedKeys := getSortedKeys(lt)
	for _, k := range sortedKeys {
		var expr string
		v := lt[k]

		if v == nil {
			return "", fmt.Errorf("cannot use null with Lt or Gt operators")
		}

		if isListType(v) {
			return "", fmt.Errorf("cannot use array or slice with Gt or Lt operators")
		}

		if reflect.ValueOf(v).Kind() == reflect.String {
			expr = fmt.Sprintf("%s %s '%s'", k, opr, v)
		} else {
			expr = fmt.Sprintf("%s %s %v", k, opr, v)
		}

		exprs = append(exprs, expr)
	}

	sql = strings.Join(exprs, " AND ")
	return sql, nil
}

// ToSQL ...
func (lt Lt) ToSQL() (sql string, err error) {
	return lt.toSQL(false, false)
}

// LtOrEq k <= v
type LtOrEq Lt

// ToSQL ...
func (ltOrEq LtOrEq) ToSQL() (sql string, err error) {
	return Lt(ltOrEq).toSQL(false, true)
}

// Gt k > v
type Gt Lt

// ToSQL ...
func (gt Gt) ToSQL() (sql string, err error) {
	return Lt(gt).toSQL(true, false)
}

// GtOrEq k >= v
type GtOrEq Lt

// ToSQL ...
func (gtOrEq GtOrEq) ToSQL() (sql string, err error) {
	return Lt(gtOrEq).toSQL(true, true)
}

// conj short for conjugation is a reusable data structure for preparing AND and OR queries which
// can be used to great effect to form deeply nested complex queries
type conj []SQLizer

func (c conj) join(separator string, defaultExpr string) (sql string, err error) {
	if len(c) == 0 {
		return defaultExpr, nil
	}

	var sqlParts []string
	for _, sqlizer := range c {
		partSQL, err := sqlizer.ToSQL()
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}

		if partSQL != "" {
			sqlParts = append(sqlParts, partSQL)
		}
	}

	if len(sqlParts) > 0 {
		sql = fmt.Sprintf("(%s)", strings.Join(sqlParts, separator))
	}

	return sql, nil
}

// And reuses conj.join to form AND clauses
type And conj

// ToSQL ...
func (a And) ToSQL() (sql string, err error) {
	return conj(a).join(" AND ", "")
}

// Or reuses conj.Join to form OR clauses 
type Or conj

// ToSQL ...
func (o Or) ToSQL() (sql string, err error) {
	return conj(o).join(" OR ", "")
}

func getSortedKeys(exp map[string]interface{}) []string {
	sortedKeys := make([]string, 0, len(exp))
	for k := range exp {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func getSortedStringKeys(exp map[string]string) []string {
	sortedKeys := make([]string, 0, len(exp))
	for k := range exp {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func isListType(val interface{}) bool {
	valVal := reflect.ValueOf(val)
	return valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice
}

// I'm putting this method on ice until I have time to write a proper parser to do this well
/**
// escapeSingleQuote is a naive escape mechanism for single quotes which are not 
// at the beginning or end of the given string 
func escapeSingleQuote(str string) string {
	return str 
	
	var singleQuote rune = '\''
	var escapeKey rune = '\\'

	var runes []rune
	var lastChar rune
	for idx, char := range str {
		// if there is a singleQuote that is not at the beginning or end
		// and lastChar is not an escapeKey
		// then append an escapeKey as the preceding character
		if idx != 0 && idx != len(str)-1 && char == singleQuote {
			if lastChar != escapeKey {
				runes = append(runes, escapeKey)
			}
		}
		runes = append(runes, char)
		lastChar = char
	}

	return string(runes)
}
*/