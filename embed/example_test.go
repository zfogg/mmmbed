package embed_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zfogg/mmmbed/embed"
)

func ExampleRenderTo() {
	input := embed.StreamAsset{Name: "message.txt", Reader: strings.NewReader("hello")}
	var output bytes.Buffer
	err := embed.RenderTo(&output, "go", []embed.StreamAsset{input}, embed.Options{GoPackage: "assets", Variables: []string{"Message"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(strings.Contains(output.String(), "package assets"))
	// Output: true
}

func Example_errors() {
	_, err := embed.Render("unknown", []embed.Asset{{Name: "x"}})
	fmt.Println(errors.Is(err, embed.ErrUnsupportedLanguage))
	// Output: true
}
