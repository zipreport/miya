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
	fmt.Println("=== Tests & Operators Examples ===")

	// Create environment
	env := miya.NewEnvironment(miya.WithAutoEscape(false))
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"."}, templateParser)
	env.SetLoader(fsLoader)

	// Prepare context data
	ctx := miya.NewContext()

	// Basic types for tests
	ctx.Set("text", "hello world")
	ctx.Set("number", 42)
	ctx.Set("integer_val", 42)
	ctx.Set("float_val", 3.14)
	ctx.Set("null_value", nil)

	// Collections
	ctx.Set("numbers", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// User data for complex examples
	ctx.Set("user", map[string]interface{}{
		"name":     "Alice",
		"age":      30,
		"role":     "admin",
		"active":   true,
		"verified": true,
	})

	// Calculation data
	ctx.Set("price", 99.99)
	ctx.Set("quantity", 3)
	ctx.Set("discount", 0.1)
	ctx.Set("total", 750)

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
	fmt.Println("=== Tests & Operators Features Demonstrated ===")
	fmt.Println()

	fmt.Println("OPERATORS:")
	fmt.Println()

	fmt.Println("1. ARITHMETIC OPERATORS (8):")
	fmt.Println("   ✓ Addition: +")
	fmt.Println("   ✓ Subtraction: -")
	fmt.Println("   ✓ Multiplication: *")
	fmt.Println("   ✓ Division: /")
	fmt.Println("   ✓ Floor Division: //")
	fmt.Println("   ✓ Modulo: %")
	fmt.Println("   ✓ Power: **")
	fmt.Println("   ✓ String Concatenation: ~")
	fmt.Println()

	fmt.Println("2. COMPARISON OPERATORS (6):")
	fmt.Println("   ✓ Equal: ==")
	fmt.Println("   ✓ Not Equal: !=")
	fmt.Println("   ✓ Less Than: <")
	fmt.Println("   ✓ Less or Equal: <=")
	fmt.Println("   ✓ Greater Than: >")
	fmt.Println("   ✓ Greater or Equal: >=")
	fmt.Println()

	fmt.Println("3. LOGICAL OPERATORS (3):")
	fmt.Println("   ✓ AND: and")
	fmt.Println("   ✓ OR: or")
	fmt.Println("   ✓ NOT: not")
	fmt.Println()

	fmt.Println("4. MEMBERSHIP OPERATORS (2):")
	fmt.Println("   ✓ In: in")
	fmt.Println("   ✓ Not In: not in")
	fmt.Println()

	fmt.Println("Total Operators: 19")
	fmt.Println()

	fmt.Println("TESTS:")
	fmt.Println()

	fmt.Println("1. TYPE TESTS (8):")
	fmt.Println("   ✓ is defined")
	fmt.Println("   ✓ is undefined")
	fmt.Println("   ✓ is none")
	fmt.Println("   ✓ is boolean")
	fmt.Println("   ✓ is string")
	fmt.Println("   ✓ is number")
	fmt.Println("   ✓ is integer")
	fmt.Println("   ✓ is float")
	fmt.Println()

	fmt.Println("2. CONTAINER TESTS (4):")
	fmt.Println("   ✓ is sequence")
	fmt.Println("   ✓ is mapping")
	fmt.Println("   ✓ is iterable")
	fmt.Println("   ✓ is callable")
	fmt.Println()

	fmt.Println("3. NUMERIC TESTS (3):")
	fmt.Println("   ✓ is even")
	fmt.Println("   ✓ is odd")
	fmt.Println("   ✓ is divisibleby(n)")
	fmt.Println()

	fmt.Println("4. STRING TESTS (7):")
	fmt.Println("   ✓ is lower")
	fmt.Println("   ✓ is upper")
	fmt.Println("   ✓ is startswith(str)")
	fmt.Println("   ✓ is endswith(str)")
	fmt.Println("   ✓ is match(regex)")
	fmt.Println("   ✓ is alpha")
	fmt.Println("   ✓ is alnum")
	fmt.Println()

	fmt.Println("5. COMPARISON TESTS (4):")
	fmt.Println("   ✓ is equalto(value)")
	fmt.Println("   ✓ is sameas(value)")
	fmt.Println("   ✓ is in(collection)")
	fmt.Println("   ✓ is contains(item)")
	fmt.Println()

	fmt.Println("6. NEGATED TESTS:")
	fmt.Println("   ✓ All tests support negation with 'is not'")
	fmt.Println("   ✓ Example: value is not defined")
	fmt.Println("   ✓ Example: number is not even")
	fmt.Println()

	fmt.Println("Total Tests: 26+")
	fmt.Println()

	fmt.Println("PRACTICAL USE CASES:")
	fmt.Println()

	fmt.Println("1. CONDITIONAL LOGIC:")
	fmt.Println("   ✓ User authentication checks")
	fmt.Println("   ✓ Permission validation")
	fmt.Println("   ✓ Data type validation")
	fmt.Println("   ✓ Empty/null checks")
	fmt.Println()

	fmt.Println("2. CALCULATIONS:")
	fmt.Println("   ✓ Price calculations")
	fmt.Println("   ✓ Discount application")
	fmt.Println("   ✓ Tax computation")
	fmt.Println("   ✓ Quantity operations")
	fmt.Println()

	fmt.Println("3. DATA VALIDATION:")
	fmt.Println("   ✓ Type checking before operations")
	fmt.Println("   ✓ Range validation")
	fmt.Println("   ✓ String format validation")
	fmt.Println("   ✓ Collection membership checks")
	fmt.Println()

	fmt.Println("4. COMPLEX EXPRESSIONS:")
	fmt.Println("   ✓ Nested conditions with and/or/not")
	fmt.Println("   ✓ Chained comparisons")
	fmt.Println("   ✓ Tests in ternary expressions")
	fmt.Println("   ✓ Tests in loops and filters")
	fmt.Println()

	fmt.Println("ALL OPERATORS AND TESTS: 100% Jinja2 Compatible ✨")
}
