package miya_test

import (
	"strings"
	"testing"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
)

// Expected outputs are Jinja2 3.1.6's, except that booleans print as true/false.
func TestDictLiterals(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{"set and read one pair", "{% set d = {'k': 'v'} %}[{{ d.k }}|{{ d|length }}]", nil, "[v|1]"},
		{"subscript", "{{ {'a': 1}['a'] }}", nil, "1"},
		{"items loop", "{% for k, v in {'x': 1}.items() %}{{ k }}={{ v }}{% endfor %}", nil, "x=1"},
		{"second pair", "{{ {'a': 1, 'b': 2}.b }}", nil, "2"},
		{"third pair", "{{ {'a': 1, 'b': 2, 'c': 3}.c }}", nil, "3"},
		{"length of two", "{{ {'a': 1, 'b': 2}|length }}", nil, "2"},
		{"trailing comma after one pair", "{{ {'a': 1,}|length }}", nil, "1"},
		{"trailing comma after two pairs", "{{ {'a': 1, 'b': 2,}|length }}", nil, "2"},
		{"empty", "{{ {}|length }}", nil, "0"},
		{"variable key", "{% set k = 'z' %}{{ {k: 1}.z }}", nil, "1"},
		{"expression key", "{{ {'a' ~ 'b': 1}.ab }}", nil, "1"},
		{"variable value", "{{ {'a': x}.a }}", map[string]interface{}{"x": 5}, "5"},
		{"integer key", "{{ {1: 'one'}[1] }}", nil, "one"},
		{"later duplicate wins", "{{ {'a': 1, 'a': 2}.a }}", nil, "2"},
		{"list value", "{{ {'x': [1, 2]}.x|length }}", nil, "2"},
		{"nested in a set tag", "{% set d = {'a': {'b': 1}} %}{{ d.a.b }}", nil, "1"},
		{"closing braces touch the tag end", "{{ [{'a': 1}][0].a}}", nil, "1"},
		{"string value containing the tag end", "{{ {'a': '}}'}.a }}", nil, "}}"},
		{"dictsort", "{{ {'b': 1, 'a': 2}|dictsort|first|first }}", nil, "a"},
		{"in", "{{ 'a' in {'a': 1} }}", nil, "true"},
		{"fresh per loop iteration", "{% for i in [1, 2] %}{% set d = {'i': i} %}{{ d.i }}{% endfor %}", nil, "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment()
			ctx := miya.NewContext()
			for k, v := range tt.data {
				ctx.Set(k, v)
			}
			result, err := env.RenderString(tt.template, ctx)
			if err != nil {
				t.Fatalf("render %q: %v", tt.template, err)
			}
			if result != tt.expected {
				t.Errorf("%q: expected %q, got %q", tt.template, tt.expected, result)
			}
		})
	}
}

func TestDictLiteralEvaluationOrder(t *testing.T) {
	var calls []string
	env := miya.NewEnvironment()
	env.AddGlobal("rec", func(args ...interface{}) (interface{}, error) {
		calls = append(calls, args[0].(string))
		return args[0], nil
	})
	result, err := env.RenderString("{{ {rec('k1'): rec('v1'), rec('k2'): rec('v2')}|length }}", miya.NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if result != "2" {
		t.Errorf("expected %q, got %q", "2", result)
	}
	if got := strings.Join(calls, ","); got != "k1,v1,k2,v2" {
		t.Errorf("expected evaluation order %q, got %q", "k1,v1,k2,v2", got)
	}
}

// A dict literal must build a new map on every evaluation: the parsed
// template is shared by every render, so a shared map would carry writes
// from one render into the next (and race between concurrent renders).
func TestDictLiteralFreshPerRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"empty", "{% set d = {} %}{{ d|length }}{{ poke(d) }}", "0"},
		{"one pair", "{% set d = {'a': 1} %}{{ d|length }}{{ poke(d) }}", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment()
			env.AddGlobal("poke", func(args ...interface{}) (interface{}, error) {
				args[0].(map[string]interface{})["poked"] = true
				return "", nil
			})
			tmpl, err := env.FromString(tt.template)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				result, err := tmpl.Render(miya.NewContext())
				if err != nil {
					t.Fatalf("render %d: %v", i, err)
				}
				if result != tt.expected {
					t.Errorf("render %d: expected %q, got %q", i, tt.expected, result)
				}
			}
		})
	}
}

// include ... with <expr> is a miya extension: the dict's pairs become
// variables in the included template.
func TestIncludeWithDictLiteral(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("part.html", "[{{ a }}]")
	sl.AddTemplate("main.html", "{% include 'part.html' with {'a': 7} %}")
	env := miya.NewEnvironment(miya.WithLoader(sl))
	tmpl, err := env.GetTemplate("main.html")
	if err != nil {
		t.Fatal(err)
	}
	result, err := tmpl.Render(miya.NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if result != "[7]" {
		t.Errorf("expected %q, got %q", "[7]", result)
	}
}

func TestDictLiteralSyntaxErrors(t *testing.T) {
	tests := []struct {
		template string
		contains string
	}{
		{"{{ {'a' 1} }}", "expected ':' after dictionary key"},
		{"{{ {'a': } }}", "unexpected token in expression"},
		{"{{ {,} }}", "unexpected token in expression"},
		{"{{ {'a': 1,,} }}", "unexpected token in expression"},
		{"{{ {'a': 1 }", "unexpected EOF"},
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%q panicked: %v", tt.template, r)
				}
			}()
			env := miya.NewEnvironment()
			_, err := env.FromString(tt.template)
			if err == nil {
				t.Fatalf("%q: expected an error", tt.template)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("%q: expected error containing %q, got %v", tt.template, tt.contains, err)
			}
		})
	}
}
