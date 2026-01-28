package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zipreport/miya/lexer"
)

// Parser parses tokens into an AST
type Parser struct {
	tokens  []*lexer.Token
	current int
	errors  []string
}

// NewParser creates a new parser with the given tokens
func NewParser(tokens []*lexer.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
		errors:  make([]string, 0),
	}
}

// Parse parses the tokens into a template AST
func (p *Parser) Parse() (*TemplateNode, error) {
	template := NewTemplateNode("", 1, 1)

	for !p.isAtEnd() {
		if p.check(lexer.TokenEOF) {
			break
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}

		if node != nil {
			template.Children = append(template.Children, node)
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(p.errors, "; "))
	}

	return template, nil
}

// parseTopLevel parses top-level template content
func (p *Parser) parseTopLevel() (Node, error) {
	switch p.peek().Type {
	case lexer.TokenText:
		return p.parseText()
	case lexer.TokenVarStart, lexer.TokenVarStartTrim:
		return p.parseVariable()
	case lexer.TokenBlockStart, lexer.TokenBlockStartTrim:
		return p.parseBlockStatement()
	case lexer.TokenCommentStart:
		return p.parseComment()
	default:
		return nil, p.error(fmt.Sprintf("unexpected token: %s", p.peek().Type))
	}
}

// parseText parses plain text content
func (p *Parser) parseText() (Node, error) {
	token := p.advance()
	return NewTextNode(token.Value, token.Line, token.Column), nil
}

// parseVariable parses variable expressions {{ ... }}
func (p *Parser) parseVariable() (Node, error) {
	startToken := p.advance() // consume {{ or {{-

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect closing }}
	if !p.check(lexer.TokenVarEnd) && !p.check(lexer.TokenVarEndTrim) {
		return nil, p.error("expected '}}' after variable expression")
	}
	p.advance()

	return NewVariableNode(expr, startToken.Line, startToken.Column), nil
}

// parseBlockStatement parses block statements {% ... %}
func (p *Parser) parseBlockStatement() (Node, error) {
	p.advance() // consume {% or {%-

	if p.isAtEnd() {
		return nil, p.error("unexpected end of input in block statement")
	}

	switch p.peek().Type {
	case lexer.TokenIf:
		return p.parseIfStatement()
	case lexer.TokenFor:
		return p.parseForStatement()
	case lexer.TokenSet:
		return p.parseSetStatement()
	case lexer.TokenBlock:
		return p.parseBlockDefinition()
	case lexer.TokenExtends:
		return p.parseExtendsStatement()
	case lexer.TokenInclude:
		return p.parseIncludeStatement()
	case lexer.TokenMacro:
		return p.parseMacroDefinition()
	case lexer.TokenCall:
		return p.parseCallBlockStatement()
	case lexer.TokenRaw:
		return p.parseRawBlock()
	case lexer.TokenAutoescape:
		return p.parseAutoescapeBlock()
	case lexer.TokenBreak:
		return p.parseBreakStatement()
	case lexer.TokenContinue:
		return p.parseContinueStatement()
	case lexer.TokenImport:
		return p.parseImportStatement()
	case lexer.TokenFrom:
		return p.parseFromStatement()
	case lexer.TokenWith:
		return p.parseWithStatement()
	case lexer.TokenDo:
		return p.parseDoStatement()
	case lexer.TokenFilter:
		return p.parseFilterBlock()
	default:
		return nil, p.error(fmt.Sprintf("unexpected block statement: %s", p.peek().Type))
	}
}

// parseComment parses comments {# ... #}
func (p *Parser) parseComment() (Node, error) {
	startToken := p.advance() // consume {#

	// Comments are handled by the lexer, this shouldn't be called
	// But if we get here, skip to the end
	for !p.check(lexer.TokenCommentEnd) && !p.isAtEnd() {
		p.advance()
	}

	if p.check(lexer.TokenCommentEnd) {
		p.advance()
	}

	return NewCommentNode("", startToken.Line, startToken.Column), nil
}

// parseIfStatement parses if/elif/else statements
func (p *Parser) parseIfStatement() (Node, error) {
	ifToken := p.advance() // consume 'if'

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after if condition")
	}
	p.advance()

	ifNode := NewIfNode(condition, ifToken.Line, ifToken.Column)

	// Parse if body
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			// Look ahead to see if this is elif, else, or endif
			if p.peekBlockType() == lexer.TokenElif {
				break
			}
			if p.peekBlockType() == lexer.TokenElse {
				break
			}
			if p.peekBlockType() == lexer.TokenEndif {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			ifNode.Body = append(ifNode.Body, node)
		}
	}

	// Parse elif and else clauses
	for p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
		blockType := p.peekBlockType()
		if blockType == lexer.TokenElif {
			p.advance() // consume {%
			p.advance() // consume elif

			elifCondition, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
				return nil, p.error("expected '%}' after elif condition")
			}
			p.advance()

			elifNode := NewIfNode(elifCondition, p.previous().Line, p.previous().Column)

			// Parse elif body
			for !p.isAtEnd() {
				if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
					nextType := p.peekBlockType()
					if nextType == lexer.TokenElif || nextType == lexer.TokenElse || nextType == lexer.TokenEndif {
						break
					}
				}

				node, err := p.parseTopLevel()
				if err != nil {
					return nil, err
				}
				if node != nil {
					elifNode.Body = append(elifNode.Body, node)
				}
			}

			ifNode.ElseIfs = append(ifNode.ElseIfs, elifNode)

		} else if blockType == lexer.TokenElse {
			p.advance() // consume {%
			p.advance() // consume else

			if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
				return nil, p.error("expected '%}' after else")
			}
			p.advance()

			// Parse else body
			for !p.isAtEnd() {
				if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
					if p.peekBlockType() == lexer.TokenEndif {
						break
					}
				}

				node, err := p.parseTopLevel()
				if err != nil {
					return nil, err
				}
				if node != nil {
					ifNode.Else = append(ifNode.Else, node)
				}
			}
			break

		} else {
			break
		}
	}

	// Expect endif
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endif %}' to close if statement")
	}
	p.advance() // consume {%

	if !p.check(lexer.TokenEndif) {
		return nil, p.error("expected 'endif'")
	}
	p.advance() // consume endif

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endif")
	}
	p.advance()

	return ifNode, nil
}

