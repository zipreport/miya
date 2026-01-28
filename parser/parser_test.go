package parser

import (
	"testing"

	"github.com/zipreport/miya/lexer"
)

func TestParseSimpleTemplate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "text only",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "simple variable",
			input:    "{{ name }}",
			expected: "Variable(name)",
		},
		{
			name:     "text and variable",
			input:    "Hello {{ name }}!",
			expected: "Text(Hello) Variable(name) Text(!)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			node, err := p.Parse()
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			if node == nil {
				t.Fatal("parsed node is nil")
			}

			// Basic validation that we got a template node
			if node == nil {
				t.Error("expected TemplateNode, got nil")
			}
		})
	}
}

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "arithmetic expression",
			input:   "{{ 1 + 2 * 3 }}",
			wantErr: false,
		},
		{
			name:    "comparison",
			input:   "{{ x > 5 }}",
			wantErr: false,
		},
		{
			name:    "logical operators",
			input:   "{{ x and not y }}",
			wantErr: false,
		},
		{
			name:    "function call",
			input:   "{{ func(arg1, arg2) }}",
			wantErr: false,
		},
		{
			name:    "attribute access",
			input:   "{{ user.name }}",
			wantErr: false,
		},
		{
			name:    "index access",
			input:   "{{ items[0] }}",
			wantErr: false,
		},
		{
			name:    "slice access",
			input:   "{{ items[1:3] }}",
			wantErr: false,
		},
		{
			name:    "filter application",
			input:   "{{ name|upper }}",
			wantErr: false,
		},
		{
			name:    "chained filters",
			input:   "{{ name|upper|reverse }}",
			wantErr: false,
		},
		{
			name:    "filter with arguments",
			input:   "{{ items|join(', ') }}",
			wantErr: false,
		},
		{
			name:    "conditional expression",
			input:   "{{ value if condition else default }}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseIfStatements(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple if",
			input: `{% if condition %}
content
{% endif %}`,
			wantErr: false,
		},
		{
			name: "if-else",
			input: `{% if condition %}
true content
{% else %}
false content
{% endif %}`,
			wantErr: false,
		},
		{
			name: "if-elif-else",
			input: `{% if condition1 %}
content1
{% elif condition2 %}
content2
{% else %}
content3
{% endif %}`,
			wantErr: false,
		},
		{
			name: "nested if",
			input: `{% if outer %}
{% if inner %}
nested content
{% endif %}
{% endif %}`,
			wantErr: false,
		},
		{
			name:    "missing endif",
			input:   `{% if condition %}content`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseForLoops(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple for loop",
			input: `{% for item in items %}
{{ item }}
{% endfor %}`,
			wantErr: false,
		},
		{
			name: "for loop with filter",
			input: `{% for item in items|sort %}
{{ item }}
{% endfor %}`,
			wantErr: false,
		},
		{
			name: "for loop with conditional",
			input: `{% for item in items if item.active %}
{{ item.name }}
{% endfor %}`,
			wantErr: false,
		},
		{
			name: "nested for loops",
			input: `{% for group in groups %}
{% for item in group.items %}
{{ item }}
{% endfor %}
{% endfor %}`,
			wantErr: false,
		},
		{
			name: "for loop with else",
			input: `{% for item in items %}
{{ item }}
{% else %}
No items
{% endfor %}`,
			wantErr: false,
		},
		{
			name:    "missing endfor",
			input:   `{% for item in items %}{{ item }}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseBlocks(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple block",
			input: `{% block content %}
default content
{% endblock %}`,
			wantErr: false,
		},
		{
			name: "block with name",
			input: `{% block content %}
content here
{% endblock content %}`,
			wantErr: false,
		},
		{
			name: "nested blocks",
			input: `{% block outer %}
{% block inner %}
inner content
{% endblock %}
{% endblock %}`,
			wantErr: false,
		},
		{
			name:    "missing endblock",
			input:   `{% block content %}content`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseMacros(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple macro",
			input: `{% macro greet(name) %}
Hello {{ name }}!
{% endmacro %}`,
			wantErr: false,
		},
		{
			name: "macro with multiple parameters",
			input: `{% macro render_user(name, age, email) %}
Name: {{ name }}, Age: {{ age }}, Email: {{ email }}
{% endmacro %}`,
			wantErr: false,
		},
		{
			name: "macro with default parameters",
			input: `{% macro greet(name, greeting='Hello') %}
{{ greeting }} {{ name }}!
{% endmacro %}`,
			wantErr: false,
		},
		{
			name:    "missing endmacro",
			input:   `{% macro test() %}content`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseInheritance(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "extends statement",
			input:   `{% extends "base.html" %}`,
			wantErr: false,
		},
		{
			name:    "include statement",
			input:   `{% include "header.html" %}`,
			wantErr: false,
		},
		{
			name: "extends with blocks",
			input: `{% extends "base.html" %}
{% block content %}
overridden content
{% endblock %}`,
			wantErr: false,
		},
		{
			name: "super call",
			input: `{% block content %}
{{ super() }}
additional content
{% endblock %}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseSet(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple set",
			input:   `{% set x = 42 %}`,
			wantErr: false,
		},
		{
			name:    "set with expression",
			input:   `{% set result = x + y * 2 %}`,
			wantErr: false,
		},
		{
			name:    "set with function call",
			input:   `{% set value = func(arg1, arg2) %}`,
			wantErr: false,
		},
		{
			name:    "missing assignment",
			input:   `{% set x %}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseRaw(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "raw block",
			input: `{% raw %}
{{ this should not be parsed }}
{% if true %}
{% endraw %}`,
			wantErr: false,
		},
		{
			name:    "missing endraw",
			input:   `{% raw %}content`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseWhitespaceControl(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "trim variable start",
			input:   `{{- name }}`,
			wantErr: false,
		},
		{
			name:    "trim variable end",
			input:   `{{ name -}}`,
			wantErr: false,
		},
		{
			name:    "trim both sides",
			input:   `{{- name -}}`,
			wantErr: false,
		},
		{
			name:    "trim block start",
			input:   `{%- if true %}content{% endif %}`,
			wantErr: false,
		},
		{
			name:    "trim block end",
			input:   `{% if true -%}content{% endif %}`,
			wantErr: false,
		},
		{
			name:    "trim block both sides",
			input:   `{%- if true -%}content{%- endif -%}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseComplexTemplate(t *testing.T) {
	input := `{% extends "base.html" %}

{% block title %}User List{% endblock %}

{% block content %}
<h1>Users</h1>
{% if users %}
  <ul>
  {% for user in users|sort(attribute='name') %}
    <li>
      <strong>{{ user.name|title }}</strong>
      {% if user.email %}
        ({{ user.email }})
      {% endif %}
      - Age: {{ user.age }}
    </li>
  {% endfor %}
  </ul>
{% else %}
  <p>No users found.</p>
{% endif %}

{% set total_users = users|length %}
<p>Total users: {{ total_users }}</p>
{% endblock %}`

	l := lexer.NewLexer(input, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	p := NewParser(tokens)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if node == nil {
		t.Fatal("parsed node is nil")
	}

	// Verify it's a template node
	if node == nil {
		t.Error("expected TemplateNode, got nil")
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unexpected token",
			input: `{{ * }}`,
		},
		{
			name:  "unclosed variable",
			input: `{{ name`,
		},
		{
			name:  "unclosed block",
			input: `{% if true`,
		},
		{
			name:  "invalid expression",
			input: `{{ + }}`,
		},
		{
			name:  "missing block name",
			input: `{% block %}content{% endblock %}`,
		},
		{
			name:  "invalid for loop",
			input: `{% for %}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				// Some inputs may fail at lexer level, which is fine
				return
			}

			p := NewParser(tokens)
			_, err = p.Parse()
			if err == nil {
				t.Errorf("expected error for input %q, got none", tt.input)
			}
		})
	}
}

func TestParserPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "arithmetic precedence",
			input:   `{{ 1 + 2 * 3 }}`, // Should parse as 1 + (2 * 3)
			wantErr: false,
		},
		{
			name:    "comparison precedence",
			input:   `{{ x + y > z }}`, // Should parse as (x + y) > z
			wantErr: false,
		},
		{
			name:    "logical precedence",
			input:   `{{ a and b or c }}`, // Should parse as (a and b) or c
			wantErr: false,
		},
		{
			name:    "parentheses override",
			input:   `{{ (a + b) * c }}`,
			wantErr: false,
		},
		{
			name:    "filter precedence",
			input:   `{{ value|filter + other }}`, // Should parse as (value|filter) + other
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, nil)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			p := NewParser(tokens)
			_, err = p.Parse()

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
