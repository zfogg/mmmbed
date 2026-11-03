package cli

import (
	"bytes"
	"github.com/zfogg/mmmbed/embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(args []string, input string) (string, error) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}
func TestStdin(t *testing.T) {
	out, err := run([]string{"-l", "python", "--name", "in.bin"}, "\x00\xff")
	if err != nil || !strings.Contains(out, "AP8=") {
		t.Fatalf("%s %v", out, err)
	}
}
func TestMultipleAndOutput(t *testing.T) {
	d := t.TempDir()
	a := filepath.Join(d, "a")
	b := filepath.Join(d, "b")
	out := filepath.Join(d, "generated.go")
	os.WriteFile(a, []byte("a"), 0600)
	os.WriteFile(b, []byte("b"), 0600)
	_, err := run([]string{"-l", "go", "-i", a, "-i", b, "-o", out}, "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil || !bytes.Contains(data, []byte("YQ==")) || !bytes.Contains(data, []byte("Yg==")) {
		t.Fatalf("%s %v", data, err)
	}
}
func TestPositionalAndErrors(t *testing.T) {
	d := t.TempDir()
	a := filepath.Join(d, "a")
	o := filepath.Join(d, "out")
	os.WriteFile(a, []byte("original"), 0600)
	if _, err := run([]string{"-l", "go", a, o}, ""); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-l", "go"}, {"-l", "go", a, a}, {"-l", "go", a, o, "-o", o}, {"-l", "bad", a}} {
		if _, err := run(args, ""); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	data, _ := os.ReadFile(a)
	if string(data) != "original" {
		t.Fatal("input overwritten")
	}
}

func TestGoPackageFlag(t *testing.T) {
	out, err := run([]string{"-l", "go", "--name", "x", "--go-package", "media"}, "abc")
	if err != nil || !strings.Contains(out, "package media\n") || !strings.Contains(out, "var Assets") {
		t.Fatalf("%s %v", out, err)
	}
	if _, err := run([]string{"-l", "python", "--name", "x", "--go-package", "media"}, "abc"); err == nil {
		t.Fatal("accepted Go package for Python")
	}
}

func TestDuplicateBasenames(t *testing.T) {
	d := t.TempDir()
	a, b := filepath.Join(d, "a"), filepath.Join(d, "b")
	for _, dir := range []string{a, b} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "same"), []byte(dir), 0600); err != nil {
			t.Fatal(err)
		}
	}
	out, err := run([]string{"-l", "go", "-i", filepath.Join(a, "same"), "-i", filepath.Join(b, "same")}, "")
	if err == nil || out != "" {
		t.Fatalf("out=%q err=%v", out, err)
	}
}

func TestVariableAndParent(t *testing.T) {
	d := t.TempDir()
	dir := filepath.Join(d, "fonts")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "font.ttf")
	if err := os.WriteFile(file, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := run([]string{"-l", "python", "--variable", "font", file}, "")
	if err != nil || !strings.Contains(out, "font_path_parent") || !strings.Contains(out, "Zm9udHM=") || strings.Contains(out, d) {
		t.Fatalf("%s %v", out, err)
	}
	if _, err := run([]string{"-l", "go", "--variable", "font", "-i", file, "-i", file}, ""); err == nil {
		t.Fatal("accepted wrong variable count")
	}
}

func TestParentOptions(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		expected string
	}{
		{[]string{"--omit-parent"}, `mmmbedDecode("")`},
		{[]string{"--parent", "fixed"}, `mmmbedDecode("Zml4ZWQ=")`},
	} {
		args := append([]string{"-l", "go", "--name", "a"}, tc.args...)
		out, err := run(args, "data")
		if err != nil || !strings.Contains(out, tc.expected) {
			t.Fatalf("%v: %s %v", args, out, err)
		}
	}
	if _, err := run([]string{"-l", "go", "--name", "a", "--parent", "x", "--omit-parent"}, ""); err == nil {
		t.Fatal("accepted conflicting flags")
	}
}

func TestLanguagesCommand(t *testing.T) {
	out, err := run([]string{"languages"}, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range embed.SortedLanguages() {
		if !strings.Contains(out, name) || languageInfo[name] == "" {
			t.Errorf("missing language %s", name)
		}
	}
}
