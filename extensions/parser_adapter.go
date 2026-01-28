package extensions

import (
	"fmt"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
)

// ParserAdapter adapts the main parser for use by extensions
type ParserAdapter struct {
	parser *parser.Parser
}

// NewParserAdapter creates a new parser adapter
func NewParserAdapter(p *parser.Parser) *ParserAdapter {
	return &ParserAdapter{parser: p}
}

// Current returns the current token
func (pa *ParserAdapter) Current() *lexer.Token {
	return pa.parser.Peek()
}

// Advance moves to the next token and returns it
func (pa *ParserAdapter) Advance() *lexer.Token {
	return pa.parser.Advance()
}

// Check returns true if the current token matches the given type
func (pa *ParserAdapter) Check(tokenType lexer.TokenType) bool {
	return pa.parser.Check(tokenType)
}

// CheckAny returns true if the current token matches any of the given types
func (pa *ParserAdapter) CheckAny(types ...lexer.TokenType) bool {
	return pa.parser.CheckAny(types...)
}

// Peek returns the current token without advancing
func (pa *ParserAdapter) Peek() *lexer.Token {
	return pa.parser.Peek()
}

// IsAtEnd returns true if we've reached the end of tokens
func (pa *ParserAdapter) IsAtEnd() bool {
	return pa.parser.IsAtEnd()
}

// ParseExpression parses an expression and returns the node
func (pa *ParserAdapter) ParseExpression() (parser.ExpressionNode, error) {
	// We'll need to expose this method in the parser
	return pa.parser.ParseExpressionPublic()
}

