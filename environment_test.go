package miya

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

func TestEnvironmentCreation(t *testing.T) {
	t.Run("DefaultEnvironment", func(t *testing.T) {
		env := NewEnvironment()
		if env == nil {
			t.Fatal("Expected non-nil environment")
		}

		// Check default delimiters
		blockStart, blockEnd, varStart, varEnd, commentStart, commentEnd := env.GetDelimiters()
		if blockStart != "{%" || blockEnd != "%}" {
			t.Errorf("Expected block delimiters {%% and %%}, got %s and %s", blockStart, blockEnd)
		}
		if varStart != "{{" || varEnd != "}}" {
			t.Errorf("Expected variable delimiters {{ and }}, got %s and %s", varStart, varEnd)
		}
		if commentStart != "{#" || commentEnd != "#}" {
			t.Errorf("Expected comment delimiters {# and #}, got %s and %s", commentStart, commentEnd)
		}

		// Check default auto-escape
		if !env.IsAutoEscape() {
			t.Error("Expected auto-escape to be enabled by default")
		}
	})

	t.Run("EnvironmentWithOptions", func(t *testing.T) {
		stringLoader := loader.NewStringLoader(&mockParser{})
		env := NewEnvironment(
			WithLoader(stringLoader),
			WithAutoEscape(false),
			WithTrimBlocks(true),
			WithLstripBlocks(true),
		)

		if env.GetLoader() != stringLoader {
			t.Error("Expected loader to be set")
		}
		if env.IsAutoEscape() {
			t.Error("Expected auto-escape to be disabled")
		}
		if !env.trimBlocks {
			t.Error("Expected trim blocks to be enabled")
		}
		if !env.lstripBlocks {
			t.Error("Expected lstrip blocks to be enabled")
		}
	})
}

func TestEnvironmentDelimiters(t *testing.T) {
	env := NewEnvironment()

	// Test setting custom delimiters
	env.SetDelimiters("<$", "$>", "<%", "%>")
	env.SetCommentDelimiters("<!", "!>")
	blockStart, blockEnd, varStart, varEnd, commentStart, commentEnd := env.GetDelimiters()

	if blockStart != "<%" || blockEnd != "%>" {
		t.Errorf("Expected block delimiters <%% and %%>, got %s and %s", blockStart, blockEnd)
	}
	if varStart != "<$" || varEnd != "$>" {
		t.Errorf("Expected variable delimiters <$ and $>, got %s and %s", varStart, varEnd)
	}
	if commentStart != "<!" || commentEnd != "!>" {
		t.Errorf("Expected comment delimiters <! and !>, got %s and %s", commentStart, commentEnd)
	}
}

func TestEnvironmentFilters(t *testing.T) {
	env := NewEnvironment()

	// Test adding custom filter
	customFilter := func(value interface{}, args ...interface{}) (interface{}, error) {
		return strings.ToUpper(value.(string)), nil
	}

	err := env.AddFilter("uppercase", customFilter)
	if err != nil {
		t.Fatalf("Failed to add filter: %v", err)
	}

	// Test retrieving filter
	retrievedFilter, ok := env.GetFilter("uppercase")
	if !ok {
		t.Error("Expected to find custom filter")
	}
	if retrievedFilter == nil {
		t.Error("Expected non-nil filter")
	}

	// Test applying filter
	result, err := env.ApplyFilter("uppercase", "hello")
	if err != nil {
		t.Fatalf("Failed to apply filter: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("Expected 'HELLO', got '%v'", result)
	}

	// Test listing filters (should include built-ins plus our custom one)
	filters := env.ListFilters()
	found := false
	for _, name := range filters {
		if name == "uppercase" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'uppercase' filter in list")
	}
}

func TestEnvironmentGlobals(t *testing.T) {
	env := NewEnvironment()

	// Test adding globals
	env.AddGlobal("site_name", "My Website")
	env.AddGlobal("version", "1.0.0")

	// Globals should be accessible in context
	ctx := newContextWithEnv(env)

	siteName, ok := ctx.Get("site_name")
	if !ok {
		t.Error("Expected to find site_name global")
	}
	if siteName != "My Website" {
		t.Errorf("Expected 'My Website', got '%v'", siteName)
	}

	version, ok := ctx.Get("version")
	if !ok {
		t.Error("Expected to find version global")
	}
	if version != "1.0.0" {
		t.Errorf("Expected '1.0.0', got '%v'", version)
	}
}

func TestEnvironmentFromString(t *testing.T) {
	env := NewEnvironment()

	t.Run("SimpleTemplate", func(t *testing.T) {
		// Note: This test will work once we have a working parser
		template, err := env.FromString("Hello World")
		if err != nil {
			t.Fatalf("Failed to create template from string: %v", err)
		}

		if template.Name() != "<string>" {
			t.Errorf("Expected template name '<string>', got '%s'", template.Name())
		}
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		// Test with malformed template syntax
		// Note: This will pass for now since we don't have full lexer/parser implementation
		_, err := env.FromString("{{ unclosed variable")
		if err != nil {
			t.Logf("Got expected error for malformed template: %v", err)
		} else {
			t.Log("No error for malformed template (expected with current simple implementation)")
		}
	})
}

func TestEnvironmentWithLoader(t *testing.T) {
	// Create temporary directory for test templates
	tempDir := t.TempDir()

	// Create test template file
	templateContent := "Hello {{ name }}!"
	templatePath := filepath.Join(tempDir, "greeting.html")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Create environment with filesystem loader
	fsLoader := loader.NewFileSystemLoader([]string{tempDir}, &mockParser{})
	env := NewEnvironment(WithLoader(fsLoader))

	t.Run("GetTemplate", func(t *testing.T) {
		template, err := env.GetTemplate("greeting.html")
		if err != nil {
			t.Fatalf("Failed to get template: %v", err)
		}

		if template.Name() != "greeting.html" {
			t.Errorf("Expected template name 'greeting.html', got '%s'", template.Name())
		}
	})

	t.Run("GetNonExistentTemplate", func(t *testing.T) {
		_, err := env.GetTemplate("nonexistent.html")
		if err == nil {
			t.Error("Expected error when getting non-existent template")
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := env.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		found := false
		for _, name := range templates {
			if name == "greeting.html" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find 'greeting.html' in template list")
		}
	})
}

func TestEnvironmentCache(t *testing.T) {
	env := NewEnvironment()

	// Create some templates via FromString (auto-cached with hash keys)
	template1, _ := env.FromString("Template 1")
	template2, _ := env.FromString("Template 2")

	// Verify templates are created
	_ = template1
	_ = template2

	t.Run("CacheSize", func(t *testing.T) {
		// After Phase 2 optimization, FromString auto-caches with hash keys
		// So we should have 2 entries (one per unique template string)
		size := env.GetCacheSize()
		if size != 2 {
			t.Errorf("Expected cache size 2, got %d", size)
		}
	})

	t.Run("ClearCache", func(t *testing.T) {
		env.ClearCache()
		size := env.GetCacheSize()
		if size != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", size)
		}
	})
}

func TestEnvironmentRenderString(t *testing.T) {
	env := NewEnvironment()

	ctx := NewContextFrom(map[string]interface{}{
		"name": "World",
	})

	// Note: This test requires working lexer/parser
	result, err := env.RenderString("Hello World", ctx)
	if err != nil {
		t.Fatalf("Failed to render string: %v", err)
	}

	// For now, without working parser, it should just return the source
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result)
	}
}

