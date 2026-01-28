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
	fmt.Println("=== Control Structures Examples ===")

	// Create environment
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context with comprehensive test data
	ctx := miya.NewContext()

	// User data for conditionals
	ctx.Set("user", map[string]interface{}{
		"name":     "Alice Johnson",
		"role":     "admin",
		"active":   true,
		"verified": true,
	})

	// Products for loops
	ctx.Set("products", []map[string]interface{}{
		{"name": "Laptop", "price": 999.99, "stock": 15},
		{"name": "Mouse", "price": 29.99, "stock": 50},
		{"name": "Keyboard", "price": 79.99, "stock": 30},
		{"name": "Monitor", "price": 299.99, "stock": 8},
		{"name": "Webcam", "price": 89.99, "stock": 25},
		{"name": "Headphones", "price": 149.99, "stock": 5},
	})

	// Configuration for dictionary iteration
	ctx.Set("config", map[string]interface{}{
		"site_name":     "Tech Store",
		"max_items":     100,
		"currency":      "USD",
		"tax_rate":      0.08,
		"free_shipping": true,
	})

	// Empty list for else clause demonstration
	ctx.Set("empty_list", []string{})

	// Variables for inline conditionals
	ctx.Set("stock", 15)
	ctx.Set("price", 99.99)
	ctx.Set("discount", 10)
	ctx.Set("total_spent", 750)

	// Categories for nested loops
	ctx.Set("categories", []map[string]interface{}{
		{
			"name":  "Input Devices",
			"items": []string{"Mouse", "Keyboard", "Trackpad"},
		},
		{
			"name":  "Output Devices",
			"items": []string{"Monitor", "Printer", "Speakers"},
		},
		{
			"name":  "Storage",
			"items": []string{"SSD", "HDD", "USB Drive"},
		},
	})

	// Render the template
	tmpl, err := env.GetTemplate("template.html")
	if err != nil {
		log.Fatal(err)
	}

	output, err := tmpl.Render(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)
	fmt.Println()

	// Print summary
	fmt.Println("=== Control Structures Features Demonstrated ===")
	fmt.Println()

	fmt.Println("1. CONDITIONAL STATEMENTS:")
	fmt.Println("   ✓ If/elif/else chains")
	fmt.Println("   ✓ Nested conditions")
	fmt.Println("   ✓ Boolean expressions")
	fmt.Println("   ✓ Comparison operators")
	fmt.Println()

	fmt.Println("2. FOR LOOPS:")
	fmt.Println("   ✓ Basic iteration")
	fmt.Println("   ✓ Loop variables (index, index0, first, last, length, revindex)")
	fmt.Println("   ✓ Conditional iteration (if clause)")
	fmt.Println("   ✓ Dictionary unpacking")
	fmt.Println("   ✓ Else clause for empty collections")
	fmt.Println("   ✓ Nested loops with loop.parent.loop")
	fmt.Println()

	fmt.Println("3. INLINE CONDITIONALS:")
	fmt.Println("   ✓ Ternary expressions: {{ 'yes' if condition else 'no' }}")
	fmt.Println("   ✓ Chained ternary for multi-way branching")
	fmt.Println("   ✓ Inline value substitution")
	fmt.Println()

	fmt.Println("4. SET STATEMENTS:")
	fmt.Println("   ✓ Variable assignment")
	fmt.Println("   ✓ Expression evaluation")
	fmt.Println("   ✓ Multiple assignment")
	fmt.Println("   ✓ Block assignment with {% set var %}...{% endset %}")
	fmt.Println()

	fmt.Println("5. WITH STATEMENTS:")
	fmt.Println("   ✓ Scoped variables")
	fmt.Println("   ✓ Multiple variable assignment")
	fmt.Println("   ✓ Nested with blocks")
	fmt.Println("   ✓ Variable isolation")
	fmt.Println()

	fmt.Println("6. COMPLEX CONTROL FLOW:")
	fmt.Println("   ✓ Nested loops")
	fmt.Println("   ✓ Combined conditions and loops")
	fmt.Println("   ✓ Loop context preservation")
	fmt.Println("   ✓ Complex boolean logic")
}
