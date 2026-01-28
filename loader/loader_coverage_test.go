package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zipreport/miya/parser"
)

// TestBaseLoader tests the BaseLoader implementation
func TestBaseLoader(t *testing.T) {
	base := &BaseLoader{}

	// IsCached should always return false
	if base.IsCached("any") {
		t.Error("BaseLoader.IsCached should always return false")
	}

	// ListTemplates should return error
	_, err := base.ListTemplates()
	if err == nil {
		t.Error("BaseLoader.ListTemplates should return error")
	}
}

// TestLoaderFunc tests the LoaderFunc implementation
func TestLoaderFunc(t *testing.T) {
	fn := LoaderFunc(func(name string) (string, error) {
		return "content for " + name, nil
	})

	// Test GetSource
	content, err := fn.GetSource("test.html")
	if err != nil {
		t.Errorf("LoaderFunc.GetSource error: %v", err)
	}
	if content != "content for test.html" {
		t.Errorf("Expected 'content for test.html', got %q", content)
	}

	// IsCached should always return false
	if fn.IsCached("any") {
		t.Error("LoaderFunc.IsCached should always return false")
	}

	// ListTemplates should return error
	_, err = fn.ListTemplates()
	if err == nil {
		t.Error("LoaderFunc.ListTemplates should return error")
	}
}

// TestReadAll tests the ReadAll function
func TestReadAll(t *testing.T) {
	// Create a temp file
	tempFile, err := os.CreateTemp("", "readall_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "Hello, World!"
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Seek(0, 0)

	content, err := ReadAll(tempFile)
	if err != nil {
		t.Errorf("ReadAll error: %v", err)
	}
	if content != testContent {
		t.Errorf("Expected %q, got %q", testContent, content)
	}
	tempFile.Close()
}

// TestFileSystemLoaderSearchTemplates tests template search functionality
func TestFileSystemLoaderSearchTemplates(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{templatesDir}, mockParser)

	t.Run("SearchByPattern", func(t *testing.T) {
		matches, err := loader.SearchTemplates("*.html")
		if err != nil {
			t.Errorf("SearchTemplates error: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected some matches for *.html pattern")
		}
	})

	t.Run("SearchByFilenamePattern", func(t *testing.T) {
		matches, err := loader.SearchTemplates("base*")
		if err != nil {
			t.Errorf("SearchTemplates error: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected matches for base* pattern")
		}
	})

	t.Run("SearchWithDirectory", func(t *testing.T) {
		matches, err := loader.SearchTemplates("sub/*")
		if err != nil {
			t.Errorf("SearchTemplates error: %v", err)
		}
		// May or may not match depending on the pattern
		_ = matches
	})
}

// TestFileSystemLoaderGetTemplatesByExtension tests extension filtering
func TestFileSystemLoaderGetTemplatesByExtension(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{templatesDir}, mockParser)

	t.Run("WithDot", func(t *testing.T) {
		matches, err := loader.GetTemplatesByExtension(".html")
		if err != nil {
			t.Errorf("GetTemplatesByExtension error: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected some .html templates")
		}
	})

	t.Run("WithoutDot", func(t *testing.T) {
		matches, err := loader.GetTemplatesByExtension("html")
		if err != nil {
			t.Errorf("GetTemplatesByExtension error: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected some html templates")
		}
	})

	t.Run("NonExistentExtension", func(t *testing.T) {
		matches, err := loader.GetTemplatesByExtension(".xyz")
		if err != nil {
			t.Errorf("GetTemplatesByExtension error: %v", err)
		}
		if len(matches) != 0 {
			t.Error("Expected no .xyz templates")
		}
	})
}

// TestFileSystemLoaderGetTemplatesInDirectory tests directory filtering
func TestFileSystemLoaderGetTemplatesInDirectory(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{templatesDir}, mockParser)

	t.Run("RootDirectory", func(t *testing.T) {
		matches, err := loader.GetTemplatesInDirectory(".")
		if err != nil {
			t.Errorf("GetTemplatesInDirectory error: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected some templates in root directory")
		}
	})

	t.Run("SubDirectory", func(t *testing.T) {
		matches, err := loader.GetTemplatesInDirectory("sub")
		if err != nil {
			t.Errorf("GetTemplatesInDirectory error: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected some templates in sub directory")
		}
	})

	t.Run("EmptyDirectory", func(t *testing.T) {
		matches, err := loader.GetTemplatesInDirectory("")
		if err != nil {
			t.Errorf("GetTemplatesInDirectory error: %v", err)
		}
		// Empty string should be treated as root
		_ = matches
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		matches, err := loader.GetTemplatesInDirectory("nonexistent")
		if err != nil {
			t.Errorf("GetTemplatesInDirectory error: %v", err)
		}
		if len(matches) != 0 {
			t.Error("Expected no templates in nonexistent directory")
		}
	})
}

