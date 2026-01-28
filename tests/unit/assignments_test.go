package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"
)

// =============================================================================
// ASSIGNMENT TESTS - CONSOLIDATED
// =============================================================================
// This file consolidates all assignment-related tests from multiple individual files:
// - advanced_assignment_test.go
// - assignment_operators_test.go
// - set_statement_test.go
// =============================================================================

// Basic Set Statement Tests
func TestSetStatement(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple variable assignment",
			template: "{% set name = 'John' %}{{ name }}",
			data:     map[string]interface{}{},
			expected: "John",
		},
		{
			name:     "expression assignment",
			template: "{% set result = value * 2 %}{{ result }}",
			data:     map[string]interface{}{"value": 21},
			expected: "42",
		},
		{
			name:     "string concatenation assignment",
			template: "{% set greeting = 'Hello ' + name %}{{ greeting }}",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "filter assignment",
			template: "{% set upper_name = name|upper %}{{ upper_name }}",
			data:     map[string]interface{}{"name": "john"},
			expected: "JOHN",
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

// Block Assignment Tests
func TestBlockAssignment(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "simple block assignment",
			template: `{%- set content -%}
<h1>Title</h1>
<p>Content</p>
{%- endset -%}
{{ content }}`,
			data:     map[string]interface{}{},
			expected: "<h1>Title</h1>\n<p>Content</p>",
		},
		{
			name: "block assignment with variables",
			template: `{%- set greeting -%}
Hello, {{ name }}! Welcome to {{ site }}.
{%- endset -%}
{{ greeting }}`,
			data:     map[string]interface{}{"name": "John", "site": "our website"},
			expected: "Hello, John! Welcome to our website.",
		},
		{
			name: "block assignment with loops",
			template: `{%- set list_html -%}
<ul>
{%- for item in items -%}
  <li>{{ item }}</li>
{%- endfor -%}
</ul>
{%- endset -%}
{{ list_html }}`,
			data:     map[string]interface{}{"items": []string{"apple", "banana", "cherry"}},
			expected: "<ul><li>apple</li><li>banana</li><li>cherry</li></ul>",
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

			if strings.TrimSpace(result) != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, strings.TrimSpace(result))
			}
		})
	}
}

// Multiple Assignment Tests
func TestMultipleAssignment(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "multiple simple assignments",
			template: `{%- set first = "John" -%}
{%- set last = "Doe" -%}
{{ first }} {{ last }}`,
			data:     map[string]interface{}{},
			expected: "John Doe",
		},
		{
			name: "assignments with dependencies",
			template: `{%- set base = 10 -%}
{%- set doubled = base * 2 -%}
{%- set result = doubled + base -%}
{{ result }}`,
			data:     map[string]interface{}{},
			expected: "30",
		},
		{
			name: "assignments with context variables",
			template: `{%- set greeting = "Hello" -%}
{%- set full_greeting = greeting + ", " + name -%}
{{ full_greeting }}`,
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello, World",
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

			if strings.TrimSpace(result) != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, strings.TrimSpace(result))
			}
		})
	}
}

// Advanced Assignment Tests
func TestAdvancedAssignments(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "assignment with conditional",
			template: `{%- set status = "active" if user.active else "inactive" -%}
{{ status }}`,
			data:     map[string]interface{}{"user": map[string]interface{}{"active": true}},
			expected: "active",
		},
		{
			name: "assignment with complex expression",
			template: `{%- set percentage = ((active / total) * 100)|round(1) if total > 0 else 0 -%}
{{ percentage }}%`,
			data:     map[string]interface{}{"active": 7, "total": 10},
			expected: "70%",
		},
		{
			name: "assignment with filter chain",
			template: `{%- set processed = text|trim|upper|reverse -%}
{{ processed }}`,
			data:     map[string]interface{}{"text": "  hello  "},
			expected: "OLLEH",
		},
		{
			name: "assignment with list comprehension-like filter",
			template: `{%- set active_users = users|selectattr('active')|list -%}
Active count: {{ active_users|length }}`,
			data: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "John", "active": true},
					map[string]interface{}{"name": "Jane", "active": false},
					map[string]interface{}{"name": "Bob", "active": true},
				},
			},
			expected: "Active count: 2",
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

			if strings.TrimSpace(result) != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, strings.TrimSpace(result))
			}
		})
	}
}

