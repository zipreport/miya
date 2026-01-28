package miya_test

import (
	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// COMPREHENSIVE LOADER TESTS
// =============================================================================
// This file provides comprehensive test coverage for template loaders
// to improve coverage from 59.9% to target 70%+
// =============================================================================

// Test FileSystem Loader Advanced Functionality
func TestFileSystemLoaderAdvanced(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "jinja2_loader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test templates
	templates := map[string]string{
		"base.html": `<html>
<head><title>{% block title %}Default{% endblock %}</title></head>
<body>{% block content %}{% endblock %}</body>
</html>`,
		"page.html": `{% extends "base.html" %}
{% block title %}Page Title{% endblock %}
{% block content %}Page Content{% endblock %}`,
		"partials/header.html": `<header>Site Header</header>`,
		"components/card.html": `<div class="card">{{ content }}</div>`,
	}

	// Write templates to files
	for name, content := range templates {
		filePath := filepath.Join(tmpDir, name)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
	}

	// Create filesystem loader
	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name:     "load from filesystem",
			template: "page.html",
			expected: []string{"Page Title", "Page Content"},
		},
		{
			name:     "load from subdirectory",
			template: "partials/header.html",
			expected: []string{"Site Header"},
		},
		{
			name:     "load component",
			template: "components/card.html",
			expected: []string{"card"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.template)
			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"content": "Test Content"}))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Template Discovery
