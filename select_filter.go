package query_builder

import "fmt"

// column represents a database column or SQL expression to be included in a SELECT clause.
// It contains the raw expression, an optional alias, parameterized arguments, and an optional associated JOIN definition.
type column struct {
	expr  string
	alias string
	args  []any

	j *join
}

type join struct {
	joinType string
	table    string
	on       string
	args     []any
}

// selectFilter manages the filtering of SELECT columns based on a predefined list of requested input fields.
// It maps external field names to their corresponding internal column definitions.
type selectFilter struct {
	inputFields []string
	mapping     map[string]column
}

// NewSelectFilter initializes and returns a new instance of selectFilter with the given input fields and an empty mapping.
func NewSelectFilter(inputFields []string) *selectFilter {
	return &selectFilter{
		inputFields: inputFields,
		mapping:     make(map[string]column),
	}
}

// AddColumn registers a basic column in the filter mapping using the column name itself as the lookup key.
// The column will be included in the SELECT clause if the column name matches a field in inputFields.
// Note: The developer is responsible for proper SQL escaping and sanitization of the column name.
func (f *selectFilter) AddColumn(col string) *selectFilter {
	f.mapping[col] = column{expr: col}
	return f
}

// AddColumnAs registers a column in the filter mapping with a specific alias.
// The column will be included in the SELECT clause if the alias matches a field in inputFields.
func (f *selectFilter) AddColumnAs(col, alias string) *selectFilter {
	f.mapping[alias] = column{expr: col, alias: alias}
	return f
}

// AddExpressionAs registers a raw SQL expression in the filter mapping with a specified alias and parameterized arguments.
// The expression will be included in the SELECT clause if the alias matches a field in inputFields.
func (f *selectFilter) AddExpressionAs(sqlExpression, alias string, args ...any) *selectFilter {
	f.mapping[alias] = column{expr: sqlExpression, alias: alias, args: args}
	return f
}

// Bind re-maps an existing column definition from its current key to a new input parameter name.
// This allows decoupling the internal database column name from the external API request parameter.
func (f *selectFilter) Bind(col, inputParam string) *selectFilter {

	if v, exists := f.mapping[col]; exists {
		delete(f.mapping, col)
		f.mapping[inputParam] = v
	}

	return f
}

// GetColumns evaluates the inputFields against the internal mapping and returns a slice of formatted SQL column strings
// (with aliases if applicable) and a flattened slice of their corresponding parameterized arguments.
// Returns nil, nil if inputFields is empty.
func (f *selectFilter) GetColumns() ([]string, []any, []join) {

	if len(f.inputFields) == 0 {
		return nil, nil, nil
	}

	// Pre-allocate with len(f.inputFields) as the maximum possible capacity
	columns := make([]string, 0, len(f.inputFields))
	args := make([]any, 0, len(f.inputFields))   // assumption about the size of the capacity
	joins := make([]join, 0, len(f.inputFields)) // assumption about the size of the capacity

	for _, inputField := range f.inputFields {
		if column, ok := f.mapping[inputField]; ok {
			if column.alias == "" {
				columns = append(columns, column.expr)
				args = append(args, column.args...)
			} else {
				columns = append(columns, fmt.Sprintf("%s AS %s", column.expr, column.alias))
				args = append(args, column.args...)
			}

			if column.j != nil {
				joins = append(joins, *column.j)
			}
		}
	}

	return columns, args, joins
}

// Join attaches a JOIN clause definition to a specific column key in the mapping.
// This enables conditional joins that are only applied if the associated column is requested in inputFields.
func (f *selectFilter) Join(columnName, joinType, table, on string, args ...any) *selectFilter {

	if col, exists := f.mapping[columnName]; exists {
		col.j = &join{
			joinType: joinType,
			table:    table,
			on:       on,
			args:     args,
		}

		f.mapping[columnName] = col
	}

	return f
}

// InnerJoin is a convenience wrapper for Join that specifies an "INNER JOIN" type.
func (f *selectFilter) InnerJoin(columnName, table, on string, args ...any) *selectFilter {
	return f.Join(columnName, "INNER JOIN", table, on, args...)
}

// LeftJoin is a convenience wrapper for Join that specifies a "LEFT JOIN" type.
func (f *selectFilter) LeftJoin(columnName, table, on string, args ...any) *selectFilter {
	return f.Join(columnName, "LEFT JOIN", table, on, args...)
}

// RightJoin is a convenience wrapper for Join that specifies a "RIGHT JOIN" type.
func (f *selectFilter) RightJoin(columnName, table, on string, args ...any) *selectFilter {
	return f.Join(columnName, "RIGHT JOIN", table, on, args...)
}

// CrossJoin is a convenience wrapper for Join that specifies a "CROSS JOIN" type.
func (f *selectFilter) CrossJoin(columnName, table, on string, args ...any) *selectFilter {
	return f.Join(columnName, "CROSS JOIN", table, on, args...)
}
