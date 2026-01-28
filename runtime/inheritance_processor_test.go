package runtime

import (
	"fmt"
	"testing"

	"github.com/zipreport/miya/parser"
)

// mockEnvironment implements EnvironmentInterface for testing
type mockEnvironment struct {
	templates map[string]*mockTemplate
	loader    interface{}
}

func newMockEnvironment() *mockEnvironment {
	return &mockEnvironment{
		templates: make(map[string]*mockTemplate),
	}
}

func (m *mockEnvironment) GetTemplate(name string) (TemplateInterface, error) {
	if tmpl, ok := m.templates[name]; ok {
		return tmpl, nil
	}
	return nil, fmt.Errorf("template not found: %s", name)
}

func (m *mockEnvironment) GetLoader() interface{} {
	return m.loader
}

func (m *mockEnvironment) addTemplate(name string, ast *parser.TemplateNode) {
	m.templates[name] = &mockTemplate{name: name, ast: ast}
}

// mockTemplate implements TemplateInterface for testing
type mockTemplate struct {
	name string
	ast  *parser.TemplateNode
}

func (m *mockTemplate) AST() *parser.TemplateNode {
	return m.ast
}

func (m *mockTemplate) Name() string {
	return m.name
}

// TestNewInheritanceProcessor tests creating a new inheritance processor
func TestNewInheritanceProcessor(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	if processor == nil {
		t.Fatal("NewInheritanceProcessor returned nil")
	}
	if processor.env != env {
		t.Error("processor env not set correctly")
	}
	if processor.cache == nil {
		t.Error("processor cache should not be nil")
	}
}

// TestNewInheritanceProcessorWithCache tests creating a processor with shared cache
func TestNewInheritanceProcessorWithCache(t *testing.T) {
	env := newMockEnvironment()
	cache := NewInheritanceCache()
	processor := NewInheritanceProcessorWithCache(env, cache)

	if processor == nil {
		t.Fatal("NewInheritanceProcessorWithCache returned nil")
	}
	if processor.cache != cache {
		t.Error("processor should use provided cache")
	}
}

// TestResolveInheritanceNoInheritance tests resolving a template without inheritance
func TestResolveInheritanceNoInheritance(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	// Create a simple template without extends
	ast := &parser.TemplateNode{
		Children: []parser.Node{
			&parser.TextNode{Content: "Hello World"},
		},
	}

	template := &mockTemplate{name: "simple.html", ast: ast}
	ctx := &simpleContext{variables: make(map[string]interface{})}

	result, err := processor.ResolveInheritance(template, ctx)
	if err != nil {
		t.Fatalf("ResolveInheritance failed: %v", err)
	}
	if result != ast {
		t.Error("expected original AST to be returned for template without inheritance")
	}
}

// TestResolveInheritanceWithExtends tests resolving a template with extends
func TestResolveInheritanceWithExtends(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	// Create base template with a block
	baseAST := &parser.TemplateNode{
		Children: []parser.Node{
			&parser.TextNode{Content: "<html>"},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.TextNode{Content: "Base content"},
				},
			},
			&parser.TextNode{Content: "</html>"},
		},
	}
	env.addTemplate("base.html", baseAST)

	// Create child template that extends base
	childAST := &parser.TemplateNode{
		Children: []parser.Node{
			&parser.ExtendsNode{
				Template: &parser.LiteralNode{Value: "base.html"},
			},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.TextNode{Content: "Child content"},
				},
			},
		},
	}

	template := &mockTemplate{name: "child.html", ast: childAST}
	ctx := &simpleContext{variables: make(map[string]interface{})}

	result, err := processor.ResolveInheritance(template, ctx)
	if err != nil {
		t.Fatalf("ResolveInheritance failed: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// TestHasInheritanceDirectives tests detecting inheritance directives
func TestHasInheritanceDirectives(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	t.Run("with extends", func(t *testing.T) {
		ast := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.ExtendsNode{
					Template: &parser.LiteralNode{Value: "base.html"},
				},
			},
		}
		if !processor.hasInheritanceDirectives(ast) {
			t.Error("expected true for template with extends")
		}
	})

	t.Run("without extends", func(t *testing.T) {
		ast := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.TextNode{Content: "Hello"},
			},
		}
		if processor.hasInheritanceDirectives(ast) {
			t.Error("expected false for template without extends")
		}
	})

	t.Run("with include", func(t *testing.T) {
		ast := &parser.TemplateNode{
			Children: []parser.Node{
				&parser.IncludeNode{
					Template: &parser.LiteralNode{Value: "partial.html"},
				},
			},
		}
		// Include should not count as inheritance
		if processor.hasInheritanceDirectives(ast) {
			t.Error("include should not count as inheritance directive")
		}
	})
}

// TestInheritanceGetCache tests getting the cache
func TestInheritanceGetCache(t *testing.T) {
	env := newMockEnvironment()
	cache := NewInheritanceCache()
	processor := NewInheritanceProcessorWithCache(env, cache)

	if processor.GetCache() != cache {
		t.Error("GetCache should return the same cache")
	}
}

// TestInheritanceInvalidateTemplate tests cache invalidation
func TestInheritanceInvalidateTemplate(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	// Just verify InvalidateTemplate doesn't panic
	processor.InvalidateTemplate("test.html")
	processor.InvalidateTemplate("nonexistent.html")
}

// TestInheritanceClearCache tests clearing all caches
func TestInheritanceClearCache(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	// Just verify ClearCache doesn't panic
	processor.ClearCache()

	// Get stats to verify cache was cleared
	stats := processor.GetCacheStats()
	if stats.HierarchyCache.Entries != 0 {
		t.Error("hierarchy cache should be cleared")
	}
}

