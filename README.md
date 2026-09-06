# go-query-builder



Query Builder is a lightweight, fluent SQL query builder for Go that bridges user input (GraphQL, REST) with safe, 
parameterized database queries. It features dynamic column filtering, automatic placeholder formatting, 
and protection against SQL injection via strict identifier validation.

## features

- 🔐 SQL Injection Protection: Parameterized queries + whitelisted column mapping
- 🔄 Dynamic Column Aliasing: Map user-facing fields to SQL columns (fullName → full_name)
- 🗄️ Multi-Dialect Support: PostgreSQL ($1), MySQL/SQLite (?), Oracle (:1), SQL Server (@p1)
- 🧩 Fluent API: Chainable methods for readable query construction
- 📊 Pagination & Sorting: Built-in support for LIMIT, OFFSET, and safe ORDER BY mapping
- ⚡ Zero Dependencies: Pure Go, no external libraries required

## Instalation

```
go get github.com/gozoro/go-query-builder
```


## Quick Start

```go
package main

import (
	"fmt"
	qub "github.com/gozoro/go-query-builder"
)

func main() {
	// 1. Define allowed fields from user input (e.g., REST/GraphQL request)
	userFields := []string{"id", "fullName", "email"}

	// 2. Create a SelectFilter to map external names to internal SQL columns
	filter := qub.NewSelectFilter(userFields).
		AddColumn("id").
		AddColumnAs("full_name", "fullName").
		AddColumnAs("email_address", "email").
		// Conditionally add a JOIN only if "fullName" is requested
		LeftJoin("fullName", "profiles p", "users.id = p.user_id")

	// 3. Build the query
	qb := qub.NewQueryBuilder().
		SelectFilter(filter).
		From("users").
		Where("status = $%d", "active").
		OrderBy("created_at DESC").
		Limit(20)

	// 4. Generate SQL + arguments
	query, args := qb.BuildSQL()

	fmt.Println("SQL:", query)
	fmt.Println("Args:", args)
}
```

##### Generated SQL (PostgreSQL):

```sql
SELECT id, full_name AS fullName, email_address AS email 
FROM users 
LEFT JOIN profiles p ON users.id = p.user_id 
WHERE status = $1 ORDER BY created_at DESC 
LIMIT $2 
```


##### Arguments:

```
Args: [active 20]
```


## Select Filter
Safely maps external API field names to internal SQL columns, preventing unauthorized column selection.

```go
// NewSelectFilter: Initializes a filter with the list of requested fields from the client
filter := qub.NewSelectFilter([]string{"id", "displayName", "total"})

// AddColumn: Registers a basic column. Key = column name. 
// Note: Developer is responsible for sanitizing the column name string.
filter.AddColumn("id")

// AddColumnAs: Registers a column with an alias. Key = alias (what the client requests)
filter.AddColumnAs("full_name", "displayName")

// AddExpressionAs: Registers a raw SQL expression with an alias and arguments
filter.AddExpressionAs("COUNT(orders.id)", "total")

// Bind: Re-maps an existing registered column to a different external input parameter name
filter.Bind("id", "userId") // Client sends "userId", maps to internal "id"

// Join / InnerJoin / LeftJoin / RightJoin / CrossJoin: 
// Attaches a JOIN to a specific column key. The JOIN is ONLY added to the final query 
// if that specific column is requested in inputFields.
filter.LeftJoin("displayName", "profiles p", "users.id = p.user_id")

// GetColumns: Evaluates inputFields and returns matched columns, args, and conditional joins
cols, args, joins := filter.GetColumns()
```

## Argument Map

Parses and sanitizes pagination and sorting parameters from HTTP/GraphQL requests.

```go
// Simulated user input from HTTP query params
reqParams := map[string]any{
	"limit":  25,
	"offset": 50,
	"sort":   "price desc",
	"search": "john",
	"tags":   []any{"go", "sql"},
	"date":   int64(1690000000),
}

// NewArgs: Parses params and applies options. Removes processed keys from the map.
args := qub.NewArgs(reqParams,
	qub.WithLimitParam("limit"),
	qub.WithLimitDefault(10),
	qub.WithLimitMax(100),
	qub.WithOffsetParam("offset"),
	qub.WithOffsetFunc(func(offset int) int { return offset * 10 }), // e.g., page number to offset
	qub.WithOrderByParams([]string{"sort"}),
	qub.WithOrderByKeyFunc(func(params []string) string { return strings.Join(params, " ") }),
)

// SetDefaultOrderBy: Fallback sorting if no rule matches
args.SetDefaultOrderBy("id ASC")

// AddOrderRule: Maps a user-facing sort string to a safe, validated SQL ORDER BY clause
args.AddOrderRule("price desc", "price DESC, id ASC")
args.AddOrderRule("name asc", "name ASC")

// Value: Retrieves a raw value by key (returns nil if absent)
searchVal := args.Value("search") // "john"

// Values: Retrieves a slice of values (e.g., for IN clauses)
tags := args.Values("tags") // []any{"go", "sql"}

// ValueLike: Safely escapes and wraps a string in '%' for SQL LIKE queries
likeVal := args.ValueLike("search") // Returns "%john%" (with escaped wildcards)

// ValueTime: Converts a Unix timestamp (int64) to time.Time
t := args.ValueTime("date") // time.Time

// Limit / Offset / OrderBy: Retrieve the final, validated pagination/sorting values
limit := args.Limit()       // 25 (capped at limitMax)
offset := args.Offset()     // 500 (50 * 10 via offsetFunc)
orderBy := args.OrderBy()   // "price DESC, id ASC" (matched via AddOrderRule)
```



## Placeholders
Standalone helper functions for generating dynamic placeholder strings.

```go
// Placeholders: Repeats a placeholder string 'count' times, separated by 'sep'
// Useful for IN (...) clauses or bulk INSERTs
ph1 := qub.Placeholders("?", ", ", 3) 
// Result: "?, ?, ?"

// IndexedPlaceholders: Generates sequentially numbered placeholders
// Format: placeholderFormat (e.g., "$%d", ":p%d"), separator, start index, count
ph2 := qub.IndexedPlaceholders("$%d", ", ", 1, 3)
// Result: "$1, $2, $3"

ph3 := qub.IndexedPlaceholders(":p%d", ", ", 5, 2)
// Result: ":p5, :p6"
```

## Security & Best Practices

1. **Never concatenate user input directly into SQL strings.** Always use AddExpressionAs with parameterized args or rely on SelectFilter mappings.
2. **Column names in `AddColumn` are not auto-escaped.** If the column name originates from user input, you must validate it against a strict whitelist or escape it using your database driver's identifier quoting function (e.g., `pq.QuoteIdentifier`).
3. **Use `ArgsMap` for sorting.** Never interpolate `ORDER BY` directly from user input. Use `AddOrderRule` to map client-facing sort keys to hardcoded, safe SQL expressions.
4. **Prefer `AndFilterWhere` / `AndFilterHaving`** over their non-And counterparts when building dynamic queries, as they preserve existing conditions instead of overwriting them.
