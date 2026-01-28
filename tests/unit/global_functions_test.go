package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"
)

func TestNewGlobalFunctions(t *testing.T) {
	env := miya.NewEnvironment()

	t.Run("cycler function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			expected string
		}{
			{
				name:     "Basic cycler usage",
				template: `{% set cycle = cycler('odd', 'even') %}{{ cycle.next() }}-{{ cycle.next() }}-{{ cycle.next() }}`,
				expected: "odd-even-odd",
			},
			{
				name:     "Cycler with numbers",
				template: `{% set cycle = cycler(1, 2, 3) %}{{ cycle.next() }},{{ cycle.next() }},{{ cycle.next() }},{{ cycle.next() }}`,
				expected: "1,2,3,1",
			},
			{
				name:     "Cycler current method",
				template: `{% set cycle = cycler('a', 'b') %}{{ cycle.current() }}-{{ cycle.next() }}-{{ cycle.current() }}`,
				expected: "a-a-b",
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
	})

	t.Run("joiner function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			expected string
		}{
			{
				name:     "Basic joiner usage",
				template: `{% set j = joiner(', ') %}{% for item in ['a', 'b', 'c'] %}{{ j() }}{{ item }}{% endfor %}`,
				expected: "a, b, c",
			},
			{
				name:     "Joiner with custom separator",
				template: `{% set j = joiner(' | ') %}{% for item in [1, 2, 3] %}{{ j() }}{{ item }}{% endfor %}`,
				expected: "1 | 2 | 3",
			},
			{
				name:     "Joiner default separator",
				template: `{% set j = joiner() %}{% for item in ['x', 'y'] %}{{ j() }}{{ item }}{% endfor %}`,
				expected: "x, y",
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
	})

	t.Run("lipsum function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			check    func(string) bool
		}{
			{
				name:     "Default lipsum",
				template: `{{ lipsum() }}`,
				check: func(result string) bool {
					return strings.Contains(result, "lorem") && strings.Contains(result, "ipsum") && len(result) > 100
				},
			},
			{
				name:     "Lipsum with paragraph count",
				template: `{{ lipsum(2) }}`,
				check: func(result string) bool {
					return strings.Count(result, "\n\n") == 1 // 2 paragraphs = 1 separator
				},
			},
			{
				name:     "Lipsum words only",
				template: `{{ lipsum(1, false) }}`,
				check: func(result string) bool {
					return strings.Contains(result, "lorem") && !strings.Contains(result, ".")
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctx := miya.NewContext()
				result, err := env.RenderString(tc.template, ctx)
				if err != nil {
					t.Fatalf("Failed to render template: %v", err)
				}

				if !tc.check(result) {
					t.Errorf("Check failed for result: %s", result)
				}
			})
		}
	})

	t.Run("enhanced dict function", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			expected string
		}{
			{
				name:     "Dict with key-value pairs",
				template: `{% set d = dict('key1', 'value1', 'key2', 'value2') %}{{ d.key1 }}-{{ d.key2 }}`,
				expected: "value1-value2",
			},
			{
				name:     "Empty dict",
				template: `{% set d = dict() %}{{ d|length }}`,
				expected: "0",
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
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
		}{
			{
				name:     "Cycler with no arguments",
				template: `{{ cycler() }}`,
			},
			{
				name:     "Dict with odd number of arguments",
				template: `{{ dict('key1', 'value1', 'key2') }}`,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctx := miya.NewContext()
				_, err := env.RenderString(tc.template, ctx)
				if err == nil {
					t.Fatalf("Expected error for template: %s", tc.template)
				}
				t.Logf("Got expected error: %v", err)
			})
		}
	})
}
