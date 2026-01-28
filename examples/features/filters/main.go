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
	fmt.Println("=== Filters Examples ===")

	// Create environment
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context with test data for all filter types
	ctx := miya.NewContext()

	// String data
	ctx.Set("text", "hello world")
	ctx.Set("long_text", "This is a very long text that will be truncated to demonstrate the truncate filter in action")

	// Numeric data
	ctx.Set("numbers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	ctx.Set("unsorted_numbers", []int{5, 2, 8, 1, 9, 3})
	ctx.Set("duplicates", []int{1, 2, 2, 3, 3, 3, 4, 4, 5})
	ctx.Set("prices", []float64{19.99, 29.99, 49.99, 99.99})

	// Collection data
	ctx.Set("fruits", []string{"apple", "banana", "cherry", "date"})

	// User data for attribute filters
	ctx.Set("users", []map[string]interface{}{
		{"name": "Alice", "active": true, "age": 30},
		{"name": "Bob", "active": false, "age": 25},
		{"name": "Charlie", "active": true, "age": 35},
		{"name": "Diana", "active": true, "age": 28},
		{"name": "Eve", "active": false, "age": 32},
	})

	// HTML/Security data
	ctx.Set("html_code", "<script>alert('XSS')</script>")
	ctx.Set("safe_html", "<strong>This is safe HTML</strong>")

	// Utility data
	ctx.Set("user_name", "Alice")
	ctx.Set("message_count", 5)
	ctx.Set("data_object", map[string]interface{}{
		"id":    123,
		"name":  "Product",
		"price": 99.99,
	})

	// Configuration for dictsort
	ctx.Set("config", map[string]interface{}{
		"site_name": "Miya Store",
		"version":   "1.0",
		"debug":     false,
		"port":      8080,
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
	fmt.Println("=== Filter Features Demonstrated ===")
	fmt.Println()

	fmt.Println("1. STRING FILTERS (16+):")
	fmt.Println("   ✓ upper, lower, capitalize, title")
	fmt.Println("   ✓ trim/strip")
	fmt.Println("   ✓ replace, truncate, center, indent")
	fmt.Println("   ✓ wordcount, wordwrap")
	fmt.Println("   ✓ split, startswith, endswith, contains")
	fmt.Println("   ✓ slugify")
	fmt.Println()

	fmt.Println("2. COLLECTION FILTERS (15+):")
	fmt.Println("   ✓ first, last, length/count")
	fmt.Println("   ✓ join, sort, reverse, unique")
	fmt.Println("   ✓ slice, batch, random")
	fmt.Println("   ✓ map (extract attributes)")
	fmt.Println("   ✓ select, reject (filter by test)")
	fmt.Println("   ✓ selectattr, rejectattr (filter by attribute)")
	fmt.Println("   ✓ groupby, items, keys, values, zip")
	fmt.Println()

	fmt.Println("3. NUMERIC FILTERS (10):")
	fmt.Println("   ✓ abs, round, int, float")
	fmt.Println("   ✓ sum, min, max")
	fmt.Println("   ✓ ceil, floor, pow")
	fmt.Println()

	fmt.Println("4. HTML/SECURITY FILTERS (7+):")
	fmt.Println("   ✓ escape/e (HTML entity encoding)")
	fmt.Println("   ✓ safe (mark as safe HTML)")
	fmt.Println("   ✓ striptags (remove HTML tags)")
	fmt.Println("   ✓ urlencode")
	fmt.Println("   ✓ urlize, urlizetruncate")
	fmt.Println("   ✓ forceescape, autoescape")
	fmt.Println()

	fmt.Println("5. UTILITY FILTERS (8+):")
	fmt.Println("   ✓ default/d (fallback values)")
	fmt.Println("   ✓ format (string formatting)")
	fmt.Println("   ✓ tojson, fromjson")
	fmt.Println("   ✓ filesizeformat")
	fmt.Println("   ✓ dictsort")
	fmt.Println("   ✓ attr, pprint, string")
	fmt.Println()

	fmt.Println("6. ADVANCED FEATURES:")
	fmt.Println("   ✓ Filter chaining: {{ value|filter1|filter2|filter3 }}")
	fmt.Println("   ✓ Filters with arguments: {{ text|truncate(30) }}")
	fmt.Println("   ✓ Filters in conditionals")
	fmt.Println("   ✓ Filters in loops")
	fmt.Println("   ✓ Filters with ternary expressions")
	fmt.Println()

	fmt.Println("TOTAL: 73+ filters with full Jinja2 compatibility")
}