// parseForStatement parses for loops
func (p *Parser) parseForStatement() (Node, error) {
	forToken := p.advance() // consume 'for'

	// Parse variable list (support multiple variables for unpacking)
	var variables []string

	if !p.check(lexer.TokenIdentifier) {
		return nil, p.error("expected variable name after 'for'")
	}

	variables = append(variables, p.advance().Value)

	// Handle multiple variables separated by commas
	for p.check(lexer.TokenComma) {
		p.advance() // consume ','
		if !p.check(lexer.TokenIdentifier) {
			return nil, p.error("expected variable name after comma")
		}
		variables = append(variables, p.advance().Value)
	}

	if !p.check(lexer.TokenIn) {
		return nil, p.error("expected 'in' after for variable(s)")
	}
	p.advance() // consume 'in'

	iterable, err := p.parseOr() // Use parseOr to avoid consuming the 'if' token
	if err != nil {
		return nil, err
	}

	// Check for optional conditional (if condition)
	var condition ExpressionNode
	if p.check(lexer.TokenIf) {
		p.advance()                  // consume 'if'
		condition, err = p.parseOr() // Use parseOr instead of parseExpression to avoid conditional parsing
		if err != nil {
			return nil, err
		}
	}

	// Check for recursive keyword
	var recursive bool
	if p.check(lexer.TokenRecursive) {
		p.advance() // consume 'recursive'
		recursive = true
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after for statement")
	}
	p.advance()

	forNode := NewForNode(variables, iterable, forToken.Line, forToken.Column)
	forNode.Condition = condition
	forNode.Recursive = recursive

	// Parse for body
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			blockType := p.peekBlockType()
			if blockType == lexer.TokenElse || blockType == lexer.TokenEndfor {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			forNode.Body = append(forNode.Body, node)
		}
	}

	// Check for else clause
	if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
		if p.peekBlockType() == lexer.TokenElse {
			p.advance() // consume {%
			p.advance() // consume else

			if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
				return nil, p.error("expected '%}' after else")
			}
			p.advance()

			// Parse else body
			for !p.isAtEnd() {
				if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
					if p.peekBlockType() == lexer.TokenEndfor {
						break
					}
				}

				node, err := p.parseTopLevel()
				if err != nil {
					return nil, err
				}
				if node != nil {
					forNode.Else = append(forNode.Else, node)
				}
			}
		}
	}

	// Expect endfor
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endfor %}' to close for statement")
	}
	p.advance() // consume {%

	if !p.check(lexer.TokenEndfor) {
		return nil, p.error("expected 'endfor'")
	}
	p.advance() // consume endfor

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endfor")
	}
	p.advance()

	return forNode, nil
}

// parseSetStatement parses set statements (supports multiple assignment and block assignment)
func (p *Parser) parseSetStatement() (Node, error) {
	setToken := p.advance() // consume 'set'

	// Parse target expression(s) - can be identifiers or attribute access
	var targets []ExpressionNode

	// Parse first target (required)
	firstTarget, err := p.parsePostfix()
	if err != nil {
		return nil, p.error("syntax error: expected assignment target after 'set'")
	}
	targets = append(targets, firstTarget)

	// Check for multiple assignment (comma-separated targets)
	for p.check(lexer.TokenComma) {
		p.advance() // consume ','
		target, err := p.parsePostfix()
		if err != nil {
			return nil, p.error("expected assignment target after comma")
		}
		targets = append(targets, target)
	}

	// Check for assignment operator or block syntax
	if p.check(lexer.TokenAssign) {
		// Regular assignment: {% set var = value %} or {% set a.b = value %}
		p.advance() // consume '='

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
			return nil, p.error("expected '%}' after set statement")
		}
		p.advance()

		return NewSetNodeWithTargets(targets, value, setToken.Line, setToken.Column), nil
	} else if p.check(lexer.TokenBlockEnd) || p.check(lexer.TokenBlockEndTrim) {
		// Block assignment: {% set var %}content{% endset %}
		if len(targets) > 1 {
			return nil, p.error("block assignment does not support multiple variables")
		}

		// Block assignment only supports simple identifiers
		if _, ok := targets[0].(*IdentifierNode); !ok {
			return nil, p.error("block assignment only supports simple variable names")
		}

		p.advance() // consume '%}'

		// Parse the body until {% endset %}
		var body []Node
		for !p.isAtEnd() {
			if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
				// Look ahead to see if this is {% endset %}
				if p.peekBlockType() == lexer.TokenEndSet {
					break
				}
			}

			node, err := p.parseTopLevel()
			if err != nil {
				return nil, err
			}
			body = append(body, node)
		}

		// Expect {% endset %}
		if p.isAtEnd() {
			return nil, p.error("expected '{% endset %}' to close set block")
		}

		p.advance() // consume '{%'
		if !p.check(lexer.TokenEndSet) {
			return nil, p.error("expected 'endset' to close set block")
		}
		p.advance() // consume 'endset'

		if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
			return nil, p.error("expected '%}' after endset")
		}
		p.advance() // consume '%}'

		// Extract the variable name from the identifier node
		varName := targets[0].(*IdentifierNode).Name
		return NewBlockSetNode(varName, body, setToken.Line, setToken.Column), nil
	} else {
		return nil, p.error("expected '=' or '%}' after set target(s)")
	}
}

