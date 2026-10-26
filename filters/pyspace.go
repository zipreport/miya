package filters

import "unicode"

// isPySpace reports whether r is whitespace in Python's str.isspace sense:
// Go's unicode.IsSpace plus the separators U+001C-U+001F.
func isPySpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)
}

// pySpaceClass is the body of a regexp character class matching the same set
// as isPySpace (Python's str.isspace).
const pySpaceClass = `[\p{Z}\t-\r\x{1c}-\x{1f}\x{85}]`
