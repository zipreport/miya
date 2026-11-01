package filters

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// pyParseInt mirrors Python's int(s, base): surrounding whitespace, an optional
// sign, an optional 0x/0o/0b prefix and single underscores between digits.
func pyParseInt(s string, base int) (int64, bool) {
	if base != 0 && (base < 2 || base > 36) {
		return 0, false
	}
	s = strings.TrimFunc(s, unicode.IsSpace)
	sign := ""
	if s != "" && (s[0] == '+' || s[0] == '-') {
		if s[0] == '-' {
			sign = "-"
		}
		s = s[1:]
	}

	prefixed := false
	if len(s) > 2 && s[0] == '0' {
		pb := 0
		switch s[1] {
		case 'x', 'X':
			pb = 16
		case 'o', 'O':
			pb = 8
		case 'b', 'B':
			pb = 2
		}
		if pb != 0 && (base == 0 || base == pb) {
			base = pb
			s = s[2:]
			prefixed = true
		}
	}
	if base == 0 {
		base = 10
		if len(s) > 1 && s[0] == '0' && strings.Trim(strings.ReplaceAll(s, "_", ""), "0") != "" {
			return 0, false
		}
	}

	if prefixed && strings.HasPrefix(s, "_") {
		s = s[1:]
	}
	if s == "" || s[0] == '_' || s[len(s)-1] == '_' || strings.Contains(s, "__") {
		return 0, false
	}
	s = strings.ReplaceAll(s, "_", "")
	if s[0] == '+' || s[0] == '-' {
		return 0, false
	}
	n, err := strconv.ParseInt(sign+s, base, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// pyParseFloat mirrors Python's float(s): surrounding whitespace, single
// underscores between digits, inf and nan, but no hexadecimal. An exponent out
// of range gives an infinity, not a failure.
func pyParseFloat(s string) (float64, bool) {
	s = strings.TrimFunc(s, unicode.IsSpace)
	if s == "" || strings.ContainsAny(s, "xXpP") {
		return 0, false
	}
	if strings.Contains(s, "_") {
		for i := 0; i < len(s); i++ {
			if s[i] != '_' {
				continue
			}
			if i == 0 || i == len(s)-1 || !isASCIIDigit(s[i-1]) || !isASCIIDigit(s[i+1]) {
				return 0, false
			}
		}
		s = strings.ReplaceAll(s, "_", "")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		var ne *strconv.NumError
		if errors.As(err, &ne) && ne.Err == strconv.ErrRange {
			return f, true
		}
		return 0, false
	}
	return f, true
}

func isASCIIDigit(c byte) bool { return c >= '0' && c <= '9' }
