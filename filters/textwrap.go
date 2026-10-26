package filters

import (
	"strings"
	"unicode"
)

const textwrapWhitespace = " \t\n\v\f\r"

// pySplitlines mirrors Python 3.14 str.splitlines() (no keepends).
func pySplitlines(s string) []string {
	var lines []string
	rs := []rune(s)
	start := 0
	for i := 0; i < len(rs); i++ {
		switch rs[i] {
		case '\n', '\v', '\f', '\x1c', '\x1d', '\x1e', '\x85', ' ', ' ':
			lines = append(lines, string(rs[start:i]))
			start = i + 1
		case '\r':
			lines = append(lines, string(rs[start:i]))
			if i+1 < len(rs) && rs[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	if start < len(rs) {
		lines = append(lines, string(rs[start:]))
	}
	return lines
}

// isTextwrapLetter mirrors the regex class [^\d\W] used by
// textwrap.TextWrapper.wordsep_re (Python 3.14).
func isTextwrapLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || unicode.Is(unicode.Nl, r) || unicode.Is(unicode.No, r)
}

func isTextwrapWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

func isTextwrapWordPunct(r rune) bool {
	return isTextwrapWord(r) || strings.ContainsRune(`!"'&.,?`, r)
}

// pyIsBlank mirrors `chunk.strip() == ”` (str.isspace for every rune).
func pyIsBlank(chunk string) bool {
	for _, r := range chunk {
		if !(unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)) {
			return false
		}
	}
	return true
}

// emDashEnd returns the end of a run of two or more '-' starting at i that is
// followed by a word rune, or -1.
func emDashEnd(rs []rune, i int) int {
	j := i
	for j < len(rs) && rs[j] == '-' {
		j++
	}
	if j-i >= 2 && j < len(rs) && isTextwrapWord(rs[j]) {
		return j
	}
	return -1
}

// hyphenSplitAt reports whether the '-' at index k ends a chunk, mirroring the
// hyphenated-word branch of wordsep_re (Python 3.14).
func hyphenSplitAt(rs []rune, k int) bool {
	letter := func(i int) bool { return i >= 0 && i < len(rs) && isTextwrapLetter(rs[i]) }
	before := (letter(k-2) && letter(k-1)) || (letter(k-3) && k-2 >= 0 && rs[k-2] == '-' && letter(k-1))
	if !before || !letter(k+1) {
		return false
	}
	return letter(k+2) || (k+2 < len(rs) && rs[k+2] == '-' && letter(k+3))
}

// textwrapChunks mirrors TextWrapper._split with wordsep_re (Python 3.14).
func textwrapChunks(line string, breakOnHyphens bool) []string {
	rs := []rune(line)
	var chunks []string
	i := 0
	for i < len(rs) {
		if strings.ContainsRune(textwrapWhitespace, rs[i]) {
			j := i
			for j < len(rs) && strings.ContainsRune(textwrapWhitespace, rs[j]) {
				j++
			}
			chunks = append(chunks, string(rs[i:j]))
			i = j
			continue
		}
		end := i
		for end < len(rs) && !strings.ContainsRune(textwrapWhitespace, rs[end]) {
			end++
		}
		if !breakOnHyphens {
			chunks = append(chunks, string(rs[i:end]))
			i = end
			continue
		}
		if i > 0 && isTextwrapWordPunct(rs[i-1]) {
			if j := emDashEnd(rs, i); j >= 0 {
				chunks = append(chunks, string(rs[i:j]))
				i = j
				continue
			}
		}
		stop := end
		for e := i + 1; e < end; e++ {
			if rs[e] == '-' && hyphenSplitAt(rs, e) {
				stop = e + 1
				break
			}
			if isTextwrapWordPunct(rs[e-1]) && emDashEnd(rs, e) >= 0 && rs[e] == '-' {
				stop = e
				break
			}
		}
		chunks = append(chunks, string(rs[i:stop]))
		i = stop
	}
	return chunks
}

// textwrapLines mirrors TextWrapper._wrap_chunks and _handle_long_word with
// drop_whitespace=True and no indents (Python 3.14). Lengths are in runes.
func textwrapLines(line string, width int, breakLongWords, breakOnHyphens bool) []string {
	chunks := textwrapChunks(line, breakOnHyphens)
	n := func(s string) int { return len([]rune(s)) }
	var lines []string
	h := 0 // chunks[h] is Python's chunks[-1]
	for h < len(chunks) {
		var cur []string
		curLen := 0
		if pyIsBlank(chunks[h]) && len(lines) > 0 {
			h++
		}
		for h < len(chunks) {
			l := n(chunks[h])
			if curLen+l > width {
				break
			}
			cur = append(cur, chunks[h])
			curLen += l
			h++
		}
		if h < len(chunks) && n(chunks[h]) > width {
			spaceLeft := width - curLen
			if width < 1 {
				spaceLeft = 1
			}
			if breakLongWords && spaceLeft > 0 {
				rc := []rune(chunks[h])
				end := spaceLeft
				if breakOnHyphens && len(rc) > spaceLeft {
					hyphen := -1
					for k := spaceLeft - 1; k >= 0; k-- {
						if rc[k] == '-' {
							hyphen = k
							break
						}
					}
					if hyphen > 0 {
						for _, c := range rc[:hyphen] {
							if c != '-' {
								end = hyphen + 1
								break
							}
						}
					}
				}
				cur = append(cur, string(rc[:end]))
				chunks[h] = string(rc[end:])
			} else if len(cur) == 0 {
				cur = append(cur, chunks[h])
				h++
			}
		}
		if len(cur) > 0 && pyIsBlank(cur[len(cur)-1]) {
			cur = cur[:len(cur)-1]
		}
		if len(cur) > 0 {
			lines = append(lines, strings.Join(cur, ""))
		}
	}
	return lines
}
