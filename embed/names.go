package embed

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func defaultVariableName(data []byte) string {
	sum := sha256.Sum256(data)
	return "mmbed_file_" + hex.EncodeToString(sum[:3])
}

var identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var reserved = map[string]bool{}

func init() {
	// Union of common keywords avoids generating identifiers that are invalid in
	// one target language. Internal names also cannot be used as asset names.
	for _, word := range strings.Fields("abstract alias and as assert async await base begin boolean break case catch char class const continue def default defer del delete do double elif else elseif end ensure enum eval except export extends extern false final finally float fn for foreach from func function global go goto if implements import in instance instanceof int interface internal is let local long map match module mut namespace native new next nil None none nonlocal not null object operator or package pass private protected pub public raise redo ref register repeat require rescue retry return self short signed sizeof static string struct super switch synchronized then this throw throws trait true try type typeof uint undef union unless unsafe unsigned until use using val var virtual void volatile when where while with yield assets Assets ASSETS Mmmbed mmmbedDecode __proto__ constructor prototype") {
		reserved[word] = true
	}
}

func variableNames(assets []Asset, requested []string) ([]string, error) {
	if len(requested) != 0 && len(requested) != len(assets) {
		return nil, fmt.Errorf("provide one --variable per input (got %d for %d inputs)", len(requested), len(assets))
	}
	names := make([]string, len(assets))
	used := map[string]bool{}
	for i, a := range assets {
		name := defaultVariableName(a.Data)
		if len(requested) != 0 {
			name = requested[i]
		}
		if name == "_" || !identifier.MatchString(name) || reserved[name] || strings.HasPrefix(name, "mmmbed_") {
			return nil, fmt.Errorf("invalid or reserved variable name %q", name)
		}
		// Reserve suffixes too, so separate variables cannot shadow metadata.
		for _, n := range []string{name, name + "_filename", name + "_path_parent", name + "_size"} {
			if used[n] {
				return nil, fmt.Errorf("variable name collision %q; specify distinct --variable names", n)
			}
			used[n] = true
		}
		names[i] = name
	}
	return names, nil
}

