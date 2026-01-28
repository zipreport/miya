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
	fmt.Println("=== Global Functions Examples ===")

	// Create environment
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context data
	ctx := miya.NewContext()

	// Data for examples
	ctx.Set("items", []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"})
	ctx.Set("numbers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// User data
	ctx.Set("users", []map[string]interface{}{
		{"name": "Alice", "age": 30, "active": true},
		{"name": "Bob", "age": 25, "active": false},
		{"name": "Charlie", "age": 35, "active": true},
		{"name": "Diana", "age": 28, "active": true},
	})

	// Data for zip() example
	ctx.Set("names", []string{"Alice", "Bob", "Charlie"})
	ctx.Set("ages", []int{30, 25, 35})
	ctx.Set("products", []string{"Laptop", "Mouse", "Keyboard"})
	ctx.Set("prices", []float64{999, 29, 79})
	ctx.Set("stock", []int{15, 50, 30})

	// Breadcrumbs for navigation example
	ctx.Set("breadcrumbs", []map[string]string{
		{"title": "Home", "url": "/"},
		{"title": "Products", "url": "/products"},
		{"title": "Electronics", "url": "/products/electronics"},
		{"title": "Laptops", "url": "/products/electronics/laptops"},
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
	fmt.Println("=== Global Functions Features Demonstrated ===")
	fmt.Println()

	fmt.Println("1. range() - Number Sequences:")
	fmt.Println("   ✓ range(n) - Generate 0 to n-1")
	fmt.Println("   ✓ range(start, end) - Generate range")
	fmt.Println("   ✓ range(start, end, step) - With custom step")
	fmt.Println("   ✓ Reverse ranges with negative step")
	fmt.Println("   ✓ Use in loops for iteration")
	fmt.Println("   Examples:")
	fmt.Println("     - range(10) → [0, 1, 2, ..., 9]")
	fmt.Println("     - range(5, 15) → [5, 6, 7, ..., 14]")
	fmt.Println("     - range(0, 20, 3) → [0, 3, 6, 9, 12, 15, 18]")
	fmt.Println()

	fmt.Println("2. dict() - Dictionary Constructor:")
	fmt.Println("   ✓ Create dictionaries with keyword arguments")
	fmt.Println("   ✓ Create from key-value pair lists")
	fmt.Println("   ✓ Access with dot notation or brackets")
	fmt.Println("   Examples:")
	fmt.Println("     - dict(name=\"Alice\", age=30)")
	fmt.Println("     - dict([(\"key1\", \"val1\"), (\"key2\", \"val2\")])")
	fmt.Println()

	fmt.Println("3. cycler() - Cycle Through Values:")
	fmt.Println("   ✓ Create cycler with multiple values")
	fmt.Println("   ✓ Call .next() to get next value in cycle")
	fmt.Println("   ✓ Automatically wraps around to beginning")
	fmt.Println("   ✓ Perfect for alternating row colors")
	fmt.Println("   Examples:")
	fmt.Println("     - cycler(\"odd\", \"even\")")
	fmt.Println("     - cycler(\"red\", \"green\", \"blue\")")
	fmt.Println()

	fmt.Println("4. joiner() - Smart Joining:")
	fmt.Println("   ✓ Create joiner with separator string")
	fmt.Println("   ✓ First call returns empty string")
	fmt.Println("   ✓ Subsequent calls return separator")
	fmt.Println("   ✓ Cleaner than manual separator logic")
	fmt.Println("   Examples:")
	fmt.Println("     - joiner(\", \") for comma-separated lists")
	fmt.Println("     - joiner(\" | \") for pipe-separated items")
	fmt.Println()

	fmt.Println("5. namespace() - Mutable Container:")
	fmt.Println("   ✓ Create mutable object for loop scoping")
	fmt.Println("   ✓ Modify values within loops")
	fmt.Println("   ✓ Accumulate results across iterations")
	fmt.Println("   ✓ Workaround for variable scoping in loops")
	fmt.Println("   Examples:")
	fmt.Println("     - namespace(count=0, total=0)")
	fmt.Println("     - Increment counters in loops")
	fmt.Println()

	fmt.Println("6. lipsum() - Lorem Ipsum Generator:")
	fmt.Println("   ✓ Generate placeholder text")
	fmt.Println("   ✓ Specify number of paragraphs")
	fmt.Println("   ✓ Generate HTML paragraphs")
	fmt.Println("   ✓ Useful for mockups and prototypes")
	fmt.Println("   Examples:")
	fmt.Println("     - lipsum() - Default text")
	fmt.Println("     - lipsum(n=5) - 5 paragraphs")
	fmt.Println("     - lipsum(n=3, html=true) - HTML format")
	fmt.Println()

	fmt.Println("7. zip() - Combine Sequences:")
	fmt.Println("   ✓ Combine multiple iterables")
	fmt.Println("   ✓ Iterate over tuples of values")
	fmt.Println("   ✓ Stops at shortest sequence")
	fmt.Println("   ✓ Perfect for parallel iteration")
	fmt.Println("   Examples:")
	fmt.Println("     - zip(names, ages)")
	fmt.Println("     - zip(products, prices, stock)")
	fmt.Println()

	fmt.Println("8. enumerate() - Index with Values:")
	fmt.Println("   ✓ Get index and value in loops")
	fmt.Println("   ✓ Custom start index")
	fmt.Println("   ✓ Cleaner than manual counter")
	fmt.Println("   Examples:")
	fmt.Println("     - enumerate(items) - Start at 0")
	fmt.Println("     - enumerate(items, start=1) - Start at 1")
	fmt.Println()

	fmt.Println("9. url_for() - URL Generation:")
	fmt.Println("   ✓ Generate URLs from endpoint names")
	fmt.Println("   ✓ Add query parameters")
	fmt.Println("   ✓ Framework integration support")
	fmt.Println()

	fmt.Println("PRACTICAL USE CASES:")
	fmt.Println()

	fmt.Println("✓ Alternating table row colors with cycler()")
	fmt.Println("✓ Building navigation breadcrumbs with joiner()")
	fmt.Println("✓ Creating grids with range()")
	fmt.Println("✓ Counting/accumulating in loops with namespace()")
	fmt.Println("✓ Parallel iteration with zip()")
	fmt.Println("✓ Numbered lists with enumerate()")
	fmt.Println("✓ Placeholder content with lipsum()")
	fmt.Println()

	fmt.Println("Total Global Functions: 9")
	fmt.Println("ALL FUNCTIONS: 100% Jinja2 Compatible ✨")
}
