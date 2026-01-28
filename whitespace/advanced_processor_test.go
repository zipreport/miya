package whitespace

import (
	"strings"
	"testing"

	"github.com/zipreport/miya/parser"
)

// Test AdvancedWhitespaceProcessor creation
func TestNewAdvancedWhitespaceProcessor(t *testing.T) {
	t.Run("Creates processor with correct settings", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(true, false, true)

		if !processor.trimBlocks {
			t.Error("Expected trimBlocks to be true")
		}
		if processor.lstripBlocks {
			t.Error("Expected lstripBlocks to be false")
		}
		if !processor.keepTrailingNewline {
			t.Error("Expected keepTrailingNewline to be true")
		}
	})
}

// Test ProcessTemplate functionality
func TestAdvancedProcessTemplate(t *testing.T) {
	t.Run("Simple template without whitespace control", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "Hello {{ name }} World"

		result := processor.ProcessTemplate(template)

		if result != template {
			t.Errorf("Expected '%s', got '%s'", template, result)
		}
	})

	t.Run("Removes trailing newline when keepTrailingNewline is false", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, false)
		template := "Hello World\n"

		result := processor.ProcessTemplate(template)

		if result != "Hello World" {
			t.Errorf("Expected 'Hello World', got '%s'", result)
		}
	})

	t.Run("Keeps trailing newline when keepTrailingNewline is true", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "Hello World\n"

		result := processor.ProcessTemplate(template)

		if result != "Hello World\n" {
			t.Errorf("Expected 'Hello World\\n', got '%s'", result)
		}
	})
}

// Test inline whitespace control processing
func TestProcessInlineWhitespaceControl(t *testing.T) {
	t.Run("Block tags with both left and right strip", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {%- if true -%}  "

		result := processor.processInlineWhitespaceControl(template)

		if result != "{% if true %}" {
			t.Errorf("Expected '{{ if true }}', got '%s'", result)
		}
	})

	t.Run("Block tags with left strip only", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {%- if true %}  "

		result := processor.processInlineWhitespaceControl(template)

		if result != "{% if true %}  " {
			t.Errorf("Expected '{{ if true }}  ', got '%s'", result)
		}
	})

	t.Run("Block tags with right strip only", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {% if true -%}  "

		result := processor.processInlineWhitespaceControl(template)

		if result != "  {% if true %}" {
			t.Errorf("Expected '  {{ if true }}', got '%s'", result)
		}
	})

	t.Run("Variable tags with both left and right strip", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {{- name -}}  "

		result := processor.processInlineWhitespaceControl(template)

		if result != "{{ name }}" {
			t.Errorf("Expected '{{ name }}', got '%s'", result)
		}
	})

	t.Run("Variable tags with left strip only", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {{- name }}  "

		result := processor.processInlineWhitespaceControl(template)

		if result != "{{ name }}  " {
			t.Errorf("Expected '{{ name }}  ', got '%s'", result)
		}
	})

	t.Run("Variable tags with right strip only", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {{ name -}}  "

		result := processor.processInlineWhitespaceControl(template)

		if result != "  {{ name }}" {
			t.Errorf("Expected '  {{ name }}', got '%s'", result)
		}
	})

	t.Run("Comment tags are removed entirely", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "Before  {# comment #}  After"

		result := processor.processInlineWhitespaceControl(template)

		if result != "BeforeAfter" {
			t.Errorf("Expected 'BeforeAfter', got '%s'", result)
		}
	})

	t.Run("Comment tags with strip modifiers are removed", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "Before  {#- comment -#}  After"

		result := processor.processInlineWhitespaceControl(template)

		if result != "BeforeAfter" {
			t.Errorf("Expected 'BeforeAfter', got '%s'", result)
		}
	})

	t.Run("Multiple tags processed correctly", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {%- for item in items -%}  {{- item.name -}}  {%- endfor -%}  "

		result := processor.processInlineWhitespaceControl(template)

		expected := "{% for item in items %}{{ item.name }}{% endfor %}"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("Complex template with mixed whitespace control", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := `
    {%- if users %}
      <ul>
      {%- for user in users %}
        <li>{{- user.name -}}</li>
      {%- endfor %}
      </ul>
    {%- endif -%}
    `

		result := processor.processInlineWhitespaceControl(template)

		// Should remove strip modifiers but keep structure
		if !strings.Contains(result, "{% if users %}") {
			t.Error("Expected processed if tag without strip modifiers")
		}
		if !strings.Contains(result, "{{ user.name }}") {
			t.Error("Expected processed variable tag without strip modifiers")
		}
	})
}

