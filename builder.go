package query_builder

import (
	"fmt"
	"strings"
)

const (
	Question = "?"    // Question is the placeholder format for MySQL, SQLite, and MariaDB.
	Dollar   = "$%d"  // Dollar is the placeholder format for PostgreSQL ($1, $2, $3...).
	Colon    = ":%d"  // Colon is the placeholder format for Oracle and some JDBC drivers (:1, :2, :3...).
	AtP      = "@p%d" // AtP is the placeholder format for SQL Server (@p1, @p2, @p3...).
)

// QueryBuilder is a fluent, chainable SQL query builder for Go.
// It supports dynamic column selection, parameterized WHERE clauses,
// JOINs, GROUP BY, ORDER BY, LIMIT, and OFFSET with automatic
// placeholder formatting for multiple database dialects.
type QueryBuilder struct {
	placeholder string

	columns []string
	table   string
	joins   []string
	where   []string
	groupBy string
	orderBy string
	limit   string
	offset  string

	selectArgs []any
	joinArgs   []any
	whereArgs  []any
}

// NewQueryBuilder initializes and returns a new QueryBuilder instance
// with default placeholder format set to Dollar ("$%d").
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		placeholder: Dollar,
	}
}

// PlaceholderFormat sets the placeholder pattern for parameterized queries.
// Accepts format strings like "$%d" (PostgreSQL), "?%d", ":p%d", etc.
// Returns the QueryBuilder for method chaining.
func (b *QueryBuilder) PlaceholderFormat(placeholder string) *QueryBuilder {
	b.placeholder = placeholder
	return b
}

// Placeholders generates a string of repeated placeholders separated by sep.
// Uses the QueryBuilder's current placeholder format and repeats it count times.
// Example: b.Placeholders(", ", 3) with Dollar → "$1, $2, $3".
func (b *QueryBuilder) Placeholders(sep string, count int) string {

	return Placeholders(b.placeholder, sep, count)
}

// Select sets the SELECT clause, replacing any previously set columns.
// Accepts a column expression (e.g., "id", "name AS username") and optional
// parameterized arguments. Returns the QueryBuilder for chaining.
func (b *QueryBuilder) Select(sel string, args ...any) *QueryBuilder {

	b.columns = make([]string, 0, 1)
	b.selectArgs = make([]any, 0, len(args))
	return b.AndSelect(sel, args...)
}

// AndSelect appends additional column expressions to the SELECT clause.
// Accepts a column expression and optional parameterized arguments.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) AndSelect(sel string, args ...any) *QueryBuilder {

	b.columns = append(b.columns, sel)
	b.selectArgs = append(b.selectArgs, args...)
	return b
}

// FilterSelect sets the SELECT clause using a Filter mapping and a list
// of requested field names. Only fields present in the Filter are included,
// with automatic aliasing (e.g., "full_name AS name"). Returns the QueryBuilder.
func (b *QueryBuilder) FilterSelect(f *Filter, inputFields []string) *QueryBuilder {

	b.columns = make([]string, 0, len(inputFields))
	b.columns = f.Filter(inputFields)

	for _, j := range f.FilterJoins(inputFields) {
		b.Join(j.joinType, j.table, j.on, j.args...)
	}

	return b
}

// AndFilterSelect appends dynamically filtered columns and their associated JOIN clauses
// to the current SELECT statement. It evaluates inputFields against the Filter mapping,
// adds matching column expressions to the query, and registers any required JOINs.
// Unlike FilterSelect, this method preserves existing columns instead of replacing them.
// Returns the QueryBuilder for method chaining.
func (b *QueryBuilder) AndFilterSelect(f *Filter, inputFields []string) *QueryBuilder {

	b.columns = append(b.columns, f.Filter(inputFields)...)

	for _, j := range f.FilterJoins(inputFields) {
		b.Join(j.joinType, j.table, j.on, j.args...)
	}

	return b
}

// From sets the main table for the SELECT query.
// Accepts a table name (optionally with alias, e.g., "users AS u").
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) From(table string) *QueryBuilder {

	b.table = table
	return b
}

// Join adds a generic JOIN clause (INNER, LEFT, RIGHT, etc.) to the query.
// Accepts join type, table name, ON condition, and optional parameterized arguments.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) Join(joinType, table, on string, args ...any) *QueryBuilder {

	b.joins = append(b.joins, fmt.Sprintf("%s %s ON %s", joinType, table, on))
	b.joinArgs = append(b.joinArgs, args...)
	return b
}

// InnerJoin adds an INNER JOIN clause. Shorthand for Join("INNER JOIN", ...).
func (b *QueryBuilder) InnerJoin(table string, on string, args ...any) *QueryBuilder {

	b.Join("INNER JOIN", table, on, args...)
	return b
}

// LeftJoin adds a LEFT JOIN clause. Shorthand for Join("LEFT JOIN", ...).
func (b *QueryBuilder) LeftJoin(table string, on string, args ...any) *QueryBuilder {

	return b.Join("LEFT JOIN", table, on, args...)
}

// RightJoin adds a RIGHT JOIN clause. Shorthand for Join("RIGHT JOIN", ...).
func (b *QueryBuilder) RightJoin(table string, on string, args ...any) *QueryBuilder {

	return b.Join("RIGHT JOIN", table, on, args...)
}

// Where sets the WHERE clause, replacing any previously set conditions.
// Accepts a SQL template with placeholders and optional parameterized arguments.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) Where(sqltpl string, args ...any) *QueryBuilder {

	b.where = make([]string, 0)
	b.AndWhere(sqltpl, args...)
	return b
}

