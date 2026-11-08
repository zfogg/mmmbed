package sqlquery

import (
	"fmt"

	sequel "github.com/zfogg/sequel/go"
)

// FormatQueries formats a copy of the queries with Sequel's SQL dialect rules.
// The input slice and its query texts are not modified, including on error.
func FormatQueries(queries []Query, dialect string) ([]Query, error) {
	var d sequel.Dialect
	switch dialect {
	case "", "standard":
		d = sequel.DialectGeneric
	case "sqlite":
		d = sequel.DialectSQLite
	case "postgres", "postgresql":
		d = sequel.DialectPostgreSQL
	case "duckdb":
		d = sequel.DialectDuckDB
	case "mysql":
		d = sequel.DialectMySQL
	default:
		return nil, fmt.Errorf("unsupported SQL formatting dialect %q", dialect)
	}
	result := append([]Query(nil), queries...)
	for i := range result {
		formatted, err := sequel.FormatSQL(d, result[i].SQL)
		if err != nil {
			return nil, fmt.Errorf("format %s (%s:%d): %w", result[i].Name, result[i].Source, result[i].Line, err)
		}
		result[i].SQL = formatted
	}
	return result, nil
}
