// Package cli implements the mmmbed command line interface.
package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/zfogg/mmmbed/embed"
)

// New returns a command configured to read stdin and write stdout by default.
func New() *cobra.Command {
	var language, name, output, goPackage, parent, collection, visibility string
	var omitParent, check, sha256 bool
	var inputs, variables []string
	cmd := &cobra.Command{Use: "mmmbed [input [output]]", Short: "Embed binary files in source code", Args: cobra.MaximumNArgs(2), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := append([]string(nil), inputs...)
			if len(args) > 0 {
				paths = append(paths, args[0])
			}
			if omitParent && cmd.Flags().Changed("parent") {
				return fmt.Errorf("--omit-parent and --parent cannot be combined")
			}
			assetParent := func(path string) string {
				if omitParent {
					return ""
				}
				if cmd.Flags().Changed("parent") {
					return parent
				}
				if path == "" {
					return ""
				}
				return parentName(path)
			}
			if len(args) == 2 {
				if output != "" {
					return fmt.Errorf("output specified twice")
				}
				output = args[1]
			}
			if goPackage != "" && !strings.EqualFold(language, "go") {
				return fmt.Errorf("--go-package is only valid for Go output")
			}
			if len(paths) == 0 && name == "" {
				return fmt.Errorf("--name is required for stdin")
			}
			if len(paths) != 0 && name != "" {
				return fmt.Errorf("--name is only valid for stdin")
			}
			if check && output == "" {
				return fmt.Errorf("--check requires an output file")
			}
			if output != "" {
				for _, path := range paths {
					in, _ := filepath.Abs(path)
					out, _ := filepath.Abs(output)
					inInfo, inErr := os.Stat(path)
					outInfo, outErr := os.Stat(output)
					if strings.EqualFold(in, out) || (inErr == nil && outErr == nil && os.SameFile(inInfo, outInfo)) {
						return fmt.Errorf("output must not overwrite an input")
					}
				}
			}
			var assets []embed.StreamAsset
			var files []*os.File
			defer func() {
				for _, f := range files {
					_ = f.Close()
				}
			}()
			for _, path := range paths {
				f, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("open %q: %w", path, err)
				}
				files = append(files, f)
				assets = append(assets, embed.StreamAsset{Name: filepath.Base(path), Parent: assetParent(path), Reader: f})
			}
			if len(paths) == 0 {
				assets = append(assets, embed.StreamAsset{Name: name, Parent: assetParent(""), Reader: cmd.InOrStdin()})
			}
			opts := embed.Options{GoPackage: goPackage, Variables: variables, Collection: collection, Visibility: visibility, SHA256: sha256}
			if check {
				existing, err := os.Open(output)
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("output %q is missing; regenerate it", output)
				}
				if err != nil {
					return fmt.Errorf("open output: %w", err)
				}
				defer existing.Close()
				cmp := &comparisonWriter{reader: existing}
				if err = embed.RenderTo(cmp, language, assets, opts); err != nil {
					if errors.Is(err, errDifferent) {
						return fmt.Errorf("output %q is stale; regenerate it", output)
					}
					return err
				}
				var extra [1]byte
				n, err := existing.Read(extra[:])
				if n > 0 {
					return fmt.Errorf("output %q is stale; regenerate it", output)
				}
				if err != nil && !errors.Is(err, io.EOF) {
					return fmt.Errorf("read output: %w", err)
				}
				return nil
			}
			if output == "" {
				return embed.RenderTo(cmd.OutOrStdout(), language, assets, opts)
			}
			// Write alongside the destination so a failed render never truncates it.
			f, err := os.CreateTemp(filepath.Dir(output), ".mmmbed-output-*")
			if err != nil {
				return fmt.Errorf("create output: %w", err)
			}
			tmp := f.Name()
			defer os.Remove(tmp)
			if info, err := os.Stat(output); err == nil {
				if err = f.Chmod(info.Mode().Perm()); err != nil {
					f.Close()
					return err
				}
			}
			if err = embed.RenderTo(f, language, assets, opts); err != nil {
				f.Close()
				return err
			}
			if err = f.Close(); err != nil {
				return fmt.Errorf("close output: %w", err)
			}
			if err = os.Rename(tmp, output); err != nil {
				return fmt.Errorf("replace output: %w", err)
			}
			return nil
		}}
	cmd.Flags().StringVarP(&language, "language", "l", "", "output language (required)")
	cmd.Flags().StringArrayVarP(&inputs, "input", "i", nil, "additional input file (repeatable)")
	cmd.Flags().BoolVar(&check, "check", false, "check whether the output file is up to date")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default stdout)")
	cmd.Flags().StringVar(&collection, "collection", "", "add a named alias for the asset collection")
	cmd.Flags().StringVar(&visibility, "visibility", "", "symbol visibility: public or private (Go, Rust, JavaScript, TypeScript)")
	cmd.Flags().BoolVar(&sha256, "sha256", false, "expose each asset SHA-256 checksum")
	cmd.Flags().StringArrayVar(&variables, "variable", nil, "variable name per input (repeatable)")
	cmd.Flags().StringVar(&goPackage, "go-package", "", "Go package name (default main)")
	cmd.Flags().StringVar(&parent, "parent", "", "override parent directory name (use empty string to omit)")
	cmd.Flags().BoolVar(&omitParent, "omit-parent", false, "omit the parent directory name")
	cmd.Flags().StringVar(&name, "name", "", "filename to associate with stdin")
	_ = cmd.MarkFlagRequired("language")
	cmd.AddCommand(&cobra.Command{Use: "languages", Short: "List supported output languages", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Supported languages")
		if _, err := fmt.Fprintln(c.OutOrStdout(), heading); err != nil {
			return err
		}
		for _, l := range embed.SortedLanguages() {
			if _, err := fmt.Fprintf(c.OutOrStdout(), "  %-12s %s\n", l, languageInfo[l]); err != nil {
				return err
			}
		}
		return nil
	}})
	cmd.Long = "Embed one or more files as source code. Supported languages: " + strings.Join(embed.SortedLanguages(), ", ") + "."
	return cmd
}

