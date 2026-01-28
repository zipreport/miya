// Step 1: Hello World
// ====================
// Learn the basics of creating an environment and rendering templates.
//
// Run: go run ./examples/tutorial/step1_hello_world

package main

import (
	"fmt"
	"log"

	"github.com/zipreport/miya"
)

func main() {
	fmt.Println("=== Step 1: Hello World ===")
	fmt.Println()

	// 1. Create a Miya environment
	// The environment holds configuration and is reusable for multiple templates.
	env := miya.NewEnvironment()

	// 2. Create a simple template from a string
	// Templates use {{ }} for variable substitution.
	template1, err := env.FromString("Hello, World!")
	if err != nil {
		log.Fatal(err)
	}

	// 3. Render the template
	// For templates without variables, pass an empty context.
	output1, err := template1.Render(miya.NewContext())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Example 1 - Static text:")
	fmt.Println(output1)
	fmt.Println()

	// 4. Template with a variable
	// Use {{ variable_name }} to insert values.
	template2, err := env.FromString("Hello, {{ name }}!")
	if err != nil {
		log.Fatal(err)
	}

	// 5. Create a context with data
	// Context holds the variables available to the template.
	ctx := miya.NewContext()
	ctx.Set("name", "Alice")

	output2, err := template2.Render(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Example 2 - With variable:")
	fmt.Println(output2)
	fmt.Println()

	// 6. Alternative: Create context from a map
	// This is convenient when you have multiple variables.
	template3, err := env.FromString("{{ greeting }}, {{ name }}! Welcome to {{ place }}.")
	if err != nil {
		log.Fatal(err)
	}

	ctx3 := miya.NewContextFrom(map[string]interface{}{
		"greeting": "Hello",
		"name":     "Bob",
		"place":    "Miya Engine",
	})

	output3, err := template3.Render(ctx3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Example 3 - Multiple variables:")
	fmt.Println(output3)
	fmt.Println()

	// Key Takeaways:
	// - Create an Environment once, reuse it for multiple templates
	// - Use FromString() to create templates from string content
	// - Use {{ variable }} syntax to insert values
	// - Pass data via Context using Set() or NewContextFrom()

	fmt.Println("=== Step 1 Complete ===")
}
