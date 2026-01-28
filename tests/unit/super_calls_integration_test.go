package miya_test

import (
	miya "github.com/zipreport/miya"
	"strings"
	"testing"

	"github.com/zipreport/miya/loader"
)

func TestSuperCallsIntegration(t *testing.T) {
	tests := []struct {
		name               string
		baseTemplate       string
		childTemplate      string
		grandChildTemplate string
		expected           string
		description        string
	}{
		{
			name:         "Simple super() with parentheses",
			baseTemplate: `<title>{% block title %}Base Title{% endblock %}</title>`,
			childTemplate: `{% extends "base.html" %}
{% block title %}Child - {{ super() }}{% endblock %}`,
			grandChildTemplate: "",
			expected:           `<title>Child - Base Title</title>`,
			description:        "{{ super() }} should resolve to parent block content",
		},
		{
			name:         "Simple super without parentheses",
			baseTemplate: `<title>{% block title %}Base Title{% endblock %}</title>`,
			childTemplate: `{% extends "base.html" %}
{% block title %}Child - {{ super }}{% endblock %}`,
			grandChildTemplate: "",
			expected:           `<title>Child - Base Title</title>`,
			description:        "{{ super }} should also resolve to parent block content",
		},
		{
			name: "Multiple super() calls in different blocks",
			baseTemplate: `<html>
<head><title>{% block title %}Base Title{% endblock %}</title></head>
<body>{% block content %}Base content{% endblock %}</body>
</html>`,
			childTemplate: `{% extends "base.html" %}
{% block title %}Child - {{ super() }}{% endblock %}
{% block content %}Child content plus {{ super() }}{% endblock %}`,
			grandChildTemplate: "",
			expected: `<html>
<head><title>Child - Base Title</title></head>
<body>Child content plus Base content</body>
</html>`,
			description: "Multiple {{ super() }} calls should work in different blocks",
		},
		{
			name:         "Super() in nested content",
			baseTemplate: `{% block content %}Base: <em>important</em>{% endblock %}`,
			childTemplate: `{% extends "base.html" %}
{% block content %}
  <div class="wrapper">
    Child prefix - {{ super() }} - Child suffix
  </div>
{% endblock %}`,
			grandChildTemplate: "",
			expected: `
  <div class="wrapper">
    Child prefix - Base: <em>important</em> - Child suffix
  </div>
`,
			description: "{{ super() }} should work within nested HTML",
		},
		{
			name:         "Three-level inheritance with super()",
			baseTemplate: `{% block message %}Base{% endblock %}`,
			childTemplate: `{% extends "base.html" %}
{% block message %}Child({{ super() }}){% endblock %}`,
			grandChildTemplate: `{% extends "child.html" %}
{% block message %}Grand{{ super() }}End{% endblock %}`,
			expected:    `GrandChild(Base)End`,
			description: "{{ super() }} should work in multi-level inheritance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create environment with DirectTemplateParser
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			directParser := loader.NewDirectTemplateParser()
			stringLoader := loader.NewStringLoader(directParser)

			// Add templates to loader
			stringLoader.AddTemplate("base.html", tt.baseTemplate)
			stringLoader.AddTemplate("child.html", tt.childTemplate)
			if tt.name == "Three-level inheritance with super()" {
				stringLoader.AddTemplate("grand.html", tt.grandChildTemplate)
			}

			env.SetLoader(stringLoader)

			// Get the template to render
			templateName := "child.html"
			if tt.name == "Three-level inheritance with super()" {
				templateName = "grand.html"
			}

			template, err := env.GetTemplate(templateName)
			if err != nil {
				t.Fatalf("Failed to load template %s: %v", templateName, err)
			}

			// Render template
			ctx := miya.NewContext()
			result, err := template.Render(ctx)
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			// Compare results (normalize whitespace)
			expected := strings.TrimSpace(tt.expected)
			actual := strings.TrimSpace(result)

			if actual != expected {
				t.Errorf("%s\nExpected:\n%q\nActual:\n%q", tt.description, expected, actual)
			}
		})
	}
}

func TestSuperCallsErrorCases(t *testing.T) {
	errorTests := []struct {
		name          string
		baseTemplate  string
		childTemplate string
		expectedError string
		description   string
	}{
		{
			name:         "Super() outside of block",
			baseTemplate: `{% block content %}Base{% endblock %}`,
			childTemplate: `{% extends "base.html" %}
{{ super() }}
{% block content %}Child{% endblock %}`,
			expectedError: "super() call outside of block context",
			description:   "{{ super() }} outside block should error",
		},
		{
			name:          "Super() in base template",
			baseTemplate:  `<title>{{ super() }}</title>`,
			childTemplate: "",
			expectedError: "super() call outside of block context",
			description:   "{{ super() }} in base template should error",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			env := miya.NewEnvironment(miya.WithAutoEscape(false))
			directParser := loader.NewDirectTemplateParser()
			stringLoader := loader.NewStringLoader(directParser)

			stringLoader.AddTemplate("base.html", tt.baseTemplate)
			if tt.childTemplate != "" {
				stringLoader.AddTemplate("child.html", tt.childTemplate)
			}

			env.SetLoader(stringLoader)

			templateName := "base.html"
			if tt.childTemplate != "" {
				templateName = "child.html"
			}

			template, err := env.GetTemplate(templateName)
			if err != nil {
				t.Fatalf("Failed to load template %s: %v", templateName, err)
			}

			ctx := miya.NewContext()
			_, err = template.Render(ctx)

			if err == nil {
				t.Fatalf("Expected error but rendering succeeded")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("%s\nExpected error containing: %q\nActual error: %q",
					tt.description, tt.expectedError, err.Error())
			}
		})
	}
}

func TestSuperCallsPerformance(t *testing.T) {
	// Test that caching works correctly with super() calls
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)

	baseTemplate := `{% block title %}Base Title{% endblock %}`
	childTemplate := `{% extends "base.html" %}
{% block title %}Child - {{ super() }}{% endblock %}`

	stringLoader.AddTemplate("base.html", baseTemplate)
	stringLoader.AddTemplate("child.html", childTemplate)
	env.SetLoader(stringLoader)

	ctx := miya.NewContext()

	// Render multiple times to test caching
	for i := 0; i < 5; i++ {
		template, err := env.GetTemplate("child.html")
		if err != nil {
			t.Fatalf("Failed to load template on iteration %d: %v", i, err)
		}

		result, err := template.Render(ctx)
		if err != nil {
			t.Fatalf("Failed to render template on iteration %d: %v", i, err)
		}

		expected := "Child - Base Title"
		if strings.TrimSpace(result) != expected {
			t.Errorf("Iteration %d: Expected %q, got %q", i, expected, result)
		}
	}

	// Check that inheritance cache is working
	stats := env.GetInheritanceCacheStats()
	if stats.HierarchyCache.Hits == 0 && stats.ResolvedCache.Hits == 0 {
		t.Errorf("Expected cache hits > 0, got hierarchy:%d resolved:%d",
			stats.HierarchyCache.Hits, stats.ResolvedCache.Hits)
	}
}
