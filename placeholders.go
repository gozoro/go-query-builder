package query_builder

import (
	"fmt"
	"strings"
)

// Placeholders returns a string consisting of the specified placeholder
// repeated count times, with each instance separated by sep.
// It is commonly used to generate parameter lists for SQL queries,
// such as IN (...) clauses or bulk INSERT statements.
func Placeholders(placeholder, sep string, count uint) string {

	placeholders := make([]string, 0, count)

	for range count {
		placeholders = append(placeholders, placeholder)
	}

	return strings.Join(placeholders, sep)
}

// IndexedPlaceholders returns a string of formatted placeholders with incrementing indices,
// separated by sep. It applies placeholderFormat (e.g., "$%d", "?%d", ":p%d") to generate
// each item, starting from the specified index and repeating count times.
// Commonly used to construct dynamic SQL parameter lists for IN (...) clauses,
// multi-row VALUES, or batch operations requiring sequentially numbered arguments.
func IndexedPlaceholders(placeholderFormat, sep string, start, count uint) string {
	placeholders := make([]string, 0, count)
	for i := range count {
		placeholders = append(placeholders, fmt.Sprintf(placeholderFormat, start+i))
	}
	return strings.Join(placeholders, sep)
}
