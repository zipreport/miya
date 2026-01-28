package inheritance

import (
	"strings"
	"testing"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// Helper function to create test templates manually
func createSimpleTemplate(name string, children []parser.Node) *parser.TemplateNode {
	template := &parser.TemplateNode{
		Name:     name,
		Children: children,
	}
	return template
}

func TestInheritanceResolver(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Create simple templates for testing resolver directly
	baseTemplate := &parser.TemplateNode{
		Name: "base.html",
		Children: []parser.Node{
			&parser.BlockNode{Name: "content", Body: []parser.Node{
				&parser.TextNode{Content: "Base content"},
			}},
		},
	}

	childTemplate := &parser.TemplateNode{
		Name: "child.html",
		Children: []parser.Node{
			&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "base.html"}},
			&parser.BlockNode{Name: "content", Body: []parser.Node{
				&parser.SuperNode{},
				&parser.TextNode{Content: " + Child content"},
			}},
		},
	}

	loader.AddTemplate("base.html", baseTemplate)
	loader.AddTemplate("child.html", childTemplate)

	resolver := NewInheritanceResolver(loader)

	// Test inheritance resolution
	ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
	if err != nil {
		t.Fatalf("Failed to resolve inheritance: %v", err)
	}

	// Check that blocks were collected correctly
	if len(ctx.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(ctx.Blocks))
	}

	contentBlock, exists := ctx.Blocks["content"]
	if !exists {
		t.Error("Content block not found")
	}

	if contentBlock.Parent == nil {
		t.Error("Parent block not linked for super() support")
	}

	// Test final template building
	finalTemplate, err := resolver.BuildFinalTemplate(ctx)
	if err != nil {
		t.Fatalf("Failed to build final template: %v", err)
	}

	if finalTemplate.Name != "child.html" {
		t.Errorf("Expected template name 'child.html', got '%s'", finalTemplate.Name)
	}
}

func TestMemoryTemplateLoader(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	template := &parser.TemplateNode{
		Name: "test.html",
		Children: []parser.Node{
			&parser.TextNode{Content: "Test content"},
		},
	}

	// Test adding and loading template
	loader.AddTemplate("test.html", template)

	loadedTemplate, err := loader.LoadTemplate("test.html")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	if loadedTemplate.Name != template.Name {
		t.Errorf("Expected template name '%s', got '%s'", template.Name, loadedTemplate.Name)
	}

	// Test loading non-existent template
	_, err = loader.LoadTemplate("nonexistent.html")
	if err == nil {
		t.Error("Expected error when loading non-existent template")
	}

	// Test template name resolution
	resolvedName := loader.ResolveTemplateName("test.html")
	if resolvedName != "test.html" {
		t.Errorf("Expected resolved name 'test.html', got '%s'", resolvedName)
	}
}

func TestCreateTemplateNameNode(t *testing.T) {
	node := CreateTemplateNameNode("test.html")

	literal, ok := node.(*parser.LiteralNode)
	if !ok {
		t.Errorf("Expected LiteralNode, got %T", node)
	}

	value, ok := literal.Value.(string)
	if !ok {
		t.Errorf("Expected string value, got %T", literal.Value)
	}

	if value != "test.html" {
		t.Errorf("Expected 'test.html', got '%s'", value)
	}
}

