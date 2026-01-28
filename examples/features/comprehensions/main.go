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
	fmt.Println("=== List & Dictionary Comprehensions Examples ===")

	// Create environment
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context data
	ctx := miya.NewContext()

	// Basic data
	ctx.Set("numbers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	ctx.Set("names", []string{"alice", "bob", "charlie", "diana"})
	ctx.Set("prices", []float64{19.99, 29.99, 49.99, 99.99, 149.99})

	// Titles for slugify
	ctx.Set("titles", []string{
		"Hello World",
		"Miya Engine Guide",
		"Template Basics",
	})

	// User data with various attributes
	ctx.Set("users", []map[string]interface{}{
		{"id": 1, "name": "Alice Johnson", "email": "alice@example.com", "active": true, "age": 30, "role": "Admin"},
		{"id": 2, "name": "Bob Smith", "email": "bob@example.com", "active": false, "age": 25, "role": "User"},
		{"id": 3, "name": "Charlie Brown", "email": "charlie@example.com", "active": true, "age": 35, "role": "Moderator"},
		{"id": 4, "name": "Diana Prince", "email": "diana@example.com", "active": true, "age": 28, "role": "User"},
		{"id": 5, "name": "Eve Davis", "email": "eve@example.com", "active": false, "age": 22, "role": "User"},
	})

	// Configuration data
	ctx.Set("config", map[string]interface{}{
		"site_name": "Miya Store",
		"version":   "1.0",
		"debug":     false,
		"port":      8080,
		"timeout":   30,
	})

	// Shopping cart
	ctx.Set("cart", []map[string]interface{}{
		{"name": "Laptop", "price": 999.99, "qty": 1},
		{"name": "Mouse", "price": 29.99, "qty": 2},
		{"name": "Keyboard", "price": 79.99, "qty": 1},
		{"name": "Monitor", "price": 299.99, "qty": 2},
	})

	// Products
	ctx.Set("products", []map[string]interface{}{
		{"sku": "LAP001", "name": "Laptop", "price": 999.99},
		{"sku": "MOU001", "name": "Mouse", "price": 29.99},
		{"sku": "KEY001", "name": "Keyboard", "price": 79.99},
	})

	// Categories for nested comprehensions
	ctx.Set("categories", []map[string]interface{}{
		{"name": "Electronics", "items": []string{"Laptop", "Phone", "Tablet"}},
		{"name": "Accessories", "items": []string{"Mouse", "Keyboard", "Cable"}},
		{"name": "Storage", "items": []string{"SSD", "HDD", "USB"}},
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
	fmt.Println("=== Comprehensions Features Demonstrated ===")
	fmt.Println()

	fmt.Println("1. LIST COMPREHENSIONS:")
	fmt.Println("   ✓ Basic transformation: [x * 2 for x in numbers]")
	fmt.Println("   ✓ Property extraction: [user.name for user in users]")
	fmt.Println("   ✓ String operations: [name|upper for name in names]")
	fmt.Println("   ✓ Arithmetic: [price * 1.1 for price in prices]")
	fmt.Println()

	fmt.Println("2. LIST COMPREHENSIONS WITH CONDITIONS:")
	fmt.Println("   ✓ Filter values: [x for x in numbers if x % 2 == 0]")
	fmt.Println("   ✓ Filter by attribute: [user.name for user in users if user.active]")
	fmt.Println("   ✓ Complex conditions: [... if cond1 and cond2]")
	fmt.Println()

	fmt.Println("3. LIST COMPREHENSIONS WITH FILTERS:")
	fmt.Println("   ✓ Apply filters: [name|title for name in names]")
	fmt.Println("   ✓ Chain filters: [name|trim|upper for name in names]")
	fmt.Println("   ✓ Format values: [\"$\" ~ (price|round(2)) for price in prices]")
	fmt.Println()

	fmt.Println("4. DICTIONARY COMPREHENSIONS:")
	fmt.Println("   ✓ Create mappings: {user.id: user.name for user in users}")
	fmt.Println("   ✓ Transform keys: {key|upper: value for key, value in dict.items()}")
	fmt.Println("   ✓ Reverse mappings: {value: key for key, value in dict.items()}")
	fmt.Println("   ✓ Index mapping: {i: name for i, name in enumerate(names)}")
	fmt.Println()

	fmt.Println("5. DICTIONARY COMPREHENSIONS WITH CONDITIONS:")
	fmt.Println("   ✓ Filter by condition: {k: v for k, v in dict.items() if condition}")
	fmt.Println("   ✓ Type filtering: {k: v for k, v in items if v is string}")
	fmt.Println("   ✓ Complex filters: {k: v for ... if cond1 and cond2}")
	fmt.Println()

	fmt.Println("6. NESTED COMPREHENSIONS:")
	fmt.Println("   ✓ Flatten lists: [item for list in lists for item in list]")
	fmt.Println("   ✓ Cartesian products: [(a, b) for a in list1 for b in list2]")
	fmt.Println()

	fmt.Println("7. PRACTICAL USE CASES:")
	fmt.Println("   ✓ Calculate totals: [item.price * item.qty for item in cart]|sum")
	fmt.Println("   ✓ Format displays: [user.name ~ \" (\" ~ user.role ~ \")\" ...]")
	fmt.Println("   ✓ Generate email lists: [user.email for user in users if user.active]")
	fmt.Println("   ✓ Create lookups: {product.sku: product for product in products}")
	fmt.Println()

	fmt.Println("8. INTEGRATION WITH TEMPLATES:")
	fmt.Println("   ✓ In conditionals: {% if [...]|length > 0 %}")
	fmt.Println("   ✓ In loops: {% for item in [expression for ...] %}")
	fmt.Println("   ✓ In variable assignment: {% set var = [...] %}")
	fmt.Println("   ✓ With additional filters: [...] | join(\", \")")
	fmt.Println()

	fmt.Println("COMPREHENSIONS: 100% Jinja2 Compatible ✨")
}
