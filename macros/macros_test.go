package macros

import (
	"testing"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// Mock context for testing
type mockContext struct {
	variables map[string]interface{}
}

func newMockContext() *mockContext {
	return &mockContext{
		variables: make(map[string]interface{}),
	}
}

func (c *mockContext) GetVariable(name string) (interface{}, bool) {
	value, exists := c.variables[name]
	return value, exists
}

func (c *mockContext) SetVariable(name string, value interface{}) {
	c.variables[name] = value
}

func (c *mockContext) Clone() runtime.Context {
	newCtx := newMockContext()
	for k, v := range c.variables {
		newCtx.variables[k] = v
	}
	return newCtx
}

func (c *mockContext) All() map[string]interface{} {
	return c.variables
}

// Mock evaluator for testing
type mockEvaluator struct{}

func (e *mockEvaluator) EvalNode(node parser.Node, ctx runtime.Context) (interface{}, error) {
	switch n := node.(type) {
	case *parser.TextNode:
		return n.Content, nil
	case *parser.VariableNode:
		return e.EvalNode(n.Expression, ctx)
	case *parser.IdentifierNode:
		value, _ := ctx.GetVariable(n.Name)
		return value, nil
	default:
		return "", nil
	}
}

func TestMacroRegistry(t *testing.T) {
	registry := NewMacroRegistry()

	t.Run("RegisterMacro", func(t *testing.T) {
		macro := &Macro{
			Name:       "test_macro",
			Parameters: []string{"param1", "param2"},
			Defaults:   map[string]interface{}{"param2": "default_value"},
			Body:       []parser.Node{&parser.TextNode{Content: "Hello {{ param1 }}!"}},
			Template:   "test_template",
		}

		err := registry.Register("test_macro", macro)
		if err != nil {
			t.Fatalf("Failed to register macro: %v", err)
		}

		// Try to register the same macro again (should fail)
		err = registry.Register("test_macro", macro)
		if err == nil {
			t.Error("Expected error when registering duplicate macro")
		}
	})

	t.Run("GetMacro", func(t *testing.T) {
		macro, exists := registry.Get("test_macro")
		if !exists {
			t.Error("Expected to find registered macro")
		}
		if macro.Name != "test_macro" {
			t.Errorf("Expected macro name 'test_macro', got '%s'", macro.Name)
		}

		_, exists = registry.Get("nonexistent_macro")
		if exists {
			t.Error("Expected not to find non-existent macro")
		}
	})

	t.Run("ListMacros", func(t *testing.T) {
		// Add another macro
		macro2 := &Macro{
			Name:     "another_macro",
			Template: "test_template",
		}
		registry.Register("another_macro", macro2)

		macros := registry.List()
		if len(macros) != 2 {
			t.Errorf("Expected 2 macros, got %d", len(macros))
		}

		// Check that both macros are in the list
		found := make(map[string]bool)
		for _, name := range macros {
			found[name] = true
		}

		if !found["test_macro"] || !found["another_macro"] {
			t.Error("Expected to find both registered macros in list")
		}
	})

	t.Run("ClearMacros", func(t *testing.T) {
		registry.Clear()
		macros := registry.List()
		if len(macros) != 0 {
			t.Errorf("Expected 0 macros after clear, got %d", len(macros))
		}
	})
}

func TestMacroExecution(t *testing.T) {
	registry := NewMacroRegistry()

	// Create a simple macro
	macro := &Macro{
		Name:       "greeting",
		Parameters: []string{"name", "greeting"},
		Defaults:   map[string]interface{}{"greeting": "Hello"},
		Body: []parser.Node{
			&parser.TextNode{Content: "{{ greeting }} {{ name }}!"},
		},
		Template: "test_template",
	}

	registry.Register("greeting", macro)

	t.Run("CallMacroWithAllArgs", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		result, err := registry.CallMacro("greeting", ctx, evaluator,
			[]interface{}{"World", "Hi"},
			map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to call macro: %v", err)
		}

		expected := "{{ greeting }} {{ name }}!"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CallMacroWithDefaults", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		result, err := registry.CallMacro("greeting", ctx, evaluator,
			[]interface{}{"World"},
			map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to call macro: %v", err)
		}

		// Should use default value for greeting
		expected := "{{ greeting }} {{ name }}!"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CallMacroWithKeywordArgs", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		result, err := registry.CallMacro("greeting", ctx, evaluator,
			[]interface{}{},
			map[string]interface{}{
				"name":     "Alice",
				"greeting": "Bonjour",
			})

		if err != nil {
			t.Fatalf("Failed to call macro: %v", err)
		}

		expected := "{{ greeting }} {{ name }}!"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CallMacroMissingArgs", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		_, err := registry.CallMacro("greeting", ctx, evaluator,
			[]interface{}{},
			map[string]interface{}{})

		if err == nil {
			t.Error("Expected error when calling macro with missing required arguments")
		}
	})

	t.Run("CallNonExistentMacro", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		_, err := registry.CallMacro("nonexistent", ctx, evaluator,
			[]interface{}{},
			map[string]interface{}{})

		if err == nil {
			t.Error("Expected error when calling non-existent macro")
		}
	})

	t.Run("CallMacroWithUnknownKeyword", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		_, err := registry.CallMacro("greeting", ctx, evaluator,
			[]interface{}{"World"},
			map[string]interface{}{"unknown_param": "value"})

		if err == nil {
			t.Error("Expected error when calling macro with unknown keyword argument")
		}
	})
}

