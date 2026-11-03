//go:build cgo

package syntax

import (
	dart "github.com/nielsenko/tree-sitter-dart/bindings/go"
	rlang "github.com/r-lib/tree-sitter-r/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
	julia "github.com/tree-sitter/tree-sitter-julia/bindings/go"
	"github.com/zfogg/mmmbed/embed"
	"testing"
)

func TestGeneratedSyntax(t *testing.T) {
	langs := map[string]*sitter.Language{"dart": sitter.NewLanguage(dart.Language()), "r": sitter.NewLanguage(rlang.Language()), "julia": sitter.NewLanguage(julia.Language())}
	for name, grammar := range langs {
		t.Run(name, func(t *testing.T) {
			src, err := embed.RenderWithOptions(name, []embed.Asset{{Name: "font.ttf", Parent: "fonts", Data: []byte{0, 1, 128, 255}}, {Name: "empty"}}, embed.Options{Variables: []string{"font", "empty_file"}, Collection: "my_collection", SHA256: true})
			if err != nil {
				t.Fatal(err)
			}
			parser := sitter.NewParser()
			defer parser.Close()
			if err := parser.SetLanguage(grammar); err != nil {
				t.Fatal(err)
			}
			tree := parser.Parse(src, nil)
			if tree == nil {
				t.Fatal("parse failed")
			}
			defer tree.Close()
			if tree.RootNode().HasError() {
				t.Fatalf("syntax errors in %s: %s", name, src)
			}
		})
	}
}
