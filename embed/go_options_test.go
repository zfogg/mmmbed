package embed

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestGoOptionsTypeCheck(t *testing.T) {
	for _, v := range []string{"public", "private"} {
		t.Run(v, func(t *testing.T) {
			src, err := RenderWithOptions("go", []Asset{{Name: "font", Data: []byte("abc")}}, Options{GoPackage: "fonts", Variables: []string{"font"}, Collection: "files", Visibility: v, SHA256: true})
			if err != nil {
				t.Fatal(err)
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "generated.go", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			cfg := types.Config{Importer: importer.Default()}
			if _, err := cfg.Check("fonts", fset, []*ast.File{f}, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}