// ParseToEnd parses all tokens until the specified end tag
func (pa *ParserAdapter) ParseToEnd(endTag string) ([]parser.Node, error) {
	var nodes []parser.Node

	for !pa.IsAtEnd() {
		// Check if we're at the end tag
		if pa.Check(lexer.TokenBlockStart) || pa.Check(lexer.TokenBlockStartTrim) {
			if pa.parser.PeekBlockTypePublic() == lexer.LookupKeyword(endTag) {
				break
			}
		}

		// Parse the next node
		node, err := pa.parser.ParseTopLevelPublic()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// ParseBlock parses a block extension body until the end tag
func (pa *ParserAdapter) ParseBlock(endTag string) ([]parser.Node, error) {
	var nodes []parser.Node

	for !pa.IsAtEnd() {
		// Check if we're at the end tag
		if pa.Check(lexer.TokenBlockStart) || pa.Check(lexer.TokenBlockStartTrim) {
			// Save current position to peek ahead
			currentPos := pa.parser.GetCurrentPosition()

			// Advance to check the tag name
			pa.Advance() // consume {% or {%-

			if pa.Check(lexer.TokenIdentifier) && pa.Peek().Value == endTag {
				// We found our end tag, reset position and break
				pa.parser.SetCurrentPosition(currentPos)
				break
			} else {
				// Not our end tag, reset position and continue parsing
				pa.parser.SetCurrentPosition(currentPos)
			}
		}

		// Parse the next node
		node, err := pa.parser.ParseTopLevelPublic()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// Error creates a parser error with the given message
func (pa *ParserAdapter) Error(message string) error {
	return pa.parser.ErrorPublic(message)
}

// ExpectBlockEnd ensures the current token is a block end (%})
func (pa *ParserAdapter) ExpectBlockEnd() error {
	if !pa.Check(lexer.TokenBlockEnd) && !pa.Check(lexer.TokenBlockEndTrim) {
		return pa.Error("expected '%%}' after tag")
	}
	pa.Advance()
	return nil
}

// ExpectEndTag consumes the end tag block
func (pa *ParserAdapter) ExpectEndTag(endTag string) error {
	if !pa.Check(lexer.TokenBlockStart) && !pa.Check(lexer.TokenBlockStartTrim) {
		return pa.Error(fmt.Sprintf("expected '{%%%% %s %%%%}' to close tag", endTag))
	}
	pa.Advance() // consume {%

	expectedToken := lexer.LookupKeyword(endTag)
	if !pa.Check(expectedToken) {
		return pa.Error(fmt.Sprintf("expected '%s'", endTag))
	}
	pa.Advance() // consume end tag

	return pa.ExpectBlockEnd()
}

// NewExtensionNode creates a new extension node
func (pa *ParserAdapter) NewExtensionNode(extensionName, tagName string, line, column int) *parser.ExtensionNode {
	return parser.NewExtensionNode(extensionName, tagName, line, column)
}

// Helper method to parse arguments until block end
func (pa *ParserAdapter) ParseArguments() ([]parser.ExpressionNode, error) {
	var args []parser.ExpressionNode

	for !pa.Check(lexer.TokenBlockEnd) && !pa.Check(lexer.TokenBlockEndTrim) && !pa.IsAtEnd() {
		arg, err := pa.ParseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Skip optional comma
		if pa.Check(lexer.TokenComma) {
			pa.Advance()
		}
	}

	return args, nil
}

// ExtensionAwareParser extends the main parser with extension support
type ExtensionAwareParser struct {
	*parser.Parser
	registry *Registry
}

// NewExtensionAwareParser creates a parser with extension support
func NewExtensionAwareParser(tokens []*lexer.Token, registry *Registry) *ExtensionAwareParser {
	return &ExtensionAwareParser{
		Parser:   parser.NewParser(tokens),
		registry: registry,
	}
}

// Parse parses the tokens into a template AST with extension support
func (eap *ExtensionAwareParser) Parse() (*parser.TemplateNode, error) {
	template := parser.NewTemplateNode("", 1, 1)

	for !eap.IsAtEnd() {
		if eap.Check(lexer.TokenEOF) {
			break
		}

		node, err := eap.ParseTopLevelPublic()
		if err != nil {
			return nil, err
		}

		if node != nil {
			template.Children = append(template.Children, node)
		}
	}

	return template, nil
}

// ParseTopLevelPublic overrides the parent method to handle custom tags
func (eap *ExtensionAwareParser) ParseTopLevelPublic() (parser.Node, error) {
	switch eap.Peek().Type {
	case lexer.TokenBlockStart, lexer.TokenBlockStartTrim:
		return eap.parseBlockStatementWithExtensions()
	default:
		// For non-block statements, use the parent implementation
		return eap.Parser.ParseTopLevelPublic()
	}
}

// parseBlockStatementWithExtensions handles block statements with extension support
func (eap *ExtensionAwareParser) parseBlockStatementWithExtensions() (parser.Node, error) {
	// Save current position before consuming the block start
	startPos := eap.GetCurrentPosition()

	eap.Advance() // consume {% or {%-

	if eap.IsAtEnd() {
		return nil, eap.ErrorPublic("unexpected end of input in block statement")
	}

	currentToken := eap.Peek()
	if currentToken.Type == lexer.TokenIdentifier {
		// Check if this is a custom tag
		if ext, ok := eap.registry.GetExtensionForTag(currentToken.Value); ok {
			return eap.parseCustomTag(ext, currentToken.Value)
		}
	}

	// For standard tags, reset position and delegate to parent parser
	eap.SetCurrentPosition(startPos)

	// Delegate directly to the parent parser's parseBlockStatement method
	return eap.Parser.ParseBlockStatementPublic()
}

// parseCustomTag handles parsing of custom tags
func (eap *ExtensionAwareParser) parseCustomTag(ext Extension, tagName string) (parser.Node, error) {
	eap.Advance() // consume the tag name

	// Create parser adapter
	adapter := NewParserAdapter(eap.Parser)

	// Let the extension parse the tag
	node, err := ext.ParseTag(tagName, adapter)
	if err != nil {
		// Get current position for error context
		current := adapter.Current()
		line, column := 1, 1
		if current != nil {
			line, column = current.Line, current.Column
		}
		return nil, NewExtensionParseError(ext.Name(), tagName, "", line, column, "failed to parse tag", err)
	}

	// If this is a block extension, we need to consume the end tag
	if ext.IsBlockExtension(tagName) {
		endTag := ext.GetEndTag(tagName)
		if endTag != "" {
			err = adapter.ExpectEndTag(endTag)
			if err != nil {
				// Get current position for error context
				current := adapter.Current()
				line, column := 1, 1
				if current != nil {
					line, column = current.Line, current.Column
				}
				return nil, NewExtensionParseError(ext.Name(), tagName, "", line, column, fmt.Sprintf("failed to parse end tag '%s'", endTag), err)
			}
		}
	}

	return node, nil
}

// Simple helper function to create tokens for testing
func CreateTokensFromString(input string) ([]*lexer.Token, error) {
	lexerInstance := lexer.NewLexer(input, lexer.DefaultConfig())
	var tokens []*lexer.Token

	for {
		token, err := lexerInstance.NextToken()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)

		if token.Type == lexer.TokenEOF {
			break
		}
	}

	return tokens, nil
}