// Assignment Operators Tests
func TestAssignmentOperators(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "arithmetic assignment",
			template: "{% set x = a + b * c %}{{ x }}",
			data:     map[string]interface{}{"a": 1, "b": 2, "c": 3},
			expected: "7",
		},
		{
			name:     "string assignment with concatenation",
			template: "{% set full = first + ' ' + last %}{{ full }}",
			data:     map[string]interface{}{"first": "John", "last": "Doe"},
			expected: "John Doe",
		},
		{
			name:     "boolean assignment with comparison",
			template: "{% set is_adult = age >= 18 %}{{ is_adult }}",
			data:     map[string]interface{}{"age": 25},
			expected: "true",
		},
		{
			name:     "assignment with logical operators",
			template: "{% set has_access = is_admin or (is_user and has_permission) %}{{ has_access }}",
			data:     map[string]interface{}{"is_admin": false, "is_user": true, "has_permission": true},
			expected: "true",
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

// Assignment Error Cases Tests
func TestAssignmentErrors(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name          string
		template      string
		data          map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:        "assignment to undefined variable",
			template:    "{% set result = undefined_var + 1 %}{{ result }}",
			data:        map[string]interface{}{},
			expectError: false, // Should handle gracefully with undefined behavior
		},
		{
			name:          "invalid assignment syntax",
			template:      "{% set = 'invalid' %}",
			data:          map[string]interface{}{},
			expectError:   true,
			errorContains: "syntax error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if test.expectError {
				if err == nil {
					// Try rendering to catch runtime errors
					_, renderErr := tmpl.Render(miya.NewContextFrom(test.data))
					if renderErr == nil {
						t.Fatalf("Expected error but none occurred")
					}
					if test.errorContains != "" && !strings.Contains(renderErr.Error(), test.errorContains) {
						t.Errorf("Expected error containing %q, got: %q", test.errorContains, renderErr.Error())
					}
				} else {
					if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
						t.Errorf("Expected error containing %q, got: %q", test.errorContains, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected parse error: %v", err)
				}

				result, renderErr := tmpl.Render(miya.NewContextFrom(test.data))
				if renderErr != nil {
					t.Fatalf("Unexpected render error: %v", renderErr)
				}

				// Just ensure it rendered something
				if len(result) == 0 {
					t.Error("Expected some output, got empty string")
				}
			}
		})
	}
}

// Assignment Scope Tests
func TestAssignmentScope(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "assignment in for loop scope",
			template: `{%- set outer = "outer" -%}
{%- for item in items -%}
  {%- set inner = "inner-" + item -%}
  {{ outer }}-{{ inner }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data:     map[string]interface{}{"items": []string{"1", "2"}},
			expected: "outer-inner-1,outer-inner-2",
		},
		{
			name: "assignment in if block scope",
			template: `{%- set base = "base" -%}
{%- if condition -%}
  {%- set conditional = base + "-conditional" -%}
  {{ conditional }}
{%- else -%}
  {{ base }}-else
{%- endif -%}`,
			data:     map[string]interface{}{"condition": true},
			expected: "base-conditional",
		},
		{
			name: "assignment persistence across blocks",
			template: `{%- set value = "initial" -%}
{%- if true -%}
  {%- set value = "modified" -%}
{%- endif -%}
{{ value }}`,
			data:     map[string]interface{}{},
			expected: "modified",
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

			if strings.TrimSpace(result) != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, strings.TrimSpace(result))
			}
		})
	}
}
