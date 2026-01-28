package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"
)

// =============================================================================
// COMPREHENSIVE RUNTIME EVALUATOR TESTS
// =============================================================================
// This file provides comprehensive test coverage for the runtime evaluation engine
// to improve coverage from 42.3% to target 70%+
// =============================================================================

// Test Expression Evaluation
func TestExpressionEvaluation(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "arithmetic expressions",
			template: "{{ (a + b) * c - d / e }}",
			data:     map[string]interface{}{"a": 5, "b": 3, "c": 2, "d": 10, "e": 2},
			expected: "11",
		},
		{
			name:     "string operations",
			template: "{{ name + ' ' + surname }}",
			data:     map[string]interface{}{"name": "John", "surname": "Doe"},
			expected: "John Doe",
		},
		{
			name:     "boolean expressions",
			template: "{{ (a > b) and (c < d) or (e == f) }}",
			data:     map[string]interface{}{"a": 5, "b": 3, "c": 2, "d": 4, "e": 1, "f": 2},
			expected: "true",
		},
		{
			name:     "comparison operations",
			template: "{{ a == b }},{{ c != d }},{{ e < f }},{{ g > h }}",
			data:     map[string]interface{}{"a": 5, "b": 5, "c": 3, "d": 4, "e": 1, "f": 2, "g": 6, "h": 4},
			expected: "true,true,true,true",
		},
		{
			name:     "nested attribute access",
			template: "{{ user.profile.settings.theme }}",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"settings": map[string]interface{}{
							"theme": "dark",
						},
					},
				},
			},
			expected: "dark",
		},
		{
			name:     "array indexing",
			template: "{{ items[0] }},{{ items[1] }},{{ items[-1] }}",
			data:     map[string]interface{}{"items": []interface{}{"first", "second", "third", "last"}},
			expected: "first,second,last",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

// Test Context Management
func TestContextManagement(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "variable scoping in loops",
			template: `{%- set outer = "outer" -%}
{%- for i in range(3) -%}
{%- set inner = "inner" + i|string -%}
{{ outer }}-{{ inner }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data:     map[string]interface{}{},
			expected: "outer-inner0,outer-inner1,outer-inner2",
		},
		{
			name: "variable scoping in loops",
			template: `{%- set counter = 0 -%}
{%- for i in range(3) -%}
{%- set counter = counter + 1 -%}
{%- endfor -%}
{{ counter }}`,
			data:     map[string]interface{}{},
			expected: "0", // Variables set in loops don't persist (scoped to loop)
		},
		{
			name: "context inheritance",
			template: `{{ parent }}
{%- with child = "child_value" -%}
  {{ parent }}-{{ child }}
{%- endwith -%}
{{ parent }}`,
			data:     map[string]interface{}{"parent": "parent_value"},
			expected: "parent_valueparent_value-child_valueparent_value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if strings.TrimSpace(result) != strings.TrimSpace(test.expected) {
				t.Errorf("Expected %q, got %q", strings.TrimSpace(test.expected), strings.TrimSpace(result))
			}
		})
	}
}

// Test Filter Application in Runtime
func TestRuntimeFilterApplication(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "dynamic filter application",
			template: "{{ value|default('fallback')|upper }}",
			data:     map[string]interface{}{},
			expected: "FALLBACK",
		},
		{
			name:     "filter with runtime arguments",
			template: "{{ text|truncate(max_length) }}",
			data:     map[string]interface{}{"text": "Hello World!", "max_length": 5},
			expected: "Hello...",
		},
		{
			name:     "conditional filter application",
			template: "{{ value|upper if apply_upper else value|lower }}",
			data:     map[string]interface{}{"value": "Hello", "apply_upper": true},
			expected: "HELLO",
		},
		{
			name:     "filter chain with runtime data",
			template: "{{ items|selectattr('active')|map('name')|join(separator) }}",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "John", "active": true},
					map[string]interface{}{"name": "Jane", "active": false},
					map[string]interface{}{"name": "Bob", "active": true},
				},
				"separator": " | ",
			},
			expected: "John | Bob",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

// Test Complex Control Flow Evaluation
func TestComplexControlFlowEvaluation(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "nested conditionals",
			template: `{%- if condition1 -%}
  {%- if condition2 -%}
    both-true
  {%- else -%}
    only-first-true
  {%- endif -%}
{%- else -%}
  {%- if condition2 -%}
    only-second-true
  {%- else -%}
    both-false
  {%- endif -%}
{%- endif -%}`,
			data:     map[string]interface{}{"condition1": true, "condition2": false},
			expected: "only-first-true",
		},
		{
			name: "conditional loops",
			template: `{%- for user in users if user.active -%}
{%- if user.role == 'admin' -%}
Admin: {{ user.name }}
{%- elif user.role == 'user' -%}
User: {{ user.name }}
{%- endif -%}
{%- endfor -%}`,
			data: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "John", "active": true, "role": "admin"},
					map[string]interface{}{"name": "Jane", "active": false, "role": "user"},
					map[string]interface{}{"name": "Bob", "active": true, "role": "user"},
				},
			},
			expected: "Admin: JohnUser: Bob",
		},
		{
			name: "loop with complex conditions",
			template: `{%- for item in items -%}
{%- if item.category == target_category and item.price < max_price -%}
{{ item.name }}: ${{ item.price }}
{%- endif -%}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "Book", "category": "education", "price": 20},
					map[string]interface{}{"name": "Laptop", "category": "electronics", "price": 1000},
					map[string]interface{}{"name": "Notebook", "category": "education", "price": 5},
				},
				"target_category": "education",
				"max_price":       15,
			},
			expected: "Notebook: $5\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if strings.TrimSpace(result) != strings.TrimSpace(test.expected) {
				t.Errorf("Expected %q, got %q", strings.TrimSpace(test.expected), strings.TrimSpace(result))
			}
		})
	}
}

