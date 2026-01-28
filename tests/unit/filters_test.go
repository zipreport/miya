package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

// =============================================================================
// FILTER TESTS - CONSOLIDATED
// =============================================================================
// This file consolidates all filter-related tests from multiple individual files:
// - math_filters_test.go
// - extension_filters_test.go
// - collection_filters_test.go
// - essential_filters_test.go
// - advanced_string_filters_test.go
// - datetime_filters_test.go
// - filter_chaining_test.go
// - default_filter_test.go
// =============================================================================

// Math Filters Tests
func TestMathFilters(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "abs filter",
			template: "{{ value|abs }}",
			data:     map[string]interface{}{"value": -42},
			expected: "42",
		},
		{
			name:     "round filter",
			template: "{{ value|round(2) }}",
			data:     map[string]interface{}{"value": 3.14159},
			expected: "3.14",
		},
		{
			name:     "ceil filter",
			template: "{{ value|ceil }}",
			data:     map[string]interface{}{"value": 3.2},
			expected: "4",
		},
		{
			name:     "floor filter",
			template: "{{ value|floor }}",
			data:     map[string]interface{}{"value": 3.8},
			expected: "3",
		},
		{
			name:     "pow filter",
			template: "{{ base|pow(exp) }}",
			data:     map[string]interface{}{"base": 2, "exp": 3},
			expected: "8",
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

// String Filters Tests
func TestAdvancedStringFilters(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "slugify filter",
			template: "{{ text|slugify }}",
			data:     map[string]interface{}{"text": "Hello World! This is a test."},
			expected: "hello-world-this-is-a-test",
		},
		{
			name:     "wordcount filter",
			template: "{{ text|wordcount }}",
			data:     map[string]interface{}{"text": "Hello world this is a test"},
			expected: "6",
		},
		{
			name:     "center filter",
			template: "{{ text|center(10, '-') }}",
			data:     map[string]interface{}{"text": "test"},
			expected: "---test---",
		},
		{
			name:     "indent filter",
			template: "{{ text|indent(2) }}",
			data:     map[string]interface{}{"text": "line1\nline2"},
			expected: "line1\n  line2",
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

// Collection Filters Tests
func TestCollectionFilters(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "unique filter",
			template: "{{ items|unique|join(',') }}",
			data:     map[string]interface{}{"items": []interface{}{1, 2, 2, 3, 1}},
			expected: "1,2,3",
		},
		{
			name:     "reverse filter",
			template: "{{ items|reverse|join(',') }}",
			data:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			expected: "3,2,1",
		},
		{
			name:     "sort filter",
			template: "{{ items|sort|join(',') }}",
			data:     map[string]interface{}{"items": []interface{}{3, 1, 2}},
			expected: "1,2,3",
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
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

// Default Filter Tests
func TestDefaultFilterWithUndefinedVariables(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "undefined variable with default",
			template: "{{ undefined_var|default('fallback') }}",
			data:     map[string]interface{}{},
			expected: "fallback",
		},
		{
			name:     "defined variable ignores default",
			template: "{{ defined_var|default('fallback') }}",
			data:     map[string]interface{}{"defined_var": "actual"},
			expected: "actual",
		},
		{
			name:     "empty string with default",
			template: "{{ empty_var|default('fallback', true) }}",
			data:     map[string]interface{}{"empty_var": ""},
			expected: "fallback",
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

// Filter Chaining Tests
func TestFilterChaining(t *testing.T) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "multiple string filters",
			template: "{{ text|upper|trim|reverse }}",
			data:     map[string]interface{}{"text": "  hello  "},
			expected: "OLLEH",
		},
		{
			name:     "collection and string filters",
			template: "{{ items|join(', ')|upper }}",
			data:     map[string]interface{}{"items": []string{"apple", "banana", "cherry"}},
			expected: "APPLE, BANANA, CHERRY",
		},
		{
			name:     "numeric and string filters",
			template: "{{ value|round(2)|string|upper }}",
			data:     map[string]interface{}{"value": 3.14159},
			expected: "3.14",
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

// Essential Filters Tests
func TestEssentialBuiltinFilters(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "length filter",
			template: "{{ items|length }}",
			data:     map[string]interface{}{"items": []int{1, 2, 3, 4, 5}},
			expected: "5",
		},
		{
			name:     "first filter",
			template: "{{ items|first }}",
			data:     map[string]interface{}{"items": []string{"apple", "banana", "cherry"}},
			expected: "apple",
		},
		{
			name:     "last filter",
			template: "{{ items|last }}",
			data:     map[string]interface{}{"items": []string{"apple", "banana", "cherry"}},
			expected: "cherry",
		},
		{
			name:     "join filter",
			template: "{{ items|join(' - ') }}",
			data:     map[string]interface{}{"items": []string{"apple", "banana", "cherry"}},
			expected: "apple - banana - cherry",
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

// Date/Time Filters Tests
func TestDateTimeFilters(t *testing.T) {
	env := miya.NewEnvironment()

	// Test with basic date operations
	tests := []struct {
		name      string
		template  string
		data      map[string]interface{}
		checkFunc func(string) bool
	}{
		{
			name:     "strftime filter",
			template: "{{ now|strftime('%Y-%m-%d') }}",
			data:     map[string]interface{}{"now": "2023-01-01T12:00:00Z"},
			checkFunc: func(result string) bool {
				return len(result) == 10 && result[4] == '-' && result[7] == '-'
			},
		},
		{
			name:     "date filter",
			template: "{{ now|date }}",
			data:     map[string]interface{}{"now": "2023-01-01T12:00:00Z"},
			checkFunc: func(result string) bool {
				return len(result) > 0 // Just check it produces output
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

// Filter Complex Expressions Tests
func TestComplexFilterExpressions(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "filter with complex expression",
			template: "{{ (items|selectattr('active')|list)|length }}",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "John", "active": true},
					map[string]interface{}{"name": "Jane", "active": false},
					map[string]interface{}{"name": "Bob", "active": true},
				},
			},
			expected: "2",
		},
		{
			name:     "nested filter operations",
			template: "{{ items|map('name')|join(' and ') }}",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "Apple"},
					map[string]interface{}{"name": "Banana"},
				},
			},
			expected: "Apple and Banana",
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