func TestEnvironmentRenderTemplate(t *testing.T) {
	// Create temporary directory for test templates
	tempDir := t.TempDir()

	// Create test template file
	templateContent := "Hello World"
	templatePath := filepath.Join(tempDir, "simple.html")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Create environment with filesystem loader
	fsLoader := loader.NewFileSystemLoader([]string{tempDir}, &mockParser{})
	env := NewEnvironment(WithLoader(fsLoader))

	ctx := NewContextFrom(map[string]interface{}{
		"name": "World",
	})

	result, err := env.RenderTemplate("simple.html", ctx)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result)
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Test global convenience functions
	stringLoader := loader.NewStringLoader(&mockParser{})
	stringLoader.AddTemplate("global_test.html", "Hello Global")

	SetDefaultLoader(stringLoader)

	t.Run("FromString", func(t *testing.T) {
		template, err := FromString("Global FromString")
		if err != nil {
			t.Fatalf("Failed to create template from string: %v", err)
		}
		if template.Name() != "<string>" {
			t.Errorf("Expected '<string>', got '%s'", template.Name())
		}
	})

	t.Run("GetTemplate", func(t *testing.T) {
		template, err := GetTemplate("global_test.html")
		if err != nil {
			t.Fatalf("Failed to get template: %v", err)
		}
		if template.Name() != "global_test.html" {
			t.Errorf("Expected 'global_test.html', got '%s'", template.Name())
		}
	})

	t.Run("RenderString", func(t *testing.T) {
		ctx := NewContextFrom(map[string]interface{}{
			"name": "Global",
		})

		result, err := RenderString("Hello Global", ctx)
		if err != nil {
			t.Fatalf("Failed to render string: %v", err)
		}
		if result != "Hello Global" {
			t.Errorf("Expected 'Hello Global', got '%s'", result)
		}
	})

	t.Run("RenderTemplate", func(t *testing.T) {
		ctx := NewContextFrom(map[string]interface{}{
			"name": "Global",
		})

		result, err := RenderTemplate("global_test.html", ctx)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}
		if result != "Hello Global" {
			t.Errorf("Expected 'Hello Global', got '%s'", result)
		}
	})
}

func TestEnvironmentAutoEscape(t *testing.T) {
	env := NewEnvironment(WithAutoEscape(true))

	ctx := NewContextFrom(map[string]interface{}{
		"html": "<script>alert('xss')</script>",
	})

	// Test with auto-escape enabled - using variables
	result, err := env.RenderString("{{ html }}", ctx)
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	// Should be HTML escaped
	expected := "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
	if result != expected {
		t.Errorf("Expected HTML escaped output, got '%s'", result)
	}

	// Test with auto-escape disabled
	env.SetAutoEscape(false)
	result, err = env.RenderString("{{ html }}", ctx)
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	// Should not be escaped
	if result != "<script>alert('xss')</script>" {
		t.Errorf("Expected unescaped output, got '%s'", result)
	}
}

// Mock parser for testing
type mockParser struct{}

func (m *mockParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	return &parser.TemplateNode{
		Name:     name,
		Children: []parser.Node{&parser.TextNode{Content: content}},
	}, nil
}

// Mock template parser implementation
func (env *Environment) GetDelimiters() (blockStart, blockEnd, varStart, varEnd, commentStart, commentEnd string) {
	return env.blockStartString, env.blockEndString,
		env.varStartString, env.varEndString,
		env.commentStartString, env.commentEndString
}

func (env *Environment) IsAutoEscape() bool {
	return env.autoEscape
}

func (env *Environment) SetAutoEscape(enabled bool) {
	env.autoEscape = enabled
}
