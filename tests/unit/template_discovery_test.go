package miya_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zipreport/miya/loader"
)

func TestTemplateDiscoveryBasic(t *testing.T) {
	// Create a temporary directory with test templates
	tempDir := t.TempDir()

	// Create test templates
	templates := map[string]string{
		"base.html":             "<html><head><title>{{ title }}</title></head><body>{% block content %}{% endblock %}</body></html>",
		"pages/index.html":      "{% extends 'base.html' %}{% block content %}<h1>Home Page</h1>{% endblock %}",
		"pages/about.html":      "{% extends 'base.html' %}{% block content %}<h1>About Page</h1>{% endblock %}",
		"components/header.j2":  "<header>{{ site_name }}</header>",
		"components/footer.j2":  "<footer>&copy; {{ year }}</footer>",
		"emails/welcome.jinja2": "<h1>Welcome {{ name }}!</h1>",
		"invalid.txt":           "This should not be listed",
	}

	for templatePath, content := range templates {
		fullPath := filepath.Join(tempDir, templatePath)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Write template file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", templatePath, err)
		}
	}

	t.Run("FileSystemLoader template discovery", func(t *testing.T) {
		// Create parser mock (we only need the interface for this test)
		var parser loader.TemplateParser = nil

		fsLoader := loader.NewFileSystemLoader([]string{tempDir}, parser)

		templates, err := fsLoader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Expected templates (should exclude invalid.txt)
		expected := []string{
			"base.html",
			"pages/index.html",
			"pages/about.html",
			"components/header.j2",
			"components/footer.j2",
			"emails/welcome.jinja2",
		}

		if len(templates) != len(expected) {
			t.Errorf("Expected %d templates, got %d: %v", len(expected), len(templates), templates)
		}

		// Check that all expected templates are found
		templateMap := make(map[string]bool)
		for _, tmpl := range templates {
			templateMap[tmpl] = true
		}

		for _, expectedTmpl := range expected {
			if !templateMap[expectedTmpl] {
				t.Errorf("Expected template %s not found in list", expectedTmpl)
			}
		}

		// Ensure invalid.txt is not in the list
		if templateMap["invalid.txt"] {
			t.Error("invalid.txt should not be in template list")
		}
	})

	t.Run("StringLoader template discovery", func(t *testing.T) {
		var parser loader.TemplateParser = nil
		stringLoader := loader.NewStringLoader(parser)

		// Add some templates
		stringLoader.AddTemplate("template1.html", "<h1>Template 1</h1>")
		stringLoader.AddTemplate("template2.html", "<h1>Template 2</h1>")
		stringLoader.AddTemplate("partial.j2", "<div>Partial content</div>")

		templates, err := stringLoader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		expected := []string{"template1.html", "template2.html", "partial.j2"}

		if len(templates) != len(expected) {
			t.Errorf("Expected %d templates, got %d: %v", len(expected), len(templates), templates)
		}

		// Check that all expected templates are found
		templateMap := make(map[string]bool)
		for _, tmpl := range templates {
			templateMap[tmpl] = true
		}

		for _, expectedTmpl := range expected {
			if !templateMap[expectedTmpl] {
				t.Errorf("Expected template %s not found in list", expectedTmpl)
			}
		}
	})

	t.Run("ChainLoader template discovery", func(t *testing.T) {
		var parser loader.TemplateParser = nil

		// Create filesystem loader
		fsLoader := loader.NewFileSystemLoader([]string{tempDir}, parser)

		// Create string loader with additional templates
		stringLoader := loader.NewStringLoader(parser)
		stringLoader.AddTemplate("memory1.html", "<h1>Memory Template 1</h1>")
		stringLoader.AddTemplate("memory2.html", "<h1>Memory Template 2</h1>")

		// Create chain loader
		chainLoader := loader.NewChainLoader(fsLoader, stringLoader)

		templates, err := chainLoader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Should include templates from both loaders
		expectedFromFS := []string{
			"base.html",
			"pages/index.html",
			"pages/about.html",
			"components/header.j2",
			"components/footer.j2",
			"emails/welcome.jinja2",
		}
		expectedFromString := []string{"memory1.html", "memory2.html"}

		expectedTotal := len(expectedFromFS) + len(expectedFromString)

		if len(templates) != expectedTotal {
			t.Errorf("Expected %d templates, got %d: %v", expectedTotal, len(templates), templates)
		}

		// Check that all expected templates are found
		templateMap := make(map[string]bool)
		for _, tmpl := range templates {
			templateMap[tmpl] = true
		}

		allExpected := append(expectedFromFS, expectedFromString...)
		for _, expectedTmpl := range allExpected {
			if !templateMap[expectedTmpl] {
				t.Errorf("Expected template %s not found in list", expectedTmpl)
			}
		}
	})

	t.Run("Template filtering by extension", func(t *testing.T) {
		var parser loader.TemplateParser = nil
		fsLoader := loader.NewFileSystemLoader([]string{tempDir}, parser)

		// Set custom extensions (only .html files)
		fsLoader.SetExtensions([]string{".html"})

		templates, err := fsLoader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Should only include .html files
		expected := []string{
			"base.html",
			"pages/index.html",
			"pages/about.html",
		}

		if len(templates) != len(expected) {
			t.Errorf("Expected %d templates, got %d: %v", len(expected), len(templates), templates)
		}

		// Check that all templates have .html extension
		for _, tmpl := range templates {
			if filepath.Ext(tmpl) != ".html" {
				t.Errorf("Template %s should not be included with .html filter", tmpl)
			}
		}
	})
}

func TestTemplateDiscoveryPatterns(t *testing.T) {
	// Create a temporary directory with various template patterns
	tempDir := t.TempDir()

	// Create nested structure
	templates := map[string]string{
		"layouts/base.html":                 "base layout",
		"layouts/admin/base.html":           "admin base layout",
		"pages/public/home.html":            "home page",
		"pages/public/about.html":           "about page",
		"pages/admin/dashboard.html":        "dashboard",
		"components/forms/login.j2":         "login form",
		"components/navigation/menu.jinja2": "navigation menu",
		"emails/notifications/welcome.html": "welcome email",
	}

	for templatePath, content := range templates {
		fullPath := filepath.Join(tempDir, templatePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", templatePath, err)
		}
	}

	t.Run("Deep nested template discovery", func(t *testing.T) {
		var parser loader.TemplateParser = nil
		fsLoader := loader.NewFileSystemLoader([]string{tempDir}, parser)

		templates, err := fsLoader.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		if len(templates) != len(templates) {
			t.Errorf("Expected %d templates, got %d", len(templates), len(templates))
		}

		// Check for specific nested templates
		expectedTemplates := []string{
			"layouts/base.html",
			"layouts/admin/base.html",
			"pages/public/home.html",
			"components/forms/login.j2",
			"components/navigation/menu.jinja2",
			"emails/notifications/welcome.html",
		}

		templateMap := make(map[string]bool)
		for _, tmpl := range templates {
			templateMap[tmpl] = true
		}

		for _, expected := range expectedTemplates {
			if !templateMap[expected] {
				t.Errorf("Expected nested template %s not found", expected)
			}
		}
	})
}
