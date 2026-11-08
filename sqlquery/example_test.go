package sqlquery_test

import (
	"fmt"
	"strings"

	"github.com/zfogg/mmmbed/sqlquery"
)

func ExampleParse() {
	queries, err := sqlquery.Parse("items.sql", strings.NewReader(`-- name: FindItem
-- access: read
-- rows: zero-or-one
SELECT name FROM items WHERE id = :id;
`))
	if err != nil {
		panic(err)
	}
	fmt.Println(queries[0].Name, queries[0].Access, queries[0].Rows)
	// Output: FindItem read zero-or-one
}