func TestBasicBlockInheritance(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Base template with a simple block
	baseTemplate := createSimpleTemplate("base.html", []parser.Node{
		&parser.TextNode{Content: "Before block: "},
		&parser.BlockNode{
			Name: "content",
			Body: []parser.Node{&parser.TextNode{Content: "Default content"}},
		},
		&parser.TextNode{Content: " :After block"},
	})

	// Child template that extends and overrides the block
	childTemplate := createSimpleTemplate("child.html", []parser.Node{
		&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "base.html"}},
		&parser.BlockNode{
			Name: "content",
			Body: []parser.Node{&parser.TextNode{Content: "Child content"}},
		},
	})

	loader.AddTemplate("base.html", baseTemplate)
	loader.AddTemplate("child.html", childTemplate)

	resolver := NewInheritanceResolver(loader)

	// Resolve inheritance
	ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
	if err != nil {
		t.Fatalf("Failed to resolve inheritance: %v", err)
	}

	// Build final template
	finalTemplate, err := resolver.BuildFinalTemplate(ctx)
	if err != nil {
		t.Fatalf("Failed to build final template: %v", err)
	}

	// Check that the final template has the expected structure
	if len(finalTemplate.Children) == 0 {
		t.Error("Final template has no children")
	}

	// Verify the child block overrode the parent block
	contentBlock, exists := ctx.Blocks["content"]
	if !exists {
		t.Error("Content block not found in resolved context")
	}

	if len(contentBlock.Body) != 1 {
		t.Errorf("Expected 1 node in content block body, got %d", len(contentBlock.Body))
	}

	textNode, ok := contentBlock.Body[0].(*parser.TextNode)
	if !ok {
		t.Errorf("Expected TextNode, got %T", contentBlock.Body[0])
	}

	if textNode.Content != "Child content" {
		t.Errorf("Expected 'Child content', got '%s'", textNode.Content)
	}
}

func TestIncludeResolution(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Template to be included
	includedTemplate := createSimpleTemplate("included.html", []parser.Node{
		&parser.TextNode{Content: "Included content"},
	})

	// Main template with include
	mainTemplate := createSimpleTemplate("main.html", []parser.Node{
		&parser.TextNode{Content: "Before include: "},
		&parser.IncludeNode{Template: &parser.LiteralNode{Value: "included.html"}},
		&parser.TextNode{Content: " :After include"},
	})

	loader.AddTemplate("included.html", includedTemplate)
	loader.AddTemplate("main.html", mainTemplate)

	resolver := NewInheritanceResolver(loader)

	// Since this template doesn't use inheritance, resolve with no parent
	ctx := &TemplateContext{
		Name:     "main.html",
		Template: mainTemplate,
		Blocks:   make(map[string]*BlockContext),
	}

	// Process the template to resolve includes
	processedNodes, err := resolver.processNodes(mainTemplate.Children, ctx)
	if err != nil {
		t.Fatalf("Failed to process nodes: %v", err)
	}

	// Check that include was resolved
	if len(processedNodes) != 3 {
		t.Errorf("Expected 3 processed nodes, got %d", len(processedNodes))
	}

	// The middle node should be the included content
	if textNode, ok := processedNodes[1].(*parser.TextNode); ok {
		if textNode.Content != "Included content" {
			t.Errorf("Expected 'Included content', got '%s'", textNode.Content)
		}
	} else {
		t.Errorf("Expected TextNode from include resolution, got %T", processedNodes[1])
	}
}

func TestSuperNodeInBlock(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Base template
	baseTemplate := createSimpleTemplate("base.html", []parser.Node{
		&parser.BlockNode{
			Name: "content",
			Body: []parser.Node{&parser.TextNode{Content: "Base content"}},
		},
	})

	// Child template with super() call
	childTemplate := createSimpleTemplate("child.html", []parser.Node{
		&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "base.html"}},
		&parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "Child: "},
				&parser.SuperNode{},
				&parser.TextNode{Content: " :End"},
			},
		},
	})

	loader.AddTemplate("base.html", baseTemplate)
	loader.AddTemplate("child.html", childTemplate)

	resolver := NewInheritanceResolver(loader)

	// Resolve inheritance
	ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
	if err != nil {
		t.Fatalf("Failed to resolve inheritance: %v", err)
	}

	// Verify that the child block has a parent reference
	contentBlock := ctx.Blocks["content"]
	if contentBlock.Parent == nil {
		t.Error("Child block should have parent reference for super() support")
	}

	// Check that parent block has the expected content
	if len(contentBlock.Parent.Body) != 1 {
		t.Errorf("Expected 1 node in parent block, got %d", len(contentBlock.Parent.Body))
	}

	parentTextNode, ok := contentBlock.Parent.Body[0].(*parser.TextNode)
	if !ok {
		t.Errorf("Expected TextNode in parent block, got %T", contentBlock.Parent.Body[0])
	}

	if parentTextNode.Content != "Base content" {
		t.Errorf("Expected 'Base content' in parent block, got '%s'", parentTextNode.Content)
	}
}

