package miya_test

import (
	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"strings"
	"testing"
)

// =============================================================================
// COMPREHENSIVE TEMPLATE INHERITANCE TESTS
// =============================================================================
// This file provides comprehensive test coverage for template inheritance
// to improve coverage from 52.4% to target 75%+
// =============================================================================

// Test Basic Inheritance Functionality
func TestBasicInheritance(t *testing.T) {
	// Create environment with string loader
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	// Define templates
	baseTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    <header>{% block header %}Default Header{% endblock %}</header>
    <main>{% block content %}Default Content{% endblock %}</main>
    <footer>{% block footer %}Default Footer{% endblock %}</footer>
</body>
</html>`

	childTemplate := `{% extends "base.html" %}
{% block title %}Child Page{% endblock %}
{% block content %}
<h1>Welcome to Child Page</h1>
<p>This is child content.</p>
{% endblock %}`

	// Setup templates
	stringLoader.AddTemplate("base.html", baseTemplate)

	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name:     "basic inheritance",
			template: childTemplate,
			expected: []string{"Child Page", "Default Header", "Welcome to Child Page", "Default Footer"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Multi-Level Inheritance
func TestMultiLevelInheritance(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	// Define templates
	baseTemplate := `<html>
<head><title>{% block title %}Base{% endblock %}</title></head>
<body>
    {% block header %}Base Header{% endblock %}
    {% block content %}Base Content{% endblock %}
    {% block footer %}Base Footer{% endblock %}
</body>
</html>`

	middleTemplate := `{% extends "base.html" %}
{% block title %}Middle - {{ super() }}{% endblock %}
{% block header %}Middle Header{% endblock %}
{% block content %}
{{ super() }}
<div>Middle Content</div>
{% endblock %}`

	childTemplate := `{% extends "middle.html" %}
{% block title %}Child - {{ super() }}{% endblock %}
{% block content %}
{{ super() }}
<div>Child Content</div>
{% endblock %}`

	// Setup templates
	stringLoader.AddTemplate("base.html", baseTemplate)
	stringLoader.AddTemplate("middle.html", middleTemplate)
	stringLoader.AddTemplate("child.html", childTemplate)

	tests := []struct {
		name         string
		templateName string
		expected     []string
	}{
		{
			name:         "three-level inheritance",
			templateName: "child.html",
			expected:     []string{"Child - Middle - Base", "Middle Header", "Base Content", "Middle Content", "Child Content", "Base Footer"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.templateName)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Super Call Functionality
func TestSuperCallFunctionality(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	baseTemplate := `{% block greeting %}Hello{% endblock %} {% block name %}World{% endblock %}!`

	childTemplate := `{% extends "base.html" %}
{% block greeting %}{{ super() }}, Good Morning{% endblock %}
{% block name %}{{ super() }} and {{ name }}{% endblock %}`

	stringLoader.AddTemplate("base.html", baseTemplate)

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "super calls with variables",
			template: childTemplate,
			data:     map[string]interface{}{"name": "John"},
			expected: "Hello, Good Morning World and John!",
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

// Test Block Overriding Patterns
func TestBlockOverridingPatterns(t *testing.T) {
	baseTemplate := `
{%- block outer -%}
  <div class="outer">
    {%- block inner -%}Default Inner{%- endblock -%}
  </div>
{%- endblock -%}`

	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name: "partial block override",
			template: `{% extends "base.html" %}
{% block inner %}Custom Inner{% endblock %}`,
			expected: []string{"<div class=\"outer\">", "Custom Inner", "</div>"},
		},
		{
			name: "complete block override",
			template: `{% extends "base.html" %}
{% block outer %}<section class="custom">Complete Override</section>{% endblock %}`,
			expected: []string{"<section class=\"custom\">Complete Override</section>"},
		},
		{
			name: "nested block with super",
			template: `{% extends "base.html" %}
{% block outer %}{{ super() }}<p>Additional content</p>{% endblock %}`,
			expected: []string{"<div class=\"outer\">", "Default Inner", "</div>", "<p>Additional content</p>"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create fresh environment for each test to avoid state pollution
			directParser := loader.NewDirectTemplateParser()
			stringLoader := loader.NewStringLoader(directParser)
			env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))
			stringLoader.AddTemplate("base.html", baseTemplate)

			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Dynamic Inheritance
func TestDynamicInheritance(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	// Setup multiple base templates
	layout1 := `<div class="layout1">{% block content %}Layout1{% endblock %}</div>`
	layout2 := `<div class="layout2">{% block content %}Layout2{% endblock %}</div>`

	stringLoader.AddTemplate("layout1.html", layout1)
	stringLoader.AddTemplate("layout2.html", layout2)

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected []string
	}{
		{
			name:     "conditional inheritance - layout1",
			template: `{% extends layout_name %}{% block content %}Dynamic Content{% endblock %}`,
			data:     map[string]interface{}{"layout_name": "layout1.html"},
			expected: []string{"layout1", "Dynamic Content"},
		},
		{
			name:     "conditional inheritance - layout2",
			template: `{% extends layout_name %}{% block content %}Dynamic Content{% endblock %}`,
			data:     map[string]interface{}{"layout_name": "layout2.html"},
			expected: []string{"layout2", "Dynamic Content"},
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

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Inheritance with Variables and Control Structures
func TestInheritanceWithControlStructures(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	baseTemplate := `
{%- block header -%}
<header>
  <h1>{{ title|default('Default Title') }}</h1>
  {%- block nav -%}{%- endblock -%}
</header>
{%- endblock -%}
{%- block content -%}Default Content{%- endblock -%}
{%- block footer -%}<footer>© {{ year|default(2023) }}</footer>{%- endblock -%}`

	childTemplate := `{% extends "base.html" %}
{% block nav %}
<nav>
  {%- for item in menu_items -%}
    <a href="{{ item.url }}">{{ item.title }}</a>
  {%- endfor -%}
</nav>
{% endblock %}
{% block content %}
{%- if user -%}
  <p>Welcome, {{ user.name }}!</p>
{%- else -%}
  <p>Please log in.</p>
{%- endif -%}
{{ super() }}
{% endblock %}`

	stringLoader.AddTemplate("base.html", baseTemplate)

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected []string
	}{
		{
			name:     "inheritance with variables",
			template: childTemplate,
			data: map[string]interface{}{
				"title": "My Site",
				"year":  2024,
				"user":  map[string]interface{}{"name": "John"},
				"menu_items": []interface{}{
					map[string]interface{}{"url": "/home", "title": "Home"},
					map[string]interface{}{"url": "/about", "title": "About"},
				},
			},
			expected: []string{"My Site", "Home", "About", "Welcome, John!", "Default Content", "© 2024"},
		},
		{
			name:     "inheritance without user",
			template: childTemplate,
			data: map[string]interface{}{
				"title":      "My Site",
				"menu_items": []interface{}{},
			},
			expected: []string{"My Site", "Please log in.", "Default Content", "© 2023"},
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

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Inheritance Error Handling
func TestInheritanceErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing parent template",
			template:      `{% extends "nonexistent.html" %}{% block content %}Child{% endblock %}`,
			expectError:   true,
			errorContains: "template not found",
		},
		{
			name:          "super call outside inheritance",
			template:      `{{ super() }}`,
			expectError:   true,
			errorContains: "super",
		},
		{
			name:        "super call in base template",
			template:    `{% block content %}{{ super() }}{% endblock %}`,
			expectError: false, // Should handle gracefully
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create fresh environment for each test to avoid cache interactions
			directParser := loader.NewDirectTemplateParser()
			stringLoader := loader.NewStringLoader(directParser)
			env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

			baseTemplate := `{% block content %}Base Content{% endblock %}`
			stringLoader.AddTemplate("base.html", baseTemplate)

			tmpl, err := env.FromString(test.template)
			if test.expectError && err != nil {
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", test.errorContains, err)
				}
				return
			}

			if err != nil && test.expectError {
				return // Parse error is acceptable for some tests
			}

			if err != nil {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			result, renderErr := tmpl.Render(miya.NewContext())
			if test.expectError {
				if renderErr == nil {
					t.Fatalf("Expected render error but got result: %q", result)
				}
				if test.errorContains != "" && !strings.Contains(renderErr.Error(), test.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", test.errorContains, renderErr)
				}
			} else {
				if renderErr != nil {
					t.Fatalf("Unexpected render error: %v", renderErr)
				}
			}
		})
	}
}

// Test Block Name Resolution
func TestBlockNameResolution(t *testing.T) {
	// Test template with multiple blocks
	baseTemplate := `
{%- block header -%}Base Header{%- endblock -%}
{%- block main -%}
  {%- block sidebar -%}Base Sidebar{%- endblock -%}
  {%- block content -%}Base Content{%- endblock -%}
{%- endblock -%}
{%- block footer -%}Base Footer{%- endblock -%}`

	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name: "selective block overrides",
			template: `{% extends "base.html" %}
{% block header %}Custom Header{% endblock %}
{% block content %}Custom Content{% endblock %}`,
			expected: []string{"Custom Header", "Base Sidebar", "Custom Content", "Base Footer"},
		},
		{
			name: "nested block structure",
			template: `{% extends "base.html" %}
{% block main %}
<div class="custom-main">
  {% block sidebar %}Custom Sidebar{% endblock %}
  {% block content %}{{ super() }} + Custom{% endblock %}
</div>
{% endblock %}`,
			expected: []string{"Base Header", "custom-main", "Custom Sidebar", "Base Content + Custom", "Base Footer"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create fresh environment for each test to avoid state pollution
			directParser := loader.NewDirectTemplateParser()
			stringLoader := loader.NewStringLoader(directParser)
			env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))
			stringLoader.AddTemplate("base.html", baseTemplate)

			tmpl, err := env.FromString(test.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Inheritance Caching
func TestInheritanceCaching(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	baseTemplate := `<base>{% block content %}Base{% endblock %}</base>`
	childTemplate := `{% extends "base.html" %}{% block content %}Child{% endblock %}`

	stringLoader.AddTemplate("base.html", baseTemplate)

	// First render
	tmpl1, err := env.FromString(childTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result1, err := tmpl1.Render(miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Second render (should use cached inheritance resolution)
	tmpl2, err := env.FromString(childTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result2, err := tmpl2.Render(miya.NewContext())
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if result1 != result2 {
		t.Errorf("Cached results should be identical: %q vs %q", result1, result2)
	}

	expected := "<base>Child</base>"
	if result1 != expected {
		t.Errorf("Expected %q, got %q", expected, result1)
	}
}

// Test Complex Inheritance Scenarios
func TestComplexInheritanceScenarios(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	// Setup a complex inheritance hierarchy
	baseTemplate := `<html>
<head>{% block head %}{% block title %}<title>Base</title>{% endblock %}{% endblock %}</head>
<body>
  {% block body %}
    {% block header %}{% endblock %}
    {% block main %}
      {% block content %}Default Content{% endblock %}
      {% block sidebar %}Default Sidebar{% endblock %}
    {% endblock %}
    {% block footer %}Default Footer{% endblock %}
  {% endblock %}
</body>
</html>`

	pageTemplate := `{% extends "base.html" %}
{% block title %}<title>{{ page_title|default('Page') }}</title>{% endblock %}
{% block header %}<header>Page Header</header>{% endblock %}
{% block footer %}<footer>Page Footer</footer>{% endblock %}`

	articleTemplate := `{% extends "page.html" %}
{% block title %}<title>{{ article.title }} - {{ super() }}</title>{% endblock %}
{% block content %}
<article>
  <h1>{{ article.title }}</h1>
  <p>{{ article.content }}</p>
</article>
{% endblock %}
{% block sidebar %}
<aside>
  <h3>Related Articles</h3>
  {% for related in article.related %}
    <p>{{ related.title }}</p>
  {% endfor %}
</aside>
{% endblock %}`

	stringLoader.AddTemplate("base.html", baseTemplate)
	stringLoader.AddTemplate("page.html", pageTemplate)

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected []string
	}{
		{
			name:     "complex nested inheritance",
			template: articleTemplate,
			data: map[string]interface{}{
				"page_title": "Blog",
				"article": map[string]interface{}{
					"title":   "Go Templates",
					"content": "Learn about Go template inheritance.",
					"related": []interface{}{
						map[string]interface{}{"title": "Jinja2 Guide"},
						map[string]interface{}{"title": "Template Best Practices"},
					},
				},
			},
			expected: []string{
				"Go Templates - <title>Blog</title>",
				"Page Header",
				"<h1>Go Templates</h1>",
				"Learn about Go template inheritance",
				"Related Articles",
				"Jinja2 Guide",
				"Template Best Practices",
				"Page Footer",
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

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}
