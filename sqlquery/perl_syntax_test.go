package sqlquery

import (
	"bytes"
	"testing"

	sitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

func TestSQLPerlSyntax(t *testing.T) {
	var output bytes.Buffer
	if err := RenderTo(&output, "perl", []Query{{Name: "Query", SQL: "SELECT :name", Access: "read", Rows: "one"}}, Options{}); err != nil {
		t.Fatal(err)
	}
	lang := grammars.PerlLanguage()
	tree, err := sitter.NewParser(lang).Parse(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if tree.RootNode().HasError() {
		t.Fatalf("invalid Perl: %s\n%s", tree.RootNode().SExpr(lang), output.String())
	}
}
