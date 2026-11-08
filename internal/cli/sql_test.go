package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLCommand(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "query.sql")
	out := filepath.Join(dir, "queries.h")
	if err := os.WriteFile(in, []byte("-- name: Lookup\n-- access: read\n-- rows: zero-or-one\nSELECT :id;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"sql", "-l", "c", "-i", dir, "-o", out}
	if _, err := run(args, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := run(append(args, "--check"), ""); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil || !strings.Contains(string(data), "MMMBED_QUERY_Lookup") {
		t.Fatalf("%s: %v", data, err)
	}
	if err := os.WriteFile(out, []byte("stale"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := run(append(args, "--check"), ""); err == nil {
		t.Fatal("accepted stale output")
	}
}

func TestSQLRejectsOutputInsideInputDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "in.sql"), []byte("-- name: Q\n-- access: read\n-- rows: one\nSELECT 1;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := run([]string{"sql", "-l", "c", "-i", dir, "-o", filepath.Join(dir, "out.sql")}, ""); err == nil {
		t.Fatal("accepted SQL output inside scanned directory")
	}
}

func TestSQLFailureDoesNotReplaceOutput(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.sql")
	out := filepath.Join(dir, "out.h")
	if err := os.WriteFile(in, []byte("-- name: Q\n-- access: read\nSELECT 1;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := run([]string{"sql", "-l", "c", in, "-o", out}, ""); err == nil {
		t.Fatal("accepted missing metadata")
	}
	data, err := os.ReadFile(out)
	if err != nil || string(data) != "keep me" {
		t.Fatalf("output changed: %q %v", data, err)
	}
	if err := os.WriteFile(in, []byte("-- name: Q\n-- access: read\n-- rows: one\nSELECT 1;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := run([]string{"sql", "-l", "c", in, "-o", in}, ""); err == nil {
		t.Fatal("accepted input as output")
	}
}
