package main

import (
	"fmt"
	"strings"

	"github.com/zipreport/miya"
)

func main() {
	// Create environment
	env := miya.NewEnvironment(
		miya.WithAutoEscape(true),
		miya.WithTrimBlocks(true),
		miya.WithLstripBlocks(true),
	)

	fmt.Println("Note: The complex templates (showcase.html, dashboard.html) use template")
	fmt.Println("inheritance ({% extends %}) which requires a FileSystemLoader.")
	fmt.Println("Below are standalone examples that demonstrate Miya features:")
	fmt.Println()

	// Demo complex features with standalone templates
	demoComplexFeatures(env)

	// Also demonstrate simple template rendering
	demoSimpleTemplates(env)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("All examples completed!")
	fmt.Println(strings.Repeat("=", 80))
}

func demoSimpleTemplates(env *miya.Environment) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Simple Template Examples")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Example 1: Basic variable substitution
	tmpl1, _ := env.FromString("Hello, {{ name|title }}!")
	ctx1 := miya.NewContextFrom(map[string]interface{}{"name": "alice"})
	output1, _ := tmpl1.Render(ctx1)
	fmt.Printf("1. Basic variable: %s\n", output1)

	// Example 2: Loop with filters
	tmpl2, _ := env.FromString("{% for item in items %}{{ loop.index }}. {{ item|upper }} {% endfor %}")
	ctx2 := miya.NewContextFrom(map[string]interface{}{
		"items": []string{"apple", "banana", "cherry"},
	})
	output2, _ := tmpl2.Render(ctx2)
	fmt.Printf("2. Loop: %s\n", output2)

	// Example 3: Conditional
	tmpl3, _ := env.FromString("{% if score > 90 %}A{% elif score > 80 %}B{% else %}C{% endif %}")
	ctx3 := miya.NewContextFrom(map[string]interface{}{"score": 95})
	output3, _ := tmpl3.Render(ctx3)
	fmt.Printf("3. Conditional: Grade = %s\n", output3)

	// Example 4: Filter chaining
	tmpl4, _ := env.FromString("{{ text|trim|upper|truncate(20) }}")
	ctx4 := miya.NewContextFrom(map[string]interface{}{
		"text": "  this is a long text that will be truncated  ",
	})
	output4, _ := tmpl4.Render(ctx4)
	fmt.Printf("4. Filter chain: %s\n", output4)

	// Example 5: Tests
	tmpl5, _ := env.FromString("{% if value is even %}Even{% else %}Odd{% endif %}")
	ctx5 := miya.NewContextFrom(map[string]interface{}{"value": 42})
	output5, _ := tmpl5.Render(ctx5)
	fmt.Printf("5. Tests: 42 is %s\n", output5)

	fmt.Println()
}

