package miya

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

func TestRuntimeIntegration(t *testing.T) {
	// Create environment
	env := NewEnvironment()

	// Create context with test data
	ctx := NewContext()
	ctx.Set("name", "World")
	ctx.Set("items", []interface{}{"apple", "banana", "cherry"})
	ctx.Set("user", map[string]interface{}{
		"name": "John",
		"age":  30,
	})

	tests := []struct {
		name     string
		ast      parser.Node
		expected string
	}{
		{
			name:     "simple text",
			ast:      parser.NewTextNode("Hello World", 1, 1),
			expected: "Hello World",
		},
		{
			name: "variable substitution",
			ast: parser.NewVariableNode(
				parser.NewIdentifierNode("name", 1, 1),
				1, 1,
			),
			expected: "World",
		},
		{
			name: "literal values",
			ast: parser.NewVariableNode(
				parser.NewLiteralNode("test", "test", 1, 1),
				1, 1,
			),
			expected: "test",
		},
		{
			name: "arithmetic",
			ast: parser.NewVariableNode(
				parser.NewBinaryOpNode(
					parser.NewLiteralNode(5, "5", 1, 1),
					"+",
					parser.NewLiteralNode(3, "3", 1, 1),
					1, 1,
				),
				1, 1,
			),
			expected: "8",
		},
		{
			name: "string concatenation",
			ast: parser.NewVariableNode(
				parser.NewBinaryOpNode(
					parser.NewLiteralNode("Hello ", "Hello ", 1, 1),
					"+",
					parser.NewIdentifierNode("name", 1, 1),
					1, 1,
				),
				1, 1,
			),
			expected: "Hello World",
		},
		{
			name: "filter application",
			ast: parser.NewVariableNode(
				parser.NewFilterNode(
					parser.NewIdentifierNode("name", 1, 1),
					"upper",
					nil,
					1, 1,
				),
				1, 1,
			),
			expected: "WORLD",
		},
		{
			name: "if statement true",
			ast: parser.NewIfNode(
				parser.NewLiteralNode(true, "true", 1, 1),
				1, 1,
			),
			expected: "yes",
		},
		{
			name: "if statement false",
			ast: parser.NewIfNode(
				parser.NewLiteralNode(false, "false", 1, 1),
				1, 1,
			),
			expected: "",
		},
		{
			name: "set variable",
			ast: parser.NewSetNode(
				"newvar",
				parser.NewLiteralNode("test value", "test value", 1, 1),
				1, 1,
			),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create template with AST
			tmpl := &Template{
				name:   "test",
				source: "",
				env:    env,
				ast:    tt.ast,
			}

			// Special setup for if statement test
			if ifNode, ok := tt.ast.(*parser.IfNode); ok {
				ifNode.Body = append(ifNode.Body, parser.NewTextNode("yes", 1, 1))
			}

			// Render template
			result, err := tmpl.Render(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}

			// Special check for set variable test - SetNode should not modify original context
			if tt.name == "set variable" {
				// SetNode execution happens in template evaluation context, not the original context
				// This is correct behavior - the original context should remain unchanged
				if _, ok := ctx.Get("newvar"); ok {
					t.Errorf("variable should not be set in original context")
				}
			}
		})
	}
}

func TestForLoopIntegration(t *testing.T) {
	env := NewEnvironment()
	ctx := NewContext()
	ctx.Set("items", []interface{}{"a", "b", "c"})

	// Create for loop: {% for item in items %}{{ item }}{% endfor %}
	forNode := parser.NewSingleForNode(
		"item",
		parser.NewIdentifierNode("items", 1, 1),
		1, 1,
	)

	// Add body: {{ item }}
	varNode := parser.NewVariableNode(
		parser.NewIdentifierNode("item", 1, 1),
		1, 1,
	)
	forNode.Body = append(forNode.Body, varNode)

	tmpl := &Template{
		name:   "test_for",
		source: "",
		env:    env,
		ast:    forNode,
	}

	result, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "abc"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestComplexTemplate(t *testing.T) {
	env := NewEnvironment()
	ctx := NewContext()
	ctx.Set("users", []interface{}{
		map[string]interface{}{"name": "Alice", "age": 25},
		map[string]interface{}{"name": "Bob", "age": 30},
	})

	// Create template node with multiple children
	template := parser.NewTemplateNode("test", 1, 1)

	// Add text: "Users: "
	template.Children = append(template.Children,
		parser.NewTextNode("Users: ", 1, 1))

	// Add for loop: {% for user in users %}{{ user.name }} ({{ user.age }}) {% endfor %}
	forNode := parser.NewSingleForNode(
		"user",
		parser.NewIdentifierNode("users", 1, 1),
		1, 1,
	)

	// Add loop body
	forNode.Body = append(forNode.Body,
		parser.NewVariableNode(
			parser.NewAttributeNode(
				parser.NewIdentifierNode("user", 1, 1),
				"name",
				1, 1,
			),
			1, 1,
		),
		parser.NewTextNode(" (", 1, 1),
		parser.NewVariableNode(
			parser.NewAttributeNode(
				parser.NewIdentifierNode("user", 1, 1),
				"age",
				1, 1,
			),
			1, 1,
		),
		parser.NewTextNode(") ", 1, 1),
	)

	template.Children = append(template.Children, forNode)

	tmpl := &Template{
		name:   "test_complex",
		source: "",
		env:    env,
		ast:    template,
	}

	result, err := tmpl.Render(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Users: Alice (25) Bob (30) "
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
