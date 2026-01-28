// Step 6: Loading Templates from Disk
// =====================================
// Learn how to organize and load templates from the filesystem.
//
// Run: go run ./examples/tutorial/step6_filesystem
//
// This step requires the templates/ directory with its template files.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	fmt.Println("=== Step 6: Loading Templates from Disk ===")
	fmt.Println()

	// Get the directory where this Go file is located
	// Templates are in the "templates" subdirectory relative to the tutorial root
	templateDir := filepath.Join("examples", "tutorial", "templates")

	// Check if templates directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		log.Fatal("Templates directory not found. Make sure to run from the project root directory.")
	}

	// 1. Create a FileSystemLoader
	// The loader knows how to find and read template files from disk.
	fmt.Println("Example 1 - Basic FileSystemLoader:")
	fmt.Println("------------------------------------")

	// Create environment first, then the loader
	env := miya.NewEnvironment(
		miya.WithAutoEscape(true), // Enable HTML auto-escaping
	)

	// Create template parser and filesystem loader
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{templateDir}, templateParser)
	fsLoader.SetExtensions([]string{".html", ".txt"})

	// Set the loader on the environment
	env.SetLoader(fsLoader)

	// Load and render a simple template
	tmpl1, err := env.GetTemplate("hello.html")
	if err != nil {
		log.Fatal("Failed to load template:", err)
	}

	ctx1 := miya.NewContextFrom(map[string]interface{}{
		"name":    "Alice",
		"message": "Welcome to Miya!",
	})

	output1, err := tmpl1.Render(ctx1)
	if err != nil {
		log.Fatal("Failed to render:", err)
	}
	fmt.Println(output1)
	fmt.Println()

	// 2. Template inheritance with filesystem
	fmt.Println("Example 2 - Inheritance from files:")
	fmt.Println("------------------------------------")

	// Load a child template that extends a base template
	// Both templates are loaded from the filesystem
	tmpl2, err := env.GetTemplate("pages/home.html")
	if err != nil {
		log.Fatal("Failed to load template:", err)
	}

	ctx2 := miya.NewContextFrom(map[string]interface{}{
		"site_name": "My Website",
		"user": map[string]interface{}{
			"name":  "Bob",
			"email": "bob@example.com",
		},
		"features": []string{
			"Fast template rendering",
			"Jinja2 compatible syntax",
			"Easy to learn",
		},
	})

	output2, err := tmpl2.Render(ctx2)
	if err != nil {
		log.Fatal("Failed to render:", err)
	}
	fmt.Println(output2)
	fmt.Println()

	// 3. Using includes from filesystem
	fmt.Println("Example 3 - Includes from files:")
	fmt.Println("---------------------------------")

	tmpl3, err := env.GetTemplate("pages/dashboard.html")
	if err != nil {
		log.Fatal("Failed to load template:", err)
	}

	ctx3 := miya.NewContextFrom(map[string]interface{}{
		"site_name": "Dashboard",
		"user": map[string]interface{}{
			"name":  "Charlie",
			"email": "charlie@example.com",
			"role":  "admin",
		},
		"stats": map[string]interface{}{
			"users":   1250,
			"orders":  847,
			"revenue": 52450.00,
			"growth":  12.5,
		},
		"recent_orders": []map[string]interface{}{
			{"id": 1001, "customer": "Alice", "total": 99.99, "status": "shipped"},
			{"id": 1002, "customer": "Bob", "total": 149.50, "status": "pending"},
			{"id": 1003, "customer": "Charlie", "total": 75.00, "status": "delivered"},
		},
	})

	output3, err := tmpl3.Render(ctx3)
	if err != nil {
		log.Fatal("Failed to render:", err)
	}
	fmt.Println(output3)
	fmt.Println()

	// 4. Loading with subdirectories
	fmt.Println("Example 4 - Organized template structure:")
	fmt.Println("------------------------------------------")
	fmt.Println("Template directory structure:")
	fmt.Println()
	printDirStructure(templateDir, "")
	fmt.Println()

	// 5. Multiple template directories (search paths)
	fmt.Println("Example 5 - Template caching:")
	fmt.Println("-----------------------------")
	fmt.Println("Templates are automatically cached after first load.")
	fmt.Println("Subsequent GetTemplate() calls return the cached version.")
	fmt.Println()

	// Load same template twice - second load uses cache
	_, _ = env.GetTemplate("hello.html") // First load (from disk)
	_, _ = env.GetTemplate("hello.html") // Second load (from cache)
	fmt.Println("Template 'hello.html' loaded (cached on first access)")
	fmt.Println()

	// 6. Error handling for missing templates
	fmt.Println("Example 6 - Error handling:")
	fmt.Println("---------------------------")

	_, err = env.GetTemplate("nonexistent.html")
	if err != nil {
		fmt.Printf("Expected error for missing template: %v\n", err)
	}
	fmt.Println()

	// Key Takeaways:
	// - Use FileSystemLoader to load templates from disk
	// - Specify allowed extensions: []string{".html", ".txt"}
	// - Templates can reference each other (extends, include, import)
	// - Organize templates in subdirectories (layouts/, pages/, includes/)
	// - Templates are cached after first load for performance
	// - Always handle errors when loading templates

	fmt.Println("=== Step 6 Complete ===")
	fmt.Println()
	fmt.Println("Recommended template organization:")
	fmt.Println("  templates/")
	fmt.Println("  ├── layouts/       # Base templates")
	fmt.Println("  ├── pages/         # Page templates (extend layouts)")
	fmt.Println("  ├── includes/      # Reusable partials")
	fmt.Println("  └── macros/        # Macro libraries")
}

// printDirStructure prints the directory tree
func printDirStructure(dir string, prefix string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		fmt.Printf("%s%s%s\n", prefix, connector, entry.Name())

		if entry.IsDir() {
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}
			printDirStructure(filepath.Join(dir, entry.Name()), newPrefix)
		}
	}
}
