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
	"database/sql"
	"log"

	qub "github.com/gozoro/go-query-builder"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// 1. Create a filter: map user fields to SQL columns
	filter := qub.NewFilter().
		AddColumn("id").
		AddAlias("full_name", "name").  // sql_column => alias
		AddAlias("email_address", "email")

	// 2. Parse user input (e.g., from GraphQL/REST)
	userFields := []string{"id", "name", "email", "hacked_field"} // "hacked_field" will be ignored

	// 3. Build the query
	qb := qub.NewQueryBuilder().
		FilterSelect(filter, userFields).
		From("users").
		Where("status = $1", "active").
		OrderBy("created_at DESC").
		Limit(20)

	// 4. Generate SQL + arguments
	sqlQuery, args := qb.BuildSQL()

	// 5. Execute safely
	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// ... scan results
}
```

##### Generated SQL (PostgreSQL):

```sql
SELECT id, full_name AS name, email_address AS email 
FROM users 
WHERE status = $1 
ORDER BY created_at DESC 
LIMIT $2
```


##### Arguments:

```
["active", 20]
```