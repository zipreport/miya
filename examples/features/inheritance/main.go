package main

import (
	"fmt"
	"log"
	"time"

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
	fmt.Println("=== Template Inheritance Examples ===")

	// Create environment with filesystem loader
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context data
	ctx := miya.NewContext()
	ctx.Set("page_title", "Template Inheritance Showcase")
	ctx.Set("description", "This example demonstrates Miya Engine's complete template inheritance system including extends, blocks, and super() calls.")
	ctx.Set("level", 2)
	ctx.Set("multilevel", "Yes")

	ctx.Set("features", []string{
		"Template Extension ({% extends %})",
		"Block Definition and Override",
		"Super Calls ({{ super() }})",
		"Multi-level Inheritance",
		"Block Name Resolution",
		"Nested Block Content",
	})

	ctx.Set("quick_links", []map[string]interface{}{
		{"title": "Documentation", "url": "/docs"},
		{"title": "Examples", "url": "/examples"},
		{"title": "GitHub", "url": "https://github.com/zipreport/miya"},
	})

	ctx.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))

	// Example 1: Basic inheritance
	fmt.Println("Example 1: Basic Template Inheritance")
	fmt.Println("--------------------------------------")

	tmpl, err := env.GetTemplate("child.html")
	if err != nil {
		log.Fatal(err)
	}

	output, err := tmpl.Render(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)
	fmt.Println()

	// Example 2: Demonstrating super() calls
	fmt.Println("Example 2: Super() Calls")
	fmt.Println("------------------------")
	fmt.Println("The child template uses {{ super() }} to include parent block content:")
	fmt.Println("- In the header block: extends the original header")
	fmt.Println("- In the navigation block: adds more links to the parent navigation")
	fmt.Println("- In the sidebar block: includes the parent sidebar below new content")
	fmt.Println("- In the footer block: keeps the copyright notice from parent")
	fmt.Println()

	// Example 3: Multi-level inheritance demonstration
	fmt.Println("Example 3: Block Override Summary")
	fmt.Println("----------------------------------")
	fmt.Println("Blocks defined in base.html:")
	fmt.Println("  - title: Sets page title")
	fmt.Println("  - extra_head: For additional CSS/JS")
	fmt.Println("  - header: Page header content")
	fmt.Println("  - navigation: Navigation menu")
	fmt.Println("  - content: Main page content")
	fmt.Println("  - sidebar: Sidebar content")
	fmt.Println("  - footer: Footer content")
	fmt.Println()
	fmt.Println("All blocks are overridden in child.html, with several using super() to preserve parent content.")
	fmt.Println()

	// Example 4: Show how inheritance works
	fmt.Println("Example 4: How It Works")
	fmt.Println("-----------------------")
	fmt.Println("1. child.html declares: {% extends \"base.html\" %}")
	fmt.Println("2. Miya Engine loads base.html as the foundation")
	fmt.Println("3. Blocks in child.html replace corresponding blocks in base.html")
	fmt.Println("4. {{ super() }} includes the parent block's content at that position")
	fmt.Println("5. The final output is rendered with all blocks resolved")
	fmt.Println()

	fmt.Println("=== Template Inheritance Features Demonstrated ===")
	fmt.Println("✓ Template Extension with {% extends %}")
	fmt.Println("✓ Block Definition and Override")
	fmt.Println("✓ Super Calls with {{ super() }}")
	fmt.Println("✓ Multi-level Block Inheritance")
	fmt.Println("✓ Selective Block Override")
	fmt.Println("✓ Parent Content Preservation")
}