// parseBlockDefinition parses block definitions
func (p *Parser) parseBlockDefinition() (Node, error) {
	blockToken := p.advance() // consume 'block'

	if !p.check(lexer.TokenIdentifier) {
		return nil, p.error("expected block name after 'block'")
	}
	blockName := p.advance().Value

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after block name")
	}
	p.advance()

	blockNode := NewBlockNode(blockName, blockToken.Line, blockToken.Column)

	// Parse block body
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			if p.peekBlockType() == lexer.TokenEndblock {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			blockNode.Body = append(blockNode.Body, node)
		}
	}

	// Expect endblock
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endblock %}' to close block")
	}
	p.advance() // consume {%

	if !p.check(lexer.TokenEndblock) {
		return nil, p.error("expected 'endblock'")
	}
	p.advance() // consume endblock

	// Optional block name
	if p.check(lexer.TokenIdentifier) {
		name := p.advance().Value
		if name != blockName {
			return nil, p.error(fmt.Sprintf("endblock name '%s' doesn't match block name '%s'", name, blockName))
		}
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endblock")
	}
	p.advance()

	return blockNode, nil
}

// parseExtendsStatement parses extends statements
func (p *Parser) parseExtendsStatement() (Node, error) {
	extendsToken := p.advance() // consume 'extends'

	template, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after extends statement")
	}
	p.advance()

	return NewExtendsNode(template, extendsToken.Line, extendsToken.Column), nil
}

// parseIncludeStatement parses include statements
func (p *Parser) parseIncludeStatement() (Node, error) {
	includeToken := p.advance() // consume 'include'

	template, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	includeNode := NewIncludeNode(template, includeToken.Line, includeToken.Column)

	// Check for optional context
	if p.check(lexer.TokenWith) {
		p.advance() // consume 'with'
		context, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		includeNode.Context = context
	}

	// Check for ignore missing
	if p.check(lexer.TokenIgnore) {
		p.advance() // consume 'ignore'
		if !p.check(lexer.TokenMissing) {
			return nil, p.error("expected 'missing' after 'ignore'")
		}
		p.advance() // consume 'missing'
		includeNode.IgnoreMissing = true
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after include statement")
	}
	p.advance()

	return includeNode, nil
}

// parseMacroDefinition parses macro definitions
func (p *Parser) parseMacroDefinition() (Node, error) {
	macroToken := p.advance() // consume 'macro'

	if !p.check(lexer.TokenIdentifier) {
		return nil, p.error("expected macro name after 'macro'")
	}
	macroName := p.advance().Value

	macroNode := NewMacroNode(macroName, macroToken.Line, macroToken.Column)

	// Parse parameters
	if p.check(lexer.TokenLeftParen) {
		p.advance() // consume '('

		for !p.check(lexer.TokenRightParen) && !p.isAtEnd() {
			if !p.check(lexer.TokenIdentifier) {
				return nil, p.error("expected parameter name")
			}
			paramName := p.advance().Value
			macroNode.Parameters = append(macroNode.Parameters, paramName)

			// Check for default value
			if p.check(lexer.TokenAssign) {
				p.advance() // consume '='
				defaultValue, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				macroNode.Defaults[paramName] = defaultValue
			}

			if p.check(lexer.TokenComma) {
				p.advance() // consume ','
			} else if !p.check(lexer.TokenRightParen) {
				return nil, p.error("expected ',' or ')' in parameter list")
			}
		}

		if !p.check(lexer.TokenRightParen) {
			return nil, p.error("expected ')' after parameter list")
		}
		p.advance() // consume ')'
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after macro declaration")
	}
	p.advance()

	// Parse macro body
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			if p.peekBlockType() == lexer.TokenEndmacro {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			macroNode.Body = append(macroNode.Body, node)
		}
	}

	// Expect endmacro
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endmacro %}' to close macro")
	}
	p.advance() // consume {%

	if !p.check(lexer.TokenEndmacro) {
		return nil, p.error("expected 'endmacro'")
	}
	p.advance() // consume endmacro

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endmacro")
	}
	p.advance()

	return macroNode, nil
}

// parseRawBlock parses raw blocks
func (p *Parser) parseRawBlock() (Node, error) {
	rawToken := p.advance() // consume 'raw'

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after raw")
	}
	p.advance()

	var content strings.Builder

	// Collect everything until {% endraw %}
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			if p.peekBlockType() == lexer.TokenEndraw {
				break
			}
		}

		token := p.advance()
		content.WriteString(token.Value)
	}

	// Expect endraw
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endraw %}' to close raw block")
	}
	p.advance() // consume {%

	if !p.check(lexer.TokenEndraw) {
		return nil, p.error("expected 'endraw'")
	}
	p.advance() // consume endraw

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endraw")
	}
	p.advance()

	return NewRawNode(content.String(), rawToken.Line, rawToken.Column), nil
}

// parseAutoescapeBlock parses autoescape blocks
func (p *Parser) parseAutoescapeBlock() (Node, error) {
	autoescapeToken := p.advance() // consume 'autoescape'

	// Parse the boolean value (true/false or on/off)
	var enabled bool
	if p.check(lexer.TokenTrue) {
		p.advance()
		enabled = true
	} else if p.check(lexer.TokenFalse) {
		p.advance()
		enabled = false
	} else if p.check(lexer.TokenIdentifier) {
		value := p.advance().Value
		switch value {
		case "on":
			enabled = true
		case "off":
			enabled = false
		default:
			return nil, p.error(fmt.Sprintf("expected 'true', 'false', 'on', or 'off' after autoescape, got '%s'", value))
		}
	} else {
		return nil, p.error("expected 'true', 'false', 'on', or 'off' after autoescape")
	}

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after autoescape")
	}
	p.advance()

	autoescapeNode := NewAutoescapeNode(enabled, autoescapeToken.Line, autoescapeToken.Column)

	// Parse body until endautoescape
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			if p.peekBlockType() == lexer.TokenEndautoescape {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			autoescapeNode.Body = append(autoescapeNode.Body, node)
		}
	}

	// Expect endautoescape
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endautoescape %}' to close autoescape block")
	}
	p.advance() // consume {%

	if !p.check(lexer.TokenEndautoescape) {
		return nil, p.error("expected 'endautoescape'")
	}
	p.advance() // consume endautoescape

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endautoescape")
	}
	p.advance()

	return autoescapeNode, nil
}

// Public methods for extension support

// ParseExpressionPublic exposes parseExpression for extensions
func (p *Parser) ParseExpressionPublic() (ExpressionNode, error) {
	return p.parseExpression()
}

