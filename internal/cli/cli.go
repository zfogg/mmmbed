// Package cli implements the mmmbed command line interface.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zfogg/mmmbed/internal/embed"
)

// New returns a command configured to read stdin and write stdout by default.
func New() *cobra.Command {
	var language, name, output, goPackage string
	var inputs, variables []string
	cmd := &cobra.Command{Use: "mmmbed [input [output]]", Short: "Embed binary files in source code", Args: cobra.MaximumNArgs(2), SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := append([]string(nil), inputs...)
			if len(args) > 0 {
				paths = append(paths, args[0])
			}
			var assets []embed.Asset
			for _, path := range paths {
				data, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("read %q: %w", path, err)
				}
				assets = append(assets, embed.Asset{Name: filepath.Base(path), Data: data, Parent: parentName(path)})
			}
			if len(paths) == 0 {
				if name == "" {
					return fmt.Errorf("--name is required for stdin")
				}
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				assets = append(assets, embed.Asset{Name: name, Data: data})
			} else if name != "" {
				return fmt.Errorf("--name is only valid for stdin")
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
			result, err := embed.RenderWithOptions(language, assets, embed.Options{GoPackage: goPackage, Variables: variables})
			if err != nil {
				return err
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
				if err := os.WriteFile(output, result, 0666); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				return nil
			}
			_, err = cmd.OutOrStdout().Write(result)
			return err
		}}
	cmd.Flags().StringVarP(&language, "language", "l", "", "output language (required)")
	cmd.Flags().StringArrayVarP(&inputs, "input", "i", nil, "additional input file (repeatable)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default stdout)")
	cmd.Flags().StringArrayVar(&variables, "variable", nil, "variable name per input (repeatable)")
	cmd.Flags().StringVar(&goPackage, "go-package", "", "Go package name (default main)")
	cmd.Flags().StringVar(&name, "name", "", "filename to associate with stdin")
	_ = cmd.MarkFlagRequired("language")
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
