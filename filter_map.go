package query_builder

import "fmt"

// Filter maintains a mapping between user-facing field names and SQL column names,
// preserving the insertion order to ensure deterministic query generation.
type Filter struct {
	mapping map[string]column
	order   []string
}

type column struct {
	name string
	j    *join
}

type join struct {
	joinType string
	table    string
	on       string
	args     []any
}

// NewFilter initializes and returns a new, empty Filter instance.
func NewFilter() *Filter {
	return &Filter{
		mapping: make(map[string]column),
		order:   make([]string, 0),
	}
}

// AddAlias registers a mapping from a user-facing field name to its underlying SQL column.
// It returns the Filter to allow for method chaining. If the alias already exists,
// its mapping is updated while preserving its original insertion order.
func (f *Filter) AddAlias(sqlColumn, aliasField string) *Filter {
	if _, exists := f.mapping[aliasField]; !exists {
		f.order = append(f.order, aliasField)
	}
	f.mapping[aliasField] = column{name: sqlColumn}
	return f
}

// AddColumn registers a SQL column where the user-facing field name matches the column name exactly.
// It returns the Filter to allow for method chaining.
func (f *Filter) AddColumn(sqlColumn string) *Filter {
	return f.AddAlias(sqlColumn, sqlColumn)
}

// AddAliasWithJoin registers a mapping from a user-facing field name to its underlying SQL column,
// and associates it with a JOIN clause to be included when this field is selected.
//   - sqlColumn: the actual column name in the database (e.g., "orders.total")
//   - aliasField: the user-facing field name (e.g., "orderTotal")
//   - joinType: the type of JOIN (e.g., "INNER JOIN", "LEFT JOIN")
//   - table: the table to join (e.g., "orders AS o")
//   - on: the ON condition with placeholders (e.g., "o.user_id = u.id AND o.status = $1")
//   - args: parameterized arguments for the ON condition
//
// If the aliasField already exists, its mapping is updated while preserving insertion order.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithJoin(sqlColumn, aliasField, joinType, table, on string, args ...any) *Filter {
	if _, exists := f.mapping[aliasField]; !exists {
		f.order = append(f.order, aliasField)
	}

	j := &join{
		joinType: joinType,
		table:    table,
		on:       on,
		args:     args,
	}

	f.mapping[aliasField] = column{name: sqlColumn, j: j}
	return f
}

// AddAliasWithInnerJoin is a shorthand for AddAliasWithJoin with joinType set to "INNER JOIN".
// Registers a column mapping with an INNER JOIN clause.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithInnerJoin(sqlColumn, aliasField, table, on string, args ...any) *Filter {

	f.AddAliasWithJoin(sqlColumn, aliasField, "INNER JOIN", table, on, args...)
	return f
}

// AddAliasWithLeftJoin is a shorthand for AddAliasWithJoin with joinType set to "LEFT JOIN".
// Registers a column mapping with a LEFT JOIN clause.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithLeftJoin(sqlColumn, aliasField, table, on string, args ...any) *Filter {

	f.AddAliasWithJoin(sqlColumn, aliasField, "LEFT JOIN", table, on, args...)
	return f
}

// AddAliasWithRightJoin is a shorthand for AddAliasWithJoin with joinType set to "RIGHT JOIN".
// Registers a column mapping with a RIGHT JOIN clause.
// Returns the Filter for method chaining.
func (f *Filter) AddAliasWithRightJoin(sqlColumn, aliasField, table, on string, args ...any) *Filter {

	f.AddAliasWithJoin(sqlColumn, aliasField, "RIGHT JOIN", table, on, args...)
	return f
}

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
			if column.name == inputName {
				columns = append(columns, column.name)
			} else {
				columns = append(columns, fmt.Sprintf("%s AS %s", column.name, inputName))
			}
		}
	}

	return columns
}

// GetSqlColumns returns a slice containing all registered SQL column names in the filter,
// preserving the order in which they were added.
func (f *Filter) GetSqlColumns() []string {
	fields := make([]string, 0, len(f.order))
	for _, alias := range f.order {
		fields = append(fields, f.mapping[alias].name)
	}
	return fields
}
