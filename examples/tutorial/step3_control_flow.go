// Step 3: Control Flow
// =====================
// Learn conditional rendering and iteration.
//
// Run: go run step3_control_flow.go

package main

import (
	"fmt"
	"log"

	"github.com/zipreport/miya"
)

func main() {
	fmt.Println("=== Step 3: Control Flow ===")
	fmt.Println()

	env := miya.NewEnvironment()

	// 1. If/Elif/Else statements
	template1 := `
{% if user.role == "admin" %}
  Admin Dashboard: Full access granted
{% elif user.role == "moderator" %}
  Moderator Panel: Limited access
{% else %}
  User View: Read-only access
{% endif %}`

	tmpl1, err := env.FromString(template1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Example 1 - Conditionals:")

	// Test with admin
	ctx1a := miya.NewContextFrom(map[string]interface{}{
		"user": map[string]interface{}{"role": "admin"},
	})
	output1a, _ := tmpl1.Render(ctx1a)
	fmt.Println("  Admin:", output1a)

	// Test with regular user
	ctx1b := miya.NewContextFrom(map[string]interface{}{
		"user": map[string]interface{}{"role": "user"},
	})
	output1b, _ := tmpl1.Render(ctx1b)
	fmt.Println("  User:", output1b)
	fmt.Println()

	// 2. Inline conditionals (ternary)
	template2 := `Status: {{ "Active" if user.active else "Inactive" }}
Access: {{ "Premium" if user.premium else "Free" }} account`

	tmpl2, _ := env.FromString(template2)
	ctx2 := miya.NewContextFrom(map[string]interface{}{
		"user": map[string]interface{}{
			"active":  true,
			"premium": false,
		},
	})

	output2, _ := tmpl2.Render(ctx2)
	fmt.Println("Example 2 - Inline conditionals:")
	fmt.Println(output2)
	fmt.Println()

	// 3. Basic for loops
	template3 := `
Shopping List:
{% for item in items %}
  - {{ item }}
{% endfor %}`

	tmpl3, _ := env.FromString(template3)
	ctx3 := miya.NewContextFrom(map[string]interface{}{
		"items": []string{"Milk", "Bread", "Eggs", "Butter"},
	})

	output3, _ := tmpl3.Render(ctx3)
	fmt.Println("Example 3 - Basic loop:")
	fmt.Println(output3)

	// 4. Loop variables
	template4 := `
Numbered List:
{% for item in items %}
  {{ loop.index }}. {{ item }}{% if loop.first %} (first){% endif %}{% if loop.last %} (last){% endif %}

{% endfor %}
Total items: {{ items|length }}`

	tmpl4, _ := env.FromString(template4)
	ctx4 := miya.NewContextFrom(map[string]interface{}{
		"items": []string{"Apple", "Banana", "Cherry"},
	})

	output4, _ := tmpl4.Render(ctx4)
	fmt.Println("Example 4 - Loop variables:")
	fmt.Println(output4)
	fmt.Println()

	// 5. Loop with else (empty list handling)
	template5 := `
{% for item in items %}
  - {{ item }}
{% else %}
  No items found.
{% endfor %}`

	tmpl5, _ := env.FromString(template5)

	fmt.Println("Example 5 - Empty list handling:")
	// With items
	ctx5a := miya.NewContextFrom(map[string]interface{}{
		"items": []string{"One", "Two"},
	})
	output5a, _ := tmpl5.Render(ctx5a)
	fmt.Println("  With items:", output5a)

	// Without items
	ctx5b := miya.NewContextFrom(map[string]interface{}{
		"items": []string{},
	})
	output5b, _ := tmpl5.Render(ctx5b)
	fmt.Println("  Empty list:", output5b)
	fmt.Println()

	// 6. Looping over maps/dictionaries
	template6 := `
User Settings:
{% for key, value in settings %}
  {{ key }}: {{ value }}
{% endfor %}`

	tmpl6, _ := env.FromString(template6)
	ctx6 := miya.NewContextFrom(map[string]interface{}{
		"settings": map[string]interface{}{
			"theme":         "dark",
			"language":      "en",
			"notifications": "enabled",
		},
	})

	output6, _ := tmpl6.Render(ctx6)
	fmt.Println("Example 6 - Dictionary iteration:")
	fmt.Println(output6)
	fmt.Println()

	// 7. Combining conditions and loops
	template7 := `
Active Users:
{% for user in users %}
{% if user.active %}
  - {{ user.name }} ({{ user.email }})
{% endif %}
{% endfor %}`

	tmpl7, _ := env.FromString(template7)
	ctx7 := miya.NewContextFrom(map[string]interface{}{
		"users": []map[string]interface{}{
			{"name": "Alice", "email": "alice@example.com", "active": true},
			{"name": "Bob", "email": "bob@example.com", "active": false},
			{"name": "Charlie", "email": "charlie@example.com", "active": true},
		},
	})

	output7, _ := tmpl7.Render(ctx7)
	fmt.Println("Example 7 - Filtered loop:")
	fmt.Println(output7)

	// Key Takeaways:
	// - Use {% if %}{% elif %}{% else %}{% endif %} for conditionals
	// - Use {{ 'a' if condition else 'b' }} for inline conditionals
	// - Use {% for item in list %}{% endfor %} for iteration
	// - Loop variables: loop.index, loop.first, loop.last, loop.length
	// - Use {% else %} inside for loops for empty list handling
	// - Use {% for key, value in dict %} for dictionary iteration

	fmt.Println("=== Step 3 Complete ===")
}
