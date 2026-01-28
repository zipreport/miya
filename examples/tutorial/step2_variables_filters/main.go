// Step 2: Variables & Filters
// ============================
// Learn how to pass complex data and transform output with filters.
//
// Run: go run ./examples/tutorial/step2_variables_filters

package main

import (
	"fmt"
	"log"

	"github.com/zipreport/miya"
)

func main() {
	fmt.Println("=== Step 2: Variables & Filters ===")
	fmt.Println()

	env := miya.NewEnvironment()

	// 1. Accessing nested data
	// Use dot notation to access nested properties.
	template1 := `
User Profile:
  Name: {{ user.name }}
  Email: {{ user.email }}
  City: {{ user.address.city }}
  Country: {{ user.address.country }}`

	tmpl1, err := env.FromString(template1)
	if err != nil {
		log.Fatal(err)
	}

	ctx1 := miya.NewContextFrom(map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Alice Johnson",
			"email": "alice@example.com",
			"address": map[string]interface{}{
				"city":    "San Francisco",
				"country": "USA",
			},
		},
	})

	output1, _ := tmpl1.Render(ctx1)
	fmt.Println("Example 1 - Nested data access:")
	fmt.Println(output1)
	fmt.Println()

	// 2. Using filters
	// Filters transform values: {{ value|filter }}
	template2 := `
String Filters:
  Original: {{ name }}
  Upper: {{ name|upper }}
  Lower: {{ name|lower }}
  Title: {{ text|title }}
  Trimmed: [{{ padded|trim }}]
  Truncated: {{ long_text|truncate(20) }}`

	tmpl2, _ := env.FromString(template2)
	ctx2 := miya.NewContextFrom(map[string]interface{}{
		"name":      "alice johnson",
		"text":      "hello world",
		"padded":    "   extra spaces   ",
		"long_text": "This is a very long text that should be truncated",
	})

	output2, _ := tmpl2.Render(ctx2)
	fmt.Println("Example 2 - String filters:")
	fmt.Println(output2)
	fmt.Println()

	// 3. Chaining filters
	// Apply multiple filters in sequence: {{ value|filter1|filter2 }}
	template3 := `
Filter Chaining:
  Original: {{ text }}
  Trimmed + Upper: {{ text|trim|upper }}
  Title + Replace: {{ text|trim|title|replace("World", "Miya") }}`

	tmpl3, _ := env.FromString(template3)
	ctx3 := miya.NewContextFrom(map[string]interface{}{
		"text": "  hello world  ",
	})

	output3, _ := tmpl3.Render(ctx3)
	fmt.Println("Example 3 - Filter chaining:")
	fmt.Println(output3)
	fmt.Println()

	// 4. Default values
	// Use |default to provide fallback for missing/empty values.
	template4 := `
Default Values:
  Existing: {{ name|default("Unknown") }}
  Missing: {{ missing_var|default("Not provided") }}
  Empty: {{ empty_string|default("Was empty", true) }}`

	tmpl4, _ := env.FromString(template4)
	ctx4 := miya.NewContextFrom(map[string]interface{}{
		"name":         "Alice",
		"empty_string": "",
	})

	output4, _ := tmpl4.Render(ctx4)
	fmt.Println("Example 4 - Default values:")
	fmt.Println(output4)
	fmt.Println()

	// 5. Numeric filters
	template5 := `
Numeric Filters:
  Original: {{ price }}
  Rounded: {{ price|round(2) }}
  As Integer: {{ price|int }}
  Absolute: {{ negative|abs }}`

	tmpl5, _ := env.FromString(template5)
	ctx5 := miya.NewContextFrom(map[string]interface{}{
		"price":    19.99567,
		"negative": -42,
	})

	output5, _ := tmpl5.Render(ctx5)
	fmt.Println("Example 5 - Numeric filters:")
	fmt.Println(output5)
	fmt.Println()

	// 6. Collection filters
	template6 := `
Collection Filters:
  Items: {{ items|join(", ") }}
  Count: {{ items|length }}
  First: {{ items|first }}
  Last: {{ items|last }}
  Sorted: {{ numbers|sort|join(", ") }}`

	tmpl6, _ := env.FromString(template6)
	ctx6 := miya.NewContextFrom(map[string]interface{}{
		"items":   []string{"Apple", "Banana", "Cherry"},
		"numbers": []int{3, 1, 4, 1, 5, 9, 2, 6},
	})

	output6, _ := tmpl6.Render(ctx6)
	fmt.Println("Example 6 - Collection filters:")
	fmt.Println(output6)
	fmt.Println()

	// Key Takeaways:
	// - Use dot notation for nested data: {{ user.address.city }}
	// - Filters transform values: {{ name|upper }}
	// - Chain filters left to right: {{ text|trim|upper }}
	// - Use |default for missing values
	// - Common filters: upper, lower, title, trim, round, join, length

	fmt.Println("=== Step 2 Complete ===")
}