func demoComplexFeatures(env *miya.Environment) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("Complex Feature Demonstrations")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Demo 1: Macros
	macroTemplate := `
{%- macro card(title, content, theme='light') -%}
<div class="card card-{{ theme }}">
	<h3>{{ title|title }}</h3>
	<p>{{ content }}</p>
</div>
{%- endmacro -%}

{{ card("Welcome", "This is a macro demo") }}
{{ card("Dark Card", "With custom theme", theme='dark') }}
`
	tmpl, _ := env.FromString(macroTemplate)
	output, _ := tmpl.Render(miya.NewContext())
	fmt.Println("1. Macros with default parameters:")
	fmt.Println(output)

	// Demo 2: Loop variables
	loopTemplate := `
{% for item in items -%}
{{ loop.index }}. {{ item }}
   {%- if loop.first %} (first){% endif %}
   {%- if loop.last %} (last){% endif %}
   - index0: {{ loop.index0 }}, revindex: {{ loop.revindex }}
{% endfor %}`
	tmpl2, _ := env.FromString(loopTemplate)
	ctx2 := miya.NewContextFrom(map[string]interface{}{
		"items": []string{"alpha", "beta", "gamma"},
	})
	output2, _ := tmpl2.Render(ctx2)
	fmt.Println("\n2. Loop variables (index, first, last, revindex):")
	fmt.Println(output2)

	// Demo 3: Nested loops with depth
	nestedTemplate := `
{% for i in range(1, 3) -%}
Outer {{ i }} (depth {{ loop.depth }}):
  {% for j in range(1, 3) -%}
  Inner {{ i }}.{{ j }} (depth {{ loop.depth }})
  {% endfor %}
{%- endfor %}`
	tmpl3, _ := env.FromString(nestedTemplate)
	output3, _ := tmpl3.Render(miya.NewContext())
	fmt.Println("\n3. Nested loops with depth:")
	fmt.Println(output3)

	// Demo 4: Tests
	testTemplate := `
Value: {{ value }}
{% if value is number %}Is a number{% endif %}
{% if value is even %}Is even{% endif %}
{% if value is divisibleby(7) %}Divisible by 7{% endif %}
`
	tmpl4, _ := env.FromString(testTemplate)
	ctx4 := miya.NewContextFrom(map[string]interface{}{"value": 42})
	output4, _ := tmpl4.Render(ctx4)
	fmt.Println("\n4. Tests (is number, is even, divisibleby):")
	fmt.Println(output4)

	// Demo 5: Filter chaining
	filterTemplate := `
Original: "{{ text }}"
Processed: "{{ text|trim|upper|truncate(25) }}"
Reversed: "{{ text|trim|reverse }}"
`
	tmpl5, _ := env.FromString(filterTemplate)
	ctx5 := miya.NewContextFrom(map[string]interface{}{
		"text": "  hello world from zinja2  ",
	})
	output5, _ := tmpl5.Render(ctx5)
	fmt.Println("\n5. Filter chaining:")
	fmt.Println(output5)

	// Demo 6: Namespace
	nsTemplate := `
{% set ns = namespace(count=0, total=0) -%}
{% for num in numbers -%}
  {% set ns.count = ns.count + 1 -%}
  {% set ns.total = ns.total + num -%}
{% endfor -%}
Processed {{ ns.count }} numbers, sum = {{ ns.total }}
`
	tmpl6, _ := env.FromString(nsTemplate)
	ctx6 := miya.NewContextFrom(map[string]interface{}{
		"numbers": []int{10, 20, 30, 40, 50},
	})
	output6, _ := tmpl6.Render(ctx6)
	fmt.Println("\n6. Namespace for mutable state:")
	fmt.Println(output6)

	// Demo 7: Slicing
	sliceTemplate := `
Array: {{ arr }}
First 3: {{ arr[:3] }}
Last 2: {{ arr[-2:] }}
Reversed: {{ arr[::-1] }}
`
	tmpl7, _ := env.FromString(sliceTemplate)
	ctx7 := miya.NewContextFrom(map[string]interface{}{
		"arr": []string{"a", "b", "c", "d", "e"},
	})
	output7, _ := tmpl7.Render(ctx7)
	fmt.Println("\n7. Array slicing:")
	fmt.Println(output7)

	// Demo 8: Global functions
	globalTemplate := `
Range: {{ range(5, 10) }}
Zip: {% for a, b in zip([1,2,3], ['a','b','c']) %}({{ a }},{{ b }}) {% endfor %}
Dict: {{ dict(name='Alice', age=30) }}
`
	tmpl8, _ := env.FromString(globalTemplate)
	output8, _ := tmpl8.Render(miya.NewContext())
	fmt.Println("\n8. Global functions (range, zip, dict):")
	fmt.Println(output8)

	// Demo 9: Filter blocks
	filterBlockTemplate := `
{% filter upper -%}
This entire block
will be uppercased
including {{ variable }}
{%- endfilter %}
`
	tmpl9, _ := env.FromString(filterBlockTemplate)
	ctx9 := miya.NewContextFrom(map[string]interface{}{"variable": "variables"})
	output9, _ := tmpl9.Render(ctx9)
	fmt.Println("\n9. Filter blocks:")
	fmt.Println(output9)

	// Demo 10: Dictionary iteration
	dictTemplate := `
{% for key, value in user -%}
{{ key }}: {{ value }}
{% endfor %}`
	tmpl10, _ := env.FromString(dictTemplate)
	ctx10 := miya.NewContextFrom(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  30,
			"city": "NYC",
		},
	})
	output10, _ := tmpl10.Render(ctx10)
	fmt.Println("\n10. Dictionary iteration:")
	fmt.Println(output10)

	fmt.Println()
}