// ParseTopLevelPublic exposes parseTopLevel for extensions
func (p *Parser) ParseTopLevelPublic() (Node, error) {
	return p.parseTopLevel()
}

// PeekBlockTypePublic exposes peekBlockType for extensions
func (p *Parser) PeekBlockTypePublic() lexer.TokenType {
	return p.peekBlockType()
}

// ErrorPublic exposes the error method for extensions
func (p *Parser) ErrorPublic(message string) error {
	return p.error(message)
}

// Additional public methods for extension parser support
// Peek exposes the peek method for extensions
func (p *Parser) Peek() *lexer.Token {
	return p.peek()
}

// Advance exposes the advance method for extensions
func (p *Parser) Advance() *lexer.Token {
	return p.advance()
}

// Check exposes the check method for extensions
func (p *Parser) Check(tokenType lexer.TokenType) bool {
	return p.check(tokenType)
}

// CheckAny exposes the checkAny method for extensions
func (p *Parser) CheckAny(types ...lexer.TokenType) bool {
	return p.checkAny(types...)
}

// IsAtEnd exposes the isAtEnd method for extensions
func (p *Parser) IsAtEnd() bool {
	return p.isAtEnd()
}

// GetCurrentPosition returns the current parser position for extensions
func (p *Parser) GetCurrentPosition() int {
	return p.current
}

// SetCurrentPosition sets the current parser position for extensions
func (p *Parser) SetCurrentPosition(pos int) {
	if pos >= 0 && pos <= len(p.tokens) {
		p.current = pos
	}
}

// GetTokens returns the token slice for extensions
func (p *Parser) GetTokens() []*lexer.Token {
	return p.tokens
}

// ParseBlockStatementPublic exposes parseBlockStatement for extensions
func (p *Parser) ParseBlockStatementPublic() (Node, error) {
	return p.parseBlockStatement()
}

// Expression parsing methods

// parseExpression parses expressions with precedence
func (p *Parser) parseExpression() (ExpressionNode, error) {
	return p.parseConditional()
}

// parseConditional parses conditional expressions (ternary operator)
func (p *Parser) parseConditional() (ExpressionNode, error) {
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if p.check(lexer.TokenIf) {
		p.advance() // consume 'if'
		condition, err := p.parseOr()
		if err != nil {
			return nil, err
		}

		if !p.check(lexer.TokenElse) {
			return nil, p.error("expected 'else' in conditional expression")
		}
		p.advance() // consume 'else'

		falseExpr, err := p.parseConditional()
		if err != nil {
			return nil, err
		}

		return NewConditionalNode(condition, expr, falseExpr, p.previous().Line, p.previous().Column), nil
	}

	return expr, nil
}

