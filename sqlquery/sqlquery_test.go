package sqlquery

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAndRender(t *testing.T) {
	input := "-- name: Lookup\n-- access: read\n-- rows: zero-or-one\nSELECT :name;\n-- name: Delete\n-- access: write\n-- rows: zero\nDELETE FROM t WHERE name = :name;\n"
	qs, err := Parse("queries.sql", strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 2 || qs[0].SQL != "SELECT :name;" {
		t.Fatalf("%+v", qs)
	}
	var out bytes.Buffer
	if err = RenderTo(&out, "c", qs, Options{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "MMMBED_QUERY_Lookup") || !strings.Contains(out.String(), ":name") {
		t.Fatal(out.String())
	}
}

func TestDirectivesInsideSQL(t *testing.T) {
	input := "/* header\n-- name: NotAQuery\n*/\n-- name: First\n-- access: read\n-- rows: one\nSELECT 'a\n-- name: NotAQuery\nb', $$x\n-- name: AlsoNot\ny$$, E'escaped\\'\n-- name: StillNot\ntext';\n/*\n-- name: Neither\n*/\n-- name: Second\n-- access: read\n-- rows: one\nSELECT 2;\n"
	qs, err := Parse("x.sql", strings.NewReader(input))
	if err != nil || len(qs) != 2 || qs[1].Name != "Second" {
		t.Fatalf("%+v: %v", qs, err)
	}
}

func TestLoadPathsAndDuplicates(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "nested")
	if err := os.Mkdir(sub, 0700); err != nil {
		t.Fatal(err)
	}
	data := []byte("-- name: Q\n-- access: read\n-- rows: many\nSELECT 1;\n")
	p := filepath.Join(sub, "q.sql")
	if err := os.WriteFile(p, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPaths([]string{root}); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPaths([]string{root, p}); err == nil {
		t.Fatal("accepted overlapping paths")
	}
	if err := os.WriteFile(filepath.Join(root, "duplicate.sql"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPaths([]string{root}); err == nil {
		t.Fatal("accepted duplicate query")
	}
}

func TestInvalidMetadata(t *testing.T) {
	for _, s := range []string{"SELECT 1", "-- name: 1bad\n", "-- name: Q\n-- access: write\nSELECT 1", "-- name: Q\n-- access: read\n-- rows: many\n"} {
		if _, err := Parse("x", strings.NewReader(s)); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}

func TestBOMAndCRLF(t *testing.T) {
	qs, err := Parse("bom.sql", strings.NewReader("\ufeff-- name: Q\r\n-- access: read\r\n-- rows: one\r\nSELECT\r\n  1;\r\n"))
	if err != nil || len(qs) != 1 || qs[0].SQL != "SELECT\r\n  1;" {
		t.Fatalf("%+v: %v", qs, err)
	}
}

func TestDollarInIdentifier(t *testing.T) {
	queries, err := Parse("pg.sql", strings.NewReader("-- name: Q\n-- access: read\n-- rows: one\nSELECT foo$bar$ FROM things;\n"))
	if err != nil || len(queries) != 1 {
		t.Fatalf("%+v: %v", queries, err)
	}
}
