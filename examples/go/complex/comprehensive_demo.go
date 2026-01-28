package main

import (
	"fmt"
	"log"
	"time"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

// simpleParser implements loader.TemplateParser interface
type simpleParser struct {
	env *miya.Environment
}

func (sp *simpleParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	template, err := sp.env.FromString(content)
	if err != nil {
		return nil, err
	}

	if templateNode, ok := template.AST().(*parser.TemplateNode); ok {
		templateNode.Name = name
		return templateNode, nil
	}

	return nil, fmt.Errorf("failed to get template AST")
}

// User represents a sample data structure
type User struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Age      int                    `json:"age"`
	Active   bool                   `json:"active"`
	Role     string                 `json:"role"`
	Created  time.Time              `json:"created"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Product represents a sample product
type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	InStock     bool     `json:"in_stock"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Rating      float64  `json:"rating"`
}

func main() {
	fmt.Println("=== Comprehensive Miya Feature Demo (FileSystem) ===")

	// Create environment with all features enabled
	env := miya.NewEnvironment(
		miya.WithAutoEscape(true),
		miya.WithStrictUndefined(true),
		miya.WithTrimBlocks(true),
		miya.WithLstripBlocks(true),
	)

	// Set up filesystem template loader with parser
	templateParser := &simpleParser{env: env}
	fsLoader := loader.NewFileSystemLoader([]string{"templates"}, templateParser)
	env.SetLoader(fsLoader)

	// Create sample data
	users := createSampleUsers()
	products := createSampleProducts()

	// Test all major feature categories
	fmt.Println("\n🚀 Running comprehensive feature tests...")

	// 1. Template Inheritance Demo
	runInheritanceDemo(env, users)

	// 2. Control Structures Demo
	runControlStructuresDemo(env, users, products)

	// 3. Filter and Test Demo
	runFiltersAndTestsDemo(env, users)

	// 4. Macro and Import Demo
	runMacrosAndImportsDemo(env, users, products)

	// 5. Advanced Features Demo
	runAdvancedFeaturesDemo(env, users)

	// 6. Whitespace and Formatting Demo
	runWhitespaceDemo(env, products)

	fmt.Println("✅ All feature demonstrations completed successfully!")
}

func createSampleUsers() []User {
	return []User{
		{
			ID:      1,
			Name:    "alice johnson",
			Email:   "alice@example.com",
			Age:     28,
			Active:  true,
			Role:    "admin",
			Created: time.Now().AddDate(0, -6, 0),
			Tags:    []string{"developer", "frontend", "react"},
			Metadata: map[string]interface{}{
				"department": "Engineering",
				"location":   "San Francisco",
				"projects":   []string{"dashboard", "mobile-app"},
			},
		},
		{
			ID:      2,
			Name:    "bob smith",
			Email:   "bob@example.com",
			Age:     35,
			Active:  true,
			Role:    "moderator",
			Created: time.Now().AddDate(-1, 0, 0),
			Tags:    []string{"backend", "golang", "api"},
			Metadata: map[string]interface{}{
				"department": "Engineering",
				"location":   "New York",
				"projects":   []string{"api-gateway", "auth-service"},
			},
		},
		{
			ID:      3,
			Name:    "carol davis",
			Email:   "carol@example.com",
			Age:     42,
			Active:  false,
			Role:    "user",
			Created: time.Now().AddDate(-2, 0, 0),
			Tags:    []string{"designer", "ui", "ux"},
			Metadata: map[string]interface{}{
				"department": "Design",
				"location":   "Remote",
				"projects":   []string{"design-system"},
			},
		},
		{
			ID:      4,
			Name:    "david wilson",
			Email:   "",
			Age:     22,
			Active:  true,
			Role:    "user",
			Created: time.Now().AddDate(0, -1, 0),
			Tags:    []string{"intern", "learning"},
			Metadata: map[string]interface{}{
				"department": "Engineering",
				"location":   "Boston",
				"projects":   []string{},
			},
		},
	}
}

func createSampleProducts() []Product {
	return []Product{
		{
			ID:          1,
			Name:        "Premium Laptop",
			Price:       1299.99,
			Category:    "Electronics",
			InStock:     true,
			Description: "High-performance laptop for professionals",
			Features:    []string{"Intel Core i7", "16GB RAM", "512GB SSD", "4K Display"},
			Rating:      4.5,
		},
		{
			ID:          2,
			Name:        "Wireless Headphones",
			Price:       199.99,
			Category:    "Electronics",
			InStock:     false,
			Description: "Premium noise-canceling headphones",
			Features:    []string{"Noise Canceling", "30hr Battery", "Bluetooth 5.0"},
			Rating:      4.2,
		},
		{
			ID:          3,
			Name:        "Coffee Maker",
			Price:       89.99,
			Category:    "Kitchen",
			InStock:     true,
			Description: "Programmable drip coffee maker",
			Features:    []string{"12-cup capacity", "Programmable", "Auto shut-off"},
			Rating:      3.8,
		},
	}
}

