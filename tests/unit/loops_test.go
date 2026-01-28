package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"
)

// =============================================================================
// LOOP TESTS - CONSOLIDATED
// =============================================================================
// This file consolidates all loop-related tests from multiple individual files:
// - advanced_loop_test.go
// - advanced_loop_variables_test.go
// - loop_control_test.go
// - advanced_for_test.go
// - recursive_for_test.go
// - dictionary_iteration_test.go
// =============================================================================

// Advanced Loop Features Tests
func TestAdvancedLoopFeatures(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "loop variables",
			template: `{%- for item in items -%}
{{ loop.index }}/{{ loop.length }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": []string{"a", "b", "c"},
			},
			expected: "1/3,2/3,3/3",
		},
		{
			name: "loop.first and loop.last",
			template: `{%- for item in items -%}
{%- if loop.first -%}[{%- endif -%}
{{ item }}
{%- if not loop.last -%},{%- endif -%}
{%- if loop.last -%}]{%- endif -%}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": []string{"x", "y", "z"},
			},
			expected: "[x,y,z]",
		},
		{
			name: "loop.index0 and loop.revindex",
			template: `{%- for item in items -%}
{{ item }}:{{ loop.index0 }}:{{ loop.revindex }}{% if not loop.last %};{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": []string{"a", "b"},
			},
			expected: "a:0:2;b:1:1",
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

// Advanced Loop Variables Tests
func TestAdvancedLoopVariables(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "nested loops with loop variables",
			template: `{%- for outer in outers -%}
{{ loop.index }}:{%- for inner in inners -%}{{ outer }}.{{ inner }}.{{ loop.index }}{% if not loop.last %},{% endif %}{%- endfor -%}
{%- if not loop.last %};{% endif -%}
{%- endfor -%}`,
			data: map[string]interface{}{
				"outers": []string{"A", "B"},
				"inners": []string{"1", "2"},
			},
			expected: "1:A.1.1,A.2.2;2:B.1.1,B.2.2",
		},
		{
			name: "loop variables in conditions",
			template: `{%- for item in items -%}
{%- if loop.index is even -%}
EVEN:{{ item }}
{%- else -%}
ODD:{{ item }}
{%- endif -%}
{%- if not loop.last -%},{%- endif -%}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": []string{"a", "b", "c", "d"},
			},
			expected: "ODD:a,EVEN:b,ODD:c,EVEN:d",
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

// Loop Control Tests (break/continue)
func TestLoopBreak(t *testing.T) {
	env := miya.NewEnvironment()

	template := `{%- for i in range(10) -%}
{%- if i == 3 -%}
{%- break -%}
{%- endif -%}
{{ i }}
{%- if not loop.last -%},{%- endif -%}
{%- endfor -%}`

	tmpl, err := env.FromString(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.Render(miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "0,1,2"
	// Remove trailing comma if present
	result = strings.TrimSuffix(result, ",")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestLoopContinue(t *testing.T) {
	env := miya.NewEnvironment()

	template := `{%- for i in range(5) -%}
{%- if i == 2 -%}
{%- continue -%}
{%- endif -%}
{{ i }}
{%- if i != 4 and i != 2 -%},{%- endif -%}
{%- endfor -%}`

	tmpl, err := env.FromString(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tmpl.Render(miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "0,1,3,4"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// For Loop with Conditions
func TestForLoopWithConditions(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "for loop with if condition",
			template: `{%- for item in items if item > 2 -%}
{{ item }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": []int{1, 2, 3, 4, 5},
			},
			expected: "3,4,5",
		},
		{
			name: "for loop with complex condition",
			template: `{%- for user in users if user.active -%}
{{ user.name }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"users": []map[string]interface{}{
					{"name": "John", "active": true},
					{"name": "Jane", "active": false},
					{"name": "Bob", "active": true},
				},
			},
			expected: "John,Bob",
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

// Variable Unpacking in For Loops
func TestVariableUnpackingInForLoops(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "key-value unpacking",
			template: `{%- for key, value in items -%}
{{ key }}:{{ value }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"items": map[string]interface{}{
					"a": 1,
					"b": 2,
				},
			},
			expected: "a:1,b:2",
		},
		{
			name: "list unpacking",
			template: `{%- for first, second in pairs -%}
{{ first }}-{{ second }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"pairs": [][]interface{}{
					{"x", "y"},
					{"a", "b"},
				},
			},
			expected: "x-y,a-b",
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

			// For map iteration tests, order is not guaranteed in Go
			// Check both possible orderings for key-value unpacking
			if test.name == "key-value unpacking" {
				if result != "a:1,b:2" && result != "b:2,a:1" {
					t.Errorf("Expected either %q or %q, got %q", "a:1,b:2", "b:2,a:1", result)
				}
			} else if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

// Recursive For Loops Tests
func TestRecursiveForLoops(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "simple recursive loop",
			template: `{%- for category in categories -%}
{{ category.name }}
{%- if category.subcategories -%}
:{%- for sub in category.subcategories -%}
{{ sub.name }}{% if not loop.last %},{% endif %}
{%- endfor -%}
{%- endif -%}
{%- if not loop.last %};{% endif -%}
{%- endfor -%}`,
			data: map[string]interface{}{
				"categories": []map[string]interface{}{
					{
						"name": "Electronics",
						"subcategories": []map[string]interface{}{
							{"name": "Laptops"},
							{"name": "Phones"},
						},
					},
					{
						"name":          "Books",
						"subcategories": nil,
					},
				},
			},
			expected: "Electronics:Laptops,Phones;Books",
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

// Dictionary Iteration Tests
func TestDictionaryIteration(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name      string
		template  string
		data      map[string]interface{}
		checkFunc func(string) bool
	}{
		{
			name: "iterate over dictionary items",
			template: `{%- for key, value in user.items() -%}
{{ key }}={{ value }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			checkFunc: func(result string) bool {
				return strings.Contains(result, "name=John") && strings.Contains(result, "age=30")
			},
		},
		{
			name: "iterate over dictionary keys",
			template: `{%- for key in user.keys() -%}
{{ key }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			checkFunc: func(result string) bool {
				return strings.Contains(result, "name") && strings.Contains(result, "age")
			},
		},
		{
			name: "iterate over dictionary values",
			template: `{%- for value in user.values() -%}
{{ value }}{% if not loop.last %},{% endif %}
{%- endfor -%}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			checkFunc: func(result string) bool {
				return strings.Contains(result, "John") && strings.Contains(result, "30")
			},
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

			if !test.checkFunc(result) {
				t.Errorf("Result %q did not pass validation", result)
			}
		})
	}
}
