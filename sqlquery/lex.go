package sqlquery

import "strings"

// sqlLex recognizes common SQL quoted strings, identifiers and comments. It is
// deliberately not a SQL parser: directives are recognized only at top level.
type sqlLex struct {
	quote         byte
	escape        bool
	dollar        string
	block         int
	hasSQL        bool
	mysql         bool
	arrayBrackets bool
}

func (l *sqlLex) normal() bool { return l.quote == 0 && l.dollar == "" && l.block == 0 }

func (l *sqlLex) scan(s string) {
	for i := 0; i < len(s); {
		if l.block > 0 {
			switch {
			case strings.HasPrefix(s[i:], "/*"):
				l.block++
				i += 2
			case strings.HasPrefix(s[i:], "*/"):
				l.block--
				i += 2
			default:
				i++
			}
			continue
		}
		if l.dollar != "" {
			if strings.HasPrefix(s[i:], l.dollar) {
				i += len(l.dollar)
				l.dollar = ""
			} else {
				i++
			}
			continue
		}
		if l.quote != 0 {
			if l.escape && s[i] == '\\' && i+1 < len(s) {
				i += 2
				continue
			}
			end := l.quote
			if end == '[' {
				end = ']'
			}
			if s[i] == end {
				if i+1 < len(s) && s[i+1] == end {
					i += 2
				} else {
					l.quote = 0
					i++
				}
			} else {
				i++
			}
			continue
		}
		if strings.HasPrefix(s[i:], "--") && (!l.mysql || i+2 == len(s) || s[i+2] <= ' ') {
			break
		}
		if l.mysql && s[i] == '#' {
			break
		}
		if strings.HasPrefix(s[i:], "/*") {
			l.block = 1
			i += 2
			continue
		}
		switch s[i] {
		case '\'', '"', '`', '[':
			if s[i] == '[' && l.arrayBrackets {
				l.hasSQL = true
				i++
				continue
			}
			l.quote = s[i]
			l.escape = (l.mysql && (s[i] == '\'' || s[i] == '"')) || (!l.mysql && s[i] == '\'' && i > 0 && (s[i-1] == 'E' || s[i-1] == 'e') && (i == 1 || !isIdentByte(s[i-2])))
			l.hasSQL = true
			i++
		case '$':
			if i > 0 && (isIdentByte(s[i-1]) || s[i-1] == '$') {
				l.hasSQL = true
				i++
				continue
			}
			j := i + 1
			for j < len(s) && ((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z') || s[j] == '_' || s[j] >= 0x80 || (j > i+1 && s[j] >= '0' && s[j] <= '9')) {
				j++
			}
			if j < len(s) && s[j] == '$' {
				l.dollar = s[i : j+1]
				i = j + 1
			} else {
				i++
			}
			l.hasSQL = true
		case ' ', '\t', '\r', '\n', '\f':
			i++
		default:
			l.hasSQL = true
			i++
		}
	}
}

func isIdentByte(c byte) bool {
	return c >= 0x80 || c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}