// TestInheritanceGetCacheStats tests getting cache statistics
func TestInheritanceGetCacheStats(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	stats := processor.GetCacheStats()

	// Just verify it returns a valid struct
	_ = stats.HierarchyCache.Hits
}

// TestIsSelfReferencingBlock tests detecting self-referencing blocks
func TestIsSelfReferencingBlock(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	t.Run("non-self-referencing", func(t *testing.T) {
		original := &parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "Original"},
			},
		}
		override := &parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "Override"},
			},
		}
		if processor.isSelfReferencingBlock(original, override) {
			t.Error("expected false for different block pointers")
		}
	})

	t.Run("self-referencing same pointer", func(t *testing.T) {
		block := &parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "Content"},
			},
		}
		// Same pointer means it's self-referencing
		if !processor.isSelfReferencingBlock(block, block) {
			t.Error("expected true for same block pointer")
		}
	})
}

// TestBlockContainsSuperCalls tests detecting super() calls in blocks
func TestBlockContainsSuperCalls(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	t.Run("with super call", func(t *testing.T) {
		block := &parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.VariableNode{
					Expression: &parser.SuperNode{},
				},
			},
		}
		if !processor.blockContainsSuperCalls(block) {
			t.Error("expected to detect super() call")
		}
	})

	t.Run("without super call", func(t *testing.T) {
		block := &parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "Hello"},
			},
		}
		if processor.blockContainsSuperCalls(block) {
			t.Error("expected not to detect super() call")
		}
	})
}

// TestCloneTemplateNode tests cloning template nodes
func TestCloneTemplateNode(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	original := &parser.TemplateNode{
		Children: []parser.Node{
			&parser.TextNode{Content: "Hello"},
			&parser.VariableNode{
				Expression: &parser.IdentifierNode{Name: "name"},
			},
		},
	}

	clone := processor.cloneTemplateNode(original)

	if clone == nil {
		t.Fatal("clone should not be nil")
	}
	if clone == original {
		t.Error("clone should be a different object")
	}
	if len(clone.Children) != len(original.Children) {
		t.Error("clone should have same number of children")
	}
}

// TestCloneNode tests cloning individual nodes
func TestCloneNode(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	t.Run("TextNode", func(t *testing.T) {
		node := &parser.TextNode{Content: "Hello"}
		clone := processor.cloneNode(node)
		if clone == nil {
			t.Fatal("clone should not be nil")
		}
		if textClone, ok := clone.(*parser.TextNode); ok {
			if textClone.Content != "Hello" {
				t.Error("content should be preserved")
			}
		} else {
			t.Error("clone should be TextNode")
		}
	})

	t.Run("VariableNode", func(t *testing.T) {
		node := &parser.VariableNode{
			Expression: &parser.IdentifierNode{Name: "var"},
		}
		clone := processor.cloneNode(node)
		if clone == nil {
			t.Fatal("clone should not be nil")
		}
	})

	t.Run("BlockNode", func(t *testing.T) {
		node := &parser.BlockNode{
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "Block content"},
			},
		}
		clone := processor.cloneNode(node)
		if clone == nil {
			t.Fatal("clone should not be nil")
		}
		if blockClone, ok := clone.(*parser.BlockNode); ok {
			if blockClone.Name != "content" {
				t.Error("block name should be preserved")
			}
		}
	})

	t.Run("IfNode", func(t *testing.T) {
		node := &parser.IfNode{
			Condition: &parser.IdentifierNode{Name: "cond"},
			Body: []parser.Node{
				&parser.TextNode{Content: "True branch"},
			},
		}
		clone := processor.cloneNode(node)
		if clone == nil {
			t.Fatal("clone should not be nil")
		}
	})

	t.Run("ForNode", func(t *testing.T) {
		node := &parser.ForNode{
			Variables: []string{"item"},
			Iterable:  &parser.IdentifierNode{Name: "items"},
			Body: []parser.Node{
				&parser.TextNode{Content: "Loop body"},
			},
		}
		clone := processor.cloneNode(node)
		if clone == nil {
			t.Fatal("clone should not be nil")
		}
	})
}

// TestReplaceBlocks tests replacing blocks in templates
func TestReplaceBlocks(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	// Create a template with blocks
	tmpl := &parser.TemplateNode{
		Children: []parser.Node{
			&parser.TextNode{Content: "Before"},
			&parser.BlockNode{
				Name: "content",
				Body: []parser.Node{
					&parser.TextNode{Content: "Original content"},
				},
			},
			&parser.TextNode{Content: "After"},
		},
	}

	// Create replacement blocks
	replacements := map[string]*parser.BlockNode{
		"content": {
			Name: "content",
			Body: []parser.Node{
				&parser.TextNode{Content: "New content"},
			},
		},
	}

	ctx := &simpleContext{variables: make(map[string]interface{})}

	err := processor.replaceBlocks(tmpl, replacements, ctx)
	if err != nil {
		t.Fatalf("replaceBlocks failed: %v", err)
	}
}

// TestDebugPrintTemplate tests the debug output function
func TestDebugPrintTemplate(t *testing.T) {
	env := newMockEnvironment()
	processor := NewInheritanceProcessor(env)

	ast := &parser.TemplateNode{
		Children: []parser.Node{
			&parser.TextNode{Content: "Hello"},
		},
	}

	// Just verify it doesn't panic
	processor.debugPrintTemplate(ast, 0)
}
