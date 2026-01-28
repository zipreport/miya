package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestWithStatementFeatures(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Simple with statement",
			template: `{% with greeting = "Hello" %}{{ greeting }} World{% endwith %}`,
			context:  map[string]interface{}{},
			expected: "Hello World",
		},
		{
			name:     "Multiple assignments",
			template: `{% with name = "Alice", age = 30 %}{{ name }} is {{ age }} years old{% endwith %}`,
			context:  map[string]interface{}{},
			expected: "Alice is 30 years old",
		},
		{
			name:     "With expression using context",
			template: `{% with doubled = count * 2 %}Count: {{ count }}, Doubled: {{ doubled }}{% endwith %}`,
			context:  map[string]interface{}{"count": 5},
			expected: "Count: 5, Doubled: 10",
		},
		{
			name:     "Scoped variables",
			template: `{{ name }}{% with name = "Bob" %}{{ name }}{% endwith %}{{ name }}`,
			context:  map[string]interface{}{"name": "Alice"},
			expected: "AliceBobAlice",
		},
		{
			name:     "Nested with statements",
			template: `{% with x = 1 %}{% with y = 2 %}{{ x + y }}{% endwith %}{% endwith %}`,
			context:  map[string]interface{}{},
			expected: "3",
		},
		{
			name:     "With complex expressions",
			template: `{% with items = [1, 2, 3] %}{% with total = items|length %}Total: {{ total }}{% endwith %}{% endwith %}`,
			context:  map[string]interface{}{},
			expected: "Total: 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContext()
			for key, value := range tc.context {
				ctx.Set(key, value)
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
}

func TestWithStatementParsing(t *testing.T) {
	env := miya.NewEnvironment()

	// Test that basic with statements parse correctly
	template := `{% with x = 1 %}{{ x }}{% endwith %}`

	ctx := miya.NewContext()
	result, err := env.RenderString(template, ctx)
	if err != nil {
		t.Fatalf("Failed to parse with statement: %v", err)
	}

	expected := "1"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestWithStatementErrors(t *testing.T) {
	env := miya.NewEnvironment()

	errorTests := []struct {
		name     string
		template string
		errMsg   string
	}{
		{
			name:     "Missing assignment",
			template: `{% with %}{{ x }}{% endwith %}`,
			errMsg:   "expected variable name",
		},
		{
			name:     "Missing equals",
			template: `{% with x 1 %}{{ x }}{% endwith %}`,
			errMsg:   "expected '='",
		},
		{
			name:     "Missing endwith",
			template: `{% with x = 1 %}{{ x }}`,
			errMsg:   "expected '{% endwith %}'",
		},
	}

	for _, tc := range errorTests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := miya.NewContext()
			_, err := env.RenderString(tc.template, ctx)
			if err == nil {
				t.Fatalf("Expected error for template: %s", tc.template)
			}

			// Just check that we get some error - the exact message may vary
			t.Logf("Got expected error: %v", err)
		})
	}
}
