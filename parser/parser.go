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
	depth   int // nesting of parsePrimary/parseBlockStatement calls
}

// maxParseDepth bounds expression and statement nesting so a deeply nested
// template returns an error instead of overflowing the goroutine stack.
// Jinja2 itself fails at about 100 levels, so no working template comes close.
const maxParseDepth = 500

func (p *Parser) enterNesting() error {
	p.depth++
	if p.depth > maxParseDepth {
		return p.error(fmt.Sprintf("template nesting too deep (limit %d)", maxParseDepth))
	}
	return nil
}

// nest counts one more level for a chain that wraps its left operand in a new
// node on each iteration (1 + 1 + 1 ...); the caller gives back *added levels.
func (p *Parser) nest(added *int) error {
	*added++
	return p.enterNesting()
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

	if err := CheckDuplicateBlocks(template); err != nil {
		return nil, err
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
	defer func() { p.depth-- }()
	if err := p.enterNesting(); err != nil {
		return nil, err
	}
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

	iterable, err := p.parseExpressionList(p.parseOr) // Use parseOr to avoid consuming the 'if' token
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
	if tuple, ok := firstTarget.(*ListNode); ok {
		// {% set (a, b) = ... %}
		targets = append(targets, tuple.Elements...)
	} else {
		targets = append(targets, firstTarget)
	}

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

		value, err := p.parseExpressionList(p.parseExpression)
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

	// "scoped" is accepted; blocks always see the variables around them
	if p.check(lexer.TokenIdentifier) && p.peek().Value == "scoped" {
		p.advance()
	}

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

	// Jinja2 order: [ignore missing] [with context | without context].
	// miya also accepts "with <expr>" and the older order
	// "with context ignore missing".
	ignoreMissing := func() error {
		if !p.check(lexer.TokenIgnore) {
			return nil
		}
		p.advance() // consume 'ignore'
		if !p.check(lexer.TokenMissing) {
			return p.error("expected 'missing' after 'ignore'")
		}
		p.advance() // consume 'missing'
		includeNode.IgnoreMissing = true
		return nil
	}
	if err := ignoreMissing(); err != nil {
		return nil, err
	}

	switch {
	case p.check(lexer.TokenIdentifier) && p.peek().Value == "without":
		p.advance() // consume 'without'
		if !p.check(lexer.TokenIdentifier) || p.peek().Value != "context" {
			return nil, p.error("expected 'context' after 'without'")
		}
		p.advance() // consume 'context'
		includeNode.WithoutContext = true
	case p.check(lexer.TokenWith):
		p.advance() // consume 'with'
		// "with context" is the keyword pair; "with context.x" is miya's
		// "with <expr>" form, as it parsed before
		if p.check(lexer.TokenIdentifier) && p.peek().Value == "context" && !p.checkNext(lexer.TokenDot) && !p.checkNext(lexer.TokenLeftBracket) && !p.checkNext(lexer.TokenPipe) {
			p.advance() // consume 'context'
			includeNode.WithContext = true
		} else {
			context, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			includeNode.Context = context
		}
	}

	if !includeNode.IgnoreMissing {
		if err := ignoreMissing(); err != nil {
			return nil, err
		}
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
		params, defaults, err := p.parseParameterList()
		if err != nil {
			return nil, err
		}
		macroNode.Parameters = params
		macroNode.Defaults = defaults
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

	usage, err := findMacroUsage(macroNode.Parameters, macroNode.Defaults, macroNode.Body)
	if err != nil {
		return nil, p.error(err.Error())
	}
	macroNode.MacroUsage = usage
	return macroNode, nil
}

// parseParameterList parses "(a, b=expr)" for macros and call blocks. It
// consumes the opening parenthesis itself.
func (p *Parser) parseParameterList() ([]string, map[string]ExpressionNode, error) {
	p.advance() // consume '('
	params := []string{}
	defaults := make(map[string]ExpressionNode)

	for !p.check(lexer.TokenRightParen) && !p.isAtEnd() {
		if !p.check(lexer.TokenIdentifier) {
			return nil, nil, p.error("expected parameter name")
		}
		paramName := p.advance().Value
		params = append(params, paramName)

		// Check for default value
		if p.check(lexer.TokenAssign) {
			p.advance() // consume '='
			defaultValue, err := p.parseExpression()
			if err != nil {
				return nil, nil, err
			}
			defaults[paramName] = defaultValue
		}

		if p.check(lexer.TokenComma) {
			p.advance() // consume ','
		} else if !p.check(lexer.TokenRightParen) {
			return nil, nil, p.error("expected ',' or ')' in parameter list")
		}
	}

	if !p.check(lexer.TokenRightParen) {
		return nil, nil, p.error("expected ')' after parameter list")
	}
	p.advance() // consume ')'
	return params, defaults, nil
}

// parseRawBlock parses raw blocks
func (p *Parser) parseRawBlock() (Node, error) {
	rawToken := p.advance() // consume 'raw'

	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after raw")
	}
	p.advance()

	// The lexer delivers the raw content as at most one text token
	content := ""
	if p.check(lexer.TokenText) {
		content = p.advance().Value
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

	return NewRawNode(content, rawToken.Line, rawToken.Column), nil
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

		// Without "else" the expression is undefined when the condition is false
		var falseExpr ExpressionNode
		if p.check(lexer.TokenElse) {
			p.advance() // consume 'else'

			if err := p.enterNesting(); err != nil {
				p.depth--
				return nil, err
			}
			falseExpr, err = p.parseConditional()
			p.depth--
			if err != nil {
				return nil, err
			}
		}

		return NewConditionalNode(condition, expr, falseExpr, p.previous().Line, p.previous().Column), nil
	}

	return expr, nil
}

// parseOr parses logical OR expressions
func (p *Parser) parseOr() (ExpressionNode, error) {
	added := 0
	defer func() { p.depth -= added }()
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenOr) {
		if err := p.nest(&added); err != nil {
			return nil, err
		}
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
	added := 0
	defer func() { p.depth -= added }()
	expr, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenAnd) {
		if err := p.nest(&added); err != nil {
			return nil, err
		}
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
		if err := p.enterNesting(); err != nil {
			p.depth--
			return nil, err
		}
		expr, err := p.parseNot()
		p.depth--
		if err != nil {
			return nil, err
		}
		return AcquireUnaryOpNode(operator.Value, expr, operator.Line, operator.Column), nil
	}

	return p.parseComparison()
}

