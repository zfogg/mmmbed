package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zfogg/mmmbed/sqlquery"
)

func newSQLCommand() *cobra.Command {
	var inputs []string
	var language, output, prefix, goPackage, dialect string
	var check, singleLine, format bool
	cmd := &cobra.Command{Use: "sql [file-or-directory ...]", Short: "Embed named SQL queries", RunE: func(cmd *cobra.Command, args []string) error {
		language = strings.ToLower(language)
		if check && output == "" {
			return fmt.Errorf("--check requires an output file")
		}
		if cmd.Flags().Changed("prefix") && language != "c" && language != "cpp" {
			return fmt.Errorf("--prefix is only valid for C and C++ output")
		}
		if cmd.Flags().Changed("go-package") && language != "go" {
			return fmt.Errorf("--go-package is only valid for Go output")
		}
		if singleLine && language != "c" && language != "cpp" {
			return fmt.Errorf("--single-line is only valid for C and C++ output")
		}
		if singleLine && format {
			return fmt.Errorf("--format and --single-line cannot be combined")
		}
		paths := append(append([]string(nil), inputs...), args...)
		// A new .sql output inside a scanned directory would become an input on
		// the next run, even if it did not exist during this run.
		if output != "" && strings.EqualFold(filepath.Ext(output), ".sql") {
			out, err := filepath.Abs(output)
			if err != nil {
				return err
			}
			for _, path := range paths {
				info, err := os.Stat(path)
				if err != nil || !info.IsDir() {
					continue
				}
				dir, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				rel, err := filepath.Rel(dir, out)
				if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					return fmt.Errorf("output must not be a .sql file inside an input directory")
				}
			}
		}
		qs, err := sqlquery.LoadPathsWithOptions(paths, sqlquery.ParseOptions{Dialect: dialect})
		if err != nil {
			return err
		}
		if format {
			qs, err = sqlquery.FormatQueries(qs, dialect)
			if err != nil {
				return err
			}
		}
		if output != "" {
			out, _ := filepath.Abs(output)
			outInfo, outErr := os.Stat(output)
			for _, q := range qs {
				in, _ := filepath.Abs(q.Source)
				inInfo, inErr := os.Stat(q.Source)
				if strings.EqualFold(in, out) || (inErr == nil && outErr == nil && os.SameFile(inInfo, outInfo)) {
					return fmt.Errorf("output must not overwrite an input")
				}
			}
		}
		var b bytes.Buffer
		if err = sqlquery.RenderTo(&b, language, qs, sqlquery.Options{Prefix: prefix, GoPackage: goPackage, SingleLine: singleLine, Formatted: format}); err != nil {
			return err
		}
		if check {
			old, err := os.ReadFile(output)
			if err != nil {
				return fmt.Errorf("read output: %w", err)
			}
			if !bytes.Equal(old, b.Bytes()) {
				return fmt.Errorf("output %q is stale; regenerate it", output)
			}
			return nil
		}
		if output == "" {
			_, err = cmd.OutOrStdout().Write(b.Bytes())
			return err
		}
		f, err := os.CreateTemp(filepath.Dir(output), ".mmmbed-output-*")
		if err != nil {
			return err
		}
		tmp := f.Name()
		defer os.Remove(tmp)
		if info, e := os.Stat(output); e == nil {
			if err = f.Chmod(info.Mode().Perm()); err != nil {
				f.Close()
				return err
			}
		}
		if _, err = f.Write(b.Bytes()); err != nil {
			f.Close()
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
		return os.Rename(tmp, output)
	}}
	cmd.Flags().StringArrayVarP(&inputs, "input", "i", nil, "SQL file or directory (repeatable)")
	cmd.Flags().StringVarP(&language, "language", "l", "", "output language")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default stdout)")
	cmd.Flags().StringVar(&prefix, "prefix", "mmmbed", "generated C identifier prefix")
	cmd.Flags().StringVar(&goPackage, "go-package", "", "Go package name (default main)")
	cmd.Flags().StringVar(&dialect, "dialect", "standard", "SQL dialect: standard, sqlite, postgres, duckdb, or mysql")
	cmd.Flags().BoolVar(&check, "check", false, "check whether output is up to date")
	cmd.Flags().BoolVar(&singleLine, "single-line", false, "join safe SQL line breaks in C/C++ output")
	cmd.Flags().BoolVar(&format, "format", false, "format SQL using Sequel")
	_ = cmd.MarkFlagRequired("language")
	return cmd
}
