package loader

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zipreport/miya/parser"
)

// MockParser implements TemplateParser for testing
type MockParser struct{}

func (m *MockParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	template := &parser.TemplateNode{
		Name:     name,
		Children: []parser.Node{&parser.TextNode{Content: content}},
	}
	return template, nil
}

// Create test files in temporary directory
func createTestTemplates(t *testing.T) string {
	tempDir := t.TempDir()

	// Create directory structure
	templatesDir := filepath.Join(tempDir, "templates")
	subDir := filepath.Join(templatesDir, "sub")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test template files
	templates := map[string]string{
		"base.html":       "<!DOCTYPE html><html>{{ content }}</html>",
		"page.html":       "{% extends 'base.html' %}{% block content %}Page Content{% endblock %}",
		"partial.jinja":   "<div>{{ message }}</div>",
		"sub/nested.html": "<p>Nested template content</p>",
		"plain.txt":       "Plain text file", // Should be ignored by default extensions
	}

	for filename, content := range templates {
		fullPath := filepath.Join(templatesDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return templatesDir
}

func TestFileSystemLoader(t *testing.T) {
	templatesDir := createTestTemplates(t)
	parser := &MockParser{}

	loader := NewFileSystemLoader([]string{templatesDir}, parser)

	t.Run("LoadExistingTemplate", func(t *testing.T) {
		template, err := loader.LoadTemplate("base.html")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		if template.Name != "base.html" {
			t.Errorf("Expected template name 'base.html', got '%s'", template.Name)
		}
	})

	t.Run("LoadTemplateWithoutExtension", func(t *testing.T) {
		template, err := loader.LoadTemplate("base")
		if err != nil {
			t.Fatalf("Failed to load template without extension: %v", err)
		}

		if template.Name != "base" {
			t.Errorf("Expected template name 'base', got '%s'", template.Name)
		}
	})

	t.Run("LoadNestedTemplate", func(t *testing.T) {
		template, err := loader.LoadTemplate("sub/nested.html")
		if err != nil {
			t.Fatalf("Failed to load nested template: %v", err)
		}

		if template.Name != "sub/nested.html" {
			t.Errorf("Expected template name 'sub/nested.html', got '%s'", template.Name)
		}
	})

	t.Run("LoadNonExistentTemplate", func(t *testing.T) {
		_, err := loader.LoadTemplate("nonexistent.html")
		if err == nil {
			t.Error("Expected error when loading non-existent template")
		}
	})

	t.Run("GetSource", func(t *testing.T) {
		content, err := loader.GetSource("base.html")
		if err != nil {
			t.Fatalf("Failed to get source: %v", err)
		}

		expected := "<!DOCTYPE html><html>{{ content }}</html>"
		if content != expected {
			t.Errorf("Expected content '%s', got '%s'", expected, content)
		}
	})

	t.Run("GetSourceWithMetadata", func(t *testing.T) {
		source, err := loader.GetSourceWithMetadata("base.html")
		if err != nil {
			t.Fatalf("Failed to get source with metadata: %v", err)
		}

		if source.Name != "base.html" {
			t.Errorf("Expected name 'base.html', got '%s'", source.Name)
		}

		if source.Content == "" {
			t.Error("Expected non-empty content")
		}

		if source.ModTime.IsZero() {
			t.Error("Expected non-zero modification time")
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := loader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		expectedTemplates := []string{"base.html", "page.html", "partial.jinja", "sub/nested.html"}

		if len(templates) != len(expectedTemplates) {
			t.Errorf("Expected %d templates, got %d", len(expectedTemplates), len(templates))
		}

		templateSet := make(map[string]bool)
		for _, template := range templates {
			templateSet[template] = true
		}

		for _, expected := range expectedTemplates {
			if !templateSet[expected] {
				t.Errorf("Expected template '%s' not found in list", expected)
			}
		}
	})

	t.Run("Caching", func(t *testing.T) {
		// Clear cache first
		loader.ClearCache()

		// Load template twice
		_, err := loader.LoadTemplate("base.html")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		if !loader.IsCached("base.html") {
			t.Error("Template should be cached after loading")
		}

		// Second load should be from cache
		_, err = loader.LoadTemplate("base.html")
		if err != nil {
			t.Fatalf("Failed to load cached template: %v", err)
		}

		// Check cache stats
		stats := loader.GetCacheStats()
		if stats.Hits < 1 {
			t.Errorf("Expected at least 1 cache hit, got %d", stats.Hits)
		}
		if stats.Size < 1 {
			t.Errorf("Expected cache size at least 1, got %d", stats.Size)
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		// Test normal name
		resolved := loader.ResolveTemplateName("template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got '%s'", resolved)
		}

		// Test path traversal protection
		resolved = loader.ResolveTemplateName("../../../etc/passwd")
		if resolved != "" {
			t.Errorf("Expected empty string for path traversal, got '%s'", resolved)
		}

		// Test leading slash removal
		resolved = loader.ResolveTemplateName("/template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got '%s'", resolved)
		}
	})

	t.Run("CustomExtensions", func(t *testing.T) {
		loader.SetExtensions([]string{".txt"})

		templates, err := loader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Should only find .txt files now
		found := false
		for _, template := range templates {
			if template == "plain.txt" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find plain.txt with custom extensions")
		}

		// Reset extensions
		loader.SetExtensions([]string{".html", ".htm", ".jinja", ".jinja2", ".j2"})
	})
}

//go:embed testdata
var testEmbedFS embed.FS

func TestEmbedLoader(t *testing.T) {
	// Create a temporary embed.FS for testing
	// Since we can't easily create embedded files in tests, we'll test the interface
	parser := &MockParser{}
	loader := NewEmbedLoader(testEmbedFS, "testdata", parser)

	t.Run("NewEmbedLoader", func(t *testing.T) {
		if loader == nil {
			t.Error("Expected non-nil embed loader")
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		resolved := loader.ResolveTemplateName("template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got '%s'", resolved)
		}

		// Test path traversal protection
		resolved = loader.ResolveTemplateName("../../../etc/passwd")
		if resolved != "" {
			t.Errorf("Expected empty string for path traversal, got '%s'", resolved)
		}
	})

	t.Run("SetExtensions", func(t *testing.T) {
		loader.SetExtensions([]string{".custom"})
		// Test that extensions were set (can't easily verify without actual files)
	})

	t.Run("ClearCache", func(t *testing.T) {
		loader.ClearCache()
		stats := loader.GetCacheStats()
		if stats.Size != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", stats.Size)
		}
	})
}

func TestStringLoader(t *testing.T) {
	parser := &MockParser{}
	loader := NewStringLoader(parser)

	t.Run("AddAndLoadTemplate", func(t *testing.T) {
		content := "<h1>{{ title }}</h1>"
		loader.AddTemplate("test.html", content)

		template, err := loader.LoadTemplate("test.html")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		if template.Name != "test.html" {
			t.Errorf("Expected template name 'test.html', got '%s'", template.Name)
		}
	})

	t.Run("GetSource", func(t *testing.T) {
		content := "<p>{{ message }}</p>"
		loader.AddTemplate("message.html", content)

		source, err := loader.GetSource("message.html")
		if err != nil {
			t.Fatalf("Failed to get source: %v", err)
		}

		if source != content {
			t.Errorf("Expected content '%s', got '%s'", content, source)
		}
	})

	t.Run("GetSourceWithMetadata", func(t *testing.T) {
		content := "<div>{{ data }}</div>"
		loader.AddTemplate("data.html", content)

		source, err := loader.GetSourceWithMetadata("data.html")
		if err != nil {
			t.Fatalf("Failed to get source with metadata: %v", err)
		}

		if source.Name != "data.html" {
			t.Errorf("Expected name 'data.html', got '%s'", source.Name)
		}

		if source.Content != content {
			t.Errorf("Expected content '%s', got '%s'", content, source.Content)
		}
	})

	t.Run("IsCached", func(t *testing.T) {
		loader.AddTemplate("cached.html", "<p>Cached</p>")

		if !loader.IsCached("cached.html") {
			t.Error("Template should be considered cached")
		}

		if loader.IsCached("nonexistent.html") {
			t.Error("Non-existent template should not be cached")
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		loader.AddTemplate("template1.html", "<p>1</p>")
		loader.AddTemplate("template2.html", "<p>2</p>")

		templates, err := loader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		if len(templates) < 2 {
			t.Errorf("Expected at least 2 templates, got %d", len(templates))
		}
	})

	t.Run("LoadNonExistentTemplate", func(t *testing.T) {
		_, err := loader.LoadTemplate("nonexistent.html")
		if err == nil {
			t.Error("Expected error when loading non-existent template")
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		resolved := loader.ResolveTemplateName("template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got '%s'", resolved)
		}
	})
}

func TestChainLoader(t *testing.T) {
	parser := &MockParser{}

	// Create string loaders with different templates
	loader1 := NewStringLoader(parser)
	loader1.AddTemplate("template1.html", "<p>From loader 1</p>")
	loader1.AddTemplate("shared.html", "<p>Shared from loader 1</p>")

	loader2 := NewStringLoader(parser)
	loader2.AddTemplate("template2.html", "<p>From loader 2</p>")
	loader2.AddTemplate("shared.html", "<p>Shared from loader 2</p>")

	chainLoader := NewChainLoader(loader1, loader2)

	t.Run("LoadFromFirstLoader", func(t *testing.T) {
		template, err := chainLoader.LoadTemplate("template1.html")
		if err != nil {
			t.Fatalf("Failed to load from first loader: %v", err)
		}

		if template.Name != "template1.html" {
			t.Errorf("Expected template name 'template1.html', got '%s'", template.Name)
		}
	})

	t.Run("LoadFromSecondLoader", func(t *testing.T) {
		template, err := chainLoader.LoadTemplate("template2.html")
		if err != nil {
			t.Fatalf("Failed to load from second loader: %v", err)
		}

		if template.Name != "template2.html" {
			t.Errorf("Expected template name 'template2.html', got '%s'", template.Name)
		}
	})

	t.Run("FirstLoaderTakesPrecedence", func(t *testing.T) {
		template, err := chainLoader.LoadTemplate("shared.html")
		if err != nil {
			t.Fatalf("Failed to load shared template: %v", err)
		}

		// Should get from first loader since it has precedence
		if template.Name != "shared.html" {
			t.Errorf("Expected template name 'shared.html', got '%s'", template.Name)
		}

		// Verify content comes from first loader
		source, err := chainLoader.GetSource("shared.html")
		if err != nil {
			t.Fatalf("Failed to get source: %v", err)
		}

		if !strings.Contains(source, "loader 1") {
			t.Error("Expected content from first loader")
		}
	})

	t.Run("AddLoader", func(t *testing.T) {
		loader3 := NewStringLoader(parser)
		loader3.AddTemplate("template3.html", "<p>From loader 3</p>")

		chainLoader.AddLoader(loader3)

		template, err := chainLoader.LoadTemplate("template3.html")
		if err != nil {
			t.Fatalf("Failed to load from added loader: %v", err)
		}

		if template.Name != "template3.html" {
			t.Errorf("Expected template name 'template3.html', got '%s'", template.Name)
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := chainLoader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Should have templates from all loaders, but no duplicates
		templateSet := make(map[string]bool)
		for _, template := range templates {
			templateSet[template] = true
		}

		expectedTemplates := []string{"template1.html", "template2.html", "template3.html", "shared.html"}
		for _, expected := range expectedTemplates {
			if !templateSet[expected] {
				t.Errorf("Expected template '%s' not found in list", expected)
			}
		}
	})

	t.Run("IsCached", func(t *testing.T) {
		// Should return true if any loader has it cached
		if !chainLoader.IsCached("template1.html") {
			t.Error("Template should be cached in chain loader")
		}

		if chainLoader.IsCached("nonexistent.html") {
			t.Error("Non-existent template should not be cached")
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		resolved := chainLoader.ResolveTemplateName("template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got '%s'", resolved)
		}
	})

	t.Run("LoadNonExistentTemplate", func(t *testing.T) {
		_, err := chainLoader.LoadTemplate("nonexistent.html")
		if err == nil {
			t.Error("Expected error when loading non-existent template from chain")
		}
	})
}

func TestLoaderFuncCompatibility(t *testing.T) {
	// Test that the original LoaderFunc still works
	loaderFunc := LoaderFunc(func(name string) (string, error) {
		if name == "test.html" {
			return "<p>Test content</p>", nil
		}
		return "", fmt.Errorf("template not found: %s", name)
	})

	t.Run("GetSource", func(t *testing.T) {
		content, err := loaderFunc.GetSource("test.html")
		if err != nil {
			t.Fatalf("Failed to get source: %v", err)
		}

		expected := "<p>Test content</p>"
		if content != expected {
			t.Errorf("Expected content '%s', got '%s'", expected, content)
		}
	})

	t.Run("IsCached", func(t *testing.T) {
		if loaderFunc.IsCached("test.html") {
			t.Error("LoaderFunc should not report templates as cached")
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		_, err := loaderFunc.ListTemplates()
		if err == nil {
			t.Error("Expected error from LoaderFunc.ListTemplates()")
		}
	})
}

func TestBaseLoaderCompatibility(t *testing.T) {
	// Test that BaseLoader still works for backward compatibility
	baseLoader := &BaseLoader{}

	t.Run("IsCached", func(t *testing.T) {
		if baseLoader.IsCached("any.html") {
			t.Error("BaseLoader should not report templates as cached")
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		_, err := baseLoader.ListTemplates()
		if err == nil {
			t.Error("Expected error from BaseLoader.ListTemplates()")
		}
	})
}

func TestReadAllFunction(t *testing.T) {
	content := "Test content for ReadAll"
	reader := strings.NewReader(content)

	result, err := ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if result != content {
		t.Errorf("Expected '%s', got '%s'", content, result)
	}
}

func TestCacheExpiration(t *testing.T) {
	templatesDir := createTestTemplates(t)
	parser := &MockParser{}

	loader := NewFileSystemLoader([]string{templatesDir}, parser)

	// Load a template
	_, err := loader.LoadTemplate("base.html")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Verify it's cached
	if !loader.IsCached("base.html") {
		t.Error("Template should be cached")
	}

	// Manually expire the cache entry for testing
	loader.cacheMutex.Lock()
	if cached, ok := loader.cache["base.html"]; ok {
		cached.expires = time.Now().Add(-1 * time.Hour) // Expire in the past
	}
	loader.cacheMutex.Unlock()

	// Should not be considered cached now
	if loader.IsCached("base.html") {
		t.Error("Template should not be cached after expiration")
	}
}

func TestSecurityFeatures(t *testing.T) {
	templatesDir := createTestTemplates(t)
	parser := &MockParser{}

	loader := NewFileSystemLoader([]string{templatesDir}, parser)

	t.Run("PathTraversalPrevention", func(t *testing.T) {
		// Try various path traversal attacks
		maliciousPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"....//....//etc/passwd",
			"/etc/passwd",
			"./../../etc/passwd",
		}

		for _, path := range maliciousPaths {
			_, err := loader.LoadTemplate(path)
			if err == nil {
				t.Errorf("Expected error for malicious path: %s", path)
			}
		}
	})

	t.Run("SymlinkHandling", func(t *testing.T) {
		// Test that symlinks are not followed by default
		loader.SetFollowLinks(false)

		// Create a symlink to a file within the template directory
		symlinkPath := filepath.Join(templatesDir, "symlink.html")
		targetPath := filepath.Join(templatesDir, "base.html")

		// Create symlink - skip test if this fails (might not have permission)
		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			t.Skipf("Cannot create symlink for testing: %v", err)
		}

		_, err := loader.LoadTemplate("symlink.html")
		// Should fail because symlinks aren't followed by default
		if err == nil {
			t.Error("Expected error when accessing symlink with followLinks=false")
		}

		// Clean up
		os.Remove(symlinkPath)
	})
}

func TestAdvancedLoaderInterface(t *testing.T) {
	parser := &MockParser{}
	loader := NewStringLoader(parser)
	loader.AddTemplate("test.html", "<p>Test</p>")

	// Test that StringLoader implements AdvancedLoader
	var advancedLoader AdvancedLoader = loader

	t.Run("LoadTemplate", func(t *testing.T) {
		template, err := advancedLoader.LoadTemplate("test.html")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}
		if template.Name != "test.html" {
			t.Errorf("Expected name 'test.html', got '%s'", template.Name)
		}
	})

	t.Run("GetSourceWithMetadata", func(t *testing.T) {
		source, err := advancedLoader.GetSourceWithMetadata("test.html")
		if err != nil {
			t.Fatalf("Failed to get source with metadata: %v", err)
		}
		if source.Name != "test.html" {
			t.Errorf("Expected name 'test.html', got '%s'", source.Name)
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		resolved := advancedLoader.ResolveTemplateName("test.html")
		if resolved != "test.html" {
			t.Errorf("Expected 'test.html', got '%s'", resolved)
		}
	})
}

func TestCachingLoaderInterface(t *testing.T) {
	templatesDir := createTestTemplates(t)
	parser := &MockParser{}

	loader := NewFileSystemLoader([]string{templatesDir}, parser)

	// Test that FileSystemLoader implements CachingLoader
	var cachingLoader CachingLoader = loader

	t.Run("ClearCache", func(t *testing.T) {
		// Load a template to populate cache
		_, err := cachingLoader.LoadTemplate("base.html")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		// Clear cache
		cachingLoader.ClearCache()

		// Verify cache is cleared
		stats := cachingLoader.GetCacheStats()
		if stats.Size != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", stats.Size)
		}
	})

	t.Run("GetCacheStats", func(t *testing.T) {
		stats := cachingLoader.GetCacheStats()
		if stats.Size < 0 {
			t.Errorf("Expected non-negative cache size, got %d", stats.Size)
		}
		if stats.Hits < 0 {
			t.Errorf("Expected non-negative hit count, got %d", stats.Hits)
		}
		if stats.Misses < 0 {
			t.Errorf("Expected non-negative miss count, got %d", stats.Misses)
		}
	})
}
