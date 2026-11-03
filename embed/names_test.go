package embed

import (
	"errors"
	"strings"
	"testing"
)

func TestDefaultVariableName(t *testing.T) {
	for _, tc := range []struct{ data, want string }{
		{"", "mmmbed_file_e3b0c4"},
		{"abc", "mmmbed_file_ba7816"},
	} {
		if got := defaultVariableName([]byte(tc.data)); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.data, got, tc.want)
		}
	}
}

func TestAliasesAllLanguages(t *testing.T) {
	for _, lang := range Languages() {
		t.Run(lang, func(t *testing.T) {
			src, err := RenderWithOptions(lang, []Asset{{Name: "a.bin", Parent: "fonts", Data: []byte("abc")}, {Name: "b.bin", Data: []byte("other")}}, Options{Variables: []string{"first", "second"}})
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"first", "first_filename", "first_path_parent", "second", "second_filename", "second_path_parent"} {
				if !strings.Contains(string(src), name) {
					t.Errorf("missing %s", name)
				}
			}
			if strings.Contains(string(src), "secret/full/path") {
				t.Fatal("unexpected full path")
			}
			def, err := Render(lang, []Asset{{Name: "a.bin", Data: []byte("abc")}})
			if err != nil || !strings.Contains(string(def), "mmmbed_file_ba7816_filename") {
				t.Fatalf("default name: %v\n%s", err, def)
			}
		})
	}
}

func TestVariableNamesValidation(t *testing.T) {
	assets := []Asset{{Name: "a", Data: []byte("a")}, {Name: "b", Data: []byte("b")}}
	for _, names := range [][]string{{"one"}, {"bad-name", "two"}, {"for", "two"}, {"same", "same"}, {"x", "x_filename"}, {"x_size", "x"}} {
		if _, err := RenderWithOptions("go", assets, Options{Variables: names}); err == nil {
			t.Errorf("accepted %q", names)
		}
	}
	src, err := Render("go", []Asset{{Name: "a"}, {Name: "b"}})
	if err != nil || !strings.Contains(string(src), "mmmbed_file_e3b0c4_2_filename") {
		t.Fatalf("identical content: %v %s", err, src)
	}
}

func TestLanguageSpecificKeywords(t *testing.T) {
	asset := []Asset{{Name: "x"}}
	for _, tc := range []struct {
		lang, name string
		valid      bool
	}{
		{"go", "class", true}, {"java", "class", false}, {"c", "match", true}, {"rust", "match", false}, {"perl", "class", true}, {"python", "class", false}, {"go", "init", false}, {"go", "nil", false}, {"elixir", "Upper", false}, {"r", "_invalid", false}, {"go", "_valid", true},
	} {
		t.Run(tc.lang+"/"+tc.name, func(t *testing.T) {
			_, err := RenderWithOptions(tc.lang, asset, Options{Variables: []string{tc.name}})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}

func TestOptionalOutput(t *testing.T) {
	for _, lang := range Languages() {
		t.Run(lang, func(t *testing.T) {
			got, err := RenderWithOptions(lang, []Asset{{Name: "x", Data: []byte("abc")}}, Options{Variables: []string{"item"}, Collection: "file_group", SHA256: true})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"file_group", "item_sha256", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"} {
				if !strings.Contains(string(got), want) {
					t.Errorf("missing %s", want)
				}
			}
		})
	}
}
func TestVisibility(t *testing.T) {
	asset := []Asset{{Name: "x", Data: []byte("abc")}}
	for _, tc := range []struct{ visibility, collection, symbol, mapName string }{{"public", "Files", "Font", "Assets"}, {"private", "files", "font", "assets"}} {
		out, err := RenderWithOptions("go", asset, Options{Variables: []string{"font"}, Collection: "files", Visibility: tc.visibility, SHA256: true})
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"var " + tc.symbol + " = " + tc.mapName + "[", "var " + tc.symbol + "_sha256", "var " + tc.collection + " = " + tc.mapName} {
			if !strings.Contains(string(out), want) {
				t.Errorf("missing %q: %s", want, out)
			}
		}
	}
	if _, err := RenderWithOptions("python", asset, Options{Visibility: "public"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("unsupported visibility accepted")
	}
	if _, err := RenderWithOptions("go", []Asset{{Name: "x"}, {Name: "y"}}, Options{Variables: []string{"foo", "Foo"}, Visibility: "public"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("visibility collision accepted")
	}
	if _, err := RenderWithOptions("go", asset, Options{Variables: []string{"x"}, Collection: "x_filename"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("collection collision accepted")
	}
}

func TestVisibilityOtherTargets(t *testing.T) {
	for _, lang := range []string{"rust", "javascript", "typescript"} {
		t.Run(lang, func(t *testing.T) {
			opts := Options{Variables: []string{"item"}, Collection: "files", SHA256: true, Visibility: "private"}
			got, err := RenderWithOptions(lang, []Asset{{Name: "x", Data: []byte("a")}}, opts)
			if err != nil {
				t.Fatal(err)
			}
			for _, bad := range []string{"pub static", "pub use", "export const", "module.exports"} {
				if strings.Contains(string(got), bad) {
					t.Errorf("private output contains %q", bad)
				}
			}
			opts.Visibility = "public"
			got, err = RenderWithOptions(lang, []Asset{{Name: "x", Data: []byte("a")}}, opts)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{"rust": "pub static", "javascript": "module.exports", "typescript": "export const"}[lang]
			if !strings.Contains(string(got), want) {
				t.Errorf("public output missing %q", want)
			}
		})
	}
}
func TestChecksumNameCollision(t *testing.T) {
	_, err := RenderWithOptions("go", []Asset{{Name: "x"}, {Name: "y"}}, Options{Variables: []string{"asset", "asset_sha256"}, SHA256: true})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected collision error: %v", err)
	}
}
