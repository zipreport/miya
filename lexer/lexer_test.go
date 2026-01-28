package lexer

import (
	"reflect"
	"testing"
)

func TestLexerPlainText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: []TokenType{TokenText, TokenEOF},
		},
		{
			name:     "text with newlines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: []TokenType{TokenText, TokenEOF},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []TokenType{TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token %d: expected %v, got %v", i, tt.expected[i], tok.Type)
				}
			}
		})
	}
}

func TestLexerVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "simple variable",
			input: "{{ name }}",
			expected: []TokenType{
				TokenVarStart, TokenIdentifier, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "variable with filter",
			input: "{{ name|upper }}",
			expected: []TokenType{
				TokenVarStart, TokenIdentifier, TokenPipe, TokenIdentifier, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "variable with dot access",
			input: "{{ user.name }}",
			expected: []TokenType{
				TokenVarStart, TokenIdentifier, TokenDot, TokenIdentifier, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "variable with bracket access",
			input: "{{ items[0] }}",
			expected: []TokenType{
				TokenVarStart, TokenIdentifier, TokenLeftBracket, TokenInteger, TokenRightBracket, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "text and variable",
			input: "Hello {{ name }}!",
			expected: []TokenType{
				TokenText, TokenVarStart, TokenIdentifier, TokenVarEnd, TokenText, TokenEOF,
			},
		},
		{
			name:  "whitespace control",
			input: "{{- name -}}",
			expected: []TokenType{
				TokenVarStartTrim, TokenIdentifier, TokenVarEndTrim, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokenTypes := make([]TokenType, len(tokens))
			for i, tok := range tokens {
				tokenTypes[i] = tok.Type
			}

			if !reflect.DeepEqual(tokenTypes, tt.expected) {
				t.Errorf("expected tokens %v, got %v", tt.expected, tokenTypes)
				for i, tok := range tokens {
					t.Logf("  [%d] %s", i, tok)
				}
			}
		})
	}
}

func TestLexerBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "if block",
			input: "{% if condition %}",
			expected: []TokenType{
				TokenBlockStart, TokenIf, TokenIdentifier, TokenBlockEnd, TokenEOF,
			},
		},
		{
			name:  "for loop",
			input: "{% for item in items %}",
			expected: []TokenType{
				TokenBlockStart, TokenFor, TokenIdentifier, TokenIn, TokenIdentifier, TokenBlockEnd, TokenEOF,
			},
		},
		{
			name:  "block with whitespace control",
			input: "{%- if true -%}",
			expected: []TokenType{
				TokenBlockStartTrim, TokenIf, TokenTrue, TokenBlockEndTrim, TokenEOF,
			},
		},
		{
			name:  "set statement",
			input: "{% set x = 10 %}",
			expected: []TokenType{
				TokenBlockStart, TokenSet, TokenIdentifier, TokenAssign, TokenInteger, TokenBlockEnd, TokenEOF,
			},
		},
		{
			name:  "extends statement",
			input: "{% extends \"base.html\" %}",
			expected: []TokenType{
				TokenBlockStart, TokenExtends, TokenString, TokenBlockEnd, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokenTypes := make([]TokenType, len(tokens))
			for i, tok := range tokens {
				tokenTypes[i] = tok.Type
			}

			if !reflect.DeepEqual(tokenTypes, tt.expected) {
				t.Errorf("expected tokens %v, got %v", tt.expected, tokenTypes)
				for i, tok := range tokens {
					t.Logf("  [%d] %s", i, tok)
				}
			}
		})
	}
}

func TestLexerComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple comment",
			input:    "{# This is a comment #}",
			expected: []TokenType{TokenEOF}, // Comments are skipped
		},
		{
			name:     "comment with text",
			input:    "Before{# comment #}After",
			expected: []TokenType{TokenText, TokenText, TokenEOF},
		},
		{
			name:     "multiline comment",
			input:    "{# This is\na multiline\ncomment #}",
			expected: []TokenType{TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokenTypes := make([]TokenType, len(tokens))
			for i, tok := range tokens {
				tokenTypes[i] = tok.Type
			}

			if !reflect.DeepEqual(tokenTypes, tt.expected) {
				t.Errorf("expected tokens %v, got %v", tt.expected, tokenTypes)
			}
		})
	}
}

func TestLexerExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "arithmetic expression",
			input: "{{ 1 + 2 * 3 }}",
			expected: []TokenType{
				TokenVarStart, TokenInteger, TokenPlus, TokenInteger, TokenMultiply, TokenInteger, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "comparison",
			input: "{% if x > 5 %}",
			expected: []TokenType{
				TokenBlockStart, TokenIf, TokenIdentifier, TokenGreater, TokenInteger, TokenBlockEnd, TokenEOF,
			},
		},
		{
			name:  "logical operators",
			input: "{% if x and not y %}",
			expected: []TokenType{
				TokenBlockStart, TokenIf, TokenIdentifier, TokenAnd, TokenNot, TokenIdentifier, TokenBlockEnd, TokenEOF,
			},
		},
		{
			name:  "string literals",
			input: `{{ "hello" + 'world' }}`,
			expected: []TokenType{
				TokenVarStart, TokenString, TokenPlus, TokenString, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "float numbers",
			input: "{{ 3.14 + 2.5e-3 }}",
			expected: []TokenType{
				TokenVarStart, TokenFloat, TokenPlus, TokenFloat, TokenVarEnd, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokenTypes := make([]TokenType, len(tokens))
			for i, tok := range tokens {
				tokenTypes[i] = tok.Type
			}

			if !reflect.DeepEqual(tokenTypes, tt.expected) {
				t.Errorf("expected tokens %v, got %v", tt.expected, tokenTypes)
				for i, tok := range tokens {
					t.Logf("  [%d] %s", i, tok)
				}
			}
		})
	}
}

func TestLexerCustomDelimiters(t *testing.T) {
	config := &LexerConfig{
		VarStartString:     "[[",
		VarEndString:       "]]",
		BlockStartString:   "<%",
		BlockEndString:     "%>",
		CommentStartString: "<#",
		CommentEndString:   "#>",
	}

	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "custom variable delimiters",
			input: "[[ name ]]",
			expected: []TokenType{
				TokenVarStart, TokenIdentifier, TokenVarEnd, TokenEOF,
			},
		},
		{
			name:  "custom block delimiters",
			input: "<% if true %>",
			expected: []TokenType{
				TokenBlockStart, TokenIf, TokenTrue, TokenBlockEnd, TokenEOF,
			},
		},
		{
			name:     "custom comment delimiters",
			input:    "<# comment #>",
			expected: []TokenType{TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, config)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tokenTypes := make([]TokenType, len(tokens))
			for i, tok := range tokens {
				tokenTypes[i] = tok.Type
			}

			if !reflect.DeepEqual(tokenTypes, tt.expected) {
				t.Errorf("expected tokens %v, got %v", tt.expected, tokenTypes)
			}
		})
	}
}

func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed variable",
			input: "{{ name",
		},
		{
			name:  "unclosed block",
			input: "{% if true",
		},
		{
			name:  "unclosed comment",
			input: "{# comment",
		},
		{
			name:  "unclosed string",
			input: `{{ "unterminated }}`,
		},
		{
			name:  "invalid character",
			input: "{% ! %}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, nil)
			_, err := l.Tokenize()
			if err == nil {
				t.Errorf("expected error for input %q, got none", tt.input)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	tok := &Token{
		Type:   TokenIdentifier,
		Value:  "name",
		Line:   1,
		Column: 5,
	}

	expected := `IDENTIFIER("name") at 1:5`
	if tok.String() != expected {
		t.Errorf("expected %q, got %q", expected, tok.String())
	}

	tok2 := &Token{
		Type:   TokenVarStart,
		Line:   2,
		Column: 1,
	}

	expected2 := `VAR_START at 2:1`
	if tok2.String() != expected2 {
		t.Errorf("expected %q, got %q", expected2, tok2.String())
	}
}
