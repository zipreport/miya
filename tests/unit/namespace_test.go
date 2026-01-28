package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestNamespaceFunction(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "BasicNamespace",
			template: `{% set ns = namespace() %}{% set ns.value = 42 %}{{ ns.value }}`,
			expected: "42",
		},
		{
			name:     "NamespaceWithInitialValues",
			template: `{% set ns = namespace(count=0) %}{% set ns.count = ns.count + 1 %}{{ ns.count }}`,
			expected: "1",
		},
		{
			name:     "NamespaceInLoop",
			template: `{% set ns = namespace(count=0) %}{% for i in range(3) %}{% set ns.count = ns.count + 1 %}{% endfor %}{{ ns.count }}`,
			expected: "3",
		},
		{
			name:     "NamespaceAcrossScopes",
			template: `{% set ns = namespace(value=0) %}{% if true %}{% set ns.value = 10 %}{% endif %}{{ ns.value }}`,
			expected: "10",
		},
		{
			name:     "MultipleNamespaces",
			template: `{% set ns1 = namespace(a=1) %}{% set ns2 = namespace(a=2) %}{{ ns1.a }},{{ ns2.a }}`,
			expected: "1,2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContext()
			result, err := env.RenderString(tc.template, ctx)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestRangeFunction(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "RangeStop",
			template: `{% for i in range(5) %}{{ i }}{% endfor %}`,
			expected: "01234",
		},
		{
			name:     "RangeStartStop",
			template: `{% for i in range(2, 5) %}{{ i }}{% endfor %}`,
			expected: "234",
		},
		{
			name:     "RangeStartStopStep",
			template: `{% for i in range(0, 10, 2) %}{{ i }}{% endfor %}`,
			expected: "02468",
		},
		{
			name:     "RangeNegativeStep",
			template: `{% for i in range(5, 0, -1) %}{{ i }}{% endfor %}`,
			expected: "54321",
		},
		{
			name:     "RangeEmpty",
			template: `{% for i in range(0) %}{{ i }}{% endfor %}empty`,
			expected: "empty",
		},
		{
			name:     "RangeInExpression",
			template: `{{ range(3)|length }}`,
			expected: "3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContext()
			result, err := env.RenderString(tc.template, ctx)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestDictFunction(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "EmptyDict",
			template: `{% set d = dict() %}{{ d }}`,
			expected: "map[]",
		},
		{
			name:     "DictWithKwargs",
			template: `{% set d = dict(a=1, b=2) %}{{ d.a }},{{ d.b }}`,
			expected: "1,2",
		},
		{
			name:     "DictAccess",
			template: `{% set d = dict(name="Alice") %}{{ d.name }}`,
			expected: "Alice",
		},
		{
			name:     "DictModification",
			template: `{% set d = dict() %}{% set d.key = "value" %}{{ d.key }}`,
			expected: "value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContext()
			result, err := env.RenderString(tc.template, ctx)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestNamespaceWithComplexScoping(t *testing.T) {
	env := miya.NewEnvironment()

	// Test namespace persisting across complex scopes
	template := `
{%- set ns = namespace(count=0, items=[]) -%}
{%- for i in range(3) -%}
  {%- if i > 0 -%}
    {%- set ns.count = ns.count + i -%}
    {%- set ns.items = ns.items + [i] -%}
  {%- endif -%}
{%- endfor -%}
Count: {{ ns.count }}, Items: {{ ns.items }}
`

	ctx := miya.NewContext()
	result, err := env.RenderString(template, ctx)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "Count: 3, Items: [1 2]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
