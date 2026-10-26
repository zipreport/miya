package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/lexer"
)

// Expected outputs and messages are Jinja2 3.1.6's: "}}" (or any end
// delimiter) only ends a tag when every ( [ { opened in it is closed.
func TestLexerBracketBalancing(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"nested dict in a print tag", "{{ {'a': {'b': 2}}.a.b }}", "2"},
		{"dict inside a list in a print tag", "{{ {'a': [1, {'b': 3}]}.a[1].b }}", "3"},
		{"nested dict in a set tag", "{% set d = {'a': {'b': 1}} %}{{ d.a.b }}", "1"},
		{"closing braces touch the tag end", "{{ [{'a': 1}][0].a}}", "1"},
		{"string value containing the tag end", "{{ {'a': '}}'}.a }}", "}}"},
		{"brackets in a for tag", "{% for i in [[1], [2]] %}{{ i[0] }}{% endfor %}", "12"},
		{"raw block with unbalanced brackets", "{% raw %}{{ f( }}{% endraw %}", "{{ f( }}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := miya.NewEnvironment().RenderString(tt.template, miya.NewContext())
			if err != nil {
				t.Fatalf("render %q: %v", tt.template, err)
			}
			if result != tt.expected {
				t.Errorf("%q: expected %q, got %q", tt.template, tt.expected, result)
			}
		})
	}
}

// Custom delimiters whose end is a single closing bracket: only balancing
// tells the delimiter apart from the dict's own "}".
func TestLexerBracketBalancingCustomDelimiters(t *testing.T) {
	tests := []struct {
		name                 string
		varStart, varEnd     string
		blockStart, blockEnd string
		template, expected   string
	}{
		{"block end is }", "{{", "}}", "{%", "}", "{% set d = {'a': 1} }{{ d.a }}", "1"},
		{"variable end is }", "${", "}", "<%", "%>", "${ {'a': 3}.a }", "3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment()
			env.SetDelimiters(tt.varStart, tt.varEnd, tt.blockStart, tt.blockEnd)
			result, err := env.RenderString(tt.template, miya.NewContext())
			if err != nil {
				t.Fatalf("render %q: %v", tt.template, err)
			}
			if result != tt.expected {
				t.Errorf("%q: expected %q, got %q", tt.template, tt.expected, result)
			}
		})
	}
}

func TestLexerBracketErrors(t *testing.T) {
	tests := []struct {
		name     string
		template string
		contains string
	}{
		{"unclosed paren in a print tag", "{{ (1 }}", "unexpected '}', expected ')'"},
		{"unclosed bracket in a print tag", "{{ x[0 }}", "unexpected '}', expected ']'"},
		{"stray closer in a print tag", "{{ x) }}", "unexpected ')' at line 1, column 6"},
		{"typo: one brace closes a print tag", "{{ x }<div>{{ y }}</div>{{ z }}", "unexpected '}' at line 1, column 7"},
		{"unclosed dict before the tag end", "{{ {'a': 1 }}<p>{{ y }}</p>{{ z }}", "unexpected '}' at line 1, column 14"},
		{"unclosed paren in a block tag", "{% set x = (1 %}", "unexpected '}', expected ')'"},
		{"stray closer in a block tag", "{% set x = 1) %}x", "unexpected ')' at line 1, column 14"},
		{"unclosed paren in an if tag", "{% if (1 %}y{% endif %}", "unexpected '}', expected ')'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := miya.NewEnvironment().FromString(tt.template)
			if err == nil {
				t.Fatalf("%q: expected an error", tt.template)
			}
			if !strings.Contains(err.Error(), "lexer error") || !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("%q: expected a lexer error containing %q, got %v", tt.template, tt.contains, err)
			}
		})
	}
}

// Inside an open bracket, "-}}" and "-%}" are a minus followed by closers,
// not trim markers. Lexed directly, so the cases do not depend on how the
// Environment treats trim markers.
func TestLexerTrimMarkerInsideBrackets(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		contains string
	}{
		{"minus before }} in a print-tag dict", "{{ {'a': 1 -}} }}", "unexpected '}' at line 1, column 15"},
		{"minus before }} in a block-tag dict", "{% set d = {'a': 1 -}} %}", "unexpected '}' at line 1, column 23"},
		{"minus before %} in a block-tag paren", "{% set x = (1 -%}", "unexpected '}', expected ')'"},
		{"plus before %} in a block-tag paren", "{% set x = (1 +%}", "unexpected '}', expected ')'"},
		{"minus before %} in a print-tag paren", "{{ (a -%} ) }}", "unexpected '}', expected ')'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%q panicked: %v", tt.src, r)
				}
			}()
			tokens, err := lexer.NewLexer(tt.src, nil).Tokenize()
			for i, tok := range tokens {
				if tok == nil {
					t.Fatalf("token %d of %q is nil", i, tt.src)
				}
			}
			if err == nil || !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("%q: expected a lexer error containing %q, got %v", tt.src, tt.contains, err)
			}
		})
	}
}

// An open bracket is reported where it was opened: at the end of the
// template, and at once when the tag's own end delimiter has no closing
// bracket in it (e.g. "%>"), so the lexer never runs on through the body.
// Only the line is pinned: columns after line 1 carry an old off-by-one.
func TestLexerUnclosedBracketPosition(t *testing.T) {
	body := strings.Repeat("It's line text\n", 300)
	tests := []struct {
		name                 string
		varStart, varEnd     string
		blockStart, blockEnd string
		template, contains   string
	}{
		{"end of template in a print tag", "{{", "}}", "{%", "%}", "a\n{{ (1", "unexpected end of template, expected ')' (opened at line 2, column "},
		{"block end %> with an unclosed paren", "${", "}", "<%", "%>", "a\n<% if (x %>\n" + body + "<% endif %>", "unexpected '%>', expected ')' (opened at line 2, column "},
		{"block end %> with an unclosed call", "${", "}", "<%", "%>", "a\n<% set y = f(x %>\n" + body, "unexpected '%>', expected ')' (opened at line 2, column "},
		{"variable end >> with an unclosed bracket", "<<", ">>", "{%", "%}", "a\n<< x[0 >>\n" + body, "unexpected '>>', expected ']' (opened at line 2, column "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment()
			env.SetDelimiters(tt.varStart, tt.varEnd, tt.blockStart, tt.blockEnd)
			_, err := env.FromString(tt.template)
			if err == nil || !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("expected a lexer error containing %q, got %v", tt.contains, err)
			}
		})
	}
}
