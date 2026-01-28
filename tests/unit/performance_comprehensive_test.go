package miya_test

import (
	"fmt"
	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// COMPREHENSIVE PERFORMANCE BENCHMARKS WITH COVERAGE TRACKING
// =============================================================================
// This file provides performance benchmarks that also help improve coverage
// by exercising various code paths under realistic load conditions
// =============================================================================

// Benchmark Template Compilation Performance
func BenchmarkTemplateCompilation(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	templates := []string{
		"Simple: {{ name }}",
		"With filter: {{ name|upper|trim }}",
		"With condition: {% if active %}{{ name }}{% endif %}",
		"With loop: {% for item in items %}{{ item }}{% endfor %}",
		"Complex: {% for user in users %}{% if user.active %}{{ user.name|title }}{% endif %}{% endfor %}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		template := templates[i%len(templates)]
		_, err := env.FromString(template)
		if err != nil {
			b.Fatalf("Failed to compile template: %v", err)
		}
	}
}

// Benchmark Template Rendering Performance
func BenchmarkTemplateRendering(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	// Pre-compile templates
	templates := map[string]*miya.Template{}
	templateStrings := map[string]string{
		"simple":      "Hello {{ name }}!",
		"filtered":    "{{ text|upper|trim }}",
		"conditional": "{% if show %}{{ message }}{% else %}Hidden{% endif %}",
		"loop":        "{% for i in range(count) %}{{ i }}{% endfor %}",
		"complex":     "{% for user in users %}{{ user.name|title }} ({{ user.age }}){% if not loop.last %}, {% endif %}{% endfor %}",
	}

	for name, tmplStr := range templateStrings {
		tmpl, err := env.FromString(tmplStr)
		if err != nil {
			b.Fatalf("Failed to compile template %s: %v", name, err)
		}
		templates[name] = tmpl
	}

	// Benchmark data
	data := map[string]interface{}{
		"name":    "John",
		"text":    "  hello world  ",
		"show":    true,
		"message": "Visible message",
		"count":   10,
		"users": []interface{}{
			map[string]interface{}{"name": "alice", "age": 25},
			map[string]interface{}{"name": "bob", "age": 30},
			map[string]interface{}{"name": "charlie", "age": 35},
		},
	}

	templateNames := []string{"simple", "filtered", "conditional", "loop", "complex"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		templateName := templateNames[i%len(templateNames)]
		tmpl := templates[templateName]
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template %s: %v", templateName, err)
		}
	}
}