// addAliases retains the collection interface and exposes each asset separately.
func addAliases(lang string, source []byte, assets []Asset, names []string) []byte {
	var b strings.Builder
	for i, a := range assets {
		n := names[i]
		key := base64.StdEncoding.EncodeToString([]byte(a.Name))
		parent := base64.StdEncoding.EncodeToString([]byte(a.Parent))
		q := strconv.Quote
		switch lang {
		case "c", "cpp":
			fmt.Fprintf(&b, "\nstatic const unsigned char *const %s = mmmbed_data_%d;\nstatic const size_t %s_size = %d;\nstatic const char *const %s_filename = (const char *)mmmbed_name_%d;\nstatic const unsigned char %s_path_parent[] = {%s%s0};\n", n, i, n, len(a.Data), n, i, n, numbers([]byte(a.Parent)), comma(a.Parent != ""))
		case "go":
			fmt.Fprintf(&b, "\nvar %s = Assets[string(mmmbedDecode(%q))]\nvar %s_filename = string(mmmbedDecode(%q))\nvar %s_path_parent = string(mmmbedDecode(%q))\n", n, key, n, key, n, parent)
		case "javascript", "typescript":
			fmt.Fprintf(&b, "\n%sconst %s = assets[Buffer.from(%q, 'base64').toString('utf8')];\n%sconst %s_filename = Buffer.from(%q, 'base64').toString('utf8');\n%sconst %s_path_parent = Buffer.from(%q, 'base64').toString('utf8');\n", exportTS(lang), n, key, exportTS(lang), n, key, exportTS(lang), n, parent)
			if lang == "javascript" {
				fmt.Fprintf(&b, "module.exports.%s = %s; module.exports.%s_filename = %s_filename; module.exports.%s_path_parent = %s_path_parent;\n", n, n, n, n, n, n)
			}
		case "python":
			fmt.Fprintf(&b, "\n%s_filename = base64.b64decode(%q).decode('utf-8')\n%s_path_parent = base64.b64decode(%q).decode('utf-8')\n%s = assets[%s_filename]\n", n, key, n, parent, n, n)
		case "rust":
			fmt.Fprintf(&b, "\n#[allow(non_upper_case_globals)]\npub static %s: &[u8] = ASSETS[%d].1;\n#[allow(non_upper_case_globals)]\npub static %s_filename: &str = ASSETS[%d].0;\n#[allow(non_upper_case_globals)]\npub static %s_path_parent: &str = %s;\n", n, i, n, i, n, rustString(a.Parent))
		case "java":
			fmt.Fprintf(&b, "  static final String %s_filename = new String(Base64.getDecoder().decode(%q), StandardCharsets.UTF_8);\n  static final String %s_path_parent = new String(Base64.getDecoder().decode(%q), StandardCharsets.UTF_8);\n  static final byte[] %s = assets.get(%s_filename);\n", n, key, n, parent, n, n)
		case "csharp":
			fmt.Fprintf(&b, " public static readonly string %s_filename = Encoding.UTF8.GetString(Convert.FromBase64String(%q));\n public static readonly string %s_path_parent = Encoding.UTF8.GetString(Convert.FromBase64String(%q));\n public static readonly byte[] %s = Assets[%s_filename];\n", n, key, n, parent, n, n)
		case "php":
			fmt.Fprintf(&b, "\n$%s_filename = base64_decode(%q);\n$%s_path_parent = base64_decode(%q);\n$%s = $assets[$%s_filename];\n", n, key, n, parent, n, n)
		case "ruby":
			fmt.Fprintf(&b, "\n%s_filename = Base64.strict_decode64(%q).force_encoding('UTF-8')\n%s_path_parent = Base64.strict_decode64(%q).force_encoding('UTF-8')\n%s = ASSETS[%s_filename]\n", n, key, n, parent, n, n)
		case "swift":
			fmt.Fprintf(&b, "\nlet %s_filename = String(data: Data(base64Encoded: %q)!, encoding: .utf8)!\nlet %s_path_parent = String(data: Data(base64Encoded: %q)!, encoding: .utf8)!\nlet %s = assets[%s_filename]!\n", n, key, n, parent, n, n)
		case "kotlin":
			fmt.Fprintf(&b, "\nval %s_filename = String(Base64.getDecoder().decode(%q), Charsets.UTF_8)\nval %s_path_parent = String(Base64.getDecoder().decode(%q), Charsets.UTF_8)\nval %s = assets.getValue(%s_filename)\n", n, key, n, parent, n, n)
		case "scala":
			fmt.Fprintf(&b, " val %s_filename = new String(Base64.getDecoder.decode(%q), StandardCharsets.UTF_8)\n val %s_path_parent = new String(Base64.getDecoder.decode(%q), StandardCharsets.UTF_8)\n val %s = assets(%s_filename)\n", n, key, n, parent, n, n)
		case "dart":
			fmt.Fprintf(&b, "\nfinal %s_filename = utf8.decode(base64.decode(%q));\nfinal %s_path_parent = utf8.decode(base64.decode(%q));\nfinal %s = assets[%s_filename]!;\n", n, key, n, parent, n, n)
		case "lua":
			fmt.Fprintf(&b, "assets[%s] = assets[%s]\nassets[%s] = %s\nassets[%s] = %s\n", q(n), luaString([]byte(a.Name)), q(n+"_filename"), luaString([]byte(a.Name)), q(n+"_path_parent"), luaString([]byte(a.Parent)))
		case "perl":
			fmt.Fprintf(&b, "\nour $%s_filename = decode('UTF-8', decode_base64(%q));\nour $%s_path_parent = decode('UTF-8', decode_base64(%q));\nour $%s = $assets{$%s_filename};\n", n, key, n, parent, n, n)
		case "r":
			fmt.Fprintf(&b, "\n%s_filename <- rawToChar(as.raw(c(%s)))\n%s_path_parent <- rawToChar(as.raw(c(%s)))\n%s <- assets[[%s_filename]]\n", n, numbers([]byte(a.Name)), n, numbers([]byte(a.Parent)), n, n)
		case "julia":
			fmt.Fprintf(&b, "\nconst %s_filename = String(base64decode(%q))\nconst %s_path_parent = String(base64decode(%q))\nconst %s = assets[%s_filename]\n", n, key, n, parent, n, n)
		case "elixir":
			fmt.Fprintf(&b, "  def %s_filename, do: Base.decode64!(%q)\n  def %s_path_parent, do: Base.decode64!(%q)\n  def %s, do: Map.fetch!(assets(), %s_filename())\n", n, key, n, parent, n, n)
		}
	}
	s := string(source)
	switch lang {
	case "java", "csharp", "elixir":
		idx := strings.LastIndex(s, "}\n")
		if lang == "elixir" {
			idx = strings.LastIndex(s, "end\n")
		}
		if idx >= 0 {
			return []byte(s[:idx] + b.String() + s[idx:])
		}
	case "scala":
		idx := strings.LastIndex(s, " }")
		if idx >= 0 {
			return []byte(s[:idx] + "\n" + b.String() + s[idx:])
		}
	case "perl":
		idx := strings.LastIndex(s, "1;\n")
		if idx >= 0 {
			return []byte(s[:idx] + b.String() + s[idx:])
		}
	case "lua":
		idx := strings.LastIndex(s, "return assets")
		if idx >= 0 {
			return []byte(s[:idx] + b.String() + s[idx:])
		}
	}
	return []byte(s + b.String())
}
func exportTS(lang string) string {
	if lang == "typescript" {
		return "export "
	}
	return ""
}
