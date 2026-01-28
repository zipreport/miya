package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestCycler(t *testing.T) {
	env := miya.NewEnvironment()

	t.Run("Basic cycler functionality", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			expected string
		}{
			{
				name:     "Basic next() calls",
				template: `{% set cycle = cycler('a', 'b') %}{{ cycle.next() }}-{{ cycle.next() }}-{{ cycle.next() }}`,
				expected: "a-b-a",
			},
			{
				name:     "Mix of next() and current()",
				template: `{% set cycle = cycler('x', 'y') %}{{ cycle.next() }}-{{ cycle.current() }}-{{ cycle.next() }}`,
				expected: "x-y-y",
			},
			{
				name:     "Cycler with single item",
				template: `{% set cycle = cycler('only') %}{{ cycle.next() }}-{{ cycle.next() }}`,
				expected: "only-only",
			},
			{
				name:     "Cycler in loop",
				template: `{% set cycle = cycler('odd', 'even') %}{% for i in range(4) %}{{ cycle.next() }}{% if not loop.last %},{% endif %}{% endfor %}`,
				expected: "odd,even,odd,even",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				result, err := env.RenderString(tc.template, miya.NewContext())
				if err != nil {
					t.Fatalf("Template failed: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	t.Run("Cycler attribute access", func(t *testing.T) {
		// Test that accessing cycler attributes works correctly
		template := `{% set cycle = cycler('test') %}{{ cycle.next is defined }}`
		result, err := env.RenderString(template, miya.NewContext())
		if err != nil {
			t.Fatalf("Attribute check failed: %v", err)
		}
		if result != "true" {
			t.Errorf("Expected 'true', got '%s' - cycler.next should be defined", result)
		}
	})

	t.Run("Cycler error cases", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
		}{
			{
				name:     "Invalid method call",
				template: `{% set cycle = cycler('a') %}{{ cycle.invalid() }}`,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := env.RenderString(tc.template, miya.NewContext())
				if err == nil {
					t.Fatalf("Expected error for template: %s", tc.template)
				}
				t.Logf("Got expected error: %v", err)
			})
		}
	})
}
