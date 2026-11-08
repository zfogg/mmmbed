package sqlquery

import (
	"strings"
	"testing"
)

func TestFormatQueries(t *testing.T) {
	for _, dialect := range []string{"standard", "sqlite", "postgres", "postgresql", "duckdb", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			input := []Query{{Name: "Lookup", SQL: "select id,name from users where id=:id", Access: "read", Rows: "one"}}
			got, err := FormatQueries(input, dialect)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(got[0].SQL, "\n") || !strings.Contains(got[0].SQL, ":id") {
				t.Fatalf("unexpected formatted SQL: %q", got[0].SQL)
			}
			if input[0].SQL != "select id,name from users where id=:id" {
				t.Fatal("modified input")
			}
		})
	}
	input := []Query{{Name: "Broken", SQL: "SELECT );", Source: "queries.sql", Line: 4}}
	if _, err := FormatQueries(input, "sqlite"); err == nil || !strings.Contains(err.Error(), "Broken") {
		t.Fatalf("expected named error, got %v", err)
	}
	if _, err := FormatQueries(input, "unknown"); err == nil {
		t.Fatal("accepted unknown dialect")
	}
}