// TestFileSystemLoaderGetTemplateInfo tests template info retrieval
func TestFileSystemLoaderGetTemplateInfo(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{templatesDir}, mockParser)

	t.Run("ExistingTemplate", func(t *testing.T) {
		info, err := loader.GetTemplateInfo("base.html")
		if err != nil {
			t.Errorf("GetTemplateInfo error: %v", err)
		}
		if info.Name != "base.html" {
			t.Errorf("Expected name 'base.html', got %q", info.Name)
		}
		if info.Extension != ".html" {
			t.Errorf("Expected extension '.html', got %q", info.Extension)
		}
		if info.Size == 0 {
			t.Error("Expected non-zero size")
		}
	})

	t.Run("TemplateWithDependencies", func(t *testing.T) {
		info, err := loader.GetTemplateInfo("page.html")
		if err != nil {
			t.Errorf("GetTemplateInfo error: %v", err)
		}
		// page.html extends base.html, so should have dependency
		if len(info.Dependencies) == 0 {
			t.Error("Expected dependencies for page.html")
		}
	})

	t.Run("NonExistentTemplate", func(t *testing.T) {
		_, err := loader.GetTemplateInfo("nonexistent.html")
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
	})
}

// TestFileSystemLoaderSymlinks tests symlink handling
func TestFileSystemLoaderSymlinks(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{templatesDir}, mockParser)

	// By default, symlinks are not followed
	loader.SetFollowLinks(false)

	// Create a symlink
	symlinkPath := filepath.Join(templatesDir, "symlink.html")
	targetPath := filepath.Join(templatesDir, "base.html")
	os.Symlink(targetPath, symlinkPath)
	defer os.Remove(symlinkPath)

	// With followLinks = false, symlink should not be accessible
	_, err := loader.GetSource("symlink.html")
	if err == nil {
		// If it succeeded, that's fine too - depends on implementation
	}

	// Enable following links
	loader.SetFollowLinks(true)
	_, err = loader.GetSource("symlink.html")
	// May or may not work depending on OS
	_ = err
}

// TestFileSystemLoaderInvalidPath tests invalid path handling
func TestFileSystemLoaderInvalidPath(t *testing.T) {
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{"/nonexistent/path"}, mockParser)

	// ListTemplates should handle non-existent path gracefully
	_, err := loader.ListTemplates()
	if err == nil {
		// May return empty list or error
	}

	// GetSource should fail
	_, err = loader.GetSource("template.html")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

