package filters

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// regexTitleCase is the regexp-based titleCase this package used before; the
// rune loop must give the same result.
var reTitleSepRef = regexp.MustCompile(`[-\p{Z}\t-\r\x{1c}-\x{1f}\x{85}({\[<]+`)

func regexTitleCase(s string) string {
	var b strings.Builder
	last := 0
	word := func(w string) {
		if w == "" {
			return
		}
		r, size := utf8.DecodeRuneInString(w)
		b.WriteString(strings.ToUpper(string(r)))
		b.WriteString(strings.ToLower(w[size:]))
	}
	for _, m := range reTitleSepRef.FindAllStringIndex(s, -1) {
		word(s[last:m[0]])
		b.WriteString(s[m[0]:m[1]])
		last = m[1]
	}
	word(s[last:])
	return b.String()
}

func TestTitleCaseMatchesRegexVersion(t *testing.T) {
	pieces := []string{
		"a", "B", "hello", "WORLD", "\u00df", "\u01c5", "\u01c6", "\u0130", "\u0131", "\ufb01", "\u00e9", "e\u0301", "1", "_", "'", ".", "!",
		" ", "  ", "-", "(", "{", "[", "<", ")", "]", ">", "\t", "\n", "\v", "\f", "\r", "\x1c", "\x1f", "\x85", "\u0085",
		"\u00a0", "\u1680", "\u2000", "\u2028", "\u2029", "\u202f", "\u3000", "\u200b", "\ufeff",
		"\xff", "\xc3", "\xe2\x80", "\u65e5\u672c", "\u03a9\u03bc\u03ad\u03b3\u03b1", "\u03a3\u03a3",
	}
	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		var b strings.Builder
		for n := rnd.Intn(8); n >= 0; n-- {
			b.WriteString(pieces[rnd.Intn(len(pieces))])
		}
		s := b.String()
		if got, want := titleCase(s), regexTitleCase(s); got != want {
			t.Fatalf("titleCase(%q) = %q, want %q", s, got, want)
		}
	}
}

func BenchmarkTitleCase(b *testing.B) {
	s := "  Sample Text Item Number 42 With Extra Spacing  "
	for i := 0; i < b.N; i++ {
		titleCase(s)
	}
}

func BenchmarkRegexTitleCase(b *testing.B) {
	s := "  Sample Text Item Number 42 With Extra Spacing  "
	for i := 0; i < b.N; i++ {
		regexTitleCase(s)
	}
}
