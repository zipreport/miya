package extensions

import (
	"strings"
	"testing"
	"time"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
)

// Test SimpleTimestampExtension
func TestSimpleTimestampExtension(t *testing.T) {
	t.Run("NewSimpleTimestampExtension creates extension correctly", func(t *testing.T) {
		ext := NewSimpleTimestampExtension()

		if ext.Name() != "timestamp" {
			t.Errorf("Expected name 'timestamp', got '%s'", ext.Name())
		}

		tags := ext.Tags()
		expectedTags := []string{"now", "timestamp"}
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}

		for i, expectedTag := range expectedTags {
			if i >= len(tags) || tags[i] != expectedTag {
				t.Errorf("Expected tag '%s' at position %d, got '%s'", expectedTag, i, tags[i])
			}
		}
	})

	t.Run("ParseTag handles 'now' tag", func(t *testing.T) {
		ext := NewSimpleTimestampExtension()

		// Create mock tokens for "{% now %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "now", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 8},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 10},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'now' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'now' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("now", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		if node == nil {
			t.Fatal("Expected non-nil node")
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode, ok := node.(*ExtensionNode)
		if !ok {
			t.Fatal("Expected ExtensionNode")
		}

		before := time.Now()
		result, err := extNode.Evaluate(mockCtx)
		after := time.Now()

		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		resultStr, ok := result.(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		// Parse the result to ensure it's a valid timestamp format
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", resultStr, time.Local)
		if err != nil {
			t.Fatalf("Invalid timestamp format: %v", err)
		}

		// Check that the time is reasonable (between before and after) with a generous 2-second window
		if parsedTime.Before(before.Add(-2*time.Second)) || parsedTime.After(after.Add(2*time.Second)) {
			t.Errorf("Timestamp %v is not between %v and %v", parsedTime, before, after)
		}
	})

	t.Run("ParseTag handles 'timestamp' tag", func(t *testing.T) {
		ext := NewSimpleTimestampExtension()

		// Create mock tokens for "{% timestamp %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "timestamp", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 14},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 16},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'timestamp' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'timestamp' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("timestamp", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode := node.(*ExtensionNode)

		before := time.Now().Unix()
		result, err := extNode.Evaluate(mockCtx)
		after := time.Now().Unix()

		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		timestamp, ok := result.(int64)
		if !ok {
			t.Fatalf("Expected int64 result, got %T", result)
		}

		// Check that the timestamp is reasonable
		if timestamp < before || timestamp > after+1 {
			t.Errorf("Timestamp %d is not between %d and %d", timestamp, before, after)
		}
	})
}

// Test HelloExtension
func TestHelloExtension(t *testing.T) {
	t.Run("NewHelloExtension creates extension correctly", func(t *testing.T) {
		ext := NewHelloExtension()

		if ext.Name() != "hello" {
			t.Errorf("Expected name 'hello', got '%s'", ext.Name())
		}

		tags := ext.Tags()
		if len(tags) != 1 || tags[0] != "hello" {
			t.Errorf("Expected tags [hello], got %v", tags)
		}
	})

	t.Run("ParseTag handles 'hello' tag", func(t *testing.T) {
		ext := NewHelloExtension()

		// Create mock tokens for "{% hello %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "hello", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 10},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 12},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'hello' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'hello' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("hello", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode := node.(*ExtensionNode)

		result, err := extNode.Evaluate(mockCtx)
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if result != "Hello from extension!" {
			t.Errorf("Expected 'Hello from extension!', got %v", result)
		}
	})
}

