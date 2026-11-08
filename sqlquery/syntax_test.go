//go:build cgo

package sqlquery

import (
	"bytes"
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

func TestSQLGeneratedSyntax(t *testing.T) {
	langs := map[string]*sitter.Language{"c": c.GetLanguage(), "cpp": cpp.GetLanguage(), "csharp": csharp.GetLanguage(), "elixir": elixir.GetLanguage(), "go": golang.GetLanguage(), "java": java.GetLanguage(), "javascript": javascript.GetLanguage(), "kotlin": kotlin.GetLanguage(), "lua": lua.GetLanguage(), "php": php.GetLanguage(), "python": python.GetLanguage(), "ruby": ruby.GetLanguage(), "rust": rust.GetLanguage(), "scala": scala.GetLanguage(), "swift": swift.GetLanguage(), "typescript": ts.GetLanguage()}
	queries := []Query{{Name: "Sample", SQL: "SELECT ':id', 'café', '$x';\n-- note", Access: "read", Rows: "many"}, {Name: "Write", SQL: "UPDATE t SET n=:n", Access: "write", Rows: "zero"}}
	for lang, grammar := range langs {
		t.Run(lang, func(t *testing.T) {
			var output bytes.Buffer
			if err := RenderTo(&output, lang, queries, Options{}); err != nil {
				t.Fatal(err)
			}
			tree, err := sitter.ParseCtx(context.Background(), output.Bytes(), grammar)
			if err != nil {
				t.Fatal(err)
			}
			if tree.HasError() {
				t.Fatalf("invalid %s: %s\n%s", lang, tree, output.String())
			}
		})
	}
}
