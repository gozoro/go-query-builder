package query_builder

import "fmt"

// Deprecated and will be removed in v1.0.9.
//
// Filter maintains a mapping between user-facing field names and SQL column names,
// preserving the insertion order to ensure deterministic query generation.
type Filter struct {
	mapping map[string]col
	order   []string
}

// Deprecated and will be removed in v1.0.9.
type col struct {
	columnName string
	aliasName  string
	j          *join
}

// Deprecated: use NewSelectFilter instead. NewFilter will be removed in v1.0.9.

// NewFilter initializes and returns a new, empty Filter instance.
func NewFilter() *Filter {
	return &Filter{
		mapping: make(map[string]col),
		order:   make([]string, 0),
	}
}

//	Deprecated and will be removed in v1.0.9.
//
// AddAliasForParam registers a mapping between an input parameter and an
// aliased SQL column.
//
// sqlColumn is the source SQL column or expression.
// alias is the alias used in SELECT, e.g. "sqlColumn AS alias".
// inputParam is the input parameter name that enables this column.
//
// If inputParam was not registered before, it is appended to the order list.
// If it was already registered, its order position is preserved and the
// existing mapping is replaced.
//
// AddAliasForParam returns the filter to allow method chaining.
func (f *Filter) AddAliasForParam(sqlColumn, aliasName, inputParam string) *Filter {

	if _, exists := f.mapping[inputParam]; !exists {
		f.order = append(f.order, inputParam)
	}
	f.mapping[inputParam] = col{columnName: sqlColumn, aliasName: aliasName}
	return f
}

// Deprecated and will be removed in v1.0.9.
//
// AddAlias registers a mapping from a user-facing field name to its underlying SQL column.
// It returns the Filter to allow for method chaining. If the alias already exists,
// its mapping is updated while preserving its original insertion order.
func (f *Filter) AddAlias(sqlColumn, aliasField string) *Filter {
	return f.AddAliasForParam(sqlColumn, aliasField, aliasField)
}

// Deprecated and will be removed in v1.0.9.
//
// AddColumn registers a SQL column where the user-facing field name matches the column name exactly.
// It returns the Filter to allow for method chaining.
func (f *Filter) AddColumn(sqlColumn string) *Filter {
	return f.AddAliasForParam(sqlColumn, sqlColumn, sqlColumn)
}

// Deprecated and will be removed in v1.0.9.
//
// AddAliasForParamWithJoin registers a mapping between an input parameter and
// an aliased SQL column, together with the JOIN required for that column.
//
// sqlColumn is the source SQL column or expression.
// aliasName is the alias used in SELECT, e.g. "sqlColumn AS aliasName".
// inputParam is the input parameter name that enables this column.
// joinType is the type of join, for example INNER JOIN or LEFT JOIN.
// table is the table to join.
// on is the join condition.
// args are optional arguments for placeholders in the join condition.
//
// If inputParam was not registered before, it is appended to the order list.
// If it was already registered, its order position is preserved and the
// existing mapping is replaced.
//
// AddAliasForParamWithJoin returns the filter to allow method chaining.
func (f *Filter) AddAliasForParamWithJoin(sqlColumn, aliasName, inputParam, joinType, table, on string, args ...any) *Filter {

	if _, exists := f.mapping[inputParam]; !exists {
		f.order = append(f.order, inputParam)
	}

	j := &join{
		joinType: joinType,
		table:    table,
		on:       on,
		args:     args,
	}

	f.mapping[inputParam] = col{columnName: sqlColumn, aliasName: aliasName, j: j}

	return f
}

// Deprecated and will be removed in v1.0.9.
//
// AddAliasWithJoin registers a mapping from a user-facing field name to its underlying SQL column,
// and associates it with a JOIN clause to be included when this field is selected.
//   - sqlColumn: the actual column name in the database (e.g., "orders.total")
//   - aliasName: the user-facing field name (e.g., "orderTotal")
//   - joinType: the type of JOIN (e.g., "INNER JOIN", "LEFT JOIN")
//   - table: the table to join (e.g., "orders AS o")
//   - on: the ON condition with placeholders (e.g., "o.user_id = u.id AND o.status = $1")
//   - args: parameterized arguments for the ON condition
//
// If the aliasName already exists, its mapping is updated while preserving insertion order.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithJoin(sqlColumn, aliasName, joinType, table, on string, args ...any) *Filter {

	f.AddAliasForParamWithJoin(sqlColumn, aliasName, aliasName, joinType, table, on, args...)
	return f
}

// Deprecated and will be removed in v1.0.9.
//
// AddAliasWithInnerJoin is a shorthand for AddAliasWithJoin with joinType set to "INNER JOIN".
// Registers a column mapping with an INNER JOIN clause.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithInnerJoin(sqlColumn, aliasField, table, on string, args ...any) *Filter {

	f.AddAliasWithJoin(sqlColumn, aliasField, "INNER JOIN", table, on, args...)
	return f
}

// Deprecated and will be removed in v1.0.9.
//
// AddAliasWithLeftJoin is a shorthand for AddAliasWithJoin with joinType set to "LEFT JOIN".
// Registers a column mapping with a LEFT JOIN clause.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithLeftJoin(sqlColumn, aliasField, table, on string, args ...any) *Filter {

	f.AddAliasWithJoin(sqlColumn, aliasField, "LEFT JOIN", table, on, args...)
	return f
}

// Deprecated and will be removed in v1.0.9.
//
// AddAliasWithRightJoin is a shorthand for AddAliasWithJoin with joinType set to "RIGHT JOIN".
// Registers a column mapping with a RIGHT JOIN clause.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithRightJoin(sqlColumn, aliasField, table, on string, args ...any) *Filter {

	f.AddAliasWithJoin(sqlColumn, aliasField, "RIGHT JOIN", table, on, args...)
	return f
}

// Deprecated and will be removed in v1.0.9.
func (f *Filter) FilterJoins(inputNames []string) []join {

	joins := make([]join, 0, len(inputNames))
	joinMap := make(map[string]struct{}, len(inputNames))

	for _, inputName := range inputNames {
		if column, ok := f.mapping[inputName]; ok {

			if column.j != nil {

				if _, exists_join := joinMap[column.j.table]; !exists_join {
					joins = append(joins, *column.j)
					joinMap[column.j.table] = struct{}{}
				}
			}
		}
	}
	return joins
}

// Deprecated and will be removed in v1.0.9.
//
// Filter processes a list of requested field names and returns a slice of valid SQL column expressions.
// The resulting slice preserves the order of the inputNames. Fields not present in the mapping are ignored.
func (f *Filter) Filter(inputNames []string) []string {
	if len(inputNames) == 0 {
		return []string{}
	}

	// Pre-allocate with len(inputNames) as the maximum possible capacity
	columns := make([]string, 0, len(inputNames))

	for _, inputName := range inputNames {
		if column, ok := f.mapping[inputName]; ok {
			if column.columnName == column.aliasName {
				columns = append(columns, column.columnName)
			} else {
				columns = append(columns, fmt.Sprintf("%s AS %s", column.columnName, column.aliasName))
			}
		}
	}

	return columns
}

// Deprecated and will be removed in v1.0.9.
//
// GetSqlColumns returns a slice containing all registered SQL column names in the filter,
// preserving the order in which they were added.
func (f *Filter) GetSqlColumns() []string {
	fields := make([]string, 0, len(f.order))
	for _, alias := range f.order {
		fields = append(fields, f.mapping[alias].columnName)
	}
	return fields
}
