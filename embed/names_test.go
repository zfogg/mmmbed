package embed

import (
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
		{"go", "class", true}, {"java", "class", false}, {"c", "match", true}, {"rust", "match", false}, {"perl", "class", true}, {"python", "class", false}, {"go", "init", false}, {"elixir", "Upper", false},
	} {
		t.Run(tc.lang+"/"+tc.name, func(t *testing.T) {
			_, err := RenderWithOptions(tc.lang, asset, Options{Variables: []string{tc.name}})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}