// AndWhere appends an additional condition to the WHERE clause using AND logic.
// Accepts a SQL template with placeholders and optional parameterized arguments.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) AndWhere(sqltpl string, args ...any) *QueryBuilder {

	b.where = append(b.where, sqltpl)
	b.whereArgs = append(b.whereArgs, args...)
	return b
}

func sliceNotNils(s []any) bool {

	for _, v := range s {
		if v != nil {
			return true
		}
	}
	return false
}

// FilterWhere sets the WHERE clause only if args are provided (non-nil).
// Useful for conditional filtering based on user input. Returns the QueryBuilder.
func (b *QueryBuilder) FilterWhere(sqltpl string, args ...any) *QueryBuilder {

	if sliceNotNils(args) {
		b.Where(sqltpl, args...)
	}
	return b
}

// AndFilterWhere appends a WHERE condition only if args are provided (non-nil).
// Useful for optional filters in dynamic queries. Returns the QueryBuilder.
func (b *QueryBuilder) AndFilterWhere(sqltpl string, args ...any) *QueryBuilder {

	if sliceNotNils(args) {
		b.AndWhere(sqltpl, args...)
	}
	return b
}

// GroupBy sets the GROUP BY clause. Accepts a comma-separated list of columns.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) GroupBy(groupBy string) *QueryBuilder {

	b.groupBy = groupBy
	return b
}

// OrderBy sets the ORDER BY clause. Accepts a column expression with direction
// (e.g., "created_at DESC", "name ASC"). Returns the QueryBuilder for chaining.
func (b *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {

	b.orderBy = orderBy
	return b
}

// Limit sets the LIMIT clause. Accepts a non-negative integer.
// The value is parameterized using the current placeholder format.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) Limit(limit int) *QueryBuilder {

	b.limit = fmt.Sprintf("%d", limit)
	return b
}

// Offset sets the OFFSET clause. Accepts a non-negative integer.
// The value is parameterized using the current placeholder format.
// Returns the QueryBuilder for chaining.
func (b *QueryBuilder) Offset(offset int) *QueryBuilder {

	b.offset = fmt.Sprintf("%d", offset)
	return b
}

// GetColumns returns the current SELECT columns and their associated arguments.
// Useful for inspecting or modifying the query before final generation.
func (b *QueryBuilder) GetColumns() ([]string, []any) {
	return b.columns, b.selectArgs
}

// GetWhere returns the combined WHERE clause as a string and its arguments.
// If no conditions are set, returns "1=1" (always true) with nil arguments.
func (b *QueryBuilder) GetWhere() (string, []any) {
	if len(b.where) == 0 {
		return "1=1", nil
	}
	return strings.Join(b.where, " AND "), b.whereArgs
}

func (b *QueryBuilder) createSelect() (string, []any) {

	if len(b.columns) == 0 {
		b.columns = append(b.columns, "*")
	}

	sel := "SELECT " + strings.Join(b.columns, ", ") + " "
	return sel, b.selectArgs
}

func (b *QueryBuilder) createFrom() string {

	return "FROM " + b.table + " "
}

func (b *QueryBuilder) createJoins() (string, []any) {

	joins := ""

	for _, join := range b.joins {
		joins += join + " "
	}

	return joins, b.joinArgs
}

func (b *QueryBuilder) createWhere() (string, []any) {
	where, whereArgs := b.GetWhere()
	return "WHERE " + where + " ", whereArgs
}

func (b *QueryBuilder) createGroupBy() string {

	if b.groupBy == "" {
		return ""
	}
	return "GROUP BY " + b.groupBy + " "
}

func (b *QueryBuilder) createOrderBy() string {

	if b.orderBy == "" {
		return ""
	}
	return "ORDER BY " + b.orderBy + " "
}

func (b *QueryBuilder) createLimit() (string, []any) {

	if b.limit == "" {
		return "", nil
	}

	limitValues := []any{b.limit}
	return "LIMIT " + b.placeholder + " ", limitValues
}

func (b *QueryBuilder) createOffset() (string, []any) {

	if b.offset == "" {
		return "", nil
	}

	i := 1

	if b.limit != "" {
		i++
	}

	offsetValues := []any{b.offset}
	return "OFFSET " + b.placeholder + " ", offsetValues
}

func (b *QueryBuilder) placenums(args []any, start int) []any {

	placenums := make([]any, 0, len(args))

	for i := range len(args) {
		placenums = append(placenums, start+int(i))
	}

	return placenums
}

// BuildSQL compiles the QueryBuilder state into a final SQL string and argument slice.
// Returns the complete SELECT query with proper clause ordering and parameterized values.
// If a numbered placeholder format is used ($%d, :%d, etc.), placeholders are expanded
// to sequential indices (e.g., $1, $2, $3) matching the argument order.
func (b *QueryBuilder) BuildSQL() (string, []any) {

	SELECT, selectArgs := b.createSelect()
	FROM := b.createFrom()
	JOIN, joinArgs := b.createJoins()
	WHERE, whereArgs := b.createWhere()
	GROUP_BY := b.createGroupBy()
	ORDER_BY := b.createOrderBy()
	LIMIT, limitArgs := b.createLimit()
	OFFSET, offsetArgs := b.createOffset()

	args := make([]any, 0, len(selectArgs)+len(joinArgs)+len(whereArgs)+len(limitArgs)+len(offsetArgs))
	args = append(args, selectArgs...)
	args = append(args, joinArgs...)
	args = append(args, whereArgs...)
	args = append(args, limitArgs...)
	args = append(args, offsetArgs...)

	query := SELECT + FROM + JOIN + WHERE + GROUP_BY + ORDER_BY + LIMIT + OFFSET

	if b.placeholder != Question {
		query = fmt.Sprintf(query, b.placenums(args, 1)...)
	}

	return query, args
}