// parseOr parses logical OR expressions
func (p *Parser) parseOr() (ExpressionNode, error) {
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenOr) {
		operator := p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseAnd parses logical AND expressions
func (p *Parser) parseAnd() (ExpressionNode, error) {
	expr, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenAnd) {
		operator := p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseNot parses logical NOT expressions
func (p *Parser) parseNot() (ExpressionNode, error) {
	if p.check(lexer.TokenNot) {
		operator := p.advance()
		expr, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return AcquireUnaryOpNode(operator.Value, expr, operator.Line, operator.Column), nil
	}

	return p.parseIs()
}

// parseIs parses 'is' test expressions
func (p *Parser) parseIs() (ExpressionNode, error) {
	expr, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	if p.check(lexer.TokenIs) {
		p.advance() // consume 'is'

		negated := false
		if p.check(lexer.TokenNot) {
			p.advance() // consume 'not'
			negated = true
		}

		if !p.check(lexer.TokenIdentifier) && !p.check(lexer.TokenNoneKeyword) {
			return nil, p.error("expected test name after 'is'")
		}
		testName := p.advance().Value

		testNode := NewTestNode(expr, testName, p.previous().Line, p.previous().Column)
		testNode.Negated = negated

		// Parse test arguments if present
		if p.check(lexer.TokenLeftParen) {
			p.advance() // consume '('

			for !p.check(lexer.TokenRightParen) && !p.isAtEnd() {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				testNode.Arguments = append(testNode.Arguments, arg)

				if p.check(lexer.TokenComma) {
					p.advance() // consume ','
				} else if !p.check(lexer.TokenRightParen) {
					return nil, p.error("expected ',' or ')' in test arguments")
				}
			}

			if !p.check(lexer.TokenRightParen) {
				return nil, p.error("expected ')' after test arguments")
			}
			p.advance() // consume ')'
		}

		return testNode, nil
	}

	return expr, nil
}

// parseComparison parses comparison expressions
func (p *Parser) parseComparison() (ExpressionNode, error) {
	expr, err := p.parseConcatenation()
	if err != nil {
		return nil, err
	}

	for p.checkAny(lexer.TokenGreater, lexer.TokenGreaterEqual, lexer.TokenLess, lexer.TokenLessEqual, lexer.TokenEqual, lexer.TokenNotEqual, lexer.TokenIn) || (p.check(lexer.TokenNot) && p.checkNext(lexer.TokenIn)) {
		if p.check(lexer.TokenNot) && p.checkNext(lexer.TokenIn) {
			// Handle "not in" operator
			notToken := p.advance() // consume 'not'
			p.advance()             // consume 'in'
			right, err := p.parseConcatenation()
			if err != nil {
				return nil, err
			}
			expr = AcquireBinaryOpNode(expr, "not in", right, notToken.Line, notToken.Column)
		} else {
			operator := p.advance()
			right, err := p.parseConcatenation()
			if err != nil {
				return nil, err
			}
			expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
		}
	}

	return expr, nil
}

// parseConcatenation parses string concatenation (~)
func (p *Parser) parseConcatenation() (ExpressionNode, error) {
	expr, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenTilde) {
		operator := p.advance()
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseAddition parses addition and subtraction
func (p *Parser) parseAddition() (ExpressionNode, error) {
	expr, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for p.checkAny(lexer.TokenPlus, lexer.TokenMinus) {
		operator := p.advance()
		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseMultiplication parses multiplication, division, and modulo
func (p *Parser) parseMultiplication() (ExpressionNode, error) {
	expr, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.checkAny(lexer.TokenMultiply, lexer.TokenDivide, lexer.TokenFloorDivide, lexer.TokenModulo) {
		operator := p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseUnary parses unary expressions
func (p *Parser) parseUnary() (ExpressionNode, error) {
	if p.checkAny(lexer.TokenMinus, lexer.TokenPlus) {
		operator := p.advance()
		expr, err := p.parsePower() // Unary should have lower precedence than power
		if err != nil {
			return nil, err
		}
		return AcquireUnaryOpNode(operator.Value, expr, operator.Line, operator.Column), nil
	}

	return p.parsePower()
}

// parsePower parses power expressions (**)
func (p *Parser) parsePower() (ExpressionNode, error) {
	expr, err := p.parsePostfix() // Power operates on postfix expressions
	if err != nil {
		return nil, err
	}

	if p.check(lexer.TokenPower) {
		operator := p.advance()
		right, err := p.parseUnary() // Right associative, and unary can bind to the right side
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parsePostfix parses postfix expressions (filters, attribute access, etc.)
func (p *Parser) parsePostfix() (ExpressionNode, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		if p.check(lexer.TokenDot) {
			p.advance() // consume '.'
			if !p.check(lexer.TokenIdentifier) {
				return nil, p.error("expected attribute name after '.'")
			}
			attr := p.advance().Value
			expr = AcquireAttributeNode(expr, attr, p.previous().Line, p.previous().Column)

		} else if p.check(lexer.TokenLeftBracket) {
			p.advance() // consume '['

			// Check for slice notation
			if p.check(lexer.TokenColon) {
				// Slice without start: [:end]
				p.advance() // consume ':'
				var end ExpressionNode
				if !p.check(lexer.TokenRightBracket) && !p.check(lexer.TokenColon) {
					end, err = p.parseExpression()
					if err != nil {
						return nil, err
					}
				}

				var step ExpressionNode
				if p.check(lexer.TokenColon) {
					p.advance() // consume ':'
					if !p.check(lexer.TokenRightBracket) {
						step, err = p.parseExpression()
						if err != nil {
							return nil, err
						}
					}
				}

				if !p.check(lexer.TokenRightBracket) {
					return nil, p.error("expected ']' after slice")
				}
				p.advance() // consume ']'

				sliceNode := NewSliceNode(expr, p.previous().Line, p.previous().Column)
				sliceNode.End = end
				sliceNode.Step = step
				expr = sliceNode

			} else {
				// Regular index or slice with start
				index, err := p.parseExpression()
				if err != nil {
					return nil, err
				}

				if p.check(lexer.TokenColon) {
					// Slice with start: [start:end] or [start:]
					p.advance() // consume ':'

					var end ExpressionNode
					if !p.check(lexer.TokenRightBracket) && !p.check(lexer.TokenColon) {
						end, err = p.parseExpression()
						if err != nil {
							return nil, err
						}
					}

					var step ExpressionNode
					if p.check(lexer.TokenColon) {
						p.advance() // consume ':'
						if !p.check(lexer.TokenRightBracket) {
							step, err = p.parseExpression()
							if err != nil {
								return nil, err
							}
						}
					}

					if !p.check(lexer.TokenRightBracket) {
						return nil, p.error("expected ']' after slice")
					}
					p.advance() // consume ']'

					sliceNode := NewSliceNode(expr, p.previous().Line, p.previous().Column)
					sliceNode.Start = index
					sliceNode.End = end
					sliceNode.Step = step
					expr = sliceNode

				} else {
					// Regular index access
					if !p.check(lexer.TokenRightBracket) {
						return nil, p.error("expected ']' after index")
					}
					p.advance() // consume ']'

					expr = AcquireGetItemNode(expr, index, p.previous().Line, p.previous().Column)
				}
			}

		} else if p.check(lexer.TokenPipe) {
			p.advance() // consume '|'
			if !p.check(lexer.TokenIdentifier) && !p.check(lexer.TokenFilter) {
				return nil, p.error("expected filter name after '|'")
			}
			filterName := p.advance().Value

			var args []ExpressionNode
			namedArgs := make(map[string]ExpressionNode)

			if p.check(lexer.TokenLeftParen) {
				p.advance() // consume '('

				for !p.check(lexer.TokenRightParen) && !p.isAtEnd() {
					// Check if this is a named argument (identifier = expression)
					if p.check(lexer.TokenIdentifier) {
						// Look ahead to see if there's an assignment
						savedPos := p.current
						argName := p.advance().Value

						if p.check(lexer.TokenAssign) {
							// This is a named argument
							p.advance()                  // consume '='
							argValue, err := p.parseOr() // Use parseOr to avoid conditional parsing issues
							if err != nil {
								return nil, err
							}
							namedArgs[argName] = argValue
						} else {
							// This is a positional argument, restore position and parse normally
							p.current = savedPos
							arg, err := p.parseOr() // Use parseOr to avoid conditional parsing issues
							if err != nil {
								return nil, err
							}
							args = append(args, arg)
						}
					} else {
						// Regular positional argument
						arg, err := p.parseOr() // Use parseOr to avoid conditional parsing issues
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
					}

					if p.check(lexer.TokenComma) {
						p.advance() // consume ','
					} else if !p.check(lexer.TokenRightParen) {
						return nil, p.error("expected ',' or ')' in filter arguments")
					}
				}

				if !p.check(lexer.TokenRightParen) {
					return nil, p.error("expected ')' after filter arguments")
				}
				p.advance() // consume ')'
			}

			filterNode := AcquireFilterNode(expr, filterName, args, p.previous().Line, p.previous().Column)
			filterNode.NamedArgs = namedArgs
			expr = filterNode

		} else if p.check(lexer.TokenLeftParen) {
			// Special case for super() - should remain SuperNode, not become CallNode
			if superNode, isSuperNode := expr.(*SuperNode); isSuperNode {
				p.advance() // consume '('

				// super() should have empty arguments
				if !p.check(lexer.TokenRightParen) {
					return nil, p.error("super() does not accept arguments")
				}
				p.advance() // consume ')'

				// Return the original SuperNode, don't wrap it in CallNode
				expr = superNode
			} else {
				// Function call for non-super expressions
				p.advance() // consume '('

				var args []ExpressionNode
				var keywords map[string]ExpressionNode

				for !p.check(lexer.TokenRightParen) && !p.isAtEnd() {
					// Check for keyword argument
					if p.check(lexer.TokenIdentifier) && p.peekNext().Type == lexer.TokenAssign {
						if keywords == nil {
							keywords = make(map[string]ExpressionNode)
						}
						key := p.advance().Value
						p.advance() // consume '='
						value, err := p.parseExpression()
						if err != nil {
							return nil, err
						}
						keywords[key] = value
					} else {
						if keywords != nil {
							return nil, p.error("positional arguments cannot follow keyword arguments")
						}
						arg, err := p.parseExpression()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)
					}

					if p.check(lexer.TokenComma) {
						p.advance() // consume ','
					} else if !p.check(lexer.TokenRightParen) {
						return nil, p.error("expected ',' or ')' in function call")
					}
				}

				if !p.check(lexer.TokenRightParen) {
					return nil, p.error("expected ')' after function arguments")
				}
				p.advance() // consume ')'

				callNode := AcquireCallNode(expr, p.previous().Line, p.previous().Column)
				callNode.Arguments = args
				callNode.Keywords = keywords
				expr = callNode
			}

		} else {
			break
		}
	}

	return expr, nil
}

// parsePrimary parses primary expressions
func (p *Parser) parsePrimary() (ExpressionNode, error) {
	switch p.peek().Type {
	case lexer.TokenTrue, lexer.TokenFalse:
		token := p.advance()
		value := token.Type == lexer.TokenTrue
		return AcquireLiteralNode(value, token.Value, token.Line, token.Column), nil

	case lexer.TokenNoneKeyword:
		token := p.advance()
		return AcquireLiteralNode(nil, token.Value, token.Line, token.Column), nil

	case lexer.TokenInteger:
		token := p.advance()
		value, err := strconv.Atoi(token.Value)
		if err != nil {
			return nil, p.error(fmt.Sprintf("invalid integer: %s", token.Value))
		}
		return AcquireLiteralNode(value, token.Value, token.Line, token.Column), nil

	case lexer.TokenFloat:
		token := p.advance()
		value, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, p.error(fmt.Sprintf("invalid float: %s", token.Value))
		}
		return AcquireLiteralNode(value, token.Value, token.Line, token.Column), nil

	case lexer.TokenString:
		token := p.advance()
		return AcquireLiteralNode(token.Value, token.Value, token.Line, token.Column), nil

	case lexer.TokenIdentifier:
		token := p.advance()
		return AcquireIdentifierNode(token.Value, token.Line, token.Column), nil

	case lexer.TokenSuper:
		token := p.advance()
		return NewSuperNode(token.Line, token.Column), nil

	case lexer.TokenLeftParen:
		p.advance() // consume '('
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.check(lexer.TokenRightParen) {
			return nil, p.error("expected ')' after expression")
		}
		p.advance() // consume ')'
		return expr, nil

	case lexer.TokenLeftBracket:
		return p.parseListLiteral()

	case lexer.TokenLeftBrace:
		return p.parseDictLiteral()

	default:
		return nil, p.error(fmt.Sprintf("unexpected token in expression: %s", p.peek().Type))
	}
}

// parseListLiteral parses list literals and comprehensions
func (p *Parser) parseListLiteral() (ExpressionNode, error) {
	startToken := p.advance() // consume '['

	if p.check(lexer.TokenRightBracket) {
		p.advance() // consume ']'
		return AcquireLiteralNode([]interface{}{}, "[]", startToken.Line, startToken.Column), nil
	}

	// Parse first expression
	firstExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for list comprehension
	if p.check(lexer.TokenFor) {
		p.advance() // consume 'for'

		if !p.check(lexer.TokenIdentifier) {
			return nil, p.error("expected variable name in list comprehension")
		}
		variable := p.advance().Value

		if !p.check(lexer.TokenIn) {
			return nil, p.error("expected 'in' in list comprehension")
		}
		p.advance() // consume 'in'

		iterable, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		compNode := NewComprehensionNode(firstExpr, variable, iterable, startToken.Line, startToken.Column)

		// Check for condition
		if p.check(lexer.TokenIf) {
			p.advance() // consume 'if'
			condition, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			compNode.Condition = condition
		}

		if !p.check(lexer.TokenRightBracket) {
			return nil, p.error("expected ']' after list comprehension")
		}
		p.advance() // consume ']'

		return compNode, nil
	}

	// Regular list literal
	elements := []ExpressionNode{firstExpr}

	for p.check(lexer.TokenComma) {
		p.advance() // consume ','

		if p.check(lexer.TokenRightBracket) {
			break // Trailing comma
		}

		element, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}

	if !p.check(lexer.TokenRightBracket) {
		return nil, p.error("expected ']' after list elements")
	}
	p.advance() // consume ']'

	return NewListNode(elements, startToken.Line, startToken.Column), nil
}

// parseDictLiteral parses dictionary literals and comprehensions
func (p *Parser) parseDictLiteral() (ExpressionNode, error) {
	startToken := p.advance() // consume '{'

	if p.check(lexer.TokenRightBrace) {
		p.advance() // consume '}'
		return AcquireLiteralNode(map[string]interface{}{}, "{}", startToken.Line, startToken.Column), nil
	}

	// Parse first key-value pair
	key, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.check(lexer.TokenColon) {
		return nil, p.error("expected ':' after dictionary key")
	}
	p.advance() // consume ':'

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for dictionary comprehension
	if p.check(lexer.TokenFor) {
		p.advance() // consume 'for'

		if !p.check(lexer.TokenIdentifier) {
			return nil, p.error("expected variable name in dict comprehension")
		}
		variable := p.advance().Value

		if !p.check(lexer.TokenIn) {
			return nil, p.error("expected 'in' in dict comprehension")
		}
		p.advance() // consume 'in'

		iterable, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		compNode := NewComprehensionNode(value, variable, iterable, startToken.Line, startToken.Column)
		compNode.IsDict = true
		compNode.KeyExpr = key

		// Check for condition
		if p.check(lexer.TokenIf) {
			p.advance() // consume 'if'
			condition, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			compNode.Condition = condition
		}

		if !p.check(lexer.TokenRightBrace) {
			return nil, p.error("expected '}' after dict comprehension")
		}
		p.advance() // consume '}'

		return compNode, nil
	}

	// Regular dictionary literal - for now return empty dict
	// In a full implementation, we'd build the actual dictionary
	pairs := make(map[string]interface{})

	if p.check(lexer.TokenComma) {
		p.advance() // consume ','

		for !p.check(lexer.TokenRightBrace) && !p.isAtEnd() {
			key, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			if !p.check(lexer.TokenColon) {
				return nil, p.error("expected ':' after dictionary key")
			}
			p.advance() // consume ':'

			value, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			// Store key-value pair (simplified)
			_ = key
			_ = value

			if p.check(lexer.TokenComma) {
				p.advance() // consume ','
			} else if !p.check(lexer.TokenRightBrace) {
				return nil, p.error("expected ',' or '}' in dictionary")
			}
		}
	}

	if !p.check(lexer.TokenRightBrace) {
		return nil, p.error("expected '}' after dictionary elements")
	}
	p.advance() // consume '}'

	return AcquireLiteralNode(pairs, "", startToken.Line, startToken.Column), nil
}

// Helper methods

func (p *Parser) peek() *lexer.Token {
	if p.isAtEnd() {
		return &lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.current]
}

func (p *Parser) peekNext() *lexer.Token {
	if p.current+1 >= len(p.tokens) {
		return &lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.current+1]
}

func (p *Parser) previous() *lexer.Token {
	if p.current == 0 {
		return &lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.current-1]
}

func (p *Parser) advance() *lexer.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens) || (p.current < len(p.tokens) && p.tokens[p.current].Type == lexer.TokenEOF)
}

func (p *Parser) check(tokenType lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

func (p *Parser) checkAny(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			return true
		}
	}
	return false
}

func (p *Parser) checkNext(tokenType lexer.TokenType) bool {
	if p.current+1 >= len(p.tokens) {
		return false
	}
	return p.tokens[p.current+1].Type == tokenType
}

func (p *Parser) peekBlockType() lexer.TokenType {
	if p.current+1 < len(p.tokens) {
		return p.tokens[p.current+1].Type
	}
	return lexer.TokenEOF
}

// parseBreakStatement parses break statements {% break %}
func (p *Parser) parseBreakStatement() (Node, error) {
	breakToken := p.advance() // consume 'break'

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after break statement")
	}
	p.advance() // consume '%}'

	return NewBreakNode(breakToken.Line, breakToken.Column), nil
}

// parseContinueStatement parses continue statements {% continue %}
func (p *Parser) parseContinueStatement() (Node, error) {
	continueToken := p.advance() // consume 'continue'

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after continue statement")
	}
	p.advance() // consume '%}'

	return NewContinueNode(continueToken.Line, continueToken.Column), nil
}

// parseImportStatement parses import statements {% import 'template' as name %}
func (p *Parser) parseImportStatement() (Node, error) {
	importToken := p.advance() // consume 'import'

	// Parse the template expression
	template, err := p.parseExpression()
	if err != nil {
		return nil, p.error("expected template expression after import")
	}

	// Expect 'as' keyword
	if !p.check(lexer.TokenAs) {
		return nil, p.error("expected 'as' after template in import statement")
	}
	p.advance() // consume 'as'

	// Parse the alias name
	if !p.check(lexer.TokenIdentifier) {
		return nil, p.error("expected identifier after 'as' in import statement")
	}
	alias := p.advance().Value

	// Expect closing %}
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after import statement")
	}
	p.advance() // consume '%}'

	return NewImportNode(importToken.Line, importToken.Column, template, alias), nil
}

