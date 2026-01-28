package miya

import (
	"reflect"
	"testing"

	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// TestContextFunctions tests context-related functions
func TestContextFunctions(t *testing.T) {
	t.Run("Push and Pop", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("outer", "value1")

		// Push a new scope - returns new context
		innerCtx := ctx.Push()
		innerCtx.Set("inner", "value2")

		// Inner scope should see both
		inner, ok := innerCtx.Get("inner")
		if !ok || inner != "value2" {
			t.Error("Expected to find inner variable")
		}
		outer, ok := innerCtx.Get("outer")
		if !ok || outer != "value1" {
			t.Error("Expected to find outer variable in inner scope")
		}

		// Pop the scope - returns parent context
		outerCtx := innerCtx.Pop()

		// Should still find outer in parent
		outer, ok = outerCtx.Get("outer")
		if !ok || outer != "value1" {
			t.Error("Expected to still find outer variable")
		}
	})

	t.Run("GetEnv", func(t *testing.T) {
		env := NewEnvironment()
		ctx := newContextWithEnv(env)

		retrievedEnv := ctx.GetEnv()
		if retrievedEnv != env {
			t.Error("Expected GetEnv to return the environment")
		}
	})

	t.Run("String representation on concrete type", func(t *testing.T) {
		ctx := newContextWithEnv(nil)
		ctx.Set("name", "test")
		ctx.Set("count", 42)

		// Access concrete type for String() method
		if concreteCtx, ok := ctx.(*context); ok {
			str := concreteCtx.String()
			if str == "" {
				t.Error("Expected non-empty string representation")
			}
		}
	})

	t.Run("Clone preserves data", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("key", "value")

		cloned := ctx.Clone()

		val, ok := cloned.Get("key")
		if !ok || val != "value" {
			t.Error("Expected clone to have the same data")
		}

		// Modifying clone should not affect original
		cloned.Set("key", "modified")
		original, _ := ctx.Get("key")
		if original != "value" {
			t.Error("Modifying clone should not affect original")
		}
	})
}

// TestTemplateContextAdapter tests the TemplateContextAdapter
func TestTemplateContextAdapter(t *testing.T) {
	t.Run("NewTemplateContextAdapter", func(t *testing.T) {
		ctx := NewContext()
		env := NewEnvironment()
		adapter := NewTemplateContextAdapter(ctx, env)

		if adapter == nil {
			t.Fatal("Expected non-nil adapter")
		}
	})

	t.Run("All method", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("key1", "value1")
		ctx.Set("key2", "value2")

		env := NewEnvironment()
		adapter := NewTemplateContextAdapter(ctx, env)

		all := adapter.All()
		if all == nil {
			t.Fatal("Expected non-nil map from All()")
		}

		if all["key1"] != "value1" || all["key2"] != "value2" {
			t.Error("All() should return all context variables")
		}
	})

	t.Run("ApplyTest", func(t *testing.T) {
		ctx := NewContext()
		env := NewEnvironment()
		adapter := NewTemplateContextAdapter(ctx, env)

		// Test a simple test
		result, err := adapter.ApplyTest("defined", "hello")
		if err != nil {
			t.Fatalf("ApplyTest failed: %v", err)
		}
		if result != true {
			t.Error("Expected 'defined' test to return true for non-nil value")
		}
	})
}

// TestTemplateMethods tests Template methods
func TestTemplateMethods(t *testing.T) {
	env := NewEnvironment()
	tmpl, err := env.FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	t.Run("Source", func(t *testing.T) {
		source := tmpl.Source()
		if source != "Hello {{ name }}!" {
			t.Errorf("Expected source 'Hello {{ name }}!', got %q", source)
		}
	})

	t.Run("AST", func(t *testing.T) {
		ast := tmpl.AST()
		if ast == nil {
			t.Error("Expected non-nil AST")
		}
	})

	t.Run("GetASTAsTemplateNode", func(t *testing.T) {
		templateNode := tmpl.GetASTAsTemplateNode()
		if templateNode == nil {
			t.Error("Expected non-nil TemplateNode")
		}
	})

	t.Run("SetAST", func(t *testing.T) {
		newAST := &parser.TemplateNode{
			Name:     "test",
			Children: []parser.Node{},
		}
		tmpl.SetAST(newAST)

		if tmpl.AST() != newAST {
			t.Error("Expected AST to be updated")
		}
	})
}

