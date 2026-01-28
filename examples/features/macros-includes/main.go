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
	fmt.Println("=== Macros & Includes Examples ===")

	// Create environment
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context data
	ctx := miya.NewContext()

	// Header/footer data
	ctx.Set("site_name", "Miya Engine Demo")
	ctx.Set("year", 2024)
	ctx.Set("company", "Miya Engine")

	ctx.Set("nav_links", []map[string]interface{}{
		{"title": "Home", "url": "/"},
		{"title": "Features", "url": "/features"},
		{"title": "Docs", "url": "/docs"},
		{"title": "Examples", "url": "/examples"},
	})

	ctx.Set("footer_links", []map[string]interface{}{
		{"title": "Privacy", "url": "/privacy"},
		{"title": "Terms", "url": "/terms"},
		{"title": "Contact", "url": "/contact"},
	})

	// User data for call block example
	ctx.Set("user", map[string]interface{}{
		"name":  "Alice Johnson",
		"email": "alice@example.com",
		"role":  "Administrator",
	})

	// Statistics data
	ctx.Set("stats", map[string]interface{}{
		"users":    1250,
		"sessions": 342,
		"orders":   89,
	})

	// Features list
	ctx.Set("features", []string{
		"Template Inheritance",
		"Macro System",
		"Include Support",
		"Filter Library",
		"Control Structures",
	})

	// Steps for ordered list
	ctx.Set("steps", []string{
		"Create environment",
		"Load templates",
		"Prepare context",
		"Render output",
	})

	// Tags
	ctx.Set("tags", []string{
		"golang",
		"templates",
		"jinja2",
		"miya",
	})

	// Products for table
	ctx.Set("products", []map[string]interface{}{
		{"name": "Laptop", "price": 999.99, "stock": 15, "status": "In Stock"},
		{"name": "Mouse", "price": 29.99, "stock": 50, "status": "In Stock"},
		{"name": "Keyboard", "price": 79.99, "stock": 0, "status": "Out of Stock"},
		{"name": "Monitor", "price": 299.99, "stock": 8, "status": "Low Stock"},
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
	fmt.Println("=== Macros & Includes Features Demonstrated ===")
	fmt.Println()

	fmt.Println("1. MACRO FEATURES:")
	fmt.Println("   ✓ Basic macro definition: {% macro name(params) %}...{% endmacro %}")
	fmt.Println("   ✓ Default parameters: {% macro name(param=\"default\") %}")
	fmt.Println("   ✓ Macros with logic (conditionals, loops)")
	fmt.Println("   ✓ Call blocks: {% call macro() %}content{% endcall %}")
	fmt.Println("   ✓ Caller function: {{ caller() }}")
	fmt.Println("   ✓ Macros with filters")
	fmt.Println("   ✓ Nested macro calls")
	fmt.Println()

	fmt.Println("2. IMPORT METHODS:")
	fmt.Println("   ✓ Namespace import: {% import \"file.html\" as namespace %}")
	fmt.Println("   ✓ Selective import: {% from \"file.html\" import macro1, macro2 %}")
	fmt.Println("   ✓ Multiple imports from same file")
	fmt.Println()

	fmt.Println("3. INCLUDE FEATURES:")
	fmt.Println("   ✓ Template inclusion: {% include \"template.html\" %}")
	fmt.Println("   ✓ Automatic context passing")
	fmt.Println("   ✓ Include with context")
	fmt.Println("   ✓ Nested includes")
	fmt.Println()

	fmt.Println("4. USE CASES DEMONSTRATED:")
	fmt.Println("   ✓ Reusable form components")
	fmt.Println("   ✓ UI component library")
	fmt.Println("   ✓ Table row generators")
	fmt.Println("   ✓ Card/panel wrappers")
	fmt.Println("   ✓ Badge and button generators")
	fmt.Println("   ✓ List renderers")
	fmt.Println("   ✓ Header/footer includes")
	fmt.Println()

	fmt.Println("5. ADVANCED FEATURES:")
	fmt.Println("   ✓ Macros in loops")
	fmt.Println("   ✓ Conditional macro calls")
	fmt.Println("   ✓ Dynamic parameter passing")
	fmt.Println("   ✓ Macro composition")
}