// parseComparison parses comparison expressions. One comparison is a
// BinaryOpNode; a chain such as a < b < c is a CompareNode, meaning
// a < b and b < c, as in Jinja2.
func (p *Parser) parseComparison() (ExpressionNode, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	var ops []CompareOperand
	for p.checkAny(lexer.TokenGreater, lexer.TokenGreaterEqual, lexer.TokenLess, lexer.TokenLessEqual, lexer.TokenEqual, lexer.TokenNotEqual, lexer.TokenIn) || (p.check(lexer.TokenNot) && p.checkNext(lexer.TokenIn)) {
		operator := p.advance()
		op := operator.Value
		if operator.Type == lexer.TokenNot {
			p.advance() // consume 'in'
			op = "not in"
		}
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		ops = append(ops, CompareOperand{Operator: op, Right: right, Line: operator.Line, Column: operator.Column})
	}

	switch len(ops) {
	case 0:
		return left, nil
	case 1:
		return AcquireBinaryOpNode(left, ops[0].Operator, ops[0].Right, ops[0].Line, ops[0].Column), nil
	}
	return NewCompareNode(left, ops, ops[0].Line, ops[0].Column), nil
}

// parseAddition parses addition and subtraction
func (p *Parser) parseAddition() (ExpressionNode, error) {
	added := 0
	defer func() { p.depth -= added }()
	expr, err := p.parseConcatenation()
	if err != nil {
		return nil, err
	}

	for p.checkAny(lexer.TokenPlus, lexer.TokenMinus) {
		if err := p.nest(&added); err != nil {
			return nil, err
		}
		operator := p.advance()
		right, err := p.parseConcatenation()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseConcatenation parses string concatenation (~), which binds tighter
// than + and - and looser than *, as in Jinja2
func (p *Parser) parseConcatenation() (ExpressionNode, error) {
	added := 0
	defer func() { p.depth -= added }()
	expr, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenTilde) {
		if err := p.nest(&added); err != nil {
			return nil, err
		}
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
	added := 0
	defer func() { p.depth -= added }()
	expr, err := p.parsePower()
	if err != nil {
		return nil, err
	}

	for p.checkAny(lexer.TokenMultiply, lexer.TokenDivide, lexer.TokenFloorDivide, lexer.TokenModulo) {
		if err := p.nest(&added); err != nil {
			return nil, err
		}
		operator := p.advance()
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parsePower parses **. As in Jinja2 it is left-associative and its operands
// are unary expressions: 2 ** 3 ** 2 is 64 and -2 ** 2 is 4.
func (p *Parser) parsePower() (ExpressionNode, error) {
	added := 0
	defer func() { p.depth -= added }()
	expr, err := p.parseUnary(true)
	if err != nil {
		return nil, err
	}

	for p.check(lexer.TokenPower) {
		if err := p.nest(&added); err != nil {
			return nil, err
		}
		operator := p.advance()
		right, err := p.parseUnary(true)
		if err != nil {
			return nil, err
		}
		expr = AcquireBinaryOpNode(expr, operator.Value, right, operator.Line, operator.Column)
	}

	return expr, nil
}

// parseUnary parses unary minus and plus. As in Jinja2, the sign binds
// before filters and tests: -x|abs is (-x)|abs.
func (p *Parser) parseUnary(withFilters bool) (ExpressionNode, error) {
	var expr ExpressionNode
	if p.checkAny(lexer.TokenMinus, lexer.TokenPlus) {
		// - - - 1 recurses here, not through parsePrimary: count the level
		operator := p.advance()
		if err := p.enterNesting(); err != nil {
			p.depth--
			return nil, err
		}
		operand, err := p.parseUnary(false)
		p.depth--
		if err != nil {
			return nil, err
		}
		expr = AcquireUnaryOpNode(operator.Value, operand, operator.Line, operator.Column)
	} else {
		primary, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		expr = primary
	}
	return p.parsePostfixOps(expr, withFilters)
}

// parsePostfix parses postfix expressions (filters, attribute access, etc.)
func (p *Parser) parsePostfix() (ExpressionNode, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	return p.parsePostfixOps(expr, true)
}

// parsePostfixOps applies the postfix operators (attribute, index, call,
// and, with withFilters, filters and tests) to expr.
func (p *Parser) parsePostfixOps(expr ExpressionNode, withFilters bool) (ExpressionNode, error) {
	var err error
	added := 0
	defer func() { p.depth -= added }()
	for {
		// every postfix operator wraps expr in a deeper node
		if p.checkAny(lexer.TokenDot, lexer.TokenLeftBracket, lexer.TokenLeftParen) || (withFilters && p.checkAny(lexer.TokenPipe, lexer.TokenIs)) {
			if err := p.nest(&added); err != nil {
				return nil, err
			}
		}
		if p.check(lexer.TokenDot) {
			p.advance() // consume '.'
			if p.check(lexer.TokenInteger) {
				// b.0 is b[0]
				tok := p.advance()
				index, convErr := strconv.Atoi(tok.Value)
				if convErr != nil {
					return nil, p.error(fmt.Sprintf("invalid integer: %s", tok.Value))
				}
				expr = AcquireGetItemNode(expr, AcquireLiteralNode(index, tok.Value, tok.Line, tok.Column), tok.Line, tok.Column)
				continue
			}
			if !isNameToken(p.peek()) {
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

		} else if withFilters && p.check(lexer.TokenIs) {
			expr, err = p.parseTest(expr)
			if err != nil {
				return nil, err
			}

		} else if withFilters && p.check(lexer.TokenPipe) {
			p.advance() // consume '|'
			if !p.check(lexer.TokenIdentifier) && !p.check(lexer.TokenFilter) {
				return nil, p.error("expected filter name after '|'")
			}
			filterName := p.advance().Value

			var args []ExpressionNode
			namedArgs := make(map[string]ExpressionNode)
			sawKeyword := false

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
							sawKeyword = true
						} else {
							if sawKeyword {
								return nil, p.error("positional argument follows keyword argument")
							}
							// This is a positional argument, restore position and parse normally
							p.current = savedPos
							arg, err := p.parseOr() // Use parseOr to avoid conditional parsing issues
							if err != nil {
								return nil, err
							}
							args = append(args, arg)
						}
					} else {
						if sawKeyword {
							return nil, p.error("positional argument follows keyword argument")
						}
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

// isNameToken reports whether tok is a name: an identifier, or a keyword,
// which Jinja2 lexes as a plain name.
func isNameToken(tok *lexer.Token) bool {
	return tok.Type == lexer.TokenIdentifier || (tok.Value != "" && lexer.LookupKeyword(tok.Value) == tok.Type)
}

// parseTest parses "is [not] name", with arguments in parentheses or one
// argument without them, as in "x is divisibleby 3".
func (p *Parser) parseTest(expr ExpressionNode) (ExpressionNode, error) {
	p.advance() // consume 'is'

	negated := false
	if p.check(lexer.TokenNot) {
		p.advance() // consume 'not'
		negated = true
	}

	if !isNameToken(p.peek()) {
		return nil, p.error("expected test name after 'is'")
	}
	nameToken := p.advance()
	testNode := NewTestNode(expr, nameToken.Value, nameToken.Line, nameToken.Column)
	testNode.Negated = negated

	switch {
	case p.check(lexer.TokenLeftParen):
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
	case p.check(lexer.TokenIs):
		return nil, p.error("cannot chain multiple tests with 'is'")
	case p.canStartTestArgument():
		arg, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		arg, err = p.parsePostfixOps(arg, false)
		if err != nil {
			return nil, err
		}
		testNode.Arguments = append(testNode.Arguments, arg)
	}

	return testNode, nil
}

// canStartTestArgument reports whether the current token starts the single
// unparenthesized argument of a test: a name, literal, list or dict, but not
// "else", "or" or "and" (as in Jinja2), nor "if", "recursive", "in" or "not",
// which miya cannot use as names and which end the test in templates miya
// accepts today: {% for v in L if v is iterable recursive %},
// {{ x is odd if t else 0 }}.
func (p *Parser) canStartTestArgument() bool {
	tok := p.peek()
	switch tok.Type {
	case lexer.TokenElse, lexer.TokenOr, lexer.TokenAnd, lexer.TokenIf, lexer.TokenRecursive, lexer.TokenIn, lexer.TokenNot:
		return false
	case lexer.TokenString, lexer.TokenInteger, lexer.TokenFloat, lexer.TokenLeftBracket, lexer.TokenLeftBrace:
		return true
	}
	return isNameToken(tok)
}

// parsePrimary parses primary expressions
func (p *Parser) parsePrimary() (ExpressionNode, error) {
	defer func() { p.depth-- }()
	if err := p.enterNesting(); err != nil {
		return nil, err
	}
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
		value, err := parseIntLiteral(token.Value)
		if err != nil {
			return nil, p.error(fmt.Sprintf("invalid integer: %s", token.Value))
		}
		return AcquireLiteralNode(value, token.Value, token.Line, token.Column), nil

	case lexer.TokenFloat:
		token := p.advance()
		value, err := strconv.ParseFloat(strings.ReplaceAll(token.Value, "_", ""), 64)
		if err != nil {
			return nil, p.error(fmt.Sprintf("invalid float: %s", token.Value))
		}
		return AcquireLiteralNode(value, token.Value, token.Line, token.Column), nil

	case lexer.TokenString:
		token := p.advance()
		value := token.Value
		// Adjacent strings join, as in Jinja2: 'a' 'b' is 'ab'
		for p.check(lexer.TokenString) {
			value += p.advance().Value
		}
		return AcquireLiteralNode(value, value, token.Line, token.Column), nil

	case lexer.TokenIdentifier:
		token := p.advance()
		return AcquireIdentifierNode(token.Value, token.Line, token.Column), nil

	case lexer.TokenSuper:
		token := p.advance()
		return NewSuperNode(token.Line, token.Column), nil

	case lexer.TokenLeftParen:
		open := p.advance() // consume '('
		if p.check(lexer.TokenRightParen) {
			p.advance() // consume ')'
			return NewListNode(nil, open.Line, open.Column), nil
		}
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.check(lexer.TokenComma) {
			// A tuple; miya represents it as a list
			elements := []ExpressionNode{expr}
			for p.check(lexer.TokenComma) {
				p.advance() // consume ','
				if p.check(lexer.TokenRightParen) {
					break
				}
				element, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, element)
			}
			expr = NewListNode(elements, open.Line, open.Column)
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

	case lexer.TokenAnd, lexer.TokenOr, lexer.TokenNot, lexer.TokenIs, lexer.TokenIn, lexer.TokenIf, lexer.TokenElse:
		return nil, p.error(fmt.Sprintf("unexpected token in expression: %s", p.peek().Type))

	default:
		if isNameToken(p.peek()) {
			// Other keywords (block, filter, missing, ...) are plain names here
			token := p.advance()
			return AcquireIdentifierNode(token.Value, token.Line, token.Column), nil
		}
		return nil, p.error(fmt.Sprintf("unexpected token in expression: %s", p.peek().Type))
	}
}

// parseIntLiteral converts an integer token: decimal, or 0x/0o/0b, with
// underscores between digits.
func parseIntLiteral(text string) (int, error) {
	digits := strings.ReplaceAll(text, "_", "")
	if len(digits) > 1 && digits[0] == '0' && strings.ContainsAny(digits[1:2], "xXoObB") {
		v, err := strconv.ParseInt(digits, 0, strconv.IntSize)
		return int(v), err
	}
	return strconv.Atoi(digits)
}

// parseExpressionList parses one expression with parse, or several separated
// by commas, which form a tuple (a list in miya): {% set a, b = 1, 2 %}.
func (p *Parser) parseExpressionList(parse func() (ExpressionNode, error)) (ExpressionNode, error) {
	first, err := parse()
	if err != nil || !p.check(lexer.TokenComma) {
		return first, err
	}
	elements := []ExpressionNode{first}
	for p.check(lexer.TokenComma) {
		p.advance() // consume ','
		element, err := parse()
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	return NewListNode(elements, first.Line(), first.Column()), nil
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

		iterable, err := p.parseOr()
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
		return NewDictNode(nil, nil, startToken.Line, startToken.Column), nil
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

		iterable, err := p.parseOr()
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

	// Regular dictionary literal
	keys := []ExpressionNode{key}
	values := []ExpressionNode{value}

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

			keys = append(keys, key)
			values = append(values, value)

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

	return NewDictNode(keys, values, startToken.Line, startToken.Column), nil
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

	withContext, err := p.parseImportContext()
	if err != nil {
		return nil, err
	}

	// Expect closing %}
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after import statement")
	}
	p.advance() // consume '%}'

	node := NewImportNode(importToken.Line, importToken.Column, template, alias)
	node.WithContext = withContext
	return node, nil
}

// parseImportContext parses an optional "with context" or "without context"
// and reports whether the import is with context.
func (p *Parser) parseImportContext() (bool, error) {
	withContext := false
	if p.check(lexer.TokenWith) {
		withContext = true
	} else if !p.check(lexer.TokenIdentifier) || p.peek().Value != "without" {
		return false, nil
	}
	p.advance() // consume 'with' or 'without'
	if !p.check(lexer.TokenIdentifier) || p.peek().Value != "context" {
		return false, p.error("expected 'context' after 'with' or 'without'")
	}
	p.advance() // consume 'context'
	return withContext, nil
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
			// As in Jinja2, a trailing comma may precede "with/without context".
			if p.check(lexer.TokenWith) || (p.check(lexer.TokenIdentifier) && p.peek().Value == "without" && p.peekNext().Value == "context") {
				break
			}
			continue
		}
		break
	}

	withContext, err := p.parseImportContext()
	if err != nil {
		return nil, err
	}

	// Expect closing %}
	if !p.check(lexer.TokenBlockEnd) && !p.check(lexer.TokenBlockEndTrim) {
		return nil, p.error("expected '%}' after from statement")
	}
	p.advance() // consume '%}'

	node := NewFromNode(fromToken.Line, fromToken.Column, template, names, aliases)
	node.WithContext = withContext
	return node, nil
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

	// Optional caller parameters: {% call(a, b=1) macro() %}
	var params []string
	var defaults map[string]ExpressionNode
	if p.check(lexer.TokenLeftParen) {
		var err error
		params, defaults, err = p.parseParameterList()
		if err != nil {
			return nil, err
		}
	}

	// Parse the call expression (function/macro call)
	callExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if _, ok := callExpr.(*CallNode); !ok {
		return nil, p.error("expected call")
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

	cb := NewCallBlockNode(callExpr, body, startToken.Line, startToken.Column)
	cb.Parameters = params
	cb.Defaults = defaults
	usage, err := findMacroUsage(params, defaults, body)
	if err != nil {
		return nil, p.error(err.Error())
	}
	cb.MacroUsage = usage
	return cb, nil
}

// parseWithStatement parses {% with var=expr, var2=expr2 %}...{% endwith %} statements
func (p *Parser) parseWithStatement() (Node, error) {
	startToken := p.advance() // consume 'with'

	// Parse assignments (var1=expr1, var2=expr2, ...)
	assignments := make(map[string]ExpressionNode)
	var names []string
	var exprs []ExpressionNode

	// A bare {% with %} has no assignments
	for !(len(names) == 0 && (p.check(lexer.TokenBlockEnd) || p.check(lexer.TokenBlockEndTrim))) {
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
		names = append(names, varName)
		exprs = append(exprs, expr)

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

	node := NewWithNode(assignments, body, startToken.Line, startToken.Column)
	node.Names = names
	node.Exprs = exprs
	return node, nil
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
		sawKeyword := false

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
					sawKeyword = true
				} else {
					if sawKeyword {
						return nil, p.error("positional argument follows keyword argument")
					}
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