// Test VersionExtension
func TestVersionExtension(t *testing.T) {
	t.Run("NewVersionExtension creates extension correctly", func(t *testing.T) {
		ext := NewVersionExtension("1.2.3")

		if ext.Name() != "version" {
			t.Errorf("Expected name 'version', got '%s'", ext.Name())
		}

		if ext.version != "1.2.3" {
			t.Errorf("Expected version '1.2.3', got '%s'", ext.version)
		}

		tags := ext.Tags()
		if len(tags) != 1 || tags[0] != "version" {
			t.Errorf("Expected tags [version], got %v", tags)
		}
	})

	t.Run("ParseTag handles 'version' tag", func(t *testing.T) {
		ext := NewVersionExtension("2.1.0")

		// Create mock tokens for "{% version %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "version", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 12},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 14},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'version' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'version' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("version", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode := node.(*ExtensionNode)

		result, err := extNode.Evaluate(mockCtx)
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		expected := "Version: 2.1.0"
		if result != expected {
			t.Errorf("Expected '%s', got %v", expected, result)
		}
	})
}

// Test HighlightExtension
func TestHighlightExtension(t *testing.T) {
	t.Run("NewHighlightExtension creates extension correctly", func(t *testing.T) {
		ext := NewHighlightExtension()

		if ext.Name() != "highlight" {
			t.Errorf("Expected name 'highlight', got '%s'", ext.Name())
		}

		tags := ext.Tags()
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}

		// Should be a block extension
		if !ext.IsBlockExtension("highlight") {
			t.Error("Expected highlight to be a block extension")
		}

		if ext.GetEndTag("highlight") != "endhighlight" {
			t.Errorf("Expected end tag 'endhighlight', got '%s'", ext.GetEndTag("highlight"))
		}
	})

	t.Run("ParseTag handles 'highlight' without language", func(t *testing.T) {
		ext := NewHighlightExtension()

		// Create mock tokens for "{% highlight %} content {% endhighlight %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "highlight", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 14},
			{Type: lexer.TokenText, Value: "code content", Line: 1, Column: 16},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 29},
			{Type: lexer.TokenIdentifier, Value: "endhighlight", Line: 1, Column: 32},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 45},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 47},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'highlight' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'highlight' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("highlight", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode := node.(*ExtensionNode)

		result, err := extNode.Evaluate(mockCtx)
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		resultStr, ok := result.(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		// Should contain the default language "text" and the content
		if !strings.Contains(resultStr, "highlight-text") {
			t.Errorf("Expected result to contain 'highlight-text', got: %s", resultStr)
		}

		if !strings.Contains(resultStr, "code content") {
			t.Errorf("Expected result to contain 'code content', got: %s", resultStr)
		}
	})

	t.Run("ParseTag handles 'highlight' with language", func(t *testing.T) {
		ext := NewHighlightExtension()

		// Create mock tokens for "{% highlight 'python' %} content {% endhighlight %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "highlight", Line: 1, Column: 4},
			{Type: lexer.TokenString, Value: "python", Line: 1, Column: 14},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 22},
			{Type: lexer.TokenText, Value: "print('hello')", Line: 1, Column: 24},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 39},
			{Type: lexer.TokenIdentifier, Value: "endhighlight", Line: 1, Column: 42},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 55},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 57},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'highlight' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'highlight' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("highlight", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode := node.(*ExtensionNode)

		result, err := extNode.Evaluate(mockCtx)
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		resultStr, ok := result.(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		// Should contain the specified language "python" and the content
		if !strings.Contains(resultStr, "highlight-python") {
			t.Errorf("Expected result to contain 'highlight-python', got: %s", resultStr)
		}

		if !strings.Contains(resultStr, "print('hello')") {
			t.Errorf("Expected result to contain Python code, got: %s", resultStr)
		}
	})
}