// Test InheritanceEvaluator functions with 0% coverage
func TestInheritanceEvaluator(t *testing.T) {
	t.Run("NewInheritanceEvaluator", func(t *testing.T) {
		loader := NewMemoryTemplateLoader()

		evaluator := NewInheritanceEvaluator(loader)
		if evaluator == nil {
			t.Fatal("NewInheritanceEvaluator returned nil")
		}
		if evaluator.resolver == nil {
			t.Error("Resolver not initialized")
		}
	})

	t.Run("EvaluateTemplate", func(t *testing.T) {
		loader := NewMemoryTemplateLoader()

		// Create simple template nodes manually
		baseTemplate := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.TextNode{Content: "Base:"},
				&parser.BlockNode{
					Name: "content",
					Body: []parser.Node{
						&parser.TextNode{Content: "Base Content"},
					},
				},
			},
		}
		loader.AddTemplate("base.html", baseTemplate)

		childTemplate := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.ExtendsNode{
					Template: &parser.LiteralNode{Value: "base.html"},
				},
				&parser.BlockNode{
					Name: "content",
					Body: []parser.Node{
						&parser.TextNode{Content: "Child Content"},
					},
				},
			},
		}
		loader.AddTemplate("child.html", childTemplate)

		evaluator := NewInheritanceEvaluator(loader)

		// Create a simple context
		ctx := &testContext{data: make(map[string]interface{})}

		result, err := evaluator.EvaluateTemplate(childTemplate, "child.html", ctx)
		if err != nil {
			t.Fatalf("EvaluateTemplate failed: %v", err)
		}

		if result == "" {
			t.Error("Expected non-empty result")
		}
	})

	t.Run("EvaluateTemplate with missing template", func(t *testing.T) {
		loader := NewMemoryTemplateLoader()
		evaluator := NewInheritanceEvaluator(loader)

		emptyTemplate := &parser.TemplateNode{
			Children: []parser.Node{},
		}

		ctx := &testContext{data: make(map[string]interface{})}

		// This should work even with empty template
		result, err := evaluator.EvaluateTemplate(emptyTemplate, "test.html", ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != "" {
			t.Error("Expected empty result for empty template")
		}
	})
}

// Test additional coverage for functions
func TestAdditionalCoverage(t *testing.T) {
	t.Run("Test inheritance resolution with complex templates", func(t *testing.T) {
		loader := NewMemoryTemplateLoader()

		// Create a template with extends and blocks
		baseTemplate := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.TextNode{Content: "Base: "},
				&parser.BlockNode{
					Name: "content",
					Body: []parser.Node{
						&parser.TextNode{Content: "default content"},
					},
				},
			},
		}
		loader.AddTemplate("base.html", baseTemplate)

		childTemplate := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.ExtendsNode{
					Template: &parser.LiteralNode{Value: "base.html"},
				},
				&parser.BlockNode{
					Name: "content",
					Body: []parser.Node{
						&parser.TextNode{Content: "overridden content"},
					},
				},
			},
		}

		resolver := NewInheritanceResolver(loader)

		// This will exercise various internal functions
		ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
		if err != nil {
			t.Fatalf("ResolveInheritance failed: %v", err)
		}

		if ctx == nil {
			t.Fatal("Expected non-nil inheritance context")
		}

		// Build final template to exercise more code paths
		finalTemplate, err := resolver.BuildFinalTemplate(ctx)
		if err != nil {
			t.Fatalf("BuildFinalTemplate failed: %v", err)
		}

		if finalTemplate == nil {
			t.Fatal("Expected non-nil final template")
		}

		if len(finalTemplate.Children) == 0 {
			t.Error("Expected final template to have children")
		}
	})
}

