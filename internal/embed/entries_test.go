package embed

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// TestGeneratedEntries checks each target's public declaration and the exact
// association of two filenames with binary and empty data, without executing it.
func TestGeneratedEntries(t *testing.T) {
	tests := map[string]struct{ declaration, first, second string }{
		"c":          {"struct mmmbed_asset", "{(const char *)mmmbed_name_0, mmmbed_data_0, 5}", "{(const char *)mmmbed_name_1, mmmbed_data_1, 0}"},
		"cpp":        {"struct mmmbed_asset", "{(const char *)mmmbed_name_0, mmmbed_data_0, 5}", "{(const char *)mmmbed_name_1, mmmbed_data_1, 0}"},
		"go":         {"var Assets = map[string][]byte{", `string(mmmbedDecode("YS7imaU=")): mmmbedDecode("AAF/gP8="),`, `string(mmmbedDecode("ZW1wdHk=")): mmmbedDecode(""),`},
		"javascript": {"module.exports = { assets };", `[Buffer.from("YS7imaU=", 'base64').toString('utf8')]: Buffer.from("AAF/gP8=", 'base64'),`, `[Buffer.from("ZW1wdHk=", 'base64').toString('utf8')]: Buffer.from("", 'base64'),`},
		"typescript": {"export const assets: Record<string, Buffer>", `[Buffer.from("YS7imaU=", 'base64').toString('utf8')]: Buffer.from("AAF/gP8=", 'base64'),`, `[Buffer.from("ZW1wdHk=", 'base64').toString('utf8')]: Buffer.from("", 'base64'),`},
		"python":     {"assets = {", `base64.b64decode("YS7imaU=").decode('utf-8'): base64.b64decode("AAF/gP8="),`, `base64.b64decode("ZW1wdHk=").decode('utf-8'): base64.b64decode(""),`},
		"rust":       {"pub static ASSETS: &[(&str, &[u8])]", `("a.♥", &[0,1,127,128,255,]),`, `("empty", &[]),`},
		"java":       {"static final Map<String, byte[]> assets", `assets.put(new String(Base64.getDecoder().decode("YS7imaU="), StandardCharsets.UTF_8), Base64.getDecoder().decode("AAF/gP8="));`, `assets.put(new String(Base64.getDecoder().decode("ZW1wdHk="), StandardCharsets.UTF_8), Base64.getDecoder().decode(""));`},
		"csharp":     {"public static readonly Dictionary<string, byte[]> Assets", `{ Encoding.UTF8.GetString(Convert.FromBase64String("YS7imaU=")), Convert.FromBase64String("AAF/gP8=") },`, `{ Encoding.UTF8.GetString(Convert.FromBase64String("ZW1wdHk=")), Convert.FromBase64String("") },`},
		"php":        {"$assets = [", `base64_decode("YS7imaU=") => base64_decode("AAF/gP8="),`, `base64_decode("ZW1wdHk=") => base64_decode(""),`},
		"ruby":       {"ASSETS = {", `Base64.strict_decode64("YS7imaU=").force_encoding('UTF-8') => Base64.strict_decode64("AAF/gP8="),`, `Base64.strict_decode64("ZW1wdHk=").force_encoding('UTF-8') => Base64.strict_decode64(""),`},
		"swift":      {"let assets: [String: Data]", `String(data: Data(base64Encoded: "YS7imaU=")!, encoding: .utf8)!: Data(base64Encoded: "AAF/gP8=")!,`, `String(data: Data(base64Encoded: "ZW1wdHk=")!, encoding: .utf8)!: Data(base64Encoded: "")!,`},
		"kotlin":     {"val assets: Map<String, ByteArray>", `String(Base64.getDecoder().decode("YS7imaU="), Charsets.UTF_8) to Base64.getDecoder().decode("AAF/gP8="),`, `String(Base64.getDecoder().decode("ZW1wdHk="), Charsets.UTF_8) to Base64.getDecoder().decode(""),`},
		"scala":      {"object Mmmbed", `new String(Base64.getDecoder.decode("YS7imaU="), StandardCharsets.UTF_8) -> Base64.getDecoder.decode("AAF/gP8="),`, `new String(Base64.getDecoder.decode("ZW1wdHk="), StandardCharsets.UTF_8) -> Base64.getDecoder.decode(""),`},
		"dart":       {"final Map<String, Uint8List> assets", `utf8.decode(base64.decode("YS7imaU=")): base64.decode("AAF/gP8="),`, `utf8.decode(base64.decode("ZW1wdHk=")): base64.decode(""),`},
		"lua":        {"return assets", `["\097\046\226\153\165"] = "\000\001\127\128\255",`, `["\101\109\112\116\121"] = "",`},
		"perl":       {"our %assets = (", `decode('UTF-8', decode_base64("YS7imaU=")) => decode_base64("AAF/gP8="),`, `decode('UTF-8', decode_base64("ZW1wdHk=")) => decode_base64(""),`},
		"r":          {"assets <- list()", `assets[[rawToChar(as.raw(c(97,46,226,153,165)))]] <- as.raw(c(0,1,127,128,255))`, `assets[[rawToChar(as.raw(c(101,109,112,116,121)))]] <- as.raw(c())`},
		"julia":      {"const assets = Dict{String, Vector{UInt8}}", `String(base64decode("YS7imaU=")) => base64decode("AAF/gP8="),`, `String(base64decode("ZW1wdHk=")) => base64decode(""),`},
		"elixir":     {"defmodule Mmmbed do", `Base.decode64!("YS7imaU=") => Base.decode64!("AAF/gP8="),`, `Base.decode64!("ZW1wdHk=") => Base.decode64!(""),`},
	}
	if len(tests) != len(Languages()) {
		t.Fatalf("cases: %d; languages: %d", len(tests), len(Languages()))
	}
	for _, lang := range Languages() {
		t.Run(lang, func(t *testing.T) {
			tc, ok := tests[lang]
			if !ok {
				t.Fatal("missing language case")
			}
			got, err := Render(lang, []Asset{{Name: "a.♥", Data: []byte{0, 1, 127, 128, 255}}, {Name: "empty"}})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{tc.declaration, tc.first, tc.second} {
				if !bytes.Contains(got, []byte(want)) {
					t.Errorf("missing %q\n%s", want, got)
				}
			}
			if strings.Index(string(got), tc.first) >= strings.Index(string(got), tc.second) {
				t.Error("entries are not in input order")
			}
			if lang == "c" || lang == "cpp" {
				for _, want := range []string{"mmmbed_name_0[] = {97,46,226,153,165,0}", "mmmbed_data_0[] = {0,1,127,128,255,0}", "mmmbed_name_1[] = {101,109,112,116,121,0}", "mmmbed_data_1[] = {0}", "assets_count = 2"} {
					if !strings.Contains(string(got), want) {
						t.Error(fmt.Sprintf("missing %q", want))
					}
				}
			}
		})
	}
}
