package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"
)

func TestMacroCallIntegration(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Basic macro with call block",
			template: `{% macro render_field(name) %}field: {{ name }} - {{ caller() }}{% endmacro %}{% call render_field('test') %}content{% endcall %}`,
			expected: `field: test - content`,
		},
		{
			name:     "Macro with parameters and call block",
			template: `{% macro wrapper(title, class) %}<div class="{{ class }}"><h1>{{ title }}</h1>{{ caller() }}</div>{% endmacro %}{% call wrapper('Hello', 'main') %}<p>World</p>{% endcall %}`,
			expected: `<div class="main"><h1>Hello</h1>&lt;p&gt;World&lt;/p&gt;</div>`,
		},
		{
			name:     "Multiple macros and calls",
			template: `{% macro bold(text) %}<b>{{ text }}: {{ caller() }}</b>{% endmacro %}{% macro italic(text) %}<i>{{ text }}: {{ caller() }}</i>{% endmacro %}{% call bold('Important') %}{% call italic('Note') %}Message{% endcall %}{% endcall %}`,
			expected: `<b>Important: &lt;i&gt;Note: Message&lt;/i&gt;</b>`,
		},
		{
			name:     "Macro with no parameters but using caller",
			template: `{% macro box() %}<div class="box">{{ caller() }}</div>{% endmacro %}{% call box() %}Hello World{% endcall %}`,
			expected: `<div class="box">Hello World</div>`,
		},
		{
			name:     "Call block with variable content",
			template: `{% set message = 'Dynamic Content' %}{% macro container() %}[{{ caller() }}]{% endmacro %}{% call container() %}{{ message }}{% endcall %}`,
			expected: `[Dynamic Content]`,
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

func TestMacroCallWithDefaults(t *testing.T) {
	env := miya.NewEnvironment()

	// Test macro with default parameters called from call block
	template := `{% macro card(title, class='default') %}<div class="{{ class }}"><h2>{{ title }}</h2>{{ caller() }}</div>{% endmacro %}{% call card('Title') %}Body content{% endcall %}`

	ctx := miya.NewContext()
	result, err := env.RenderString(template, ctx)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := `<div class="default"><h2>Title</h2>Body content</div>`
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestMacroCallErrorHandling(t *testing.T) {
	env := miya.NewEnvironment()

	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "Macro using caller without call block",
			template: `{% macro test() %}{{ caller() }}{% endmacro %}{{ test() }}`,
		},
		{
			name:     "Call block with undefined macro",
			template: `{% call undefined_macro() %}content{% endcall %}`,
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
}
