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
	return "mmmbed_file_" + hex.EncodeToString(sum[:3])
}

var identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Language keywords differ: for example, `match` is valid in C but reserved
// in Rust. Generated helper names remain reserved in every target.
var reserved = map[string]bool{}
var keywords = map[string]map[string]bool{}

func init() {
	for _, w := range strings.Fields("assets Assets ASSETS Mmmbed mmmbedDecode __proto__ constructor prototype") {
		reserved[w] = true
	}
	lists := map[string]string{
		"c":          "auto break case char const continue default do double else enum extern float for goto if inline int long register restrict return short signed sizeof static struct switch typedef union unsigned void volatile while _Alignas _Alignof _Atomic _Bool _Complex _Generic _Imaginary _Noreturn _Static_assert _Thread_local",
		"cpp":        "alignas alignof and and_eq asm auto bitand bitor bool break case catch char class compl concept const constexpr const_cast continue co_await co_return co_yield decltype default delete do double dynamic_cast else enum explicit export extern false float for friend goto if inline int long mutable namespace new noexcept not not_eq nullptr operator or or_eq private protected public register reinterpret_cast requires return short signed sizeof static static_assert static_cast struct switch template this thread_local throw true try typedef typeid typename union unsigned using virtual void volatile wchar_t while xor xor_eq",
		"go":         "break default func interface select case defer go map struct chan else goto package switch const fallthrough if range type continue for import return var init nil",
		"javascript": "await break case catch class const continue debugger default delete do else enum export extends false finally for function if implements import in instanceof interface let new null package private protected public return static super switch this throw true try typeof var void while with yield",
		"typescript": "abstract any as asserts async await bigint boolean break case catch class const constructor continue debugger declare default delete do else enum export extends false finally for from function get if implements import in infer instanceof interface is keyof let module namespace never new null number object of package private protected public readonly require return set static string super switch symbol this throw true try type typeof undefined unique unknown var void while with yield",
		"python":     "False None True and as assert async await break class continue def del elif else except finally for from global if import in is lambda nonlocal not or pass raise return try while with yield",
		"rust":       "as async await break const continue crate dyn else enum extern false fn for if impl in let loop match mod move mut pub ref return self Self static struct super trait true type unsafe use where while abstract become box do final macro override priv typeof unsized virtual yield try",
		"java":       "abstract assert boolean break byte case catch char class const continue default do double else enum extends final finally float for goto if implements import instanceof int interface long native new package private protected public return short static strictfp super switch synchronized this throw throws transient try void volatile while true false null var yield record sealed permits non-sealed",
		"csharp":     "abstract as base bool break byte case catch char checked class const continue decimal default delegate do double else enum event explicit extern false finally fixed float for foreach goto if implicit in int interface internal is lock long namespace new null object operator out override params private protected public readonly ref return sbyte sealed short sizeof stackalloc static string struct switch this throw true try typeof uint ulong unchecked unsafe ushort using virtual void volatile while",
		"php":        "abstract and array as break callable case catch class clone const continue declare default do echo else elseif empty enddeclare endfor endforeach endif endswitch endwhile eval exit extends final finally fn for foreach function global goto if implements include include_once instanceof insteadof interface isset list match namespace new or print private protected public readonly require require_once return static switch throw trait try unset use var while xor yield",
		"ruby":       "BEGIN END alias and begin break case class def defined do else elsif end ensure false for if in module next nil not or redo rescue retry return self super then true undef unless until when while yield",
		"swift":      "associatedtype class deinit enum extension fileprivate func import init inout internal let open operator private protocol public rethrows static struct subscript typealias var break case continue default defer do else fallthrough for guard if in repeat return switch where while as Any catch false is nil super self Self throw throws true try",
		"kotlin":     "as break class continue do else false for fun if in interface is null object package return super this throw true try typealias typeof val var when while",
		"scala":      "abstract case catch class def do else enum export extends false final finally for forSome given if implicit import lazy match new null object override package private protected return sealed super then this throw trait try true type val var while with yield",
		"dart":       "abstract as assert async await break case catch class const continue covariant default deferred do dynamic else enum export extends extension external factory false final finally for Function get if implements import in interface is late library mixin new null on operator part required rethrow return set static super switch sync this throw true try typedef var void while with yield",
		"lua":        "and break do else elseif end false for function goto if in local nil not or repeat return then true until while",
		"perl":       "", // Perl variables have a sigil and may use keyword spelling.
		"r":          "if else repeat while function for in next break TRUE FALSE NULL Inf NaN NA NA_integer_ NA_real_ NA_complex_ NA_character_",
		"julia":      "baremodule begin break catch const continue do else elseif end export false finally for function global if import let local macro module quote return struct true try using while",
		"elixir":     "after and catch do else end false fn in nil not or rescue true when",
	}
	for lang, list := range lists {
		set := map[string]bool{}
		for _, w := range strings.Fields(list) {
			set[w] = true
		}
		keywords[lang] = set
	}
}

func variableNames(lang string, assets []Asset, requested []string) ([]string, error) {
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
		if name == "_" || !identifier.MatchString(name) || reserved[name] || keywords[lang][name] || (len(requested) != 0 && strings.HasPrefix(name, "mmmbed_")) {
			return nil, fmt.Errorf("invalid or reserved variable name %q", name)
		}
		conflicts := func(candidate string) bool {
			for _, n := range []string{candidate, candidate + "_filename", candidate + "_path_parent", candidate + "_size"} {
				if used[n] {
					return true
				}
			}
			return false
		}
		if len(requested) == 0 {
			base := name
			for suffix := 2; conflicts(name); suffix++ {
				name = fmt.Sprintf("%s_%d", base, suffix)
			}
		} else if conflicts(name) {
			return nil, fmt.Errorf("variable name collision %q", name)
		}
		for _, n := range []string{name, name + "_filename", name + "_path_parent", name + "_size"} {
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
