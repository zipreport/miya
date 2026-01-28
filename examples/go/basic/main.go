package main

import (
	"fmt"
	"log"

	miya "github.com/zipreport/miya"
)

func main() {
	// Create a new Jinja2 environment
	env := miya.NewEnvironment()

	// Example 1: Simple variable substitution
	fmt.Println("=== Example 1: Simple Variables ===")
	simpleExample(env)

	// Example 2: Using filters
	fmt.Println("\n=== Example 2: Filters ===")
	filterExample(env)

	// Example 3: Control structures
	fmt.Println("\n=== Example 3: Control Structures ===")
	controlExample(env)

	// Example 4: Loops
	fmt.Println("\n=== Example 4: Loops ===")
	loopExample(env)

	// Example 5: Macros
	fmt.Println("\n=== Example 5: Macros ===")
	macroExample(env)
}

func simpleExample(env *miya.Environment) {
	template := `Hello {{ name }}! Welcome to {{ place }}.`

	ctx := miya.NewContext()
	ctx.Set("name", "Alice")
	ctx.Set("place", "Wonderland")

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func filterExample(env *miya.Environment) {
	template := `Original: {{ text }}
Upper: {{ text|upper }}
Lower: {{ text|lower }}
Title: {{ text|title }}
Length: {{ text|length }}
Tags: {{ tags|join(", ") }}
First tag: {{ tags|first }}
Last tag: {{ tags|last }}`

	ctx := miya.NewContext()
	ctx.Set("text", "Hello WORLD from Jinja2")
	ctx.Set("tags", []string{"go", "jinja2", "templates"})

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func controlExample(env *miya.Environment) {
	template := `
User: {{ user.name }}
{% if user.is_admin %}
  Status: Administrator
{% elif user.is_premium %}
  Status: Premium User
{% else %}
  Status: Regular User
{% endif %}

{% if user.age >= 18 %}
  Access: Full
{% else %}
  Access: Restricted
{% endif %}

{% if items %}
  Items found: {{ items|length }}
{% else %}
  No items found
{% endif %}`

	ctx := miya.NewContext()
	ctx.Set("user", map[string]interface{}{
		"name":       "John Doe",
		"is_admin":   false,
		"is_premium": true,
		"age":        25,
	})
	ctx.Set("items", []string{"item1", "item2", "item3"})

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func loopExample(env *miya.Environment) {
	template := `
Products:
{% for product in products %}
  {{ loop.index }}. {{ product.name }} - ${{ product.price }}
     In stock: {% if product.in_stock %}Yes{% else %}No{% endif %}
     {% if loop.first %}(Featured){% endif %}
     {% if loop.last %}(Last item){% endif %}
{% endfor %}

Categories:
{% for category in categories %}
  {{ category.name }}:
  {% for item in category.items %}
    - {{ item }}
  {% endfor %}
{% endfor %}

Even numbers:
{% for num in numbers %}
  {% if num is even %}
    {{ num }}{% if not loop.last %}, {% endif %}
  {% endif %}
{% endfor %}`

	ctx := miya.NewContext()
	ctx.Set("products", []map[string]interface{}{
		{"name": "Laptop", "price": 999.99, "in_stock": true},
		{"name": "Mouse", "price": 25.50, "in_stock": true},
		{"name": "Keyboard", "price": 75.00, "in_stock": false},
	})
	ctx.Set("categories", []map[string]interface{}{
		{
			"name":  "Electronics",
			"items": []string{"Laptop", "Phone", "Tablet"},
		},
		{
			"name":  "Books",
			"items": []string{"Fiction", "Non-fiction", "Comics"},
		},
	})
	ctx.Set("numbers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func macroExample(env *miya.Environment) {
	template := `
{% macro greeting(name, title="") %}
  {% if title %}
    Hello {{ title }} {{ name }}!
  {% else %}
    Hello {{ name }}!
  {% endif %}
{% endmacro %}

{% macro product_card(product) %}
<div class="product">
  <h3>{{ product.name }}</h3>
  <p>Price: ${{ product.price }}</p>
  <p>{{ product.description }}</p>
</div>
{% endmacro %}

{{ greeting("Alice") }}
{{ greeting("Bob", "Dr.") }}

{% for p in products %}
{{ product_card(p) }}
{% endfor %}`

	ctx := miya.NewContext()
	ctx.Set("products", []map[string]interface{}{
		{
			"name":        "Go Book",
			"price":       39.99,
			"description": "Learn Go programming",
		},
		{
			"name":        "Template Guide",
			"price":       29.99,
			"description": "Master template engines",
		},
	})

	result, err := env.RenderString(template, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
