package compat_demo_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

// benchParser implements loader.TemplateParser interface
type benchParser struct {
	env *miya.Environment
}

func (sp *benchParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	template, err := sp.env.FromString(content)
	if err != nil {
		return nil, err
	}

	if templateNode, ok := template.AST().(*parser.TemplateNode); ok {
		templateNode.Name = name
		return templateNode, nil
	}

	return nil, err
}

func getContextData() map[string]interface{} {
	return map[string]interface{}{
		// Basic variables
		"name":          "Jinja2 User",
		"count":         42,
		"is_active":     true,
		"lang":          "en",
		"author":        "Miya Engine Team",
		"page_title":    "Complete Jinja2 Feature Test",
		"current_year":  time.Now().Year(),
		"jinja_version": "3.1.x",

		// User object for nested access
		"user": map[string]interface{}{
			"name":  "Alice Johnson",
			"email": "alice@example.com",
			"role":  "Administrator",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "San Francisco",
				"state":  "CA",
				"zip":    "94102",
			},
		},

		// Lists and arrays
		"colors":     []string{"red", "green", "blue", "yellow", "purple"},
		"fruits":     []string{"apple", "banana", "cherry", "date", "elderberry"},
		"numbers":    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 50, 75, 100},
		"empty_list": []interface{}{},

		// Tree structure for recursive loops
		"tree": []map[string]interface{}{
			{
				"name": "Root",
				"children": []map[string]interface{}{
					{
						"name": "Child 1",
						"children": []map[string]interface{}{
							{"name": "Grandchild 1.1", "children": []map[string]interface{}{}},
							{"name": "Grandchild 1.2", "children": []map[string]interface{}{}},
						},
					},
					{
						"name": "Child 2",
						"children": []map[string]interface{}{
							{"name": "Grandchild 2.1", "children": []map[string]interface{}{}},
						},
					},
					{"name": "Child 3", "children": []map[string]interface{}{}},
				},
			},
		},

		// Products for complex examples
		"products": []map[string]interface{}{
			{"name": "Laptop Pro", "category": "Electronics", "price": 1299.99, "stock": 15, "rating": 4.5},
			{"name": "Wireless Mouse", "category": "Electronics", "price": 29.99, "stock": 150, "rating": 4.2},
			{"name": "Mechanical Keyboard", "category": "Electronics", "price": 149.99, "stock": 45, "rating": 4.8},
			{"name": "Office Desk", "category": "Furniture", "price": 399.99, "stock": 8, "rating": 4.3},
			{"name": "Ergonomic Chair", "category": "Furniture", "price": 599.99, "stock": 0, "rating": 4.7},
			{"name": "Monitor Stand", "category": "Furniture", "price": 49.99, "stock": 30, "rating": 4.1},
			{"name": "USB-C Hub", "category": "Electronics", "price": 79.99, "stock": 75, "rating": 4.4},
			{"name": "Desk Lamp", "category": "Furniture", "price": 89.99, "stock": 22, "rating": 4.6},
			{"name": "Webcam HD", "category": "Electronics", "price": 129.99, "stock": 50, "rating": 4.3},
			{"name": "Bookshelf", "category": "Furniture", "price": 199.99, "stock": 5, "rating": 4.5},
		},

		// People list for advanced filtering
		"people": []map[string]interface{}{
			{"name": "Alice", "age": 30, "city": "NYC", "active": true},
			{"name": "Bob", "age": 25, "city": "SF", "active": true},
			{"name": "Charlie", "age": 35, "city": "NYC", "active": false},
			{"name": "Diana", "age": 28, "city": "LA", "active": true},
			{"name": "Eve", "age": 32, "city": "SF", "active": true},
			{"name": "Frank", "age": 29, "city": "NYC", "active": false},
		},

		// For template includes
		"show_menu":   true,
		"context_var": "This variable is available in included templates",
		"note_text":   "This is a custom note passed to the include.",
	}
}

// BenchmarkTemplateCompilation benchmarks template loading and compilation
func BenchmarkTemplateCompilation(b *testing.B) {
	templatesDir := "../../examples/compat/python/templates"
	absPath, _ := filepath.Abs(templatesDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := miya.NewEnvironment(
			miya.WithAutoEscape(true),
			miya.WithTrimBlocks(true),
			miya.WithLstripBlocks(true),
		)

		templateParser := &benchParser{env: env}
		fsLoader := loader.NewFileSystemLoader([]string{absPath}, templateParser)
		env.SetLoader(fsLoader)

		_, _ = env.GetTemplate("benchmark_simple.html")
	}
}

// BenchmarkTemplateRendering benchmarks template rendering
func BenchmarkTemplateRendering(b *testing.B) {
	templatesDir := "../../examples/compat/python/templates"
	absPath, _ := filepath.Abs(templatesDir)

	env := miya.NewEnvironment(
		miya.WithAutoEscape(true),
		miya.WithTrimBlocks(true),
		miya.WithLstripBlocks(true),
	)

	templateParser := &benchParser{env: env}
	fsLoader := loader.NewFileSystemLoader([]string{absPath}, templateParser)
	env.SetLoader(fsLoader)

	template, err := env.GetTemplate("benchmark_simple.html")
	if err != nil || template == nil {
		b.Skipf("Skipping benchmark: template not found at %s", absPath)
		return
	}
	context := miya.NewContextFrom(getContextData())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = template.Render(context)
	}
}

// BenchmarkTemplateRenderingWithAllocation benchmarks rendering with memory stats
func BenchmarkTemplateRenderingWithAllocation(b *testing.B) {
	templatesDir := "../../examples/compat/python/templates"
	absPath, _ := filepath.Abs(templatesDir)

	env := miya.NewEnvironment(
		miya.WithAutoEscape(true),
		miya.WithTrimBlocks(true),
		miya.WithLstripBlocks(true),
	)

	templateParser := &benchParser{env: env}
	fsLoader := loader.NewFileSystemLoader([]string{absPath}, templateParser)
	env.SetLoader(fsLoader)

	template, err := env.GetTemplate("benchmark_simple.html")
	if err != nil || template == nil {
		b.Skipf("Skipping benchmark: template not found at %s", absPath)
		return
	}
	context := miya.NewContextFrom(getContextData())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = template.Render(context)
	}
}

// BenchmarkContextCreation benchmarks context creation overhead
func BenchmarkContextCreation(b *testing.B) {
	data := getContextData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = miya.NewContextFrom(data)
	}
}