// Benchmark Large Template Rendering
func BenchmarkLargeTemplateRendering(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	// Create a large, complex template
	largeTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>{{ page_title|default('Default Title') }}</title>
    <meta name="description" content="{{ page_description|truncate(160) }}">
</head>
<body>
    <header>
        <nav>
            {% for item in navigation %}
                <a href="{{ item.url }}"{% if item.active %} class="active"{% endif %}>
                    {{ item.title|title }}
                </a>
            {% endfor %}
        </nav>
    </header>
    
    <main>
        <section class="hero">
            <h1>{{ hero.title }}</h1>
            <p>{{ hero.description }}</p>
        </section>
        
        <section class="content">
            {% for section in content_sections %}
                <div class="section">
                    <h2>{{ section.title }}</h2>
                    {% if section.items %}
                        <ul>
                            {% for item in section.items %}
                                <li>
                                    <strong>{{ item.name }}</strong>: {{ item.description|truncate(100) }}
                                    {% if item.tags %}
                                        <div class="tags">
                                            {% for tag in item.tags %}
                                                <span class="tag">{{ tag }}</span>
                                            {% endfor %}
                                        </div>
                                    {% endif %}
                                </li>
                            {% endfor %}
                        </ul>
                    {% endif %}
                </div>
            {% endfor %}
        </section>
        
        <section class="sidebar">
            <h3>Recent Posts</h3>
            {% for post in recent_posts %}
                <article class="post-preview">
                    <h4><a href="{{ post.url }}">{{ post.title }}</a></h4>
                    <p>{{ post.excerpt|truncate(150) }}</p>
                    <div class="meta">
                        By {{ post.author.name }} on {{ post.date|date('%Y-%m-%d') }}
                    </div>
                </article>
            {% endfor %}
        </section>
    </main>
    
    <footer>
        <div class="footer-content">
            {% for column in footer_columns %}
                <div class="footer-column">
                    <h4>{{ column.title }}</h4>
                    <ul>
                        {% for link in column.links %}
                            <li><a href="{{ link.url }}">{{ link.text }}</a></li>
                        {% endfor %}
                    </ul>
                </div>
            {% endfor %}
        </div>
        <div class="copyright">
            &copy; {{ current_year }} {{ site_name }}. All rights reserved.
        </div>
    </footer>
</body>
</html>`

	tmpl, err := env.FromString(largeTemplate)
	if err != nil {
		b.Fatalf("Failed to compile large template: %v", err)
	}

	// Create large dataset
	data := map[string]interface{}{
		"page_title":       "Large Page Test",
		"page_description": "This is a comprehensive page with lots of content to test template rendering performance",
		"site_name":        "Benchmark Site",
		"current_year":     2024,
		"hero": map[string]interface{}{
			"title":       "Welcome to Performance Testing",
			"description": "Testing template rendering performance with large datasets",
		},
		"navigation":       generateNavItems(10),
		"content_sections": generateContentSections(5, 20),
		"recent_posts":     generatePosts(15),
		"footer_columns":   generateFooterColumns(4, 5),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render large template: %v", err)
		}
	}
}

// Benchmark Inheritance Performance
func BenchmarkInheritancePerformance(b *testing.B) {
	// Create temporary directory for templates
	tmpDir, err := os.MkdirTemp("", "jinja2_inheritance_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create inheritance chain templates
	templates := map[string]string{
		"base.html": `<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default{% endblock %}</title>
    {% block extra_head %}{% endblock %}
</head>
<body class="{% block body_class %}{% endblock %}">
    <header>{% block header %}Default Header{% endblock %}</header>
    <nav>{% block navigation %}{% endblock %}</nav>
    <main>
        {% block breadcrumb %}{% endblock %}
        <div class="content">
            {% block content %}Default Content{% endblock %}
        </div>
        {% block sidebar %}{% endblock %}
    </main>
    <footer>{% block footer %}Default Footer{% endblock %}</footer>
    {% block extra_js %}{% endblock %}
</body>
</html>`,

		"layout.html": `{% extends "base.html" %}
{% block header %}Site Header - {{ site_name }}{% endblock %}
{% block navigation %}
<nav>
    {% for item in nav_items %}
        <a href="{{ item.url }}">{{ item.title }}</a>
    {% endfor %}
</nav>
{% endblock %}
{% block footer %}Site Footer - {{ current_year }}{% endblock %}`,

		"page.html": `{% extends "layout.html" %}
{% block title %}{{ page.title }} - {{ super() }}{% endblock %}
{% block body_class %}{{ page.type }}-page{% endblock %}
{% block breadcrumb %}
<nav class="breadcrumb">
    {% for crumb in breadcrumbs %}
        {% if not loop.last %}
            <a href="{{ crumb.url }}">{{ crumb.title }}</a> &gt; 
        {% else %}
            <span>{{ crumb.title }}</span>
        {% endif %}
    {% endfor %}
</nav>
{% endblock %}
{% block content %}
<h1>{{ page.title }}</h1>
<div class="page-content">{{ page.content|safe }}</div>
{% endblock %}
{% block sidebar %}
<aside>
    {% for widget in sidebar_widgets %}
        <div class="widget">
            <h3>{{ widget.title }}</h3>
            <div class="widget-content">{{ widget.content }}</div>
        </div>
    {% endfor %}
</aside>
{% endblock %}`,
	}

	// Write templates to files
	for name, content := range templates {
		filePath := filepath.Join(tmpDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to write template %s: %v", name, err)
		}
	}

	// Create environment with filesystem loader
	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(false))

	// Pre-compile the page template
	tmpl, err := env.GetTemplate("page.html")
	if err != nil {
		b.Fatalf("Failed to load page template: %v", err)
	}

	// Benchmark data
	data := map[string]interface{}{
		"site_name":    "Benchmark Site",
		"current_year": 2024,
		"page": map[string]interface{}{
			"title":   "Performance Test Page",
			"type":    "test",
			"content": "<p>This is a test page for benchmarking template inheritance performance.</p>",
		},
		"nav_items": []interface{}{
			map[string]interface{}{"url": "/", "title": "Home"},
			map[string]interface{}{"url": "/about", "title": "About"},
			map[string]interface{}{"url": "/contact", "title": "Contact"},
		},
		"breadcrumbs": []interface{}{
			map[string]interface{}{"url": "/", "title": "Home"},
			map[string]interface{}{"url": "/category", "title": "Category"},
			map[string]interface{}{"url": "", "title": "Current Page"},
		},
		"sidebar_widgets": []interface{}{
			map[string]interface{}{"title": "Recent Posts", "content": "List of recent posts"},
			map[string]interface{}{"title": "Categories", "content": "List of categories"},
			map[string]interface{}{"title": "Tags", "content": "Tag cloud"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render inheritance template: %v", err)
		}
	}
}

// Benchmark Filter Chain Performance
func BenchmarkFilterChainPerformance(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	filterTests := []struct {
		name     string
		template string
		data     map[string]interface{}
	}{
		{
			name:     "simple_filter",
			template: "{{ text|upper }}",
			data:     map[string]interface{}{"text": "hello world"},
		},
		{
			name:     "double_filter",
			template: "{{ text|upper|trim }}",
			data:     map[string]interface{}{"text": "  hello world  "},
		},
		{
			name:     "triple_filter",
			template: "{{ text|trim|upper|reverse }}",
			data:     map[string]interface{}{"text": "  hello world  "},
		},
		{
			name:     "complex_filter_chain",
			template: "{{ items|selectattr('active')|map('name')|sort|join(', ')|upper }}",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "alice", "active": true},
					map[string]interface{}{"name": "bob", "active": false},
					map[string]interface{}{"name": "charlie", "active": true},
				},
			},
		},
	}

	// Pre-compile templates
	templates := make([]*miya.Template, len(filterTests))
	for i, test := range filterTests {
		tmpl, err := env.FromString(test.template)
		if err != nil {
			b.Fatalf("Failed to compile template %s: %v", test.name, err)
		}
		templates[i] = tmpl
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testIdx := i % len(filterTests)
		test := filterTests[testIdx]
		tmpl := templates[testIdx]

		_, err := tmpl.Render(miya.NewContextFrom(test.data))
		if err != nil {
			b.Fatalf("Failed to render template %s: %v", test.name, err)
		}
	}
}

// Benchmark Concurrent Template Rendering
func BenchmarkConcurrentRendering(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	template := `
{% for user in users %}
    <div class="user-{{ loop.index }}">
        <h3>{{ user.name|title }}</h3>
        <p>Age: {{ user.age }}</p>
        {% if user.skills %}
            <ul>
                {% for skill in user.skills %}
                    <li>{{ skill|upper }}</li>
                {% endfor %}
            </ul>
        {% endif %}
    </div>
{% endfor %}`

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to compile template: %v", err)
	}

	// Generate test data
	users := make([]interface{}, 50)
	for i := 0; i < 50; i++ {
		users[i] = map[string]interface{}{
			"name": fmt.Sprintf("User%d", i),
			"age":  20 + (i % 40),
			"skills": []interface{}{
				fmt.Sprintf("skill%d", i%5),
				fmt.Sprintf("skill%d", (i+1)%5),
			},
		}
	}

	data := map[string]interface{}{"users": users}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tmpl.Render(miya.NewContextFrom(data))
			if err != nil {
				b.Fatalf("Failed to render template: %v", err)
			}
		}
	})
}

// Benchmark Memory Usage Patterns
func BenchmarkMemoryUsagePatterns(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	// Template with various memory-intensive operations
	template := `
{% set processed_items = [] %}
{% for item in items %}
    {% set item_data = {
        'id': item.id,
        'processed_name': item.name|upper|trim,
        'category': item.category|default('uncategorized'),
        'tags': item.tags|join(', ') if item.tags else 'no tags'
    } %}
    {% set _ = processed_items.append(item_data) %}
{% endfor %}

<div class="results">
    <h1>Processed {{ processed_items|length }} items</h1>
    {% for item in processed_items %}
        <div class="item-{{ loop.index0 }}">
            <h2>{{ item.processed_name }}</h2>
            <p>Category: {{ item.category }}</p>
            <p>Tags: {{ item.tags }}</p>
        </div>
    {% endfor %}
</div>

<div class="summary">
    <p>Categories found: {{ processed_items|map(attribute='category')|unique|list|join(', ') }}</p>
    <p>Total items: {{ processed_items|length }}</p>
</div>`

	tmpl, err := env.FromString(template)
	if err != nil {
		b.Fatalf("Failed to compile memory benchmark template: %v", err)
	}

	// Generate data with varying sizes
	generateData := func(size int) map[string]interface{} {
		items := make([]interface{}, size)
		for i := 0; i < size; i++ {
			tags := make([]interface{}, i%5+1)
			for j := range tags {
				tags[j] = fmt.Sprintf("tag%d", j)
			}

			items[i] = map[string]interface{}{
				"id":       i,
				"name":     fmt.Sprintf("  Item %d  ", i),
				"category": fmt.Sprintf("category%d", i%3),
				"tags":     tags,
			}
		}
		return map[string]interface{}{"items": items}
	}

	dataSizes := []int{10, 50, 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := dataSizes[i%len(dataSizes)]
		data := generateData(size)

		_, err := tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render memory benchmark template: %v", err)
		}
	}
}

// Benchmark Template Caching Performance
func BenchmarkTemplateCaching(b *testing.B) {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	templates := []string{
		"Template 1: {{ value1 }}",
		"Template 2: {{ value2|upper }}",
		"Template 3: {% if condition %}{{ message }}{% endif %}",
		"Template 4: {% for item in items %}{{ item }}{% endfor %}",
		"Template 5: {{ data|length }} items found",
	}

	data := map[string]interface{}{
		"value1":    "test1",
		"value2":    "test2",
		"condition": true,
		"message":   "conditional message",
		"items":     []interface{}{"a", "b", "c"},
		"data":      []interface{}{1, 2, 3, 4, 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		templateStr := templates[i%len(templates)]

		// This will test template caching - same template strings should be cached
		tmpl, err := env.FromString(templateStr)
		if err != nil {
			b.Fatalf("Failed to compile template: %v", err)
		}

		_, err = tmpl.Render(miya.NewContextFrom(data))
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

// Benchmark Error Handling Performance
func BenchmarkErrorHandling(b *testing.B) {
	env := miya.NewEnvironment(miya.WithStrictUndefined(true))

	errorTemplates := []struct {
		template string
		data     map[string]interface{}
	}{
		{
			template: "{{ undefined_variable }}",
			data:     map[string]interface{}{},
		},
		{
			template: "{{ user.nonexistent_field }}",
			data:     map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
		},
		{
			template: "{{ items[100] }}",
			data:     map[string]interface{}{"items": []interface{}{"a", "b", "c"}},
		},
		{
			template: "{{ value|nonexistent_filter }}",
			data:     map[string]interface{}{"value": "test"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		test := errorTemplates[i%len(errorTemplates)]

		tmpl, err := env.FromString(test.template)
		if err != nil {
			// Some templates may fail at parse time
			continue
		}

		// This should produce an error, which we expect
		_, err = tmpl.Render(miya.NewContextFrom(test.data))
		if err == nil {
			b.Fatalf("Expected error but none occurred for template: %s", test.template)
		}
	}
}

// Helper functions for generating test data

func generateNavItems(count int) []interface{} {
	items := make([]interface{}, count)
	for i := 0; i < count; i++ {
		items[i] = map[string]interface{}{
			"url":    fmt.Sprintf("/page%d", i),
			"title":  fmt.Sprintf("Page %d", i),
			"active": i == 0,
		}
	}
	return items
}

func generateContentSections(sectionCount, itemsPerSection int) []interface{} {
	sections := make([]interface{}, sectionCount)
	for i := 0; i < sectionCount; i++ {
		items := make([]interface{}, itemsPerSection)
		for j := 0; j < itemsPerSection; j++ {
			tags := make([]interface{}, j%3+1)
			for k := range tags {
				tags[k] = fmt.Sprintf("tag%d", k)
			}

			items[j] = map[string]interface{}{
				"name":        fmt.Sprintf("Item %d-%d", i, j),
				"description": fmt.Sprintf("This is the description for item %d in section %d", j, i),
				"tags":        tags,
			}
		}

		sections[i] = map[string]interface{}{
			"title": fmt.Sprintf("Section %d", i),
			"items": items,
		}
	}
	return sections
}

func generatePosts(count int) []interface{} {
	posts := make([]interface{}, count)
	for i := 0; i < count; i++ {
		posts[i] = map[string]interface{}{
			"title":   fmt.Sprintf("Blog Post %d", i),
			"excerpt": fmt.Sprintf("This is the excerpt for blog post %d. It contains some sample text to test rendering performance.", i),
			"url":     fmt.Sprintf("/blog/post-%d", i),
			"date":    "2024-01-15",
			"author":  map[string]interface{}{"name": fmt.Sprintf("Author %d", i%5)},
		}
	}
	return posts
}

func generateFooterColumns(columnCount, linksPerColumn int) []interface{} {
	columns := make([]interface{}, columnCount)
	for i := 0; i < columnCount; i++ {
		links := make([]interface{}, linksPerColumn)
		for j := 0; j < linksPerColumn; j++ {
			links[j] = map[string]interface{}{
				"url":  fmt.Sprintf("/footer-link-%d-%d", i, j),
				"text": fmt.Sprintf("Link %d-%d", i, j),
			}
		}

		columns[i] = map[string]interface{}{
			"title": fmt.Sprintf("Column %d", i),
			"links": links,
		}
	}
	return columns
}
