package miya_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zipreport/miya/loader"
)

func TestAdvancedTemplateDiscovery(t *testing.T) {
	// Create a temporary directory with test templates
	tempDir := t.TempDir()

	// Create comprehensive test template structure
	templates := map[string]string{
		"base.html":                     `<html><head><title>{{ title }}</title></head><body>{% block content %}{% endblock %}</body></html>`,
		"layouts/admin.html":            `{% extends "base.html" %}{% block content %}<div class="admin">{{ super() }}</div>{% endblock %}`,
		"layouts/public.html":           `{% extends "base.html" %}{% block content %}<div class="public">{{ super() }}</div>{% endblock %}`,
		"pages/home.html":               `{% extends "layouts/public.html" %}{% block content %}<h1>Home</h1>{% include "components/header.j2" %}{% endblock %}`,
		"pages/about.html":              `{% extends "layouts/public.html" %}{% block content %}<h1>About</h1>{% endblock %}`,
		"pages/admin/dashboard.html":    `{% extends "layouts/admin.html" %}{% import "macros/admin.j2" as admin %}{% block content %}<h1>Dashboard</h1>{% endblock %}`,
		"components/header.j2":          `<header>{{ site_name }}</header>`,
		"components/footer.j2":          `<footer>&copy; {{ year }}</footer>`,
		"components/forms/login.jinja2": `<form>{% from "macros/forms.j2" import input %}</form>`,
		"macros/admin.j2":               `{% macro admin_menu() %}<nav>Admin Menu</nav>{% endmacro %}`,
		"macros/forms.j2":               `{% macro input(name, type="text") %}<input name="{{ name }}" type="{{ type }}">{% endmacro %}`,
		"emails/welcome.html":           `<h1>Welcome {{ name }}!</h1>{% include "components/footer.j2" %}`,
		"reports/monthly.jinja":         `{% extends "layouts/admin.html" %}{% block content %}<h1>Monthly Report</h1>{% endblock %}`,
		"invalid.txt":                   "This should not be listed",
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

	// Create filesystem loader for testing
	var parser loader.TemplateParser = nil
	fsLoader := loader.NewFileSystemLoader([]string{tempDir}, parser)

	t.Run("Search templates by pattern", func(t *testing.T) {
		// Test various search patterns
		tests := []struct {
			pattern  string
			expected []string
		}{
			{
				pattern:  "*.html",
				expected: []string{"base.html", "layouts/admin.html", "layouts/public.html", "pages/home.html", "pages/about.html", "pages/admin/dashboard.html", "emails/welcome.html"},
			},
			{
				pattern:  "pages/*.html",
				expected: []string{"pages/home.html", "pages/about.html"},
			},
			{
				pattern:  "components/*",
				expected: []string{"components/header.j2", "components/footer.j2"},
			},
			{
				pattern:  "macros/*.j2",
				expected: []string{"macros/admin.j2", "macros/forms.j2"},
			},
		}

		for _, tc := range tests {
			t.Run("Pattern: "+tc.pattern, func(t *testing.T) {
				results, err := fsLoader.SearchTemplates(tc.pattern)
				if err != nil {
					t.Fatalf("Failed to search templates: %v", err)
				}

				if len(results) != len(tc.expected) {
					t.Errorf("Expected %d results, got %d: %v", len(tc.expected), len(results), results)
					return
				}

				// Convert to map for easier checking
				resultMap := make(map[string]bool)
				for _, result := range results {
					resultMap[result] = true
				}

				for _, expected := range tc.expected {
					if !resultMap[expected] {
						t.Errorf("Expected template %s not found in results", expected)
					}
				}
			})
		}
	})

	t.Run("Get templates by extension", func(t *testing.T) {
		tests := []struct {
			extension string
			expected  []string
		}{
			{
				extension: ".html",
				expected:  []string{"base.html", "layouts/admin.html", "layouts/public.html", "pages/home.html", "pages/about.html", "pages/admin/dashboard.html", "emails/welcome.html"},
			},
			{
				extension: "j2", // Test without leading dot
				expected:  []string{"components/header.j2", "components/footer.j2", "macros/admin.j2", "macros/forms.j2"},
			},
			{
				extension: ".jinja2",
				expected:  []string{"components/forms/login.jinja2"},
			},
			{
				extension: ".jinja",
				expected:  []string{"reports/monthly.jinja"},
			},
		}

		for _, tc := range tests {
			t.Run("Extension: "+tc.extension, func(t *testing.T) {
				results, err := fsLoader.GetTemplatesByExtension(tc.extension)
				if err != nil {
					t.Fatalf("Failed to get templates by extension: %v", err)
				}

				if len(results) != len(tc.expected) {
					t.Errorf("Expected %d results, got %d: %v", len(tc.expected), len(results), results)
					return
				}

				// Convert to map for easier checking
				resultMap := make(map[string]bool)
				for _, result := range results {
					resultMap[result] = true
				}

				for _, expected := range tc.expected {
					if !resultMap[expected] {
						t.Errorf("Expected template %s not found in results", expected)
					}
				}
			})
		}
	})

	t.Run("Get templates in directory", func(t *testing.T) {
		tests := []struct {
			directory string
			expected  []string
		}{
			{
				directory: ".",
				expected:  []string{"base.html"},
			},
			{
				directory: "pages",
				expected:  []string{"pages/home.html", "pages/about.html"},
			},
			{
				directory: "components",
				expected:  []string{"components/header.j2", "components/footer.j2"},
			},
			{
				directory: "macros",
				expected:  []string{"macros/admin.j2", "macros/forms.j2"},
			},
			{
				directory: "pages/admin",
				expected:  []string{"pages/admin/dashboard.html"},
			},
		}

		for _, tc := range tests {
			t.Run("Directory: "+tc.directory, func(t *testing.T) {
				results, err := fsLoader.GetTemplatesInDirectory(tc.directory)
				if err != nil {
					t.Fatalf("Failed to get templates in directory: %v", err)
				}

				if len(results) != len(tc.expected) {
					t.Errorf("Expected %d results, got %d: %v", len(tc.expected), len(results), results)
					return
				}

				// Convert to map for easier checking
				resultMap := make(map[string]bool)
				for _, result := range results {
					resultMap[result] = true
				}

				for _, expected := range tc.expected {
					if !resultMap[expected] {
						t.Errorf("Expected template %s not found in results", expected)
					}
				}
			})
		}
	})

	t.Run("Get template info", func(t *testing.T) {
		tests := []struct {
			templateName      string
			expectedExtension string
			expectedDirectory string
			expectedDepsCount int
		}{
			{
				templateName:      "base.html",
				expectedExtension: ".html",
				expectedDirectory: ".",
				expectedDepsCount: 0,
			},
			{
				templateName:      "pages/home.html",
				expectedExtension: ".html",
				expectedDirectory: "pages",
				expectedDepsCount: 2, // extends layouts/public.html, includes components/header.j2
			},
			{
				templateName:      "pages/admin/dashboard.html",
				expectedExtension: ".html",
				expectedDirectory: "pages/admin",
				expectedDepsCount: 2, // extends layouts/admin.html, imports macros/admin.j2
			},
			{
				templateName:      "components/forms/login.jinja2",
				expectedExtension: ".jinja2",
				expectedDirectory: "components/forms",
				expectedDepsCount: 1, // from macros/forms.j2
			},
		}

		for _, tc := range tests {
			t.Run("Template: "+tc.templateName, func(t *testing.T) {
				info, err := fsLoader.GetTemplateInfo(tc.templateName)
				if err != nil {
					t.Fatalf("Failed to get template info: %v", err)
				}

				if info.Name != tc.templateName {
					t.Errorf("Expected name %s, got %s", tc.templateName, info.Name)
				}

				if info.Extension != tc.expectedExtension {
					t.Errorf("Expected extension %s, got %s", tc.expectedExtension, info.Extension)
				}

				if info.Directory != tc.expectedDirectory {
					t.Errorf("Expected directory %s, got %s", tc.expectedDirectory, info.Directory)
				}

				if info.Size <= 0 {
					t.Error("Expected positive file size")
				}

				if info.ModTime.IsZero() {
					t.Error("Expected non-zero modification time")
				}

				if len(info.Dependencies) != tc.expectedDepsCount {
					t.Errorf("Expected %d dependencies, got %d: %v", tc.expectedDepsCount, len(info.Dependencies), info.Dependencies)
				}

				// Verify that the template path exists
				if _, err := os.Stat(info.Path); os.IsNotExist(err) {
					t.Errorf("Template path %s does not exist", info.Path)
				}
			})
		}
	})

	t.Run("Template dependency extraction", func(t *testing.T) {
		tests := []struct {
			templateName string
			expectedDeps []string
		}{
			{
				templateName: "pages/home.html",
				expectedDeps: []string{"layouts/public.html", "components/header.j2"},
			},
			{
				templateName: "pages/admin/dashboard.html",
				expectedDeps: []string{"layouts/admin.html", "macros/admin.j2"},
			},
			{
				templateName: "components/forms/login.jinja2",
				expectedDeps: []string{"macros/forms.j2"},
			},
			{
				templateName: "emails/welcome.html",
				expectedDeps: []string{"components/footer.j2"},
			},
		}

		for _, tc := range tests {
			t.Run("Dependencies for: "+tc.templateName, func(t *testing.T) {
				info, err := fsLoader.GetTemplateInfo(tc.templateName)
				if err != nil {
					t.Fatalf("Failed to get template info: %v", err)
				}

				if len(info.Dependencies) != len(tc.expectedDeps) {
					t.Errorf("Expected %d dependencies, got %d: %v", len(tc.expectedDeps), len(info.Dependencies), info.Dependencies)
					return
				}

				depMap := make(map[string]bool)
				for _, dep := range info.Dependencies {
					depMap[dep] = true
				}

				for _, expectedDep := range tc.expectedDeps {
					if !depMap[expectedDep] {
						t.Errorf("Expected dependency %s not found", expectedDep)
					}
				}
			})
		}
	})
}