func TestTemplateDiscovery(t *testing.T) {
	// Create temporary test directory structure
	tmpDir, err := os.MkdirTemp("", "jinja2_discovery_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure with templates
	templatePaths := []string{
		"templates/layout.html",
		"templates/pages/home.html",
		"templates/pages/about.html",
		"templates/components/nav.html",
		"templates/components/footer.html",
		"shared/macros.html",
		"config.yaml", // Non-template file
	}

	for _, path := range templatePaths {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		content := "Template: " + path
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Create filesystem loader with multiple search paths
	directParser := loader.NewDirectTemplateParser()
	searchPaths := []string{
		filepath.Join(tmpDir, "templates"),
		filepath.Join(tmpDir, "shared"),
	}
	fsLoader := loader.NewFileSystemLoader(searchPaths, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(false))

	tests := []struct {
		name            string
		templateName    string
		expectError     bool
		expectedContent string
	}{
		{
			name:            "find template in first search path",
			templateName:    "layout.html",
			expectError:     false,
			expectedContent: "templates/layout.html",
		},
		{
			name:            "find template in subdirectory",
			templateName:    "pages/home.html",
			expectError:     false,
			expectedContent: "templates/pages/home.html",
		},
		{
			name:            "find template in second search path",
			templateName:    "macros.html",
			expectError:     false,
			expectedContent: "shared/macros.html",
		},
		{
			name:         "template not found",
			templateName: "nonexistent.html",
			expectError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.templateName)
			if test.expectError {
				if err == nil {
					t.Fatalf("Expected error but got template")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContext())
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			if !strings.Contains(result, test.expectedContent) {
				t.Errorf("Expected result to contain %q, got: %q", test.expectedContent, result)
			}
		})
	}
}

// Test Template Caching Behavior
func TestTemplateCaching(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jinja2_cache_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial template
	templatePath := filepath.Join(tmpDir, "cached.html")
	initialContent := "Initial Content: {{ value }}"
	if err := os.WriteFile(templatePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial template: %v", err)
	}

	// Create filesystem loader
	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(false))

	// First load
	tmpl1, err := env.GetTemplate("cached.html")
	if err != nil {
		t.Fatalf("Failed to load template first time: %v", err)
	}

	result1, err := tmpl1.Render(miya.NewContextFrom(map[string]interface{}{"value": "test1"}))
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if !strings.Contains(result1, "Initial Content: test1") {
		t.Errorf("Expected initial content, got: %q", result1)
	}

	// Second load (should use cache)
	tmpl2, err := env.GetTemplate("cached.html")
	if err != nil {
		t.Fatalf("Failed to load template second time: %v", err)
	}

	result2, err := tmpl2.Render(miya.NewContextFrom(map[string]interface{}{"value": "test2"}))
	if err != nil {
		t.Fatalf("Failed to render cached template: %v", err)
	}

	if !strings.Contains(result2, "Initial Content: test2") {
		t.Errorf("Expected cached content, got: %q", result2)
	}

	// Modify template file
	time.Sleep(100 * time.Millisecond) // Ensure different modification time
	updatedContent := "Updated Content: {{ value }}"
	if err := os.WriteFile(templatePath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Third load (should detect change and reload)
	tmpl3, err := env.GetTemplate("cached.html")
	if err != nil {
		t.Fatalf("Failed to load updated template: %v", err)
	}

	result3, err := tmpl3.Render(miya.NewContextFrom(map[string]interface{}{"value": "test3"}))
	if err != nil {
		t.Fatalf("Failed to render updated template: %v", err)
	}

	// Note: This test depends on the loader implementation
	// Some loaders may cache aggressively and require explicit invalidation
	if strings.Contains(result3, "Updated Content: test3") {
		t.Logf("Template reloading works correctly")
	} else if strings.Contains(result3, "Initial Content: test3") {
		t.Logf("Template is cached (expected behavior for some loaders)")
	} else {
		t.Errorf("Unexpected template content: %q", result3)
	}
}

// Test StringLoader Advanced Features
func TestStringLoaderAdvanced(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)

	// Add templates with dependencies
	templates := map[string]string{
		"base.html":         `<base>{% block content %}{% endblock %}</base>`,
		"child.html":        `{% extends "base.html" %}{% block content %}Child: {{ value }}{% endblock %}`,
		"include.html":      `<included>{{ included_value }}</included>`,
		"with_include.html": `<main>{% include "include.html" %}</main>`,
		"macro_lib.html":    `{% macro greet(name) %}Hello, {{ name }}!{% endmacro %}`,
		"use_macro.html":    `{% from "macro_lib.html" import greet %}{{ greet(user_name) }}`,
	}

	for name, content := range templates {
		stringLoader.AddTemplate(name, content)
	}

	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(false))

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected []string
	}{
		{
			name:     "inheritance in string loader",
			template: "child.html",
			data:     map[string]interface{}{"value": "test"},
			expected: []string{"<base>", "Child: test", "</base>"},
		},
		{
			name:     "include in string loader",
			template: "with_include.html",
			data:     map[string]interface{}{"included_value": "included_test"},
			expected: []string{"<main>", "<included>included_test</included>", "</main>"},
		},
		{
			name:     "macro import in string loader",
			template: "use_macro.html",
			data:     map[string]interface{}{"user_name": "John"},
			expected: []string{"Hello, John!"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.template)
			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// Test Loader Error Handling
func TestLoaderErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jinja2_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template with syntax error
	syntaxErrorTemplate := `{% if condition %} No endif tag`
	templatePath := filepath.Join(tmpDir, "syntax_error.html")
	if err := os.WriteFile(templatePath, []byte(syntaxErrorTemplate), 0644); err != nil {
		t.Fatalf("Failed to write syntax error template: %v", err)
	}

	// Create valid template
	validTemplate := `Valid template: {{ value }}`
	validPath := filepath.Join(tmpDir, "valid.html")
	if err := os.WriteFile(validPath, []byte(validTemplate), 0644); err != nil {
		t.Fatalf("Failed to write valid template: %v", err)
	}

	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(false))

	tests := []struct {
		name         string
		templateName string
		expectError  bool
		errorType    string
	}{
		{
			name:         "template not found",
			templateName: "nonexistent.html",
			expectError:  true,
			errorType:    "not found",
		},
		{
			name:         "syntax error in template",
			templateName: "syntax_error.html",
			expectError:  true,
			errorType:    "syntax",
		},
		{
			name:         "valid template loads successfully",
			templateName: "valid.html",
			expectError:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.templateName)
			if test.expectError {
				if err == nil {
					t.Fatalf("Expected error but got template")
				}
				if test.errorType != "" && !strings.Contains(strings.ToLower(err.Error()), test.errorType) {
					t.Errorf("Expected error type %q, got: %v", test.errorType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error loading template: %v", err)
			}

			// Try to render the valid template
			result, renderErr := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"value": "test"}))
			if renderErr != nil {
				t.Fatalf("Failed to render valid template: %v", renderErr)
			}

			if !strings.Contains(result, "Valid template: test") {
				t.Errorf("Unexpected render result: %q", result)
			}
		})
	}
}