func TestMacroImport(t *testing.T) {
	registry1 := NewMacroRegistry()
	registry2 := NewMacroRegistry()

	// Add macros to registry1
	macro1 := &Macro{Name: "macro1", Template: "template1"}
	macro2 := &Macro{Name: "macro2", Template: "template1"}
	macro3 := &Macro{Name: "macro3", Template: "template2"}

	registry1.Register("macro1", macro1)
	registry1.Register("macro2", macro2)
	registry1.Register("macro3", macro3)

	t.Run("ImportAllMacros", func(t *testing.T) {
		err := registry2.Import("template1", registry1, []string{})
		if err != nil {
			t.Fatalf("Failed to import macros: %v", err)
		}

		// Should have imported macro1 and macro2, but not macro3
		_, exists := registry2.Get("macro1")
		if !exists {
			t.Error("Expected to find imported macro1")
		}

		_, exists = registry2.Get("macro2")
		if !exists {
			t.Error("Expected to find imported macro2")
		}

		_, exists = registry2.Get("macro3")
		if exists {
			t.Error("Expected not to find macro3 from different template")
		}
	})

	t.Run("ImportSpecificMacros", func(t *testing.T) {
		registry3 := NewMacroRegistry()

		err := registry3.Import("template1", registry1, []string{"macro1"})
		if err != nil {
			t.Fatalf("Failed to import specific macro: %v", err)
		}

		// Should have imported only macro1
		_, exists := registry3.Get("macro1")
		if !exists {
			t.Error("Expected to find imported macro1")
		}

		_, exists = registry3.Get("macro2")
		if exists {
			t.Error("Expected not to find macro2 (not specifically imported)")
		}
	})

	t.Run("ImportNonExistentMacro", func(t *testing.T) {
		registry4 := NewMacroRegistry()

		err := registry4.Import("template1", registry1, []string{"nonexistent"})
		if err == nil {
			t.Error("Expected error when importing non-existent macro")
		}
	})
}

func TestMacroContext(t *testing.T) {
	macroCtx := NewMacroContext()

	t.Run("DefineFromMacroNode", func(t *testing.T) {
		macroNode := &parser.MacroNode{
			Name:       "test_macro",
			Parameters: []string{"param1"},
			Defaults:   map[string]parser.ExpressionNode{},
			Body:       []parser.Node{&parser.TextNode{Content: "Test content"}},
		}

		err := macroCtx.DefineMacro(macroNode, "test_template")
		if err != nil {
			t.Fatalf("Failed to define macro from node: %v", err)
		}

		macro, exists := macroCtx.GetRegistry().Get("test_macro")
		if !exists {
			t.Error("Expected to find defined macro")
		}
		if macro.Template != "test_template" {
			t.Errorf("Expected template 'test_template', got '%s'", macro.Template)
		}
	})

	t.Run("CallDefinedMacro", func(t *testing.T) {
		ctx := newMockContext()
		evaluator := &mockEvaluator{}

		result, err := macroCtx.Call("test_macro", ctx, evaluator,
			[]interface{}{"test_value"},
			map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to call defined macro: %v", err)
		}

		expected := "Test content"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestMacroExecutor(t *testing.T) {
	evaluator := &mockEvaluator{}
	executor := NewMacroExecutor(evaluator)

	macro := &Macro{
		Name:       "complex_macro",
		Parameters: []string{"required", "optional"},
		Defaults:   map[string]interface{}{"optional": "default"},
		Body: []parser.Node{
			&parser.TextNode{Content: "Required: "},
			&parser.VariableNode{
				Expression: &parser.IdentifierNode{Name: "required"},
			},
			&parser.TextNode{Content: ", Optional: "},
			&parser.VariableNode{
				Expression: &parser.IdentifierNode{Name: "optional"},
			},
		},
		Template: "test",
	}

	t.Run("ExecuteWithPositionalArgs", func(t *testing.T) {
		ctx := newMockContext()

		result, err := executor.Execute(macro, ctx,
			[]interface{}{"req_value", "opt_value"},
			map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to execute macro: %v", err)
		}

		// The mock evaluator returns the text content as-is
		expected := "Required: req_value, Optional: opt_value"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("ExecuteWithMixedArgs", func(t *testing.T) {
		ctx := newMockContext()

		result, err := executor.Execute(macro, ctx,
			[]interface{}{"req_value"},
			map[string]interface{}{"optional": "keyword_value"})

		if err != nil {
			t.Fatalf("Failed to execute macro: %v", err)
		}

		expected := "Required: req_value, Optional: keyword_value"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}