// parentName returns the immediate directory name without exposing its path.
func parentName(path string) string {
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return ""
	}
	base := filepath.Base(dir)
	if base == string(filepath.Separator) || base == "." {
		return ""
	}
	return base
}

var languageInfo = map[string]string{
	"c": ".c — C compiler", "cpp": ".cpp — C++ compiler", "go": ".go — Go", "javascript": ".js — Node.js", "typescript": ".ts — Node.js and Node types", "python": ".py — Python 3", "rust": ".rs — Rust", "java": ".java — Java", "csharp": ".cs — .NET", "php": ".php — PHP", "ruby": ".rb — Ruby", "swift": ".swift — Swift", "kotlin": ".kt — Kotlin/JVM", "scala": ".scala — Scala/JVM", "dart": ".dart — Dart", "lua": ".lua — Lua", "perl": ".pl — Perl", "r": ".R — R", "julia": ".jl — Julia", "elixir": ".ex — Elixir",
}

var errDifferent = errors.New("output differs")

type comparisonWriter struct{ reader io.Reader }

func (c *comparisonWriter) Write(p []byte) (int, error) {
	const chunk = 32 * 1024
	var buffer [chunk]byte
	total := 0
	for len(p) > 0 {
		n := len(p)
		if n > len(buffer) {
			n = len(buffer)
		}
		count, err := io.ReadFull(c.reader, buffer[:n])
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return total, errDifferent
			}
			return total, err
		}
		if !bytes.Equal(p[:n], buffer[:count]) {
			return total, errDifferent
		}
		total += n
		p = p[n:]
	}
	return total, nil
}