// Test global whitespace application
func TestApplyGlobalWhitespace(t *testing.T) {
	t.Run("trimBlocks removes newlines after block tags", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(true, false, true)
		template := "{% if true %}\nContent\n{% endif %}\nAfter"

		result := processor.applyGlobalWhitespace(template)

		expected := "{% if true %}Content\n{% endif %}After"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("trimBlocks handles CRLF line endings", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(true, false, true)
		template := "{% for item in items %}\r\nContent"

		result := processor.applyGlobalWhitespace(template)

		expected := "{% for item in items %}Content"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("lstripBlocks removes whitespace before block tags", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, true, true)
		template := "Before\n    {% if true %}\nContent"

		result := processor.applyGlobalWhitespace(template)

		expected := "Before\n{% if true %}\nContent"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("Both trimBlocks and lstripBlocks work together", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(true, true, true)
		template := "Before\n    {% if true %}\n    Content\n  {% endif %}\nAfter"

		result := processor.applyGlobalWhitespace(template)

		expected := "Before\n{% if true %}    Content\n{% endif %}After"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

// Test whitespace control parsing
func TestParseWhitespaceControl(t *testing.T) {
	t.Run("No whitespace control modifiers", func(t *testing.T) {
		content, control := ParseWhitespaceControl("if true")

		if content != "if true" {
			t.Errorf("Expected 'if true', got '%s'", content)
		}
		if control.LeftStrip || control.RightStrip {
			t.Error("Expected no whitespace control")
		}
	})

	t.Run("Left strip modifier", func(t *testing.T) {
		content, control := ParseWhitespaceControl("-if true")

		if content != "if true" {
			t.Errorf("Expected 'if true', got '%s'", content)
		}
		if !control.LeftStrip {
			t.Error("Expected LeftStrip to be true")
		}
		if control.RightStrip {
			t.Error("Expected RightStrip to be false")
		}
	})

	t.Run("Right strip modifier", func(t *testing.T) {
		content, control := ParseWhitespaceControl("if true-")

		if content != "if true" {
			t.Errorf("Expected 'if true', got '%s'", content)
		}
		if control.LeftStrip {
			t.Error("Expected LeftStrip to be false")
		}
		if !control.RightStrip {
			t.Error("Expected RightStrip to be true")
		}
	})

	t.Run("Both strip modifiers", func(t *testing.T) {
		content, control := ParseWhitespaceControl("-if true-")

		if content != "if true" {
			t.Errorf("Expected 'if true', got '%s'", content)
		}
		if !control.LeftStrip {
			t.Error("Expected LeftStrip to be true")
		}
		if !control.RightStrip {
			t.Error("Expected RightStrip to be true")
		}
	})

	t.Run("Strips and trims whitespace around content", func(t *testing.T) {
		content, control := ParseWhitespaceControl("-  if true  -")

		if content != "if true" {
			t.Errorf("Expected 'if true', got '%s'", content)
		}
		if !control.LeftStrip || !control.RightStrip {
			t.Error("Expected both strip modifiers")
		}
	})
}

// Test ApplyWhitespaceControl function
func TestApplyWhitespaceControl(t *testing.T) {
	t.Run("Empty nodes returns empty", func(t *testing.T) {
		nodes := []parser.Node{}
		controls := []WhitespaceControl{}

		result := ApplyWhitespaceControl(nodes, controls)

		if len(result) != 0 {
			t.Errorf("Expected 0 nodes, got %d", len(result))
		}
	})

	t.Run("Right strip removes left whitespace from following text", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.IfNode{}, // This would have right strip
			&parser.TextNode{Content: "   \t  Hello"},
		}
		controls := []WhitespaceControl{
			{RightStrip: true},
		}

		result := ApplyWhitespaceControl(nodes, controls)

		if len(result) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(result))
		}

		if textNode, ok := result[1].(*parser.TextNode); ok {
			if textNode.Content != "Hello" {
				t.Errorf("Expected 'Hello', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode")
		}
	})

	t.Run("Left strip removes right whitespace from preceding text", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello   \t  "},
			&parser.IfNode{}, // This would have left strip
		}
		controls := []WhitespaceControl{
			{LeftStrip: true},
		}

		result := ApplyWhitespaceControl(nodes, controls)

		if textNode, ok := result[0].(*parser.TextNode); ok {
			if textNode.Content != "Hello" {
				t.Errorf("Expected 'Hello', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode")
		}
	})

	t.Run("Removes empty text nodes after processing", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.TextNode{Content: "   \t  "},
			&parser.IfNode{},
		}
		controls := []WhitespaceControl{
			{LeftStrip: true},
		}

		result := ApplyWhitespaceControl(nodes, controls)

		// Should remove empty text node
		if len(result) != 1 {
			t.Errorf("Expected 1 node (empty text removed), got %d", len(result))
		}

		if _, ok := result[0].(*parser.IfNode); !ok {
			t.Error("Expected IfNode to remain")
		}
	})

	t.Run("Both strips work on same text node", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.IfNode{}, // Previous with right strip
			&parser.TextNode{Content: "   Hello   "},
			&parser.ForNode{}, // Next with left strip
		}
		controls := []WhitespaceControl{
			{RightStrip: true},
			{LeftStrip: true},
		}

		result := ApplyWhitespaceControl(nodes, controls)

		if textNode, ok := result[1].(*parser.TextNode); ok {
			if textNode.Content != "Hello" {
				t.Errorf("Expected 'Hello', got '%s'", textNode.Content)
			}
		} else {
			t.Error("Expected TextNode")
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("StripWhitespaceAroundTags processes inline control", func(t *testing.T) {
		template := "  {%- if true -%}  {{- name -}}  {%- endif -%}  "

		result := StripWhitespaceAroundTags(template)

		expected := "{% if true %}{{ name }}{% endif %}"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CompactWhitespace removes excessive whitespace", func(t *testing.T) {
		text := "  Hello    World  \n\n\n  Multiple   Spaces  \n  "

		result := CompactWhitespace(text)

		expected := "Hello World\nMultiple Spaces"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CompactWhitespace handles tabs and mixed whitespace", func(t *testing.T) {
		text := "  Hello\t\t\tWorld  \n \t \n  End  "

		result := CompactWhitespace(text)

		expected := "Hello World\nEnd"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CompactWhitespace preserves single newlines", func(t *testing.T) {
		text := "Line 1\nLine 2\nLine 3"

		result := CompactWhitespace(text)

		expected := "Line 1\nLine 2\nLine 3"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("CompactWhitespace handles empty and whitespace-only strings", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"", ""},
			{"   ", ""},
			{"  \n  \n  ", ""},
			{"\t\n\r\n  ", ""},
		}

		for _, tc := range testCases {
			result := CompactWhitespace(tc.input)
			if result != tc.expected {
				t.Errorf("For input '%s', expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})
}

// Test complex integration scenarios
func TestComplexWhitespaceScenarios(t *testing.T) {
	t.Run("Template with mixed inline and global controls", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(true, true, false)
		template := `
    {%- if users %}
      <ul>
      {% for user in users -%}
        <li>{{ user.name }}</li>
      {% endfor %}
      </ul>
    {% endif -%}
    `

		result := processor.ProcessTemplate(template)

		// Should process inline controls and apply global settings
		if strings.Contains(result, "{%-") || strings.Contains(result, "-%}") {
			t.Error("Expected inline strip modifiers to be processed")
		}

		// Should not end with newline (keepTrailingNewline = false)
		if strings.HasSuffix(result, "\n") {
			t.Error("Expected trailing newline to be removed")
		}
	})

	t.Run("Nested template structures", func(t *testing.T) {
		processor := NewAdvancedWhitespaceProcessor(false, false, true)
		template := "  {%- for i in range(3) -%}  {%- if i > 0 -%}  ,  {%- endif -%}  {{ i }}  {%- endfor -%}  "

		result := processor.processInlineWhitespaceControl(template)

		expected := "{% for i in range(3) %}{% if i > 0 %},{% endif %}{{ i }}{% endfor %}"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}