// parseFromStatement parses from-import statements {% from 'template' import name1, name2 %}
func (p *Parser) parseFromStatement() (Node, error) {
	fromToken := p.advance() // consume 'from'

	// Parse the template expression
	template, err := p.parseExpression()
	if err != nil {
		return nil, p.error("expected template expression after from")
	}

	// Expect 'import' keyword
	if !p.check(lexer.TokenImport) {
		return nil, p.error("expected 'import' after template in from statement")
	}
	p.advance() // consume 'import'

	// Parse the list of names to import
	var names []string
	aliases := make(map[string]string)

	for {
		if !p.check(lexer.TokenIdentifier) {
			return nil, p.error("expected identifier in import list")
		}
		name := p.advance().Value
		names = append(names, name)

		// Check for optional 'as' alias
		if p.check(lexer.TokenAs) {
			p.advance() // consume 'as'
			if !p.check(lexer.TokenIdentifier) {
				return nil, p.error("expected identifier after 'as'")
			}
			alias := p.advance().Value
			aliases[name] = alias
		}

		// Check for comma to continue or end
		if p.check(lexer.TokenComma) {
			p.advance() // consume ','
			continue
		}
		break
	}

	// Expect closing %}
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after from statement")
	}
	p.advance() // consume '%}'

	return NewFromNode(fromToken.Line, fromToken.Column, template, names, aliases), nil
}

