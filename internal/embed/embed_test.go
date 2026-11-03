package embed

import (
	"bytes"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllLanguages(t *testing.T) {
	assets := []Asset{{Name: "a ♥.bin", Data: []byte{0, 1, 127, 128, 255}}, {Name: "empty", Data: []byte{}}}
	if len(Languages()) != 20 {
		t.Fatalf("got %d languages", len(Languages()))
	}
	for _, lang := range Languages() {
		t.Run(lang, func(t *testing.T) {
			got, err := Render(lang, assets)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) == 0 {
				t.Fatal("empty source")
			}
			again, err := Render(lang, assets)
			if err != nil || !bytes.Equal(got, again) {
				t.Fatal("non-deterministic output")
			}
			if lang != "c" && lang != "cpp" && lang != "rust" && lang != "lua" && lang != "r" {
				for _, a := range assets {
					for _, s := range []string{base64.StdEncoding.EncodeToString([]byte(a.Name)), base64.StdEncoding.EncodeToString(a.Data)} {
						if !bytes.Contains(got, []byte(s)) {
							t.Fatalf("missing encoded %q", s)
						}
					}
				}
			}
		})
	}
}
func TestRejectInvalid(t *testing.T) {
	for _, tc := range []struct {
		lang   string
		assets []Asset
	}{{"unknown", []Asset{{Name: "x"}}}, {"go", nil}, {"go", []Asset{{Name: "../x"}}}, {"go", []Asset{{Name: "x"}, {Name: "x"}}}} {
		if _, err := Render(tc.lang, tc.assets); err == nil {
			t.Fatalf("accepted %+v", tc)
		}
	}
}
func TestJavaLargeLiteral(t *testing.T) {
	out, err := Render("java", []Asset{{Name: "large", Data: make([]byte, 100000)}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("String.join")) {
		t.Fatal("large literal was not split")
	}
}

// Exercise generated source through the target runtime where available.
func TestGeneratedPrograms(t *testing.T) {
	tests := []struct {
		lang, tool, file, script string
		args                     []string
	}{
		{"c", "gcc", "asset.c", `#include <string.h>
int main(void) { return assets_count != 2 || strcmp(assets[0].name, "font.ttf") || assets[0].size != 4 || memcmp(assets[0].data, "\0\1\200\377", 4) || assets[1].size; }`, nil},
		{"cpp", "g++", "asset.cpp", `#include <cstring>
int main() { return assets_count != 2 || std::strcmp(assets[0].name, "font.ttf") || assets[0].size != 4 || std::memcmp(assets[0].data, "\0\1\200\377", 4) || assets[1].size; }`, nil},
		{"go", "go", "asset.go", `func main() { if len(assets)!=2 || string(assets["font.ttf"]) != "\x00\x01\x80\xff" || len(assets["empty"])!=0 { panic("bad") } }`, []string{"run", "asset.go"}},
		{"python", "python", "asset.py", `assert assets['font.ttf'] == bytes([0,1,128,255]) and assets['empty'] == b''`, []string{"asset.py"}},
		{"javascript", "node", "asset.js", `if (!assets['font.ttf'].equals(Buffer.from([0,1,128,255])) || assets.empty.length) throw Error('bad');`, []string{"asset.js"}},
		{"rust", "rustc", "asset.rs", `fn main() { assert_eq!(ASSETS[0].0, "font.ttf"); assert_eq!(ASSETS[0].1, &[0,1,128,255]); assert!(ASSETS[1].1.is_empty()); }`, nil},
	}
	for _, tc := range tests {
		t.Run(tc.lang, func(t *testing.T) {
			tool, err := exec.LookPath(tc.tool)
			if err != nil {
				t.Skipf("%s unavailable", tc.tool)
			}
			dir := t.TempDir()
			src, err := Render(tc.lang, []Asset{{Name: "font.ttf", Data: []byte{0, 1, 128, 255}}, {Name: "empty"}})
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, tc.file), append(src, []byte("\n"+tc.script+"\n")...), 0600); err != nil {
				t.Fatal(err)
			}
			args := tc.args
			if args == nil {
				args = []string{tc.file, "-o", "program.exe"}
			}
			cmd := exec.Command(tool, args...)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("compile/run: %v\n%s", err, out)
			}
			if tc.args == nil {
				cmd = exec.Command(filepath.Join(dir, "program.exe"))
				cmd.Dir = dir
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("execute: %v\n%s", err, out)
				}
			}
		})
	}
}
func TestRustNameEscapes(t *testing.T) {
	s := rustString("a\x01\n\"♥")
	if !strings.Contains(s, `\u{1}`) {
		t.Fatal(s)
	}
}
