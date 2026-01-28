// Step 4: Template Inheritance
// =============================
// Learn how to create reusable layouts with blocks.
//
// Run: go run ./examples/tutorial/step4_inheritance

package main

import (
	"fmt"
	"log"

	"github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

// SimpleTemplateParser implements loader.TemplateParser interface
type SimpleTemplateParser struct {
	env *miya.Environment
}

func NewSimpleTemplateParser(env *miya.Environment) *SimpleTemplateParser {
	return &SimpleTemplateParser{env: env}
}

func (stp *SimpleTemplateParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	template, err := stp.env.FromString(content)
	if err != nil {
		return nil, err
	}

	if templateNode, ok := template.AST().(*parser.TemplateNode); ok {
		templateNode.Name = name
		return templateNode, nil
	}

	return nil, fmt.Errorf("failed to extract template node")
}

func main() {
	fmt.Println("=== Step 4: Template Inheritance ===")
	fmt.Println()

	// Template inheritance requires a loader so templates can reference each other.
	// We'll use a StringLoader to define templates in code.

	// 1. Define a base template
	// Base templates define the structure with {% block %} placeholders.
	baseTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
    {% block head %}{% endblock %}
</head>
<body>
    <header>
        {% block header %}
        <h1>My Website</h1>
        <nav>Home | About | Contact</nav>
        {% endblock %}
    </header>

    <main>
        {% block content %}
        <p>Default content goes here.</p>
        {% endblock %}
    </main>

    <footer>
        {% block footer %}
        <p>Copyright 2024</p>
        {% endblock %}
    </footer>
</body>
</html>`

	// 2. Define a child template
	// Child templates use {% extends %} and override specific blocks.
	homeTemplate := `{% extends "base.html" %}

{% block title %}Home - My Website{% endblock %}

{% block content %}
<h2>Welcome!</h2>
<p>This is the home page content.</p>
<ul>
{% for item in features %}
    <li>{{ item }}</li>
{% endfor %}
</ul>
{% endblock %}`

	// 3. Another child template demonstrating super()
	// super() includes the parent block's content.
	aboutTemplate := `{% extends "base.html" %}

{% block title %}About Us{% endblock %}

{% block header %}
{{ super() }}
<p class="subtitle">Learn more about us</p>
{% endblock %}

{% block content %}
<h2>About Our Company</h2>
<p>{{ description }}</p>
{% endblock %}

{% block footer %}
{{ super() }}
<p>Contact: {{ email }}</p>
{% endblock %}`

	// 4. Create an environment and string loader with our templates
	env := miya.NewEnvironment()
	templateParser := NewSimpleTemplateParser(env)
	stringLoader := loader.NewStringLoader(templateParser)
	stringLoader.AddTemplate("base.html", baseTemplate)
	stringLoader.AddTemplate("home.html", homeTemplate)
	stringLoader.AddTemplate("about.html", aboutTemplate)

	// 5. Set the loader on the environment
	env.SetLoader(stringLoader)

	// 6. Render the home page
	fmt.Println("Example 1 - Home Page (basic inheritance):")
	fmt.Println("-------------------------------------------")

	homeTmpl, err := env.GetTemplate("home.html")
	if err != nil {
		log.Fatal(err)
	}

	homeCtx := miya.NewContextFrom(map[string]interface{}{
		"features": []string{"Fast rendering", "Jinja2 compatible", "Easy to use"},
	})

	homeOutput, err := homeTmpl.Render(homeCtx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(homeOutput)
	fmt.Println()

	// 7. Render the about page (demonstrates super())
	fmt.Println("Example 2 - About Page (using super()):")
	fmt.Println("----------------------------------------")

	aboutTmpl, err := env.GetTemplate("about.html")
	if err != nil {
		log.Fatal(err)
	}

	aboutCtx := miya.NewContextFrom(map[string]interface{}{
		"description": "We build fast template engines for Go developers.",
		"email":       "hello@example.com",
	})

	aboutOutput, err := aboutTmpl.Render(aboutCtx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(aboutOutput)
	fmt.Println()

	// 8. Multi-level inheritance example
	fmt.Println("Example 3 - Multi-level inheritance:")
	fmt.Println("-------------------------------------")

	// Base -> Section -> Page
	sectionTemplate := `{% extends "base.html" %}

{% block header %}
{{ super() }}
<nav class="section-nav">Section: {{ section_name }}</nav>
{% endblock %}`

	articleTemplate := `{% extends "section.html" %}

{% block title %}{{ article.title }} - {{ section_name }}{% endblock %}

{% block content %}
<article>
    <h2>{{ article.title }}</h2>
    <p class="meta">By {{ article.author }}</p>
    <div>{{ article.body }}</div>
</article>
{% endblock %}`

	stringLoader.AddTemplate("section.html", sectionTemplate)
	stringLoader.AddTemplate("article.html", articleTemplate)

	articleTmpl, err := env.GetTemplate("article.html")
	if err != nil {
		log.Fatal(err)
	}

	articleCtx := miya.NewContextFrom(map[string]interface{}{
		"section_name": "Technology",
		"article": map[string]interface{}{
			"title":  "Getting Started with Miya",
			"author": "Jane Developer",
			"body":   "Miya is a powerful template engine for Go...",
		},
	})

	articleOutput, err := articleTmpl.Render(articleCtx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(articleOutput)

	// Key Takeaways:
	// - Base templates define structure with {% block name %}...{% endblock %}
	// - Child templates use {% extends "parent.html" %}
	// - Override blocks by redefining them in child templates
	// - Use {{ super() }} to include parent's block content
	// - Inheritance can be multi-level (grandparent -> parent -> child)
	// - Requires a Loader (FileSystemLoader or MemoryLoader)

	fmt.Println("=== Step 4 Complete ===")
}
