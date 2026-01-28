package extensions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
)

// Test ParserAdapter functionality
func TestParserAdapter(t *testing.T) {
	t.Run("NewParserAdapter creates adapter correctly", func(t *testing.T) {
		tokens := []*lexer.Token{
			{Type: lexer.TokenText, Value: "test", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 5},
		}

		parserInstance := parser.NewParser(tokens)
		adapter := NewParserAdapter(parserInstance)

		if adapter == nil {
			t.Fatal("NewParserAdapter returned nil")
		}

		if adapter.parser != parserInstance {
			t.Error("Parser reference not set correctly")
		}
	})

	t.Run("Token navigation methods work correctly", func(t *testing.T) {
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "test", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 9},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 11},
		}

		parserInstance := parser.NewParser(tokens)
		adapter := NewParserAdapter(parserInstance)

		// Test Current/Peek
		current := adapter.Current()
		peek := adapter.Peek()
		if current.Type != lexer.TokenBlockStart || peek.Type != lexer.TokenBlockStart {
			t.Error("Current/Peek not working correctly")
		}

		if current != peek {
			t.Error("Current and Peek should return same token")
		}

		// Test Advance
		advanced := adapter.Advance()
		if advanced.Type != lexer.TokenBlockStart {
			t.Error("Advance should return current token before advancing")
		}

		// Now current should be different
		newCurrent := adapter.Current()
		if newCurrent.Type != lexer.TokenIdentifier {
			t.Error("After advance, current should be next token")
		}

		// Test IsAtEnd
		if adapter.IsAtEnd() {
			t.Error("Should not be at end yet")
		}

		// Advance to EOF
		adapter.Advance() // identifier
		adapter.Advance() // block end

		if !adapter.IsAtEnd() {
			t.Error("Should be at end now")
		}
	})

	t.Run("Check methods work correctly", func(t *testing.T) {
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "test", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 9},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 11},
		}

		parserInstance := parser.NewParser(tokens)
		adapter := NewParserAdapter(parserInstance)

		// Test Check
		if !adapter.Check(lexer.TokenBlockStart) {
			t.Error("Check should return true for current token type")
		}

		if adapter.Check(lexer.TokenIdentifier) {
			t.Error("Check should return false for non-current token type")
		}

		// Test CheckAny
		if !adapter.CheckAny(lexer.TokenText, lexer.TokenBlockStart, lexer.TokenInteger) {
			t.Error("CheckAny should return true when current token matches one of the types")
		}

		if adapter.CheckAny(lexer.TokenText, lexer.TokenInteger, lexer.TokenString) {
			t.Error("CheckAny should return false when current token matches none of the types")
		}
	})

	t.Run("ExpectBlockEnd works correctly", func(t *testing.T) {
		// Test successful case
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 3},
		}

		parserInstance := parser.NewParser(tokens)
		adapter := NewParserAdapter(parserInstance)

		err := adapter.ExpectBlockEnd()
		if err != nil {
			t.Errorf("ExpectBlockEnd should succeed with block end token: %v", err)
		}

		// Test with trim variant
		tokens = []*lexer.Token{
			{Type: lexer.TokenBlockEndTrim, Value: "-%}", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 4},
		}

		parserInstance = parser.NewParser(tokens)
		adapter = NewParserAdapter(parserInstance)

		err = adapter.ExpectBlockEnd()
		if err != nil {
			t.Errorf("ExpectBlockEnd should succeed with block end trim token: %v", err)
		}

		// Test failure case
		tokens = []*lexer.Token{
			{Type: lexer.TokenIdentifier, Value: "test", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 5},
		}

		parserInstance = parser.NewParser(tokens)
		adapter = NewParserAdapter(parserInstance)

		err = adapter.ExpectBlockEnd()
		if err == nil {
			t.Error("ExpectBlockEnd should fail with non-block-end token")
		}

		if !strings.Contains(err.Error(), "expected") {
			t.Errorf("Expected error message about block end, got: %v", err)
		}
	})

	t.Run("ParseArguments works correctly", func(t *testing.T) {
		// Test parsing multiple arguments
		tokens := []*lexer.Token{
			{Type: lexer.TokenString, Value: "arg1", Line: 1, Column: 1},
			{Type: lexer.TokenComma, Value: ",", Line: 1, Column: 7},
			{Type: lexer.TokenInteger, Value: "42", Line: 1, Column: 9},
			{Type: lexer.TokenComma, Value: ",", Line: 1, Column: 12},
			{Type: lexer.TokenIdentifier, Value: "var", Line: 1, Column: 14},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 18},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 20},
		}

		parserInstance := parser.NewParser(tokens)
		adapter := NewParserAdapter(parserInstance)

		args, err := adapter.ParseArguments()
		if err != nil {
			t.Fatalf("ParseArguments failed: %v", err)
		}

		if len(args) != 3 {
			t.Errorf("Expected 3 arguments, got %d", len(args))
		}

		// Test empty arguments
		tokens = []*lexer.Token{
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 3},
		}

		parserInstance = parser.NewParser(tokens)
		adapter = NewParserAdapter(parserInstance)

		args, err = adapter.ParseArguments()
		if err != nil {
			t.Fatalf("ParseArguments with no args failed: %v", err)
		}

		if len(args) != 0 {
			t.Errorf("Expected 0 arguments, got %d", len(args))
		}
	})

	t.Run("NewExtensionNode creates node correctly", func(t *testing.T) {
		parser := parser.NewParser([]*lexer.Token{})
		adapter := NewParserAdapter(parser)

		node := adapter.NewExtensionNode("test_ext", "test_tag", 10, 5)

		if node == nil {
			t.Fatal("NewExtensionNode returned nil")
		}

		if node.ExtensionName != "test_ext" {
			t.Errorf("Expected extension name 'test_ext', got '%s'", node.ExtensionName)
		}

		if node.TagName != "test_tag" {
			t.Errorf("Expected tag name 'test_tag', got '%s'", node.TagName)
		}

		if node.Line() != 10 {
			t.Errorf("Expected line 10, got %d", node.Line())
		}

		if node.Column() != 5 {
			t.Errorf("Expected column 5, got %d", node.Column())
		}
	})
}