func (p *Parser) error(message string) error {
	token := p.peek()
	fullMsg := fmt.Sprintf("%s at line %d, column %d", message, token.Line, token.Column)
	p.errors = append(p.errors, fullMsg)
	return fmt.Errorf("%s", fullMsg)
}

// GetErrors returns any errors encountered during parsing
func (p *Parser) GetErrors() []string {
	return p.errors
}

// parseCallBlockStatement parses {% call expression %}...{% endcall %} statements
func (p *Parser) parseCallBlockStatement() (Node, error) {
	startToken := p.advance() // consume 'call'

	// Parse the call expression (function/macro call)
	callExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect {% %}
	if !p.check(lexer.TokenBlockEnd) {
		return nil, p.error("expected '%}' after call expression")
	}
	p.advance() // consume '%}'

	// Parse body until {% endcall %}
	var body []Node
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			blockType := p.peekBlockType()
			if blockType == lexer.TokenEndcall {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
	}

	// Consume the {% endcall %} block
	if !(p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim)) {
		return nil, p.error("expected '{% endcall %}' to close call block")
	}
	p.advance() // consume '{%'

	if !p.check(lexer.TokenEndcall) {
		return nil, p.error("expected 'endcall' after '{%'")
	}
	p.advance() // consume 'endcall'

	if !(p.check(lexer.TokenBlockEnd) || p.check(lexer.TokenBlockEndTrim)) {
		return nil, p.error("expected '%}' after endcall")
	}
	p.advance() // consume '%}'

	return NewCallBlockNode(callExpr, body, startToken.Line, startToken.Column), nil
}

