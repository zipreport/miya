package miya_test

import (
	miya "github.com/zipreport/miya"
	"testing"

	"github.com/zipreport/miya/runtime"
)

func TestEnhancedAutoEscape(t *testing.T) {
	t.Run("Different escape contexts", func(t *testing.T) {
		testValue := `<script>alert("test")</script>`

		tests := []struct {
			context  runtime.EscapeContext
			expected string
		}{
			{
				context:  runtime.EscapeContextHTML,
				expected: "&lt;script&gt;alert(&#34;test&#34;)&lt;/script&gt;",
			},
			{
				context:  runtime.EscapeContextXML,
				expected: "&lt;script&gt;alert(&quot;test&quot;)&lt;/script&gt;",
			},
			{
				context:  runtime.EscapeContextJS,
				expected: "\\u003cscript\\u003ealert(\\\"test\\\")\\u003c/script\\u003e",
			},
			{
				context:  runtime.EscapeContextURL,
				expected: "%3Cscript%3Ealert%28%22test%22%29%3C%2Fscript%3E",
			},
			{
				context:  runtime.EscapeContextNone,
				expected: testValue, // No escaping
			},
		}

		config := runtime.DefaultAutoEscapeConfig()
		escaper := runtime.NewAutoEscaper(config)

		for _, tc := range tests {
			t.Run(string(tc.context), func(t *testing.T) {
				result := escaper.Escape(testValue, tc.context)
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	t.Run("SafeValue handling", func(t *testing.T) {
		config := runtime.DefaultAutoEscapeConfig()
		escaper := runtime.NewAutoEscaper(config)

		unsafeHTML := "<b>Bold</b>"
		safeHTML := runtime.SafeValue{Value: "<b>Bold</b>"}

		// Unsafe value should be escaped
		escapedUnsafe := escaper.Escape(unsafeHTML, runtime.EscapeContextHTML)
		if escapedUnsafe != "&lt;b&gt;Bold&lt;/b&gt;" {
			t.Errorf("Expected unsafe HTML to be escaped, got: %s", escapedUnsafe)
		}

		// Safe value should not be escaped
		escapedSafe := escaper.Escape(safeHTML, runtime.EscapeContextHTML)
		if escapedSafe != "<b>Bold</b>" {
			t.Errorf("Expected safe HTML to not be escaped, got: %s", escapedSafe)
		}
	})

	t.Run("Context detection from filename", func(t *testing.T) {
		config := runtime.DefaultAutoEscapeConfig()
		escaper := runtime.NewAutoEscaper(config)

		tests := []struct {
			filename string
			expected runtime.EscapeContext
		}{
			{"template.html", runtime.EscapeContextHTML},
			{"page.htm", runtime.EscapeContextHTML},
			{"data.xml", runtime.EscapeContextXML},
			{"script.js", runtime.EscapeContextJS},
			{"style.css", runtime.EscapeContextCSS},
			{"api.json", runtime.EscapeContextJSON},
			{"document.xhtml", runtime.EscapeContextXHTML},
			{"unknown.txt", runtime.EscapeContextHTML}, // Default
		}

		for _, tc := range tests {
			t.Run(tc.filename, func(t *testing.T) {
				context := escaper.DetectContext(tc.filename)
				if context != tc.expected {
					t.Errorf("Expected context %s for %s, got %s", tc.expected, tc.filename, context)
				}
			})
		}
	})

	t.Run("JavaScript context specific escaping", func(t *testing.T) {
		config := runtime.DefaultAutoEscapeConfig()
		escaper := runtime.NewAutoEscaper(config)

		tests := []struct {
			input    string
			expected string
		}{
			{"alert('hello')", "alert(\\'hello\\')"},
			{"var x = \"test\"", "var x = \\\"test\\\""},
			{"new\nline", "new\\nline"},
			{"tab\there", "tab\\there"},
			{"</script>", "\\u003c/script\\u003e"},
			{"<script>", "\\u003cscript\\u003e"},
		}

		for _, tc := range tests {
			result := escaper.Escape(tc.input, runtime.EscapeContextJS)
			if result != tc.expected {
				t.Errorf("JS escape '%s': expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})

	t.Run("CSS context escaping", func(t *testing.T) {
		config := runtime.DefaultAutoEscapeConfig()
		escaper := runtime.NewAutoEscaper(config)

		tests := []struct {
			input    string
			expected string
		}{
			{"font-family: 'Arial'", "font-family: \\'Arial\\'"},
			{"content: \"Hello\"", "content: \\\"Hello\\\""},
			{"color: red;\nbackground: blue", "color: red;\\A background: blue"},
		}

		for _, tc := range tests {
			result := escaper.Escape(tc.input, runtime.EscapeContextCSS)
			if result != tc.expected {
				t.Errorf("CSS escape '%s': expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})
}

func TestAutoEscapeConfiguration(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		config := runtime.DefaultAutoEscapeConfig()

		if !config.Enabled {
			t.Error("Expected auto-escape to be enabled by default")
		}

		if config.Context != runtime.EscapeContextHTML {
			t.Error("Expected default context to be HTML")
		}

		if len(config.Extensions) == 0 {
			t.Error("Expected default extensions to be configured")
		}
	})

	t.Run("Disable auto-escape", func(t *testing.T) {
		config := runtime.DefaultAutoEscapeConfig()
		config.Enabled = false
		escaper := runtime.NewAutoEscaper(config)

		result := escaper.Escape("<script>", runtime.EscapeContextHTML)
		if result != "<script>" {
			t.Errorf("Expected no escaping when disabled, got: %s", result)
		}
	})

	t.Run("Context wrapper functionality", func(t *testing.T) {
		ctx := miya.NewContext()
		ctx.Set("test", "value")

		config := runtime.DefaultAutoEscapeConfig()
		escaper := runtime.NewAutoEscaper(config)
		wrapper := runtime.NewContextWrapperFromInterface(ctx, escaper, runtime.EscapeContextJS)

		if !wrapper.IsAutoescapeEnabled() {
			t.Error("Expected auto-escaping to be enabled")
		}

		if wrapper.GetEscapeContext() != runtime.EscapeContextJS {
			t.Error("Expected JS context")
		}

		if wrapper.GetAutoEscaper() != escaper {
			t.Error("Expected correct auto-escaper")
		}

		// Test cloning preserves settings
		cloned := wrapper.Clone()
		if clonedWrapper, ok := cloned.(*runtime.ContextWrapper); ok {
			if clonedWrapper.GetEscapeContext() != runtime.EscapeContextJS {
				t.Error("Expected cloned context to preserve escape context")
			}
		} else {
			t.Error("Expected cloned context to be a ContextWrapper")
		}
	})
}
