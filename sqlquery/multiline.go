package sqlquery

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

func formattedSQLLiteral(lang, sql string, formatted bool, fallback string) string {
	if !formatted {
		return fallback
	}
	for _, r := range sql {
		if (r < 32 && r != '\n' && r != '\r' && r != '\t') || r == '\u2028' || r == '\u2029' {
			return fallback
		}
	}
	return multilineSQLLiteral(lang, sql)
}

// multilineSQLLiteral keeps the SQL bytes exactly while exposing its lines in source.
func multilineSQLLiteral(lang, sql string) string {
	lines := strings.SplitAfter(sql, "\n")
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		var part string
		switch lang {
		case "c", "cpp":
			part = cString(line)
		case "go":
			part = strconv.Quote(line)
		case "rust":
			part = rustString(line)
		case "lua":
			part = luaString(line)
		default:
			var encoded bytes.Buffer
			encoder := json.NewEncoder(&encoded)
			encoder.SetEscapeHTML(false)
			_ = encoder.Encode(line)
			part = strings.TrimSuffix(encoded.String(), "\n")
			switch lang {
			case "php", "kotlin", "dart", "julia":
				part = strings.ReplaceAll(part, "$", `\$`)
			case "perl":
				part = strings.ReplaceAll(strings.ReplaceAll(part, "$", `\$`), "@", `\@`)
			case "ruby", "elixir":
				part = strings.ReplaceAll(part, "#{", `\#{`)
			}
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return `""`
	}
	switch lang {
	case "c", "cpp":
		return strings.Join(parts, "\n    ")
	case "python":
		return "(" + strings.Join(parts, "\n        ") + ")"
	case "rust":
		return "concat!(" + strings.Join(parts, ",\n        ") + ")"
	case "r":
		return "paste0(" + strings.Join(parts, ",\n        ") + ")"
	case "lua":
		return "(" + strings.Join(parts, " ..\n        ") + ")"
	case "php", "perl":
		return "(" + strings.Join(parts, " .\n        ") + ")"
	case "elixir":
		return "(" + strings.Join(parts, " <>\n        ") + ")"
	default:
		return "(" + strings.Join(parts, " +\n        ") + ")"
	}
}
