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
	fmt.Println("=== Advanced Features Examples ===")

	// Create environment with advanced options
	env := miya.NewEnvironment(
		miya.WithAutoEscape(true), // Enable HTML auto-escaping
	)

	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context data
	ctx := miya.NewContext()

	// Basic data
	ctx.Set("name", "Miya Engine")
	ctx.Set("numbers", []int{1, 2, 3, 4, 5})
	ctx.Set("items", []string{"Apple", "Banana", "Cherry"})

	// User data for filter blocks
	ctx.Set("users", []map[string]interface{}{
		{"name": "Alice Johnson", "active": true, "age": 30},
		{"name": "Bob Smith", "active": false, "age": 25},
		{"name": "Charlie Brown", "active": true, "age": 35},
	})

	// HTML snippets for autoescape examples
	ctx.Set("html_snippet", "<strong>Bold Text</strong>")
	ctx.Set("dangerous_html", "<script>alert('XSS')</script>")

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
	fmt.Println("=== Advanced Features Demonstrated ===")
	fmt.Println()

	fmt.Println("1. FILTER BLOCKS:")
	fmt.Println("   ✓ Single filter application: {% filter upper %}...{% endfilter %}")
	fmt.Println("   ✓ Chained filters: {% filter trim|upper|replace(...) %}")
	fmt.Println("   ✓ Filter blocks with logic (loops, conditionals)")
	fmt.Println("   ✓ Nested filter blocks")
	fmt.Println("   ✓ All 73+ filters supported in filter blocks")
	fmt.Println()

	fmt.Println("2. DO STATEMENTS:")
	fmt.Println("   ✓ Execute expressions without output: {% do expression %}")
	fmt.Println("   ✓ Side effects without rendering")
	fmt.Println("   ✓ Filter application: {% do value|filter %}")
	fmt.Println("   ✓ Complex expressions: {% do (a * b)|round %}")
	fmt.Println("   ✓ Whitespace control support: {%- do ... -%}")
	fmt.Println()

	fmt.Println("3. WHITESPACE CONTROL:")
	fmt.Println("   ✓ Left strip: {%- statement %} or {{- variable }}")
	fmt.Println("   ✓ Right strip: {% statement -%} or {{ variable -}}")
	fmt.Println("   ✓ Both sides: {%- statement -%}")
	fmt.Println("   ✓ Works with all block statements")
	fmt.Println("   ✓ Works with variable expressions")
	fmt.Println("   ✓ Clean output formatting")
	fmt.Println()

	fmt.Println("4. RAW BLOCKS:")
	fmt.Println("   ✓ Escape template syntax: {% raw %}...{% endraw %}")
	fmt.Println("   ✓ Display literal {{ }} and {% %}")
	fmt.Println("   ✓ Perfect for documentation")
	fmt.Println("   ✓ Show template examples in output")
	fmt.Println()

	fmt.Println("5. AUTOESCAPE CONTROL:")
	fmt.Println("   ✓ Auto HTML escaping (default enabled)")
	fmt.Println("   ✓ Explicit control: {% autoescape true/false %}")
	fmt.Println("   ✓ Safe filter: {{ html|safe }}")
	fmt.Println("   ✓ Escape filter: {{ text|escape }}")
	fmt.Println("   ✓ Force escape: {{ text|forceescape }}")
	fmt.Println("   ✓ XSS prevention")
	fmt.Println()

	fmt.Println("6. ENVIRONMENT CONFIGURATION:")
	fmt.Println("   ✓ WithAutoEscape(bool) - HTML auto-escaping")
	fmt.Println("   ✓ WithStrictUndefined(bool) - Error on undefined variables")
	fmt.Println("   ✓ WithTrimBlocks(bool) - Remove first newline after blocks")
	fmt.Println("   ✓ WithLstripBlocks(bool) - Strip leading whitespace")
	fmt.Println("   ✓ Custom delimiters support")
	fmt.Println()

	fmt.Println("7. COMBINING FEATURES:")
	fmt.Println("   ✓ Filter blocks + whitespace control")
	fmt.Println("   ✓ Filter blocks + comprehensions")
	fmt.Println("   ✓ Do statements + filters")
	fmt.Println("   ✓ Whitespace control + conditionals")
	fmt.Println("   ✓ All features work together seamlessly")
	fmt.Println()

	fmt.Println("ADVANCED FEATURES: 100% Jinja2 Compatible ✨")
	fmt.Println()

	// Show configuration examples
	fmt.Println("=== Configuration Examples ===")
	fmt.Println()
	fmt.Println("// Create environment with options")
	fmt.Println("env := miya.NewEnvironment(")
	fmt.Println("    miya.WithAutoEscape(true),      // Enable HTML escaping")
	fmt.Println("    miya.WithStrictUndefined(true), // Strict undefined handling")
	fmt.Println("    miya.WithTrimBlocks(true),      // Trim block newlines")
	fmt.Println("    miya.WithLstripBlocks(true),    // Strip leading whitespace")
	fmt.Println(")")
	fmt.Println()

	fmt.Println("// Use different loaders")
	fmt.Println("loader := miya.NewFileSystemLoader(\"templates/\")")
	fmt.Println("env.SetLoader(loader)")
	fmt.Println()

	fmt.Println("// Register custom filters")
	fmt.Println("env.AddFilter(\"custom\", customFilterFunc)")
}