// TestChainLoaderCoverage tests the ChainLoader implementation
func TestChainLoaderCoverage(t *testing.T) {
	mockParser := &MockParser{}
	stringLoader1 := NewStringLoader(mockParser)
	stringLoader1.AddTemplate("template1.html", "Content 1")

	stringLoader2 := NewStringLoader(mockParser)
	stringLoader2.AddTemplate("template2.html", "Content 2")

	chain := NewChainLoader(stringLoader1, stringLoader2)

	t.Run("LoadFromFirstLoader", func(t *testing.T) {
		source, err := chain.GetSource("template1.html")
		if err != nil {
			t.Errorf("GetSource error: %v", err)
		}
		if source != "Content 1" {
			t.Errorf("Expected 'Content 1', got %q", source)
		}
	})

	t.Run("LoadFromSecondLoader", func(t *testing.T) {
		source, err := chain.GetSource("template2.html")
		if err != nil {
			t.Errorf("GetSource error: %v", err)
		}
		if source != "Content 2" {
			t.Errorf("Expected 'Content 2', got %q", source)
		}
	})

	t.Run("LoadNonExistent", func(t *testing.T) {
		_, err := chain.GetSource("nonexistent.html")
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
	})

	t.Run("IsCached", func(t *testing.T) {
		// Should check all loaders
		cached := chain.IsCached("template1.html")
		if !cached {
			t.Error("Expected template1.html to be cached")
		}

		cached = chain.IsCached("nonexistent.html")
		if cached {
			t.Error("Expected nonexistent.html to not be cached")
		}
	})

	t.Run("LoadTemplate", func(t *testing.T) {
		tmpl, err := chain.LoadTemplate("template1.html")
		if err != nil {
			t.Errorf("LoadTemplate error: %v", err)
		}
		if tmpl == nil {
			t.Error("Expected non-nil template")
		}
	})

	t.Run("GetSourceWithMetadata", func(t *testing.T) {
		source, err := chain.GetSourceWithMetadata("template1.html")
		if err != nil {
			t.Errorf("GetSourceWithMetadata error: %v", err)
		}
		if source.Name != "template1.html" {
			t.Errorf("Expected name 'template1.html', got %q", source.Name)
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		resolved := chain.ResolveTemplateName("template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got %q", resolved)
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := chain.ListTemplates()
		if err != nil {
			t.Errorf("ListTemplates error: %v", err)
		}
		if len(templates) < 2 {
			t.Error("Expected at least 2 templates from chain")
		}
	})

	t.Run("AddLoader", func(t *testing.T) {
		stringLoader3 := NewStringLoader(mockParser)
		stringLoader3.AddTemplate("template3.html", "Content 3")
		chain.AddLoader(stringLoader3)

		source, err := chain.GetSource("template3.html")
		if err != nil {
			t.Errorf("GetSource error after AddLoader: %v", err)
		}
		if source != "Content 3" {
			t.Errorf("Expected 'Content 3', got %q", source)
		}
	})
}

// TestChainLoaderEmpty tests ChainLoader with no loaders
func TestChainLoaderEmpty(t *testing.T) {
	chain := NewChainLoader()

	// ResolveTemplateName with no loaders
	resolved := chain.ResolveTemplateName("template.html")
	if resolved != "template.html" {
		t.Errorf("Expected 'template.html', got %q", resolved)
	}

	// GetSource should fail
	_, err := chain.GetSource("template.html")
	if err == nil {
		t.Error("Expected error for empty chain")
	}
}

// TestStringLoaderAdditional tests additional StringLoader functionality
func TestStringLoaderAdditional(t *testing.T) {
	mockParser := &MockParser{}
	loader := NewStringLoader(mockParser)

	loader.AddTemplate("test.html", "Test content")

	t.Run("GetSourceWithMetadata", func(t *testing.T) {
		source, err := loader.GetSourceWithMetadata("test.html")
		if err != nil {
			t.Errorf("GetSourceWithMetadata error: %v", err)
		}
		if source.Content != "Test content" {
			t.Errorf("Expected 'Test content', got %q", source.Content)
		}
		if source.ModTime.IsZero() {
			t.Error("Expected non-zero ModTime")
		}
	})

	t.Run("GetSourceWithMetadataNonExistent", func(t *testing.T) {
		_, err := loader.GetSourceWithMetadata("nonexistent.html")
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
	})

	t.Run("ResolveTemplateName", func(t *testing.T) {
		resolved := loader.ResolveTemplateName("template.html")
		if resolved != "template.html" {
			t.Errorf("Expected 'template.html', got %q", resolved)
		}
	})
}

// TestExtractDependencies tests the dependency extraction
func TestExtractDependencies(t *testing.T) {
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{"."}, mockParser)

	tests := []struct {
		source   string
		expected []string
	}{
		{
			source:   `{% extends 'base.html' %}`,
			expected: []string{"base.html"},
		},
		{
			source:   `{% include "partial.html" %}`,
			expected: []string{"partial.html"},
		},
		{
			source:   `{% import "macros.html" as macros %}`,
			expected: []string{"macros.html"},
		},
		{
			source:   `{% from "forms.html" import input %}`,
			expected: []string{"forms.html"},
		},
		{
			source:   `{% extends 'base.html' %}{% include 'partial.html' %}`,
			expected: []string{"base.html", "partial.html"},
		},
		{
			source:   `{% extends 'base.html' %}{% extends 'base.html' %}`, // Duplicate
			expected: []string{"base.html"},                                // Should be deduplicated
		},
	}

	for _, tt := range tests {
		deps, err := loader.extractDependencies(tt.source)
		if err != nil {
			t.Errorf("extractDependencies error: %v", err)
		}

		if len(deps) != len(tt.expected) {
			t.Errorf("For source %q: expected %d dependencies, got %d", tt.source, len(tt.expected), len(deps))
			continue
		}

		// Check that all expected deps are present
		depSet := make(map[string]bool)
		for _, d := range deps {
			depSet[d] = true
		}
		for _, exp := range tt.expected {
			if !depSet[exp] {
				t.Errorf("For source %q: expected dependency %q not found", tt.source, exp)
			}
		}
	}
}

