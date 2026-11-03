package cli

import (
	"bytes"
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