// Test CacheExtension
func TestCacheExtension(t *testing.T) {
	t.Run("NewCacheExtension creates extension correctly", func(t *testing.T) {
		ext := NewCacheExtension()

		if ext.Name() != "cache" {
			t.Errorf("Expected name 'cache', got '%s'", ext.Name())
		}

		// Should be a block extension
		if !ext.IsBlockExtension("cache") {
			t.Error("Expected cache to be a block extension")
		}

		if ext.GetEndTag("cache") != "endcache" {
			t.Errorf("Expected end tag 'endcache', got '%s'", ext.GetEndTag("cache"))
		}
	})

	t.Run("ParseTag handles 'cache' with timeout", func(t *testing.T) {
		ext := NewCacheExtension()

		// Create mock tokens for "{% cache 300 %} content {% endcache %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "cache", Line: 1, Column: 4},
			{Type: lexer.TokenInteger, Value: "300", Line: 1, Column: 10},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 14},
			{Type: lexer.TokenText, Value: "cached content", Line: 1, Column: 16},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 31},
			{Type: lexer.TokenIdentifier, Value: "endcache", Line: 1, Column: 34},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 43},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 45},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'cache' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'cache' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("cache", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}
		extNode := node.(*ExtensionNode)

		result, err := extNode.Evaluate(mockCtx)
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if result != "cached content" {
			t.Errorf("Expected 'cached content', got %v", result)
		}
	})

	t.Run("ParseTag handles 'cache' with timeout and key", func(t *testing.T) {
		ext := NewCacheExtension()

		// Create mock tokens for "{% cache 300 'my_key' %} content {% endcache %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "cache", Line: 1, Column: 4},
			{Type: lexer.TokenInteger, Value: "300", Line: 1, Column: 10},
			{Type: lexer.TokenString, Value: "my_key", Line: 1, Column: 14},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 22},
			{Type: lexer.TokenText, Value: "keyed content", Line: 1, Column: 24},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 38},
			{Type: lexer.TokenIdentifier, Value: "endcache", Line: 1, Column: 41},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 50},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 52},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		// Move to the 'cache' token
		adapter.Advance() // consume {%
		adapter.Advance() // consume 'cache' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("cache", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		// Verify the node has the expected arguments
		extNode := node.(*ExtensionNode)
		if len(extNode.Arguments) != 2 {
			t.Errorf("Expected 2 arguments, got %d", len(extNode.Arguments))
		}

		// Test evaluation
		mockCtx := &MockRuntimeContext{}

		result, err := extNode.Evaluate(mockCtx)
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if result != "keyed content" {
			t.Errorf("Expected 'keyed content', got %v", result)
		}
	})

	t.Run("ParseTag handles unknown tag", func(t *testing.T) {
		ext := NewHighlightExtension()

		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "unknown", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 12},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 14},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		adapter.Advance() // consume {%
		adapter.Advance() // consume 'unknown' tag name (simulating ExtensionAwareParser behavior)

		_, err := ext.ParseTag("unknown", adapter)
		if err == nil {
			t.Error("Expected error for unknown tag")
		}

		if !strings.Contains(err.Error(), "unknown highlight tag") {
			t.Errorf("Expected 'unknown highlight tag' in error, got: %v", err)
		}
	})
}

// Test error handling in extensions
func TestExtensionErrorHandling(t *testing.T) {
	t.Run("Evaluation with invalid context type", func(t *testing.T) {
		ext := NewHighlightExtension()

		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "highlight", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 14},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 16},
			{Type: lexer.TokenIdentifier, Value: "endhighlight", Line: 1, Column: 19},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 32},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 34},
		}

		parser := parser.NewParser(tokens)
		adapter := NewParserAdapter(parser)

		adapter.Advance() // consume {%
		adapter.Advance() // consume 'highlight' tag name (simulating ExtensionAwareParser behavior)

		node, err := ext.ParseTag("highlight", adapter)
		if err != nil {
			t.Fatalf("ParseTag failed: %v", err)
		}

		extNode := node.(*ExtensionNode)

		// Try to evaluate with wrong context type
		_, err = extNode.Evaluate("wrong context type")
		if err == nil {
			t.Error("Expected error with wrong context type")
		}

		if !strings.Contains(err.Error(), "invalid context type") {
			t.Errorf("Expected 'invalid context type' in error, got: %v", err)
		}
	})
}