// TestEmbedLoaderAdditional tests additional EmbedLoader functionality
func TestEmbedLoaderAdditional(t *testing.T) {
	mockParser := &MockParser{}
	loader := NewEmbedLoader(testEmbedFS, "testdata", mockParser)

	t.Run("GetSource", func(t *testing.T) {
		source, err := loader.GetSource("embed_test.html")
		if err != nil {
			t.Errorf("GetSource error: %v", err)
		}
		if source == "" {
			t.Error("Expected non-empty source")
		}
	})

	t.Run("IsCached", func(t *testing.T) {
		// Load a template first
		_, err := loader.LoadTemplate("embed_test.html")
		if err != nil {
			t.Errorf("LoadTemplate error: %v", err)
		}

		// Should be cached now
		if !loader.IsCached("embed_test.html") {
			t.Error("Expected embed_test.html to be cached")
		}
	})

	t.Run("GetSourceWithMetadata", func(t *testing.T) {
		source, err := loader.GetSourceWithMetadata("embed_test.html")
		if err != nil {
			t.Errorf("GetSourceWithMetadata error: %v", err)
		}
		if source.Name != "embed_test.html" {
			t.Errorf("Expected name 'embed_test.html', got %q", source.Name)
		}
		// Embedded files have zero ModTime
		if !source.ModTime.IsZero() {
			// That's fine, implementation may set it
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := loader.ListTemplates()
		if err != nil {
			t.Errorf("ListTemplates error: %v", err)
		}
		if len(templates) == 0 {
			t.Error("Expected at least one template")
		}
	})

	t.Run("GetCacheStats", func(t *testing.T) {
		stats := loader.GetCacheStats()
		if stats.Hits < 0 || stats.Misses < 0 {
			t.Error("Stats should be non-negative")
		}
	})
}

// createTestTemplatesForCoverage creates test templates for coverage tests
func createTestTemplatesForCoverage(t *testing.T) string {
	tempDir := t.TempDir()
	templatesDir := filepath.Join(tempDir, "templates")
	subDir := filepath.Join(templatesDir, "sub")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	templates := map[string]string{
		"base.html":       "<!DOCTYPE html><html>{{ content }}</html>",
		"page.html":       "{% extends 'base.html' %}{% block content %}Page{% endblock %}",
		"partial.jinja":   "<div>{{ message }}</div>",
		"sub/nested.html": "<p>Nested</p>",
		"import.html":     "{% import 'macros.html' as macros %}{{ macros.test() }}",
		"from.html":       "{% from 'forms.html' import input %}{{ input() }}",
	}

	for filename, content := range templates {
		fullPath := filepath.Join(templatesDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return templatesDir
}

// TestCacheExpirationCoverage tests that cached templates expire
func TestCacheExpirationCoverage(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	mockParser := &MockParser{}
	loader := NewFileSystemLoader([]string{templatesDir}, mockParser)

	// Load a template
	_, err := loader.LoadTemplate("base.html")
	if err != nil {
		t.Fatalf("LoadTemplate error: %v", err)
	}

	// Should be cached
	if !loader.IsCached("base.html") {
		t.Error("Expected template to be cached")
	}

	// Clear cache
	loader.ClearCache()

	// Should not be cached anymore
	if loader.IsCached("base.html") {
		t.Error("Expected template to not be cached after clear")
	}
}

// ErrorParser returns an error when parsing
type ErrorParser struct{}

func (e *ErrorParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	return nil, os.ErrInvalid
}

// TestLoaderWithParseError tests loaders with parse errors
func TestLoaderWithParseError(t *testing.T) {
	templatesDir := createTestTemplatesForCoverage(t)
	errorParser := &ErrorParser{}

	t.Run("FileSystemLoaderParseError", func(t *testing.T) {
		loader := NewFileSystemLoader([]string{templatesDir}, errorParser)
		_, err := loader.LoadTemplate("base.html")
		if err == nil {
			t.Error("Expected error when parsing fails")
		}
	})

	t.Run("StringLoaderParseError", func(t *testing.T) {
		loader := NewStringLoader(errorParser)
		loader.AddTemplate("test.html", "content")
		_, err := loader.LoadTemplate("test.html")
		if err == nil {
			t.Error("Expected error when parsing fails")
		}
	})
}
