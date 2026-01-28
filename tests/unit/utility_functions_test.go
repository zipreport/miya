package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"
)

func TestUtilityFunctions(t *testing.T) {
	env := miya.NewEnvironment()

	t.Run("zip function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			data     map[string]interface{}
			expected string
		}{
			{
				name:     "Basic zip with two lists",
				template: `{% for item in zip([1, 2, 3], ['a', 'b', 'c']) %}{{ item[0] }}{{ item[1] }}{% endfor %}`,
				expected: "1a2b3c",
			},
			{
				name:     "Zip with different lengths",
				template: `{% for item in zip([1, 2, 3, 4], ['a', 'b']) %}{{ item[0] }}{{ item[1] }}{% endfor %}`,
				expected: "1a2b",
			},
			{
				name:     "Zip with variables",
				template: `{% for item in zip(numbers, letters) %}{{ item[0] }}-{{ item[1] }}{% if not loop.last %},{% endif %}{% endfor %}`,
				data: map[string]interface{}{
					"numbers": []int{1, 2, 3},
					"letters": []string{"a", "b", "c"},
				},
				expected: "1-a,2-b,3-c",
			},
			{
				name:     "Zip with empty list",
				template: `{% for item in zip([], [1, 2, 3]) %}{{ item[0] }}{% endfor %}empty`,
				expected: "empty",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctx := miya.NewContext()
				for k, v := range tc.data {
					ctx.Set(k, v)
				}
				result, err := env.RenderString(tc.template, ctx)
				if err != nil {
					t.Fatalf("Failed to render template: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	t.Run("enumerate function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			data     map[string]interface{}
			expected string
		}{
			{
				name:     "Basic enumerate",
				template: `{% for item in enumerate(['a', 'b', 'c']) %}{{ item[0] }}:{{ item[1] }}{% if not loop.last %},{% endif %}{% endfor %}`,
				expected: "0:a,1:b,2:c",
			},
			{
				name:     "Enumerate with start parameter",
				template: `{% for item in enumerate(['x', 'y', 'z'], 1) %}{{ item[0] }}:{{ item[1] }}{% if not loop.last %},{% endif %}{% endfor %}`,
				expected: "1:x,2:y,3:z",
			},
			{
				name:     "Enumerate with variable",
				template: `{% for item in enumerate(items) %}[{{ item[0] }}]{{ item[1] }}{% endfor %}`,
				data: map[string]interface{}{
					"items": []string{"first", "second"},
				},
				expected: "[0]first[1]second",
			},
			{
				name:     "Enumerate empty list",
				template: `{% for item in enumerate([]) %}{{ item[0] }}{% endfor %}none`,
				expected: "none",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctx := miya.NewContext()
				for k, v := range tc.data {
					ctx.Set(k, v)
				}
				result, err := env.RenderString(tc.template, ctx)
				if err != nil {
					t.Fatalf("Failed to render template: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	t.Run("url_for function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			data     map[string]interface{}
			expected string
		}{
			{
				name:     "Basic url_for",
				template: `{{ url_for('home') }}`,
				expected: "/home",
			},
			{
				name:     "url_for with parameters",
				template: `{{ url_for('user', 'id', 123, 'tab', 'profile') }}`,
				expected: "/user?id=123&amp;tab=profile", // HTML escaped due to auto-escaping
			},
			{
				name:     "url_for with variable",
				template: `{{ url_for(endpoint) }}`,
				data: map[string]interface{}{
					"endpoint": "dashboard",
				},
				expected: "/dashboard",
			},
			{
				name:     "url_for with dynamic parameters",
				template: `{{ url_for('article', 'id', article_id, 'section', section) }}`,
				data: map[string]interface{}{
					"article_id": 456,
					"section":    "comments",
				},
				expected: "/article?id=456&amp;section=comments", // HTML escaped due to auto-escaping
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctx := miya.NewContext()
				for k, v := range tc.data {
					ctx.Set(k, v)
				}
				result, err := env.RenderString(tc.template, ctx)
				if err != nil {
					t.Fatalf("Failed to render template: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	t.Run("Combined utility functions", func(t *testing.T) {
		// Simplified template to avoid whitespace issues
		template := `{%- for item in enumerate(zip(names, ages)) -%}{{ item[0] + 1 }}. {{ item[1][0] }} ({{ item[1][1] }} years old){%- if not loop.last %}, {% endif -%}{%- endfor -%}`

		ctx := miya.NewContext()
		ctx.Set("names", []string{"Alice", "Bob", "Charlie"})
		ctx.Set("ages", []int{25, 30, 35})

		result, err := env.RenderString(template, ctx)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		expected := "1. Alice (25 years old), 2. Bob (30 years old), 3. Charlie (35 years old)"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestUtilityFunctionErrors(t *testing.T) {
	env := miya.NewEnvironment()

	t.Run("zip with non-iterable", func(t *testing.T) {
		template := `{{ zip(123, ['a', 'b']) }}`
		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for non-iterable argument to zip")
		}
		if !strings.Contains(err.Error(), "not iterable") {
			t.Errorf("Expected 'not iterable' error, got: %v", err)
		}
	})

	t.Run("enumerate without arguments", func(t *testing.T) {
		template := `{{ enumerate() }}`
		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for enumerate without arguments")
		}
		if !strings.Contains(err.Error(), "requires at least one argument") {
			t.Errorf("Expected 'requires at least one argument' error, got: %v", err)
		}
	})

	t.Run("url_for without arguments", func(t *testing.T) {
		template := `{{ url_for() }}`
		ctx := miya.NewContext()
		_, err := env.RenderString(template, ctx)
		if err == nil {
			t.Fatal("Expected error for url_for without arguments")
		}
		if !strings.Contains(err.Error(), "requires at least one argument") {
			t.Errorf("Expected 'requires at least one argument' error, got: %v", err)
		}
	})
}