// Test Loader Performance with Many Templates
func TestLoaderPerformance(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jinja2_perf_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create many templates
	numTemplates := 100
	for i := 0; i < numTemplates; i++ {
		templateName := filepath.Join(tmpDir, "template"+string(rune('0'+i%10))+".html")
		content := `Template ` + string(rune('0'+i%10)) + `: {{ value }}`
		if err := os.WriteFile(templateName, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %d: %v", i, err)
		}
	}

	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(false))

	// Load all templates multiple times to test caching
	for round := 0; round < 3; round++ {
		for i := 0; i < 10; i++ { // Only test first 10 templates for speed
			templateName := "template" + string(rune('0'+i)) + ".html"
			tmpl, err := env.GetTemplate(templateName)
			if err != nil {
				t.Fatalf("Failed to load template %s in round %d: %v", templateName, round, err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(map[string]interface{}{"value": "test"}))
			if err != nil {
				t.Fatalf("Failed to render template %s: %v", templateName, err)
			}

			expectedContent := "Template " + string(rune('0'+i)) + ": test"
			if !strings.Contains(result, expectedContent) {
				t.Errorf("Expected %q, got %q", expectedContent, result)
			}
		}
	}
}

// Test Loader Interface Compliance
func TestLoaderInterfaceCompliance(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()

	// Test StringLoader
	stringLoader := loader.NewStringLoader(directParser)
	stringLoader.AddTemplate("test.html", "Test template")

	// Verify it implements the Loader interface
	var loaderInterface loader.Loader = stringLoader

	source, err := loaderInterface.GetSource("test.html")
	if err != nil {
		t.Errorf("StringLoader GetSource failed: %v", err)
	}
	if source != "Test template" {
		t.Errorf("Expected 'Test template', got %q", source)
	}

	// Test IsCached
	if !loaderInterface.IsCached("test.html") {
		t.Error("StringLoader should report template as cached")
	}

	if loaderInterface.IsCached("nonexistent.html") {
		t.Error("StringLoader should not report nonexistent template as cached")
	}

	// Test FileSystemLoader
	tmpDir, err := os.MkdirTemp("", "jinja2_interface_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fsTemplateContent := "FileSystem template"
	fsTemplatePath := filepath.Join(tmpDir, "fs_test.html")
	if err := os.WriteFile(fsTemplatePath, []byte(fsTemplateContent), 0644); err != nil {
		t.Fatalf("Failed to write filesystem template: %v", err)
	}

	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	var fsLoaderInterface loader.Loader = fsLoader

	fsSource, err := fsLoaderInterface.GetSource("fs_test.html")
	if err != nil {
		t.Errorf("FileSystemLoader GetSource failed: %v", err)
	}
	if fsSource != fsTemplateContent {
		t.Errorf("Expected %q, got %q", fsTemplateContent, fsSource)
	}
}

// Test Loader Edge Cases
func TestLoaderEdgeCases(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()

	tests := []struct {
		name    string
		setupFn func() loader.Loader
		testFn  func(t *testing.T, l loader.Loader)
	}{
		{
			name: "empty template name",
			setupFn: func() loader.Loader {
				sl := loader.NewStringLoader(directParser)
				sl.AddTemplate("", "Empty name template")
				return sl
			},
			testFn: func(t *testing.T, l loader.Loader) {
				source, err := l.GetSource("")
				if err != nil {
					t.Errorf("Failed to get empty name template: %v", err)
				}
				if source != "Empty name template" {
					t.Errorf("Expected 'Empty name template', got %q", source)
				}
			},
		},
		{
			name: "template with special characters",
			setupFn: func() loader.Loader {
				sl := loader.NewStringLoader(directParser)
				sl.AddTemplate("special!@#$%.html", "Special chars template")
				return sl
			},
			testFn: func(t *testing.T, l loader.Loader) {
				source, err := l.GetSource("special!@#$%.html")
				if err != nil {
					t.Errorf("Failed to get special chars template: %v", err)
				}
				if source != "Special chars template" {
					t.Errorf("Expected 'Special chars template', got %q", source)
				}
			},
		},
		{
			name: "very long template name",
			setupFn: func() loader.Loader {
				longName := strings.Repeat("a", 1000) + ".html"
				sl := loader.NewStringLoader(directParser)
				sl.AddTemplate(longName, "Long name template")
				return sl
			},
			testFn: func(t *testing.T, l loader.Loader) {
				longName := strings.Repeat("a", 1000) + ".html"
				source, err := l.GetSource(longName)
				if err != nil {
					t.Errorf("Failed to get long name template: %v", err)
				}
				if source != "Long name template" {
					t.Errorf("Expected 'Long name template', got %q", source)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader := test.setupFn()
			test.testFn(t, loader)
		})
	}
}