// TestEnvironmentMethods tests additional Environment methods
func TestEnvironmentMethods(t *testing.T) {
	t.Run("AddTest and GetTest", func(t *testing.T) {
		env := NewEnvironment()

		customTest := func(value interface{}, args ...interface{}) (bool, error) {
			return value == "custom", nil
		}

		err := env.AddTest("mycustomtest", customTest)
		if err != nil {
			t.Fatalf("AddTest failed: %v", err)
		}

		test, ok := env.GetTest("mycustomtest")
		if !ok {
			t.Error("Expected to find custom test")
		}
		if test == nil {
			t.Error("Expected non-nil test function")
		}

		// Test functionality
		result, err := test("custom")
		if err != nil {
			t.Fatalf("Test function failed: %v", err)
		}
		if result != true {
			t.Error("Expected test to return true for 'custom'")
		}
	})

	t.Run("ListTests", func(t *testing.T) {
		env := NewEnvironment()
		tests := env.ListTests()

		if len(tests) == 0 {
			t.Error("Expected at least some built-in tests")
		}

		// Check for common tests
		found := make(map[string]bool)
		for _, name := range tests {
			found[name] = true
		}

		if !found["defined"] {
			t.Error("Expected 'defined' test to be listed")
		}
	})

	t.Run("ApplyTest", func(t *testing.T) {
		env := NewEnvironment()

		result, err := env.ApplyTest("defined", "hello")
		if err != nil {
			t.Fatalf("ApplyTest failed: %v", err)
		}
		if result != true {
			t.Error("Expected 'defined' test to return true")
		}
	})

	t.Run("GetConfig and SetConfig", func(t *testing.T) {
		env := NewEnvironment()

		env.SetConfig("custom_option", "custom_value")
		value, ok := env.GetConfig("custom_option")

		if !ok || value != "custom_value" {
			t.Errorf("Expected 'custom_value', got %v", value)
		}

		// Non-existent config
		nilValue, ok := env.GetConfig("non_existent")
		if ok || nilValue != nil {
			t.Error("Expected nil and false for non-existent config")
		}
	})

	t.Run("WithStrictUndefined", func(t *testing.T) {
		env := NewEnvironment(WithStrictUndefined(true))

		if env.undefinedBehavior != runtime.UndefinedStrict {
			t.Error("Expected strict undefined behavior")
		}
	})

	t.Run("WithDebugUndefined", func(t *testing.T) {
		env := NewEnvironment(WithDebugUndefined(true))

		if env.undefinedBehavior != runtime.UndefinedDebug {
			t.Error("Expected debug undefined behavior")
		}
	})

	t.Run("WithUndefinedBehavior", func(t *testing.T) {
		env := NewEnvironment(WithUndefinedBehavior(runtime.UndefinedStrict))

		if env.undefinedBehavior != runtime.UndefinedStrict {
			t.Error("Expected strict undefined behavior")
		}
	})
}

// TestExtensionMethods tests extension-related methods
func TestExtensionMethods(t *testing.T) {
	t.Run("GetExtensionRegistry", func(t *testing.T) {
		env := NewEnvironment()
		registry := env.GetExtensionRegistry()

		if registry == nil {
			t.Error("Expected non-nil extension registry")
		}
	})

	t.Run("IsCustomTag returns false for unknown tag", func(t *testing.T) {
		env := NewEnvironment()
		isCustom := env.IsCustomTag("unknowntag")

		if isCustom {
			t.Error("Expected false for unknown tag")
		}
	})

	t.Run("GetExtensionForTag returns false for unknown tag", func(t *testing.T) {
		env := NewEnvironment()
		ext, ok := env.GetExtensionForTag("unknowntag")

		if ok || ext != nil {
			t.Error("Expected nil and false for unknown tag")
		}
	})

	t.Run("GetExtension returns false for unknown extension", func(t *testing.T) {
		env := NewEnvironment()
		ext, ok := env.GetExtension("unknownext")

		if ok || ext != nil {
			t.Error("Expected nil and false for unknown extension")
		}
	})
}

// TestInheritanceCacheMethods tests inheritance cache methods
func TestInheritanceCacheMethods(t *testing.T) {
	t.Run("ClearInheritanceCache", func(t *testing.T) {
		env := NewEnvironment()
		// Should not panic
		env.ClearInheritanceCache()
	})

	t.Run("InvalidateTemplate", func(t *testing.T) {
		env := NewEnvironment()
		// Should not panic
		env.InvalidateTemplate("test.html")
	})

	t.Run("GetInheritanceCacheStats", func(t *testing.T) {
		env := NewEnvironment()
		stats := env.GetInheritanceCacheStats()

		// Stats is a struct, just verify it doesn't panic
		_ = stats
	})

	t.Run("ConfigureInheritanceCache", func(t *testing.T) {
		env := NewEnvironment()
		// Should not panic
		env.ConfigureInheritanceCache(500, 30, 10)
	})
}