// Test error conditions to exercise more code paths
func TestErrorConditions(t *testing.T) {
	t.Run("Test with invalid templates", func(t *testing.T) {
		loader := NewMemoryTemplateLoader()

		// Test with template that extends non-existent base
		childTemplate := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.ExtendsNode{
					Template: &parser.LiteralNode{Value: "nonexistent.html"},
				},
			},
		}

		resolver := NewInheritanceResolver(loader)

		_, err := resolver.ResolveInheritance(childTemplate, "child.html")
		if err == nil {
			t.Error("Expected error when extending non-existent template")
		}
	})

	t.Run("Test with empty templates", func(t *testing.T) {
		loader := NewMemoryTemplateLoader()

		emptyTemplate := &parser.TemplateNode{
			Children: []parser.Node{},
		}

		resolver := NewInheritanceResolver(loader)

		ctx, err := resolver.ResolveInheritance(emptyTemplate, "empty.html")
		if err != nil {
			t.Errorf("Unexpected error with empty template: %v", err)
		}

		if ctx == nil {
			t.Error("Expected non-nil context for empty template")
		}

		finalTemplate, err := resolver.BuildFinalTemplate(ctx)
		if err != nil {
			t.Errorf("Unexpected error building empty template: %v", err)
		}

		if finalTemplate == nil {
			t.Error("Expected non-nil final template")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Test context implementation
type testContext struct {
	data map[string]interface{}
}

func (c *testContext) Get(name string) interface{} {
	return c.data[name]
}

func (c *testContext) Set(name string, value interface{}) {
	c.data[name] = value
}

func (c *testContext) Has(name string) bool {
	_, exists := c.data[name]
	return exists
}

func (c *testContext) Clone() runtime.Context {
	newData := make(map[string]interface{})
	for k, v := range c.data {
		newData[k] = v
	}
	return &testContext{data: newData}
}

func (c *testContext) GetVariable(key string) (interface{}, bool) {
	val, exists := c.data[key]
	return val, exists
}

func (c *testContext) SetVariable(key string, value interface{}) {
	c.data[key] = value
}

func (c *testContext) All() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// TestProcessNodeWithIfNode tests processing templates with IfNode
func TestProcessNodeWithIfNode(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Template with if node inside a block
	baseTemplate := &parser.TemplateNode{
		Name: "base.html",
		Children: []parser.Node{
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.IfNode{
						Condition: &parser.LiteralNode{Value: true},
						Body: []parser.Node{
							&parser.TextNode{Content: "True branch"},
						},
						Else: []parser.Node{
							&parser.TextNode{Content: "False branch"},
						},
						ElseIfs: []*parser.IfNode{
							{
								Condition: &parser.LiteralNode{Value: false},
								Body: []parser.Node{
									&parser.TextNode{Content: "Elif branch"},
								},
							},
						},
					},
				},
			},
		},
	}

	loader.AddTemplate("base.html", baseTemplate)

	resolver := NewInheritanceResolver(loader)

	ctx := &TemplateContext{
		Name:     "base.html",
		Template: baseTemplate,
		Blocks:   make(map[string]*BlockContext),
	}

	// Process nodes to exercise processNode with IfNode
	processedNodes, err := resolver.processNodes(baseTemplate.Children, ctx)
	if err != nil {
		t.Fatalf("processNodes failed: %v", err)
	}

	if len(processedNodes) == 0 {
		t.Error("Expected processed nodes")
	}
}

// TestProcessNodeWithForNode tests processing templates with ForNode
func TestProcessNodeWithForNode(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Template with for node inside a block
	baseTemplate := &parser.TemplateNode{
		Name: "base.html",
		Children: []parser.Node{
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.ForNode{
						Variables: []string{"item"},
						Iterable:  &parser.LiteralNode{Value: []interface{}{1, 2, 3}},
						Body: []parser.Node{
							&parser.TextNode{Content: "Item"},
						},
						Else: []parser.Node{
							&parser.TextNode{Content: "No items"},
						},
					},
				},
			},
		},
	}

	loader.AddTemplate("base.html", baseTemplate)

	resolver := NewInheritanceResolver(loader)

	ctx := &TemplateContext{
		Name:     "base.html",
		Template: baseTemplate,
		Blocks:   make(map[string]*BlockContext),
	}

	// Process nodes to exercise processNode with ForNode
	processedNodes, err := resolver.processNodes(baseTemplate.Children, ctx)
	if err != nil {
		t.Fatalf("processNodes failed: %v", err)
	}

	if len(processedNodes) == 0 {
		t.Error("Expected processed nodes")
	}
}

