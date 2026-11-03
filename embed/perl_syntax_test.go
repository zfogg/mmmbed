package embed

import (
	sitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
	"testing"
)

func TestPerlSyntax(t *testing.T) {
	src, err := RenderWithOptions("perl", []Asset{{Name: "font.ttf", Parent: "fonts", Data: []byte{0, 1, 128, 255}}, {Name: "empty"}}, Options{Variables: []string{"font", "empty_file"}})
	if err != nil {
		t.Fatal(err)
	}
	language := grammars.PerlLanguage()
	parser := sitter.NewParser(language)
	tree, err := parser.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if tree.RootNode().HasError() {
		t.Fatalf("invalid Perl: %s\n%s", tree.RootNode().SExpr(language), src)
	}
	broken, err := parser.Parse([]byte("my $x = ;"))
	if err != nil {
		t.Fatal(err)
	}
	if !broken.RootNode().HasError() {
		t.Fatal("parser accepted malformed Perl")
	}
}