func getShowcaseContext() map[string]interface{} {
	return map[string]interface{}{
		"site_name":      "Miya Showcase",
		"username":       "alice johnson",
		"role":           "admin",
		"is_active":      true,
		"language":       "en",
		"author":         "Miya Team",
		"copyright_year": 2024,
		"company_name":   "Miya Project",
		"font_family":    "system-ui, sans-serif",
		"bg_color":       "#ffffff",
		"navigation_items": []map[string]interface{}{
			{"title": "Home", "url": "/"},
			{"title": "Features", "url": "/features"},
			{"title": "Documentation", "url": "/docs"},
			{"title": "Examples", "url": "/examples"},
		},
		"messages": []map[string]interface{}{
			{"type": "success", "text": "Welcome to Miya!"},
			{"type": "info", "text": "Explore the features below"},
		},
	}
}

func getDashboardContext() map[string]interface{} {
	return map[string]interface{}{
		"site_name": "Admin Dashboard",
		"current_user": map[string]interface{}{
			"name":   "alice johnson",
			"email":  "alice@example.com",
			"avatar": "/img/alice.jpg",
		},
		"stats": []map[string]interface{}{
			{
				"label":  "Total Sales",
				"value":  12548,
				"prefix": "$",
				"change": 12.5,
				"color":  "#10b981",
			},
			{
				"label":  "New Users",
				"value":  342,
				"change": 8.2,
				"color":  "#3b82f6",
			},
			{
				"label":  "Orders",
				"value":  1429,
				"change": -3.1,
				"color":  "#f59e0b",
			},
			{
				"label":  "Revenue",
				"value":  45678,
				"prefix": "$",
				"change": 15.7,
				"color":  "#8b5cf6",
			},
		},
		"activities": []map[string]interface{}{
			{"user": "john doe", "action": "created", "target": "New Product", "time": "5 min ago"},
			{"user": "jane smith", "action": "updated", "target": "User Profile", "time": "12 min ago"},
			{"user": "bob wilson", "action": "deleted", "target": "Old Record", "time": "1 hour ago"},
			{"user": "alice brown", "action": "uploaded", "target": "Product Image", "time": "2 hours ago"},
			{"user": "charlie green", "action": "commented on", "target": "Issue #42", "time": "3 hours ago"},
		},
		"products": []map[string]interface{}{
			{"name": "Premium Widget", "sales": 1250, "revenue": 24500, "growth": 15.2},
			{"name": "Standard Widget", "sales": 3420, "revenue": 68400, "growth": 8.7},
			{"name": "Basic Widget", "sales": 5680, "revenue": 34080, "growth": -2.3},
			{"name": "Deluxe Widget", "sales": 890, "revenue": 26700, "growth": 22.1},
			{"name": "Mini Widget", "sales": 2340, "revenue": 11700, "growth": 5.4},
		},
		"sales_trend": []map[string]interface{}{
			{"label": "Mon", "value": 120},
			{"label": "Tue", "value": 150},
			{"label": "Wed", "value": 180},
			{"label": "Thu", "value": 140},
			{"label": "Fri", "value": 200},
			{"label": "Sat", "value": 170},
			{"label": "Sun", "value": 130},
		},
		"team_members": []map[string]interface{}{
			{
				"name":   "alice johnson",
				"role":   "project manager",
				"status": "online",
				"tasks":  map[string]int{"completed": 12, "total": 15},
			},
			{
				"name":   "bob smith",
				"role":   "developer",
				"status": "online",
				"tasks":  map[string]int{"completed": 8, "total": 10},
			},
			{
				"name":   "carol white",
				"role":   "designer",
				"status": "offline",
				"tasks":  map[string]int{"completed": 15, "total": 15},
			},
			{
				"name":   "david brown",
				"role":   "developer",
				"status": "online",
				"tasks":  map[string]int{"completed": 5, "total": 12},
			},
		},
	}
}