// TestProcessNodeWithSuperInIfNode tests super() inside an IfNode
func TestProcessNodeWithSuperInIfNode(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Base template with a block
	baseTemplate := &parser.TemplateNode{
		Name: "base.html",
		Children: []parser.Node{
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.TextNode{Content: "Base content"},
				},
			},
		},
	}

	// Child template with super() inside an if node
	childTemplate := &parser.TemplateNode{
		Name: "child.html",
		Children: []parser.Node{
			&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "base.html"}},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.IfNode{
						Condition: &parser.LiteralNode{Value: true},
						Body: []parser.Node{
							&parser.TextNode{Content: "Before: "},
							&parser.SuperNode{},
							&parser.TextNode{Content: " :After"},
						},
						Else: []parser.Node{
							&parser.TextNode{Content: "Else branch"},
						},
						ElseIfs: []*parser.IfNode{
							{
								Condition: &parser.LiteralNode{Value: false},
								Body: []parser.Node{
									&parser.SuperNode{},
								},
							},
						},
					},
				},
			},
		},
	}

	loader.AddTemplate("base.html", baseTemplate)
	loader.AddTemplate("child.html", childTemplate)

	resolver := NewInheritanceResolver(loader)

	// Resolve inheritance
	ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
	if err != nil {
		t.Fatalf("ResolveInheritance failed: %v", err)
	}

	// Build final template
	finalTemplate, err := resolver.BuildFinalTemplate(ctx)
	if err != nil {
		t.Fatalf("BuildFinalTemplate failed: %v", err)
	}

	if finalTemplate == nil {
		t.Fatal("Expected non-nil final template")
	}
}

// TestProcessNodeWithSuperInForNode tests super() inside a ForNode
func TestProcessNodeWithSuperInForNode(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Base template with a block
	baseTemplate := &parser.TemplateNode{
		Name: "base.html",
		Children: []parser.Node{
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.TextNode{Content: "Base content"},
				},
			},
		},
	}

	// Child template with super() inside a for node
	childTemplate := &parser.TemplateNode{
		Name: "child.html",
		Children: []parser.Node{
			&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "base.html"}},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.ForNode{
						Variables: []string{"item"},
						Iterable:  &parser.LiteralNode{Value: []interface{}{1}},
						Body: []parser.Node{
							&parser.SuperNode{},
						},
						Else: []parser.Node{
							&parser.TextNode{Content: "No items"},
						},
					},
				},
			},
		},
	}

	loader.AddTemplate("base.html", baseTemplate)
	loader.AddTemplate("child.html", childTemplate)

	resolver := NewInheritanceResolver(loader)

	// Resolve inheritance
	ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
	if err != nil {
		t.Fatalf("ResolveInheritance failed: %v", err)
	}

	// Build final template
	finalTemplate, err := resolver.BuildFinalTemplate(ctx)
	if err != nil {
		t.Fatalf("BuildFinalTemplate failed: %v", err)
	}

	if finalTemplate == nil {
		t.Fatal("Expected non-nil final template")
	}
}

