package runtime

import (
	"strings"
	"testing"

	"github.com/zipreport/miya/parser"
)

// TestImportedNamespace tests the ImportedNamespace type
func TestImportedNamespace(t *testing.T) {
	t.Run("Get macro", func(t *testing.T) {
		macro := &TemplateMacro{
			Name:       "test_macro",
			Parameters: []string{"arg1"},
			Defaults:   make(map[string]interface{}),
			Body:       []parser.Node{},
			Context:    &simpleContext{variables: make(map[string]interface{})},
		}

		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       map[string]*TemplateMacro{"test_macro": macro},
			Variables:    make(map[string]interface{}),
			Context:      &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		result, ok := in.Get("test_macro")
		if !ok {
			t.Error("expected to find macro")
		}
		if result == nil {
			t.Error("expected non-nil function for macro")
		}
	})

	t.Run("Get variable", func(t *testing.T) {
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       make(map[string]*TemplateMacro),
			Variables:    map[string]interface{}{"my_var": "my_value"},
			Context:      &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		result, ok := in.Get("my_var")
		if !ok {
			t.Error("expected to find variable")
		}
		if result != "my_value" {
			t.Errorf("Get('my_var') = %v, want 'my_value'", result)
		}
	})

	t.Run("Get __template__", func(t *testing.T) {
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       make(map[string]*TemplateMacro),
			Variables:    make(map[string]interface{}),
			Context:      &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		result, ok := in.Get("__template__")
		if !ok {
			t.Error("expected to find __template__")
		}
		if result != "test.html" {
			t.Errorf("Get('__template__') = %v, want 'test.html'", result)
		}
	})

	t.Run("Get __imported__", func(t *testing.T) {
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       make(map[string]*TemplateMacro),
			Variables:    make(map[string]interface{}),
			Context:      &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		result, ok := in.Get("__imported__")
		if !ok {
			t.Error("expected to find __imported__")
		}
		if result != true {
			t.Errorf("Get('__imported__') = %v, want true", result)
		}
	})

	t.Run("Get not found", func(t *testing.T) {
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       make(map[string]*TemplateMacro),
			Variables:    make(map[string]interface{}),
			Context:      &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		_, ok := in.Get("nonexistent")
		if ok {
			t.Error("expected not to find nonexistent key")
		}
	})

	t.Run("Set", func(t *testing.T) {
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       make(map[string]*TemplateMacro),
			Variables:    make(map[string]interface{}),
			Context:      &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		in.Set("new_var", "new_value")

		result, ok := in.Get("new_var")
		if !ok {
			t.Error("expected to find set variable")
		}
		if result != "new_value" {
			t.Errorf("Get after Set = %v, want 'new_value'", result)
		}
	})

	t.Run("String", func(t *testing.T) {
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros: map[string]*TemplateMacro{
				"macro1": {},
				"macro2": {},
			},
			Variables: map[string]interface{}{
				"var1": 1,
			},
			Context: &simpleContext{variables: make(map[string]interface{})},
		}

		in := &ImportedNamespace{
			namespace: ns,
			evaluator: NewEvaluator(),
		}

		str := in.String()
		if !strings.Contains(str, "test.html") {
			t.Error("String should contain template name")
		}
		if !strings.Contains(str, "2 macros") {
			t.Error("String should contain macro count")
		}
		if !strings.Contains(str, "1 variables") {
			t.Error("String should contain variable count")
		}
	})
}

// TestTemplateMacroCallCoverage tests the TemplateMacro type
func TestTemplateMacroCallCoverage(t *testing.T) {
	t.Run("Call with args", func(t *testing.T) {
		// Create a simple text node for the macro body
		body := []parser.Node{
			&parser.TextNode{Content: "Hello, "},
			&parser.VariableNode{
				Expression: &parser.IdentifierNode{Name: "name"},
			},
		}

		macro := &TemplateMacro{
			Name:       "greet",
			Parameters: []string{"name"},
			Defaults:   make(map[string]interface{}),
			Body:       body,
			Context:    &simpleContext{variables: make(map[string]interface{})},
		}

		e := NewEvaluator()
		ctx := &simpleContext{variables: make(map[string]interface{})}
		ctx.SetVariable("name", "World") // Set a default in case

		result, err := macro.Call(e, ctx, []interface{}{"Alice"}, nil)
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		// The result should contain something (macro execution)
		_ = result
	})

	t.Run("Call with defaults", func(t *testing.T) {
		body := []parser.Node{
			&parser.TextNode{Content: "Value: "},
		}

		macro := &TemplateMacro{
			Name:       "show",
			Parameters: []string{"value"},
			Defaults: map[string]interface{}{
				"value": "default",
			},
			Body:    body,
			Context: &simpleContext{variables: make(map[string]interface{})},
		}

		e := NewEvaluator()
		ctx := &simpleContext{variables: make(map[string]interface{})}

		// Call without arguments - should use default
		result, err := macro.Call(e, ctx, []interface{}{}, nil)
		if err != nil {
			t.Fatalf("Call with defaults failed: %v", err)
		}
		_ = result
	})

	t.Run("Call with kwargs", func(t *testing.T) {
		body := []parser.Node{
			&parser.TextNode{Content: "Test"},
		}

		macro := &TemplateMacro{
			Name:       "test",
			Parameters: []string{"a", "b"},
			Defaults: map[string]interface{}{
				"a": "default_a",
				"b": "default_b",
			},
			Body:    body,
			Context: &simpleContext{variables: make(map[string]interface{})},
		}

		e := NewEvaluator()
		ctx := &simpleContext{variables: make(map[string]interface{})}

		kwargs := map[string]interface{}{
			"a": 1,
			"b": 2,
		}

		result, err := macro.Call(e, ctx, nil, kwargs)
		if err != nil {
			t.Fatalf("Call with kwargs failed: %v", err)
		}
		_ = result
	})
}

// TestImportSystem tests the ImportSystem type
func TestImportSystem(t *testing.T) {
	t.Run("GetImportedNamespace", func(t *testing.T) {
		e := NewEvaluator()
		is := NewImportSystem(nil, e)
		ns := &TemplateNamespace{
			TemplateName: "test.html",
			Macros:       make(map[string]*TemplateMacro),
			Variables:    make(map[string]interface{}),
		}

		result := is.GetImportedNamespace(ns)
		if result == nil {
			t.Error("expected non-nil result")
		}
	})
}
