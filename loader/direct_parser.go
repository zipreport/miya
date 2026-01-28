package loader

import (
	"fmt"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
)

// DirectTemplateParser parses templates without Environment coupling
// This eliminates circular dependencies in template loading
type DirectTemplateParser struct {
	lexerConfig *lexer.LexerConfig
}

// NewDirectTemplateParser creates a parser that works independently of Environment
func NewDirectTemplateParser() *DirectTemplateParser {
	return &DirectTemplateParser{
		lexerConfig: &lexer.LexerConfig{
			// Use standard Jinja2 delimiters
			VarStartString:     "{{",
			VarEndString:       "}}",
			BlockStartString:   "{%",
			BlockEndString:     "%}",
			CommentStartString: "{#",
			CommentEndString:   "#}",
			TrimBlocks:         false,
			LstripBlocks:       false,
		},
	}
}

// NewDirectTemplateParserWithConfig creates a parser with custom lexer configuration
func NewDirectTemplateParserWithConfig(config *lexer.LexerConfig) *DirectTemplateParser {
	return &DirectTemplateParser{
		lexerConfig: config,
	}
}

// ParseTemplate parses template content directly without Environment involvement
func (p *DirectTemplateParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	// Create lexer with configuration
	l := lexer.NewLexer(content, p.lexerConfig)

	// Tokenize the source
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error in template %s: %v", name, err)
	}

	// Parse tokens into AST
	parserInstance := parser.NewParser(tokens)
	ast, err := parserInstance.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser error in template %s: %v", name, err)
	}

	// Set template name in AST
	ast.Name = name

	return ast, nil
}

// SetLexerConfig updates the lexer configuration
func (p *DirectTemplateParser) SetLexerConfig(config *lexer.LexerConfig) {
	p.lexerConfig = config
}

// GetLexerConfig returns the current lexer configuration
func (p *DirectTemplateParser) GetLexerConfig() *lexer.LexerConfig {
	return p.lexerConfig
}