// Test Runtime Error Handling
func TestRuntimeErrorHandling(t *testing.T) {
	env := miya.NewEnvironment(miya.WithStrictUndefined(true))

	tests := []struct {
		name        string
		template    string
		data        map[string]interface{}
		expectError bool
		errorMatch  string
	}{
		{
			name:        "undefined variable access",
			template:    "{{ undefined_variable }}",
			data:        map[string]interface{}{},
			expectError: true,
			errorMatch:  "undefined",
		},
		{
			name:        "invalid attribute access",
			template:    "{{ user.nonexistent }}",
			data:        map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
			expectError: true,
			errorMatch:  "undefined",
		},
		{
			// Out-of-bounds array access returns undefined (empty string) in Jinja2
			// This is intentional behavior, not an error
			name:        "invalid array index",
			template:    "{{ items[10] }}",
			data:        map[string]interface{}{"items": []interface{}{"a", "b", "c"}},
			expectError: false, // Returns silent undefined, not an error
		},
		{
			name:        "type mismatch in operations",
			template:    "{{ string_var + number_var }}",
			data:        map[string]interface{}{"string_var": "hello", "number_var": 42},
			expectError: true, // Should produce TypeError
			errorMatch:  "TypeError",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				if test.expectError {
					return // Parse error is acceptable
				}
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if test.expectError {
				if err == nil {
					t.Fatalf("Expected error but got result: %q", result)
				}
				if test.errorMatch != "" && !strings.Contains(err.Error(), test.errorMatch) {
					t.Errorf("Expected error containing %q, got: %v", test.errorMatch, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test Variable Resolution Performance
func TestVariableResolutionPerformance(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	// Create a deeply nested context
	deepData := make(map[string]interface{})
	current := deepData
	for i := 0; i < 10; i++ {
		nested := make(map[string]interface{})
		current["level"] = nested
		current = nested
	}
	current["value"] = "deep_value"

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "deep variable access",
			template: "{{ level.level.level.level.level.level.level.level.level.level.value }}",
			data:     deepData,
			expected: "deep_value",
		},
		{
			name:     "multiple variable lookups",
			template: "{{ a }}{{ b }}{{ c }}{{ d }}{{ e }}",
			data:     map[string]interface{}{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5"},
			expected: "12345",
		},
		{
			name:     "repeated variable access",
			template: "{{ user.name }}-{{ user.name }}-{{ user.name }}",
			data:     map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
			expected: "John-John-John",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

// Test Runtime Type Handling
func TestRuntimeTypeHandling(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "integer operations",
			template: "{{ a + b }},{{ a * b }},{{ (a / b)|int }}",
			data:     map[string]interface{}{"a": 10, "b": 3},
			expected: "13,30,3",
		},
		{
			name:     "float operations",
			template: "{{ (a + b)|round(2) }}",
			data:     map[string]interface{}{"a": 3.14159, "b": 2.71828},
			expected: "5.86",
		},
		{
			name:     "boolean operations",
			template: "{{ true and false }},{{ true or false }},{{ not true }}",
			data:     map[string]interface{}{},
			expected: "false,true,false",
		},
		{
			name:     "mixed type comparisons",
			template: "{{ 5 == '5' }},{{ [] is empty }},{{ {} is mapping }}",
			data:     map[string]interface{}{},
			expected: "false,true,true",
		},
		{
			name:     "nil/null handling",
			template: "{{ null_value is none }},{{ null_value|default('fallback') }}",
			data:     map[string]interface{}{"null_value": nil},
			expected: "true,fallback",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

// Test Runtime Template Caching
func TestRuntimeTemplateCaching(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	template := "{{ value|upper }}"
	data := map[string]interface{}{"value": "test"}

	// First render
	tmpl1, err := env.FromString(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result1, err := tmpl1.Render(miya.NewContextFrom(data))
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Second render of same template (should use cache)
	tmpl2, err := env.FromString(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result2, err := tmpl2.Render(miya.NewContextFrom(data))
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if result1 != result2 {
		t.Errorf("Results should be identical: %q vs %q", result1, result2)
	}

	if result1 != "TEST" {
		t.Errorf("Expected 'TEST', got %q", result1)
	}
}

// Test Complex Data Structure Evaluation
func TestComplexDataStructureEvaluation(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	complexData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"id":   1,
				"name": "John",
				"profile": map[string]interface{}{
					"age":   30,
					"email": "john@example.com",
					"tags":  []interface{}{"developer", "golang", "templates"},
				},
			},
			map[string]interface{}{
				"id":   2,
				"name": "Jane",
				"profile": map[string]interface{}{
					"age":   25,
					"email": "jane@example.com",
					"tags":  []interface{}{"designer", "ui", "ux"},
				},
			},
		},
		"config": map[string]interface{}{
			"site_name": "Test Site",
			"features":  map[string]interface{}{"search": true, "comments": false},
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "nested array access",
			template: "{{ users[0].profile.tags[1] }}",
			expected: "golang",
		},
		{
			name:     "deep object navigation",
			template: "{{ config.features.search }}",
			expected: "true",
		},
		{
			name: "complex iteration",
			template: `{%- for user in users -%}
{{ user.name }}: {{ user.profile.tags|join(', ') }}
{% endfor -%}`,
			expected: "John: developer, golang, templates\nJane: designer, ui, ux\n",
		},
		{
			name:     "mixed access patterns",
			template: "{{ users[0].name }} ({{ users[0].profile.age }}) - {{ config.site_name }}",
			expected: "John (30) - Test Site",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(complexData))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if strings.TrimSpace(result) != strings.TrimSpace(test.expected) {
				t.Errorf("Expected %q, got %q", strings.TrimSpace(test.expected), strings.TrimSpace(result))
			}
		})
	}
}