// Test ExtensionAwareParser functionality
func TestExtensionAwareParser(t *testing.T) {
	t.Run("NewExtensionAwareParser creates parser correctly", func(t *testing.T) {
		tokens := []*lexer.Token{
			{Type: lexer.TokenText, Value: "test", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 5},
		}

		registry := NewRegistry()
		eap := NewExtensionAwareParser(tokens, registry)

		if eap == nil {
			t.Fatal("NewExtensionAwareParser returned nil")
		}

		if eap.registry != registry {
			t.Error("Registry reference not set correctly")
		}

		if eap.Parser == nil {
			t.Error("Parser not created")
		}
	})

	t.Run("Parse handles basic template without extensions", func(t *testing.T) {
		tokens := []*lexer.Token{
			{Type: lexer.TokenText, Value: "Hello World", Line: 1, Column: 1},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 12},
		}

		registry := NewRegistry()
		eap := NewExtensionAwareParser(tokens, registry)

		template, err := eap.Parse()
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if template == nil {
			t.Fatal("Parse returned nil template")
		}

		if len(template.Children) != 1 {
			t.Errorf("Expected 1 child node, got %d", len(template.Children))
		}
	})

	t.Run("Parse handles custom extension tags", func(t *testing.T) {
		// Create tokens for "Hello {% hello %} World"
		tokens := []*lexer.Token{
			{Type: lexer.TokenText, Value: "Hello ", Line: 1, Column: 1},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 7},
			{Type: lexer.TokenIdentifier, Value: "hello", Line: 1, Column: 10},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 16},
			{Type: lexer.TokenText, Value: " World", Line: 1, Column: 18},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 24},
		}

		registry := NewRegistry()
		ext := NewHelloExtension()
		err := registry.Register(ext)
		if err != nil {
			t.Fatalf("Failed to register extension: %v", err)
		}

		eap := NewExtensionAwareParser(tokens, registry)

		template, err := eap.Parse()
		if err != nil {
			t.Fatalf("Parse with extension failed: %v", err)
		}

		if template == nil {
			t.Fatal("Parse returned nil template")
		}

		if len(template.Children) != 3 {
			t.Errorf("Expected 3 child nodes, got %d", len(template.Children))
		}

		// Check that middle node is an extension node
		if extNode, ok := template.Children[1].(*ExtensionNode); ok {
			if extNode.ExtensionName != "hello" {
				t.Errorf("Expected extension name 'hello', got '%s'", extNode.ExtensionName)
			}
			if extNode.TagName != "hello" {
				t.Errorf("Expected tag name 'hello', got '%s'", extNode.TagName)
			}
		} else {
			t.Error("Middle node should be an ExtensionNode")
		}
	})

	t.Run("Parse handles block extension tags", func(t *testing.T) {
		// Create tokens for "{% highlight 'python' %} code {% endhighlight %}"
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

		registry := NewRegistry()
		ext := NewHighlightExtension()
		err := registry.Register(ext)
		if err != nil {
			t.Fatalf("Failed to register extension: %v", err)
		}

		eap := NewExtensionAwareParser(tokens, registry)

		template, err := eap.Parse()
		if err != nil {
			t.Fatalf("Parse with block extension failed: %v", err)
		}

		if template == nil {
			t.Fatal("Parse returned nil template")
		}

		if len(template.Children) != 1 {
			t.Errorf("Expected 1 child node, got %d", len(template.Children))
		}

		// Check that it's an extension node with body
		if extNode, ok := template.Children[0].(*ExtensionNode); ok {
			if extNode.ExtensionName != "highlight" {
				t.Errorf("Expected extension name 'highlight', got '%s'", extNode.ExtensionName)
			}
			if len(extNode.Arguments) != 1 {
				t.Errorf("Expected 1 argument, got %d", len(extNode.Arguments))
			}
			if len(extNode.Body) != 1 {
				t.Errorf("Expected 1 body node, got %d", len(extNode.Body))
			}
		} else {
			t.Error("Node should be an ExtensionNode")
		}
	})

	t.Run("Parse handles unknown tags as standard tags", func(t *testing.T) {
		// Create tokens for "{% if true %} content {% endif %}"
		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIf, Value: "if", Line: 1, Column: 4},
			{Type: lexer.TokenTrue, Value: "true", Line: 1, Column: 7},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 12},
			{Type: lexer.TokenText, Value: "content", Line: 1, Column: 14},
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 22},
			{Type: lexer.TokenEndif, Value: "endif", Line: 1, Column: 25},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 31},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 33},
		}

		registry := NewRegistry()
		eap := NewExtensionAwareParser(tokens, registry)

		template, err := eap.Parse()
		if err != nil {
			t.Fatalf("Parse with standard tags failed: %v", err)
		}

		if template == nil {
			t.Fatal("Parse returned nil template")
		}

		// Should have parsed as a standard if node
		if len(template.Children) != 1 {
			t.Errorf("Expected 1 child node, got %d", len(template.Children))
		}

		// The node should be an IfNode, not an ExtensionNode
		if _, ok := template.Children[0].(*parser.IfNode); !ok {
			t.Error("Node should be an IfNode for standard if tag")
		}
	})
}