// TestCollectBlocksFromNodeWithNestedStructures tests block collection from nested nodes
func TestCollectBlocksFromNodeWithNestedStructures(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Template with nested structure containing blocks
	template := &parser.TemplateNode{
		Name: "test.html",
		Children: []parser.Node{
			&parser.IfNode{
				Condition: &parser.LiteralNode{Value: true},
				Body: []parser.Node{
					&parser.BlockNode{
						Name: "inner_block",
						Body: []parser.Node{
							&parser.TextNode{Content: "Inside if"},
						},
					},
				},
				Else: []parser.Node{
					&parser.BlockNode{
						Name: "else_block",
						Body: []parser.Node{
							&parser.TextNode{Content: "In else"},
						},
					},
				},
			},
			&parser.ForNode{
				Variables: []string{"item"},
				Iterable:  &parser.LiteralNode{Value: []interface{}{1}},
				Body: []parser.Node{
					&parser.BlockNode{
						Name: "for_block",
						Body: []parser.Node{
							&parser.TextNode{Content: "In for"},
						},
					},
				},
			},
		},
	}

	loader.AddTemplate("test.html", template)

	resolver := NewInheritanceResolver(loader)

	// Create a proper TemplateContext with Blocks map
	ctx := &TemplateContext{
		Name:     "test.html",
		Template: template,
		Blocks:   make(map[string]*BlockContext),
	}
	err := resolver.collectBlocks(template.Children, ctx, "test.html")
	if err != nil {
		t.Fatalf("collectBlocks failed: %v", err)
	}

	// Check that blocks inside if, else, and for were collected
	if _, exists := ctx.Blocks["inner_block"]; !exists {
		t.Error("Expected inner_block to be collected")
	}
	if _, exists := ctx.Blocks["else_block"]; !exists {
		t.Error("Expected else_block to be collected")
	}
	if _, exists := ctx.Blocks["for_block"]; !exists {
		t.Error("Expected for_block to be collected")
	}
}

// TestEvaluateTemplateNameWithIdentifier tests template name evaluation with identifiers
func TestEvaluateTemplateNameWithIdentifier(t *testing.T) {
	loader := NewMemoryTemplateLoader()
	resolver := NewInheritanceResolver(loader)

	// Test with literal node containing string
	literalNode := &parser.LiteralNode{Value: "template.html"}
	name := resolver.evaluateTemplateName(literalNode)
	if name != "template.html" {
		t.Errorf("Expected 'template.html', got %q", name)
	}

	// Test with non-string literal (falls back to string conversion)
	intNode := &parser.LiteralNode{Value: 123}
	name = resolver.evaluateTemplateName(intNode)
	if name == "" {
		t.Errorf("Expected non-empty string for non-string literal")
	}

	// Test with non-literal node (falls back to string conversion)
	textNode := &parser.TextNode{Content: "text"}
	name = resolver.evaluateTemplateName(textNode)
	if name == "" {
		t.Errorf("Expected non-empty string for TextNode")
	}
}

// TestMultipleLevelInheritance tests three levels of template inheritance
func TestMultipleLevelInheritance(t *testing.T) {
	loader := NewMemoryTemplateLoader()

	// Grand-base template
	grandBase := &parser.TemplateNode{
		Name: "grandbase.html",
		Children: []parser.Node{
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.TextNode{Content: "Grand base"},
				},
			},
		},
	}

	// Base template extending grand-base
	baseTemplate := &parser.TemplateNode{
		Name: "base.html",
		Children: []parser.Node{
			&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "grandbase.html"}},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.SuperNode{},
					&parser.TextNode{Content: " + Base"},
				},
			},
		},
	}

	// Child template extending base
	childTemplate := &parser.TemplateNode{
		Name: "child.html",
		Children: []parser.Node{
			&parser.ExtendsNode{Template: &parser.LiteralNode{Value: "base.html"}},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.SuperNode{},
					&parser.TextNode{Content: " + Child"},
				},
			},
		},
	}

	loader.AddTemplate("grandbase.html", grandBase)
	loader.AddTemplate("base.html", baseTemplate)
	loader.AddTemplate("child.html", childTemplate)

	resolver := NewInheritanceResolver(loader)

	// Resolve inheritance
	ctx, err := resolver.ResolveInheritance(childTemplate, "child.html")
	if err != nil {
		t.Fatalf("ResolveInheritance failed: %v", err)
	}

	// Verify block chain is correctly set up
	contentBlock := ctx.Blocks["content"]
	if contentBlock == nil {
		t.Fatal("Expected content block")
	}
	if contentBlock.Parent == nil {
		t.Fatal("Expected content block to have parent")
	}
	if contentBlock.Parent.Parent == nil {
		t.Fatal("Expected parent block to have its own parent (3-level chain)")
	}

	// Build final template
	finalTemplate, err := resolver.BuildFinalTemplate(ctx)
	if err != nil {
		t.Fatalf("BuildFinalTemplate failed: %v", err)
	}

	if finalTemplate == nil {
		t.Fatal("Expected non-nil final template")
	}
}
