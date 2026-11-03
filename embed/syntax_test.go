//go:build cgo

package embed

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/swift"
	ts "github.com/smacker/go-tree-sitter/typescript/typescript"
)

func TestGeneratedSyntax(t *testing.T) {
	langs := map[string]*sitter.Language{
		"c": c.GetLanguage(), "cpp": cpp.GetLanguage(), "csharp": csharp.GetLanguage(), "elixir": elixir.GetLanguage(), "go": golang.GetLanguage(), "java": java.GetLanguage(), "javascript": javascript.GetLanguage(), "kotlin": kotlin.GetLanguage(), "lua": lua.GetLanguage(), "php": php.GetLanguage(), "python": python.GetLanguage(), "ruby": ruby.GetLanguage(), "rust": rust.GetLanguage(), "scala": scala.GetLanguage(), "swift": swift.GetLanguage(), "typescript": ts.GetLanguage(),
	}
	for name, grammar := range langs {
		t.Run(name, func(t *testing.T) {
			src, err := RenderWithOptions(name, []Asset{{Name: "font ♥.ttf", Parent: "fonts", Data: []byte{0, 1, 128, 255}}, {Name: "empty"}}, Options{Variables: []string{"font", "empty_file"}})
			if err != nil {
				t.Fatal(err)
			}
			tree, err := sitter.ParseCtx(context.Background(), src, grammar)
			if err != nil {
				t.Fatal(err)
			}

			if tree.HasError() {
				t.Fatalf("syntax errors in generated %s:\n%s\n%s", name, tree, src)
			}
		})
	}
}