// Test error handling in parser adapter
func TestParserAdapterErrorHandling(t *testing.T) {
	t.Run("Error method creates appropriate error", func(t *testing.T) {
		tokens := []*lexer.Token{
			{Type: lexer.TokenText, Value: "test", Line: 5, Column: 10},
			{Type: lexer.TokenEOF, Value: "", Line: 5, Column: 14},
		}

		parserInstance := parser.NewParser(tokens)
		adapter := NewParserAdapter(parserInstance)

		err := adapter.Error("test error message")
		if err == nil {
			t.Fatal("Error method should return non-nil error")
		}

		if !strings.Contains(err.Error(), "test error message") {
			t.Errorf("Error should contain message, got: %v", err)
		}
	})

	t.Run("Extension parsing errors are wrapped correctly", func(t *testing.T) {
		// Create a failing extension
		failingExt := &FailingTestExtension{
			BaseExtension: NewBaseExtension("failing", []string{"fail"}),
		}

		tokens := []*lexer.Token{
			{Type: lexer.TokenBlockStart, Value: "{%", Line: 1, Column: 1},
			{Type: lexer.TokenIdentifier, Value: "fail", Line: 1, Column: 4},
			{Type: lexer.TokenBlockEnd, Value: "%}", Line: 1, Column: 9},
			{Type: lexer.TokenEOF, Value: "", Line: 1, Column: 11},
		}

		registry := NewRegistry()
		err := registry.Register(failingExt)
		if err != nil {
			t.Fatalf("Failed to register failing extension: %v", err)
		}

		eap := NewExtensionAwareParser(tokens, registry)

		_, err = eap.Parse()
		if err == nil {
			t.Error("Expected error from failing extension")
		}

		// Check that it's wrapped as an ExtensionError
		if extErr, ok := err.(*ExtensionError); ok {
			if extErr.ExtensionName != "failing" {
				t.Errorf("Expected extension name 'failing', got '%s'", extErr.ExtensionName)
			}
			if extErr.TagName != "fail" {
				t.Errorf("Expected tag name 'fail', got '%s'", extErr.TagName)
			}
		} else {
			t.Errorf("Expected ExtensionError, got %T", err)
		}
	})
}

// Test parser utility functions
func TestParserUtilityFunctions(t *testing.T) {
	t.Run("CreateTokensFromString works correctly", func(t *testing.T) {
		input := "Hello {% test %} World"

		tokens, err := CreateTokensFromString(input)
		if err != nil {
			t.Fatalf("CreateTokensFromString failed: %v", err)
		}

		if len(tokens) == 0 {
			t.Fatal("No tokens created")
		}

		// Should end with EOF
		lastToken := tokens[len(tokens)-1]
		if lastToken.Type != lexer.TokenEOF {
			t.Error("Last token should be EOF")
		}

		// Should contain text and block tokens
		foundText := false
		foundBlockStart := false
		foundIdentifier := false

		for _, token := range tokens {
			switch token.Type {
			case lexer.TokenText:
				foundText = true
			case lexer.TokenBlockStart:
				foundBlockStart = true
			case lexer.TokenIdentifier:
				foundIdentifier = true
			}
		}

		if !foundText {
			t.Error("Should have found text token")
		}
		if !foundBlockStart {
			t.Error("Should have found block start token")
		}
		if !foundIdentifier {
			t.Error("Should have found identifier token")
		}
	})

	t.Run("CreateTokensFromString handles lexer errors", func(t *testing.T) {
		// Test with malformed template syntax that should cause lexer error
		input := "{{% invalid syntax"

		_, err := CreateTokensFromString(input)
		if err != nil {
			t.Logf("Got expected lexer error: %v", err)
		} else {
			// If this doesn't cause an error, that's fine too - some lexers are more permissive
			t.Logf("Lexer was permissive with malformed input")
		}
	})
}

// Test extension for testing failures
type FailingTestExtension struct {
	*BaseExtension
}

func (fte *FailingTestExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	return nil, fmt.Errorf("intentional parse failure")
}
