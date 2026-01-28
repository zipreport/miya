package main

import (
	"fmt"
	"log"
	"os"

	miya "github.com/zipreport/miya"
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
	// Example 1: Template inheritance
	fmt.Println("=== Example 1: Template Inheritance ===")
	inheritanceExample()

	// Example 2: Recursive loops
	fmt.Println("\n=== Example 2: Recursive Loops ===")
	recursiveExample()

	// Example 3: Custom filters
	fmt.Println("\n=== Example 3: Custom Filters ===")
	customFilterExample()

	// Example 4: Auto-escaping
	fmt.Println("\n=== Example 4: Auto-escaping ===")
	autoEscapeExample()

	// Example 5: Advanced features
	fmt.Println("\n=== Example 5: Advanced Features ===")
	advancedFeaturesExample()
}

func inheritanceExample() {
	// Create templates directory
	os.MkdirAll("templates", 0755)

	// Create base template
	baseTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    <header>
        {% block header %}
        <h1>Welcome</h1>
        {% endblock %}
    </header>
    
    <main>
        {% block content %}
        <!-- Main content goes here -->
        {% endblock %}
    </main>
    
    <footer>
        {% block footer %}
        <p>&copy; 2024 Example Corp</p>
        {% endblock %}
    </footer>
</body>
</html>`

	// Create child template
	childTemplate := `{% extends "base.html" %}

{% block title %}Home Page{% endblock %}

{% block header %}
    {{ super() }}
    <nav>Home | About | Contact</nav>
{% endblock %}

{% block content %}
    <h2>Welcome to our site!</h2>
    <p>This is the home page content.</p>
{% endblock %}`

	// Save templates
	os.WriteFile("templates/base.html", []byte(baseTemplate), 0644)
	os.WriteFile("templates/home.html", []byte(childTemplate), 0644)

	// Create environment with file loader
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"templates"}, templateParser)
	env.SetLoader(fsLoader)

	// Render child template
	tmpl, err := env.GetTemplate("home.html")
	if err != nil {
		log.Fatal(err)
	}

	ctx := miya.NewContext()
	result, err := tmpl.Render(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

	// Cleanup
	os.RemoveAll("templates")
}

func recursiveExample() {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	// Note: Miya doesn't support Python-style string repetition ("  " * n)
	// We use a workaround with the range function to create indentation
	template := `
Menu Structure:
{% for item in menu recursive %}
{% for _ in range(loop.depth0) %}  {% endfor %}{{ item.title }}
{%- if item.children %}
{{ loop(item.children) }}
{%- endif %}
{% endfor %}`

	ctx := miya.NewContext()
	ctx.Set("menu", []map[string]interface{}{
		{
			"title": "Home",
		},
		{
			"title": "Products",
			"children": []map[string]interface{}{
				{
					"title": "Electronics",
					"children": []map[string]interface{}{
						{"title": "Laptops"},
						{"title": "Phones"},
					},
				},
				{
					"title": "Books",
					"children": []map[string]interface{}{
						{"title": "Fiction"},
						{"title": "Non-fiction"},
					},
				},
			},
		},
		{
			"title": "About",
		},
	})

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func customFilterExample() {
	env := miya.NewEnvironment()

	// Add custom filter
	env.AddFilter("reverse", func(value interface{}, args ...interface{}) (interface{}, error) {
		str, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("reverse filter requires a string")
		}

		runes := []rune(str)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	})

	// Add custom test
	env.AddTest("palindrome", func(value interface{}, args ...interface{}) (bool, error) {
		str, ok := value.(string)
		if !ok {
			return false, nil
		}

		runes := []rune(str)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			if runes[i] != runes[j] {
				return false, nil
			}
		}
		return true, nil
	})

	template := `
Original: {{ text }}
Reversed: {{ text|reverse }}
Is "{{ word1 }}" a palindrome? {% if word1 is palindrome %}Yes{% else %}No{% endif %}
Is "{{ word2 }}" a palindrome? {% if word2 is palindrome %}Yes{% else %}No{% endif %}`

	ctx := miya.NewContext()
	ctx.Set("text", "Hello World")
	ctx.Set("word1", "racecar")
	ctx.Set("word2", "hello")

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func autoEscapeExample() {
	// Environment with auto-escape enabled (default)
	envEscaped := miya.NewEnvironment()

	// Environment with auto-escape disabled
	envRaw := miya.NewEnvironment(miya.WithAutoEscape(false))

	template := `
User input: {{ user_input }}
HTML content: {{ html_content }}
Safe content: {{ safe_content|safe }}
Escaped content: {{ content|escape }}`

	ctx := miya.NewContext()
	ctx.Set("user_input", "<script>alert('XSS')</script>")
	ctx.Set("html_content", "<b>Bold text</b>")
	ctx.Set("safe_content", "<i>Italic text</i>")
	ctx.Set("content", "<span>Some text</span>")

	fmt.Println("With auto-escape:")
	result, err := envEscaped.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

	fmt.Println("\nWithout auto-escape:")
	result, err = envRaw.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func advancedFeaturesExample() {
	env := miya.NewEnvironment(miya.WithAutoEscape(false))

	template := `
{# This is a comment that won't appear in output #}

{# Set variables #}
{% set name = "Alice" %}
{% set age = 30 %}
{% set items = ["apple", "banana", "orange"] %}

Name: {{ name }}, Age: {{ age }}

{# Conditional expressions (ternary) #}
Status: {{ "Adult" if age >= 18 else "Minor" }}

{# List comprehensions #}
{% set doubled = [x * 2 for x in [1, 2, 3, 4, 5]] %}
Doubled: {{ doubled|join(", ") }}

{# Dictionary comprehensions #}
{% set fruit_lengths = {fruit: fruit|length for fruit in items} %}
Fruit lengths:
{% for fruit, length in fruit_lengths.items() %}
  {{ fruit }}: {{ length }} characters
{% endfor %}

{# With statement #}
{% with greeting = "Hello", subject = name %}
  {{ greeting }}, {{ subject }}!
{% endwith %}

{# Global functions #}
Range: {% for i in range(1, 6) %}{{ i }}{% if not loop.last %}, {% endif %}{% endfor %}

{# Cycler function #}
{% set cycle = cycler("odd", "even") %}
Cycling: {{ cycle.next() }}, {{ cycle.next() }}, {{ cycle.next() }}

{# Namespace #}
{% set ns = namespace(total=0) %}
{% for price in [10, 20, 30] %}
  {% set ns.total = ns.total + price %}
{% endfor %}
Total: {{ ns.total }}

{# Raw block - no processing #}
{% raw %}
  This {{ variable }} won't be processed
  Neither will {% this %} block tag
{% endraw %}`

	ctx := miya.NewContext()

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