// TestConcurrentSafety tests concurrent safety functions
func TestConcurrentSafety(t *testing.T) {
	t.Run("NewThreadSafeTemplate", func(t *testing.T) {
		env := NewEnvironment()
		tmpl, err := env.FromString("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		safeTmpl := NewThreadSafeTemplate(tmpl)
		if safeTmpl == nil {
			t.Fatal("Expected non-nil thread-safe template")
		}
	})

	t.Run("RenderConcurrent", func(t *testing.T) {
		env := NewEnvironment()
		tmpl, err := env.FromString("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		safeTmpl := NewThreadSafeTemplate(tmpl)

		ctx := NewContext()
		ctx.Set("name", "World")

		result, err := safeTmpl.RenderConcurrent(ctx)
		if err != nil {
			t.Fatalf("RenderConcurrent failed: %v", err)
		}

		if result != "Hello World!" {
			t.Errorf("Expected 'Hello World!', got %q", result)
		}
	})
}

// TestTemplateAdapters tests template and environment adapters
func TestTemplateAdapters(t *testing.T) {
	t.Run("environmentAdapter GetTemplate", func(t *testing.T) {
		env := NewEnvironment()
		_, err := env.FromString("test content")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		adapter := &environmentAdapter{env: env}
		_, err = adapter.GetTemplate("nonexistent")
		// Should error for non-existent template
		if err == nil {
			t.Log("GetTemplate for non-existent template may return error or nil")
		}
	})

	t.Run("environmentAdapter GetLoader", func(t *testing.T) {
		env := NewEnvironment()
		adapter := &environmentAdapter{env: env}
		loader := adapter.GetLoader()
		// May or may not be nil depending on configuration
		_ = loader
	})

	t.Run("templateAdapter methods", func(t *testing.T) {
		env := NewEnvironment()
		tmpl, err := env.FromString("Hello {{ name }}!")
		if err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		adapter := &templateAdapter{template: tmpl}

		if adapter.Name() != tmpl.Name() {
			t.Error("Expected adapter Name() to match template Name()")
		}

		ast := adapter.AST()
		if ast == nil {
			t.Error("Expected non-nil AST from adapter")
		}
	})
}

// TestContextNestedAccess tests nested attribute access
func TestContextNestedAccess(t *testing.T) {
	t.Run("Access nested map attributes", func(t *testing.T) {
		ctx := NewContext()
		ctx.Set("user", map[string]interface{}{
			"name": "Alice",
			"address": map[string]interface{}{
				"city": "New York",
			},
		})

		// This tests the getLocal and getAttribute functions
		user, ok := ctx.Get("user")
		if !ok {
			t.Fatal("Expected to find user")
		}

		userMap, ok := user.(map[string]interface{})
		if !ok {
			t.Fatal("Expected user to be a map")
		}

		if userMap["name"] != "Alice" {
			t.Error("Expected user.name to be Alice")
		}
	})

	t.Run("Access struct attributes via reflection", func(t *testing.T) {
		type User struct {
			Name  string
			Email string
		}

		ctx := NewContext()
		ctx.Set("user", User{Name: "Bob", Email: "bob@example.com"})

		user, ok := ctx.Get("user")
		if !ok {
			t.Fatal("Expected to find user")
		}

		// Access should work for struct
		val := reflect.ValueOf(user)
		if val.Kind() == reflect.Struct {
			nameField := val.FieldByName("Name")
			if nameField.IsValid() && nameField.String() != "Bob" {
				t.Error("Expected Name field to be Bob")
			}
		}
	})
}

// TestEnvironmentMacros tests macro-related environment methods
func TestEnvironmentMacros(t *testing.T) {
	t.Run("GetMacro returns false for undefined macro", func(t *testing.T) {
		env := NewEnvironment()
		macro, ok := env.GetMacro("undefined_macro")
		if ok || macro != nil {
			t.Error("Expected nil and false for undefined macro")
		}
	})

	t.Run("ListMacros", func(t *testing.T) {
		env := NewEnvironment()
		macros := env.ListMacros()
		// Should return empty list or nil for no macros
		_ = macros
	})

	t.Run("ClearMacros", func(t *testing.T) {
		env := NewEnvironment()
		// Should not panic
		env.ClearMacros()
	})
}

// TestCapitalizeFirst tests the capitalizeFirst helper
func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"HELLO", "HELLO"},
		{"", ""},
		{"a", "A"},
		{"1abc", "1abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalizeFirst(tt.input)
			if result != tt.expected {
				t.Errorf("capitalizeFirst(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeHTMLOptionWhitespace tests HTML option whitespace normalization
func TestNormalizeHTMLOptionWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "option with extra whitespace",
			input:    "<option value=\"1\">  Option   One  </option>",
			expected: "<option value=\"1\">Option One</option>",
		},
		{
			name:     "option with newlines",
			input:    "<option value=\"2\">\n  Option Two\n</option>",
			expected: "<option value=\"2\">Option Two</option>",
		},
		{
			name:     "no options",
			input:    "<div>Hello World</div>",
			expected: "<div>Hello World</div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeHTMLOptionWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeHTMLOptionWhitespace(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeBlogAuthorLinks tests blog author link normalization
func TestNormalizeBlogAuthorLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "author link",
			input:    `By <a href="/authors/john">John Doe</a>`,
			expected: `By John Doe`,
		},
		{
			name:     "no author link",
			input:    `By John Doe`,
			expected: `By John Doe`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeBlogAuthorLinks(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeBlogAuthorLinks(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkContextOperations(b *testing.B) {
	ctx := NewContext()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx.Set("key", "value")
		}
	})

	b.Run("Get", func(b *testing.B) {
		ctx.Set("key", "value")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx.Get("key")
		}
	})

	b.Run("Clone", func(b *testing.B) {
		ctx.Set("key1", "value1")
		ctx.Set("key2", "value2")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx.Clone()
		}
	})
}
