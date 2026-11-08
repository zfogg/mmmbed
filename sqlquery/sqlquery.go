// Package sqlquery loads named SQL queries without interpreting their SQL dialect.
package sqlquery

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Query contains a SQL template and its declared execution characteristics.
type Query struct {
	Name, SQL, Access, Rows, Source string
	Line                            int
}

// Access and row-count values are declarations, not runtime enforcement.
const (
	AccessRead    = "read"
	AccessWrite   = "write"
	RowsZero      = "zero"
	RowsZeroOrOne = "zero-or-one"
	RowsOne       = "one"
	RowsMany      = "many"
)

var identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Parse reads queries from a source. Directive lines delimit queries; SQL is
// otherwise retained, including comments and line endings.
func Parse(source string, r io.Reader) ([]Query, error) {
	if r == nil {
		return nil, fmt.Errorf("%s: nil SQL reader", source)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	if !utf8.Valid(b) || strings.ContainsRune(string(b), 0) {
		return nil, fmt.Errorf("%s: SQL must be UTF-8 without NUL", source)
	}
	b = bytes.TrimPrefix(b, []byte{0xef, 0xbb, 0xbf})
	var out []Query
	var q *Query
	var body strings.Builder
	metadata := false
	var lex sqlLex
	seen := map[string]int{}
	finish := func() error {
		if q == nil {
			return nil
		}
		if !lex.normal() {
			return fmt.Errorf("%s:%d: unterminated quoted string or block comment in %s", source, q.Line, q.Name)
		}
		q.SQL = strings.TrimSpace(body.String())
		body.Reset()
		if q.Access == "" || q.Rows == "" {
			return fmt.Errorf("%s:%d: query %s requires access and rows", source, q.Line, q.Name)
		}
		if !lex.hasSQL {
			return fmt.Errorf("%s:%d: query %s has no SQL", source, q.Line, q.Name)
		}
		out = append(out, *q)
		return nil
	}
	reader := bufio.NewReader(strings.NewReader(string(b)))
	line := 0
	for {
		s, e := reader.ReadString('\n')
		if len(s) != 0 {
			line++
			trimmed := strings.TrimSpace(s)
			if lex.normal() && strings.HasPrefix(trimmed, "-- name:") {
				if err := finish(); err != nil {
					return nil, err
				}
				name := strings.TrimSpace(strings.TrimPrefix(trimmed, "-- name:"))
				if !identifier.MatchString(name) {
					return nil, fmt.Errorf("%s:%d: invalid query name %q", source, line, name)
				}
				if first, ok := seen[name]; ok {
					return nil, fmt.Errorf("%s:%d: duplicate query %s (first at line %d)", source, line, name, first)
				}
				seen[name] = line
				q = &Query{Name: name, Source: source, Line: line}
				metadata = true
				lex = sqlLex{}
			} else if q == nil {
				lex.scan(s)
				if lex.hasSQL {
					return nil, fmt.Errorf("%s:%d: SQL before first query", source, line)
				}
			} else if metadata && lex.normal() && (strings.HasPrefix(trimmed, "-- access:") || strings.HasPrefix(trimmed, "-- rows:")) {
				parts := strings.SplitN(strings.TrimPrefix(trimmed, "-- "), ":", 2)
				value := strings.TrimSpace(parts[1])
				switch parts[0] {
				case "access":
					if q.Access != "" || (value != "read" && value != "write") {
						return nil, fmt.Errorf("%s:%d: invalid or duplicate access %q", source, line, value)
					}
					q.Access = value
				case "rows":
					if q.Rows != "" || (value != "zero" && value != "zero-or-one" && value != "one" && value != "many") {
						return nil, fmt.Errorf("%s:%d: invalid or duplicate rows %q", source, line, value)
					}
					q.Rows = value
				}
			} else {
				body.WriteString(s)
				lex.scan(s)
				if lex.hasSQL {
					metadata = false
				}
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, fmt.Errorf("%s: %w", source, e)
		}
	}
	if err := finish(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no queries", source)
	}
	return out, nil
}

// LoadPaths recursively discovers .sql files. Directory symlinks are not followed.
func LoadPaths(paths []string) ([]Query, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no input paths")
	}
	var files []string
	seen := map[string]bool{}
	for _, p := range paths {
		info, err := os.Lstat(p)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlink input %q is not supported", p)
		}
		if info.IsDir() {
			err = filepath.WalkDir(p, func(path string, d fs.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.Type()&os.ModeSymlink != 0 {
					return nil
				}
				if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".sql") {
					info, err := d.Info()
					if err != nil {
						return err
					}
					if !info.Mode().IsRegular() {
						return fmt.Errorf("input %q must be a regular file", path)
					}
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if !info.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(p), ".sql") {
				return nil, fmt.Errorf("input %q must be a regular .sql file or directory", p)
			}
			files = append(files, p)
		}
	}
	sort.Strings(files)
	var result []Query
	names := map[string]Query{}
	var infos []os.FileInfo
	for _, p := range files {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		if seen[abs] {
			return nil, fmt.Errorf("overlapping input %q", p)
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		for _, prev := range infos {
			if os.SameFile(info, prev) {
				return nil, fmt.Errorf("overlapping input %q", p)
			}
		}
		infos = append(infos, info)
		seen[abs] = true
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		queries, err := Parse(p, f)
		closeErr := f.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		for _, q := range queries {
			if prev, ok := names[q.Name]; ok {
				return nil, fmt.Errorf("%s:%d: duplicate query %s (first at %s:%d)", q.Source, q.Line, q.Name, prev.Source, prev.Line)
			}
			names[q.Name] = q
			result = append(result, q)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no SQL files found")
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
