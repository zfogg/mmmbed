package embed

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRenderToMatchesRender(t *testing.T) {
	for _, lang := range Languages() {
		t.Run(lang, func(t *testing.T) {
			assets := []Asset{{Name: "font.ttf", Parent: "fonts", Data: []byte{0, 1, 128, 255}}, {Name: "empty"}, {Name: "same", Data: []byte{0, 1, 128, 255}}}
			stream := make([]StreamAsset, len(assets))
			for i, a := range assets {
				stream[i] = StreamAsset{Name: a.Name, Parent: a.Parent, Reader: bytes.NewReader(a.Data)}
			}
			for _, opt := range []Options{{}, {Variables: []string{"font", "empty_file", "same_file"}}} {
				want, err := RenderWithOptions(lang, assets, opt)
				if err != nil {
					t.Fatal(err)
				}
				var got bytes.Buffer
				if err = RenderTo(&got, lang, stream, opt); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(want, got.Bytes()) {
					t.Fatalf("output differs\nwant: %s\ngot: %s", want, got.Bytes())
				}
				for i, a := range assets {
					stream[i].Reader = bytes.NewReader(a.Data)
				}
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestRenderToErrors(t *testing.T) {
	if err := RenderTo(io.Discard, "go", []StreamAsset{{Name: "x", Reader: failingReader{}}}, Options{}); err == nil {
		t.Fatal("missing read error")
	}
	if err := RenderTo(failingWriter{}, "go", []StreamAsset{{Name: "x", Reader: strings.NewReader("abc")}}, Options{}); err == nil {
		t.Fatal("missing write error")
	}
}
func TestRenderToLargeJavaChunks(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 20000)
	var out bytes.Buffer
	if err := RenderTo(&out, "java", []StreamAsset{{Name: "large", Reader: bytes.NewReader(data)}}, Options{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "String.join(") {
		t.Fatal("large string not split")
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

type countWriter struct{ n int64 }

func (w *countWriter) Write(p []byte) (int, error) { w.n += int64(len(p)); return len(p), nil }
func TestRenderToLargeStream(t *testing.T) {
	const size = 16 << 20
	var out countWriter
	if err := RenderTo(&out, "go", []StreamAsset{{Name: "large", Reader: io.LimitReader(zeroReader{}, size)}}, Options{}); err != nil {
		t.Fatal(err)
	}
	if out.n < size {
		t.Fatalf("output too short: %d", out.n)
	}
}