// parseWithStatement parses {% with var=expr, var2=expr2 %}...{% endwith %} statements
func (p *Parser) parseWithStatement() (Node, error) {
	startToken := p.advance() // consume 'with'

	// Parse assignments (var1=expr1, var2=expr2, ...)
	assignments := make(map[string]ExpressionNode)

	for {
		// Parse variable name
		if !p.check(lexer.TokenIdentifier) {
			return nil, p.error("expected variable name in with statement")
		}
		varName := p.advance().Value

		// Expect '='
		if !p.check(lexer.TokenAssign) {
			return nil, p.error("expected '=' after variable name in with statement")
		}
		p.advance() // consume '='

		// Parse expression
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		assignments[varName] = expr

		// Check for more assignments
		if p.check(lexer.TokenComma) {
			p.advance() // consume ','
			continue
		}
		break
	}

	// Expect {% %}
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after with assignments")
	}
	p.advance() // consume '%}'

	// Parse body until {% endwith %}
	var body []Node
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			blockType := p.peekBlockType()
			if blockType == lexer.TokenEndwith {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
	}

	// Consume the {% endwith %} block
	if !(p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim)) {
		return nil, p.error("expected '{% endwith %}' to close with block")
	}
	p.advance() // consume '{%'

	if !p.check(lexer.TokenEndwith) {
		return nil, p.error("expected 'endwith' after '{%'")
	}
	p.advance() // consume 'endwith'

	if !(p.check(lexer.TokenBlockEnd) || p.check(lexer.TokenBlockEndTrim)) {
		return nil, p.error("expected '%}' after endwith")
	}
	p.advance() // consume '%}'

	return NewWithNode(assignments, body, startToken.Line, startToken.Column), nil
}

// parseDoStatement parses do statements {% do expression %}
func (p *Parser) parseDoStatement() (Node, error) {
	startToken := p.advance() // consume 'do'

	// Parse the expression to execute
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect block end
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after do expression")
	}
	p.advance() // consume '%}'

	return NewDoNode(expr, startToken.Line, startToken.Column), nil
}

// parseFilterBlock parses filter blocks {% filter upper|trim %}...{% endfilter %}
func (p *Parser) parseFilterBlock() (Node, error) {
	filterToken := p.advance() // consume 'filter'

	// Parse the filter chain
	var filterChain []FilterNode

	// First filter is required
	if !p.check(lexer.TokenIdentifier) && !p.check(lexer.TokenFilter) {
		return nil, p.error("expected filter name after 'filter'")
	}

	for {
		if !p.check(lexer.TokenIdentifier) && !p.check(lexer.TokenFilter) {
			break
		}

		filterName := p.advance().Value
		var args []ExpressionNode
		namedArgs := make(map[string]ExpressionNode)

		// Parse filter arguments if present
		if p.check(lexer.TokenLeftParen) {
			p.advance() // consume '('

			for !p.check(lexer.TokenRightParen) && !p.isAtEnd() {
				// Check for keyword argument
				if p.check(lexer.TokenIdentifier) && p.peekNext().Type == lexer.TokenAssign {
					if namedArgs == nil {
						namedArgs = make(map[string]ExpressionNode)
					}
					name := p.advance().Value
					p.advance() // consume '='
					expr, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					namedArgs[name] = expr
				} else {
					expr, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, expr)
				}

				if p.check(lexer.TokenComma) {
					p.advance()
				} else {
					break
				}
			}

			if !p.check(lexer.TokenRightParen) {
				return nil, p.error("expected ')' after filter arguments")
			}
			p.advance() // consume ')'
		}

		// Create a dummy identifier node as the expression (will be replaced during evaluation)
		dummyExpr := AcquireIdentifierNode("__filter_block_content__", filterToken.Line, filterToken.Column)
		filterNode := AcquireFilterNode(dummyExpr, filterName, args, filterToken.Line, filterToken.Column)
		filterNode.NamedArgs = namedArgs
		filterChain = append(filterChain, *filterNode)

		// Check for more filters in chain
		if p.check(lexer.TokenPipe) {
			p.advance() // consume '|'
		} else {
			break
		}
	}

	// Expect block end
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after filter chain")
	}
	p.advance() // consume '%}'

	// Create the filter block node
	filterBlockNode := NewFilterBlockNode(filterChain, filterToken.Line, filterToken.Column)

	// Parse block body until {% endfilter %}
	for !p.isAtEnd() {
		if p.check(lexer.TokenBlockStart) || p.check(lexer.TokenBlockStartTrim) {
			if p.peekBlockType() == lexer.TokenEndfilter {
				break
			}
		}

		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		filterBlockNode.Body = append(filterBlockNode.Body, node)
	}

	// Expect {% endfilter %}
	if !p.check(lexer.TokenBlockStart) && !p.check(lexer.TokenBlockStartTrim) {
		return nil, p.error("expected '{% endfilter %}' to close filter block")
	}
	p.advance() // consume '{%'

	if !p.check(lexer.TokenEndfilter) {
		return nil, p.error("expected 'endfilter' to close filter block")
	}
	p.advance() // consume 'endfilter'

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after endfilter")
	}
	p.advance() // consume '%}'

	return filterBlockNode, nil
}