func runInheritanceDemo(env *miya.Environment, users []User) {
	fmt.Println("📋 1. Template Inheritance Demo")

	template, err := env.GetTemplate("dashboard.html")
	if err != nil {
		log.Printf("❌ Failed to load dashboard template: %v", err)
		return
	}

	ctx := miya.NewContext()
	ctx.Set("user", users[0])
	ctx.Set("current_time", time.Now())

	result, err := template.Render(ctx)
	if err != nil {
		log.Printf("❌ Failed to render dashboard: %v", err)
		return
	}

	fmt.Printf("✅ Dashboard rendered successfully (%d chars)\n", len(result))
	fmt.Printf("Preview: %s...\n\n", result[:min(200, len(result))])
}

func runControlStructuresDemo(env *miya.Environment, users []User, products []Product) {
	fmt.Println("🔄 2. Control Structures Demo")

	template, err := env.GetTemplate("control_demo.html")
	if err != nil {
		log.Printf("❌ Failed to load control demo template: %v", err)
		return
	}

	ctx := miya.NewContext()
	// Convert to interface{} slice for template engine
	userInterfaces := make([]interface{}, len(users))
	for i, user := range users {
		userInterfaces[i] = user
	}
	ctx.Set("users", userInterfaces)
	ctx.Set("products", products)

	result, err := template.Render(ctx)
	if err != nil {
		log.Printf("❌ Failed to render control demo: %v", err)
		return
	}

	fmt.Printf("✅ Control structures demo rendered (%d chars)\n", len(result))
	fmt.Printf("Preview: %s...\n\n", result[:min(300, len(result))])
}

func runFiltersAndTestsDemo(env *miya.Environment, users []User) {
	fmt.Println("🔧 3. Filters and Tests Demo")

	template, err := env.GetTemplate("filters_demo.html")
	if err != nil {
		log.Printf("❌ Failed to load filters demo template: %v", err)
		return
	}

	ctx := miya.NewContext()
	// Convert to interface{} slice for template engine
	userInterfaces := make([]interface{}, len(users))
	for i, user := range users {
		userInterfaces[i] = user
	}
	ctx.Set("users", userInterfaces)

	result, err := template.Render(ctx)
	if err != nil {
		log.Printf("❌ Failed to render filters demo: %v", err)
		return
	}

	fmt.Printf("✅ Filters and tests demo rendered (%d chars)\n", len(result))
	fmt.Printf("Preview: %s...\n\n", result[:min(250, len(result))])
}

func runMacrosAndImportsDemo(env *miya.Environment, users []User, products []Product) {
	fmt.Println("📦 4. Macros and Imports Demo")

	// Test macro functionality via dashboard template (which imports utilities)
	template, err := env.GetTemplate("dashboard.html")
	if err != nil {
		log.Printf("❌ Failed to load template for macro demo: %v", err)
		return
	}

	ctx := miya.NewContext()
	ctx.Set("user", users[1])
	ctx.Set("products", products)
	ctx.Set("current_time", time.Now())

	result, err := template.Render(ctx)
	if err != nil {
		log.Printf("❌ Failed to render macro demo: %v", err)
		return
	}

	fmt.Printf("✅ Macros and imports demo completed (%d chars)\n", len(result))

	// Test direct macro import
	macroTemplate, err := env.GetTemplate("macro_test.html")
	if err != nil {
		log.Printf("❌ Failed to load macro test: %v", err)
		return
	}

	macroResult, err := macroTemplate.Render(miya.NewContext())
	if err != nil {
		log.Printf("❌ Failed to render macro test: %v", err)
		return
	}

	fmt.Printf("✅ Direct macro test rendered (%d chars)\n", len(macroResult))
	fmt.Printf("Preview: %s...\n\n", macroResult[:min(200, len(macroResult))])
}

func runAdvancedFeaturesDemo(env *miya.Environment, users []User) {
	fmt.Println("🚀 5. Advanced Features Demo")

	template, err := env.GetTemplate("advanced_demo.html")
	if err != nil {
		log.Printf("❌ Failed to load advanced demo template: %v", err)
		return
	}

	ctx := miya.NewContext()
	// Convert to interface{} slice for template engine
	userInterfaces := make([]interface{}, len(users))
	for i, user := range users {
		userInterfaces[i] = user
	}
	ctx.Set("users", userInterfaces)
	ctx.Set("current_time", time.Now())

	result, err := template.Render(ctx)
	if err != nil {
		log.Printf("❌ Failed to render advanced demo: %v", err)
		return
	}

	fmt.Printf("✅ Advanced features demo rendered (%d chars)\n", len(result))
	fmt.Printf("Preview: %s...\n\n", result[:min(300, len(result))])
}

func runWhitespaceDemo(env *miya.Environment, products []Product) {
	fmt.Println("⭐ 6. Whitespace Control Demo")

	template, err := env.GetTemplate("whitespace_demo.html")
	if err != nil {
		log.Printf("❌ Failed to load whitespace demo template: %v", err)
		return
	}

	ctx := miya.NewContext()
	ctx.Set("products", products)

	result, err := template.Render(ctx)
	if err != nil {
		log.Printf("❌ Failed to render whitespace demo: %v", err)
		return
	}

	fmt.Printf("✅ Whitespace control demo rendered (%d chars)\n", len(result))
	fmt.Printf("Clean output preview:\n%s\n\n", result[:min(400, len(result))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
