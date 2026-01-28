// Step 5: Macros & Components
// ============================
// Learn how to create reusable template components with macros.
//
// Run: go run ./examples/tutorial/step5_macros

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
	fmt.Println("=== Step 5: Macros & Components ===")
	fmt.Println()

	// 1. Basic macro definition and usage
	// Macros are like functions - they take parameters and return rendered content.
	template1 := `
{# Define a macro #}
{% macro greeting(name, title="Mr.") %}
Hello, {{ title }} {{ name }}!
{% endmacro %}

{# Use the macro #}
{{ greeting("Smith") }}
{{ greeting("Johnson", "Dr.") }}
{{ greeting("Alice", "Ms.") }}`

	env := miya.NewEnvironment()
	tmpl1, err := env.FromString(template1)
	if err != nil {
		log.Fatal(err)
	}

	output1, _ := tmpl1.Render(miya.NewContext())
	fmt.Println("Example 1 - Basic macros:")
	fmt.Println(output1)
	fmt.Println()

	// 2. Macros for UI components
	template2 := `
{# Button component macro #}
{% macro button(text, type="primary", size="md") %}
<button class="btn btn-{{ type }} btn-{{ size }}">{{ text }}</button>
{% endmacro %}

{# Alert component macro #}
{% macro alert(message, level="info") %}
<div class="alert alert-{{ level }}">
    {{ message }}
</div>
{% endmacro %}

{# Use the components #}
{{ button("Submit") }}
{{ button("Cancel", "secondary") }}
{{ button("Delete", "danger", "sm") }}

{{ alert("Operation successful!", "success") }}
{{ alert("Please review your input.", "warning") }}`

	tmpl2, _ := env.FromString(template2)
	output2, _ := tmpl2.Render(miya.NewContext())
	fmt.Println("Example 2 - UI components:")
	fmt.Println(output2)
	fmt.Println()

	// 3. Macros with logic
	template3 := `
{% macro user_badge(user) %}
<span class="badge badge-{{ 'success' if user.active else 'secondary' }}">
    {{ user.name }}
    {% if user.role == "admin" %}(Admin){% endif %}
</span>
{% endmacro %}

{% for user in users %}
{{ user_badge(user) }}
{% endfor %}`

	tmpl3, _ := env.FromString(template3)
	ctx3 := miya.NewContextFrom(map[string]interface{}{
		"users": []map[string]interface{}{
			{"name": "Alice", "active": true, "role": "admin"},
			{"name": "Bob", "active": true, "role": "user"},
			{"name": "Charlie", "active": false, "role": "user"},
		},
	})

	output3, _ := tmpl3.Render(ctx3)
	fmt.Println("Example 3 - Macros with logic:")
	fmt.Println(output3)
	fmt.Println()

	// 4. Importing macros from separate templates
	// In real projects, macros are defined in separate files and imported.
	fmt.Println("Example 4 - Importing macros:")
	fmt.Println("------------------------------")

	// Create a macro library
	macroLibrary := `
{% macro input(name, type="text", placeholder="", required=false) %}
<input type="{{ type }}" name="{{ name }}" placeholder="{{ placeholder }}"
       class="form-control"{% if required %} required{% endif %}>
{% endmacro %}

{% macro select(name, options, selected="") %}
<select name="{{ name }}" class="form-control">
{% for opt in options %}
    <option value="{{ opt.value }}"{% if opt.value == selected %} selected{% endif %}>
        {{ opt.label }}
    </option>
{% endfor %}
</select>
{% endmacro %}

{% macro form_group(label, field) %}
<div class="form-group">
    <label>{{ label }}</label>
    {{ field }}
</div>
{% endmacro %}`

	// Template that uses the macro library
	formTemplate := `{% import "forms.html" as forms %}

<form>
    {{ forms.input("username", placeholder="Enter username", required=true) }}
    {{ forms.input("email", type="email", placeholder="Enter email") }}
    {{ forms.input("password", type="password", required=true) }}

    {{ forms.select("country", countries, selected="us") }}
</form>`

	// Set up string loader
	envWithLoader := miya.NewEnvironment()
	templateParser := NewSimpleTemplateParser(envWithLoader)
	stringLoader := loader.NewStringLoader(templateParser)
	stringLoader.AddTemplate("forms.html", macroLibrary)
	stringLoader.AddTemplate("register.html", formTemplate)
	envWithLoader.SetLoader(stringLoader)
	tmpl4, err := envWithLoader.GetTemplate("register.html")
	if err != nil {
		log.Fatal(err)
	}

	ctx4 := miya.NewContextFrom(map[string]interface{}{
		"countries": []map[string]interface{}{
			{"value": "us", "label": "United States"},
			{"value": "uk", "label": "United Kingdom"},
			{"value": "ca", "label": "Canada"},
		},
	})

	output4, _ := tmpl4.Render(ctx4)
	fmt.Println(output4)
	fmt.Println()

	// 5. Selective import with 'from'
	fmt.Println("Example 5 - Selective import:")
	fmt.Println("------------------------------")

	selectiveTemplate := `{% from "forms.html" import input, select %}

{# Now use directly without namespace #}
{{ input("search", placeholder="Search...") }}
{{ select("sort", sort_options) }}`

	stringLoader.AddTemplate("search.html", selectiveTemplate)
	tmpl5, _ := envWithLoader.GetTemplate("search.html")

	ctx5 := miya.NewContextFrom(map[string]interface{}{
		"sort_options": []map[string]interface{}{
			{"value": "date", "label": "Date"},
			{"value": "name", "label": "Name"},
			{"value": "price", "label": "Price"},
		},
	})

	output5, _ := tmpl5.Render(ctx5)
	fmt.Println(output5)
	fmt.Println()

	// 6. Building a component library pattern
	fmt.Println("Example 6 - Complete component usage:")
	fmt.Println("--------------------------------------")

	cardMacros := `
{% macro card(title, footer="") %}
<div class="card">
    <div class="card-header">{{ title }}</div>
    <div class="card-body">
        {{ caller() }}
    </div>
    {% if footer %}
    <div class="card-footer">{{ footer }}</div>
    {% endif %}
</div>
{% endmacro %}

{% macro list_group(items) %}
<ul class="list-group">
{% for item in items %}
    <li class="list-group-item">{{ item }}</li>
{% endfor %}
</ul>
{% endmacro %}`

	pageTemplate := `{% import "cards.html" as ui %}

{{ ui.list_group(features) }}

{% call ui.card("User Profile", "Last updated: today") %}
<p>Name: {{ user.name }}</p>
<p>Email: {{ user.email }}</p>
{% endcall %}`

	stringLoader.AddTemplate("cards.html", cardMacros)
	stringLoader.AddTemplate("page.html", pageTemplate)

	tmpl6, _ := envWithLoader.GetTemplate("page.html")
	ctx6 := miya.NewContextFrom(map[string]interface{}{
		"features": []string{"Fast", "Flexible", "Easy"},
		"user": map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		},
	})

	output6, _ := tmpl6.Render(ctx6)
	fmt.Println(output6)

	// Key Takeaways:
	// - Define macros with {% macro name(params) %}...{% endmacro %}
	// - Parameters can have default values: {% macro btn(text, type="primary") %}
	// - Import macros: {% import "file.html" as alias %}
	// - Selective import: {% from "file.html" import macro1, macro2 %}
	// - Use {% call macro() %}content{% endcall %} with {{ caller() }} for wrapped content
	// - Macros are great for UI components: buttons, forms, cards, alerts

	fmt.Println("\n=== Step 5 Complete ===")
	fmt.Println("\nCongratulations! You've completed the Miya Engine tutorial.")
	fmt.Println("Check out examples/features/ for more advanced examples.")
}
