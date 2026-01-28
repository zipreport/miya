package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type LexerConfig struct {
	VarStartString     string
	VarEndString       string
	BlockStartString   string
	BlockEndString     string
	CommentStartString string
	CommentEndString   string
	TrimBlocks         bool
	LstripBlocks       bool
}

func DefaultConfig() *LexerConfig {
	return &LexerConfig{
		VarStartString:     "{{",
		VarEndString:       "}}",
		BlockStartString:   "{%",
		BlockEndString:     "%}",
		CommentStartString: "{#",
		CommentEndString:   "#}",
	}
}

type Lexer struct {
	input  string
	config *LexerConfig

	pos     int // current position in input
	readPos int // current reading position (after current char)
	line    int
	column  int

	ch byte // current char

	state lexerState
}

type lexerState int

const (
	stateText lexerState = iota
	stateVariable
	stateBlock
	stateComment
)

func NewLexer(input string, config *LexerConfig) *Lexer {
	if config == nil {
		config = DefaultConfig()
	}

	l := &Lexer{
		input:  input,
		config: config,
		line:   1,
		column: 1,
		state:  stateText,
	}

	l.readChar()
	return l
}

func (l *Lexer) NextToken() (*Token, error) {
	switch l.state {
	case stateText:
		return l.lexText()
	case stateVariable:
		return l.lexVariable()
	case stateBlock:
		return l.lexBlock()
	case stateComment:
		return l.lexComment()
	default:
		return nil, fmt.Errorf("unexpected lexer state: %v", l.state)
	}
}

func (l *Lexer) Tokenize() ([]*Token, error) {
	// Pre-allocate with estimated capacity based on input length
	// Average template has roughly 1 token per 20-30 characters
	estimatedTokens := len(l.input)/20 + 16
	tokens := make([]*Token, 0, estimatedTokens)

	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, tok)

		if tok.Type == TokenEOF {
			break
		}
	}

	return tokens, nil
}

func (l *Lexer) lexText() (*Token, error) {
	if l.ch == 0 {
		return l.makeToken(TokenEOF, ""), nil
	}

	startPos := l.pos
	startLine := l.line
	startColumn := l.column

	for l.ch != 0 {
		// Check for delimiter starts
		if l.peekString(l.config.VarStartString) {
			if l.pos > startPos {
				// Return accumulated text
				text := l.input[startPos:l.pos]
				return &Token{
					Type:   TokenText,
					Value:  text,
					Line:   startLine,
					Column: startColumn,
				}, nil
			}
			// Switch to variable mode
			return l.lexVarStart()
		}

		if l.peekString(l.config.BlockStartString) {
			if l.pos > startPos {
				// Return accumulated text
				text := l.input[startPos:l.pos]
				return &Token{
					Type:   TokenText,
					Value:  text,
					Line:   startLine,
					Column: startColumn,
				}, nil
			}
			// Switch to block mode
			return l.lexBlockStart()
		}

		if l.peekString(l.config.CommentStartString) {
			if l.pos > startPos {
				// Return accumulated text
				text := l.input[startPos:l.pos]
				return &Token{
					Type:   TokenText,
					Value:  text,
					Line:   startLine,
					Column: startColumn,
				}, nil
			}
			// Switch to comment mode
			return l.lexCommentStart()
		}

		l.readChar()
	}

	// Return any remaining text
	if l.pos > startPos {
		text := l.input[startPos:l.pos]
		return &Token{
			Type:   TokenText,
			Value:  text,
			Line:   startLine,
			Column: startColumn,
		}, nil
	}

	return l.makeToken(TokenEOF, ""), nil
}

func (l *Lexer) lexVarStart() (*Token, error) {
	line := l.line
	column := l.column

	// Check for trim variant {{-
	trimRight := false
	if l.peekString(l.config.VarStartString + "-") {
		l.consumeString(l.config.VarStartString + "-")
		trimRight = true
		l.state = stateVariable
		return &Token{
			Type:      TokenVarStartTrim,
			Value:     l.config.VarStartString + "-",
			Line:      line,
			Column:    column,
			TrimRight: trimRight,
		}, nil
	}

	l.consumeString(l.config.VarStartString)
	l.state = stateVariable
	return &Token{
		Type:   TokenVarStart,
		Value:  l.config.VarStartString,
		Line:   line,
		Column: column,
	}, nil
}

func (l *Lexer) lexBlockStart() (*Token, error) {
	line := l.line
	column := l.column

	// Check for trim variant {%-
	trimRight := false
	if l.peekString(l.config.BlockStartString + "-") {
		l.consumeString(l.config.BlockStartString + "-")
		trimRight = true
		l.state = stateBlock
		return &Token{
			Type:      TokenBlockStartTrim,
			Value:     l.config.BlockStartString + "-",
			Line:      line,
			Column:    column,
			TrimRight: trimRight,
		}, nil
	}

	l.consumeString(l.config.BlockStartString)
	l.state = stateBlock
	return &Token{
		Type:   TokenBlockStart,
		Value:  l.config.BlockStartString,
		Line:   line,
		Column: column,
	}, nil
}

func (l *Lexer) lexCommentStart() (*Token, error) {
	startLine := l.line
	startColumn := l.column

	l.consumeString(l.config.CommentStartString)

	// Skip everything until comment end
	for !l.peekString(l.config.CommentEndString) && l.ch != 0 {
		l.readChar()
	}

	if l.ch == 0 {
		return nil, fmt.Errorf("unclosed comment at line %d, column %d", startLine, startColumn)
	}

	l.consumeString(l.config.CommentEndString)
	l.state = stateText

	// Comments are skipped, continue to next token
	return l.NextToken()
}

func (l *Lexer) lexVariable() (*Token, error) {
	l.skipWhitespace()

	// Check for variable end
	if l.peekString("-" + l.config.VarEndString) {
		line := l.line
		column := l.column
		l.consumeString("-" + l.config.VarEndString)
		l.state = stateText
		return &Token{
			Type:     TokenVarEndTrim,
			Value:    "-" + l.config.VarEndString,
			Line:     line,
			Column:   column,
			TrimLeft: true,
		}, nil
	}

	if l.peekString(l.config.VarEndString) {
		line := l.line
		column := l.column
		l.consumeString(l.config.VarEndString)
		l.state = stateText
		return &Token{
			Type:   TokenVarEnd,
			Value:  l.config.VarEndString,
			Line:   line,
			Column: column,
		}, nil
	}

	return l.lexExpression()
}

func (l *Lexer) lexBlock() (*Token, error) {
	l.skipWhitespace()

	// Check for block end
	if l.peekString("-" + l.config.BlockEndString) {
		line := l.line
		column := l.column
		l.consumeString("-" + l.config.BlockEndString)
		l.state = stateText
		return &Token{
			Type:     TokenBlockEndTrim,
			Value:    "-" + l.config.BlockEndString,
			Line:     line,
			Column:   column,
			TrimLeft: true,
		}, nil
	}

	if l.peekString(l.config.BlockEndString) {
		line := l.line
		column := l.column
		l.consumeString(l.config.BlockEndString)
		l.state = stateText
		return &Token{
			Type:   TokenBlockEnd,
			Value:  l.config.BlockEndString,
			Line:   line,
			Column: column,
		}, nil
	}

	return l.lexExpression()
}

func (l *Lexer) lexComment() (*Token, error) {
	// Comments are handled in lexCommentStart
	return nil, fmt.Errorf("unexpected call to lexComment")
}

func (l *Lexer) lexExpression() (*Token, error) {
	l.skipWhitespace()

	line := l.line
	column := l.column

	switch l.ch {
	case 0:
		return nil, fmt.Errorf("unexpected EOF in expression at line %d, column %d", line, column)

	// Single character tokens
	case '=':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return l.makeTokenAt(TokenEqual, "==", line, column), nil
		}
		return l.makeTokenAt(TokenAssign, "=", line, column), nil

	case '+':
		l.readChar()
		return l.makeTokenAt(TokenPlus, "+", line, column), nil

	case '-':
		// Could be minus or start of end delimiter trim
		if l.peekString("-"+l.config.VarEndString) || l.peekString("-"+l.config.BlockEndString) {
			// This will be handled by lexVariable or lexBlock
			return nil, nil
		}
		l.readChar()
		return l.makeTokenAt(TokenMinus, "-", line, column), nil

	case '*':
		l.readChar()
		if l.ch == '*' {
			l.readChar()
			return l.makeTokenAt(TokenPower, "**", line, column), nil
		}
		return l.makeTokenAt(TokenMultiply, "*", line, column), nil

	case '/':
		l.readChar()
		if l.ch == '/' {
			l.readChar()
			return l.makeTokenAt(TokenFloorDivide, "//", line, column), nil
		}
		return l.makeTokenAt(TokenDivide, "/", line, column), nil

	case '%':
		l.readChar()
		return l.makeTokenAt(TokenModulo, "%", line, column), nil

	case '|':
		l.readChar()
		return l.makeTokenAt(TokenPipe, "|", line, column), nil

	case '~':
		l.readChar()
		return l.makeTokenAt(TokenTilde, "~", line, column), nil

	case '<':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return l.makeTokenAt(TokenLessEqual, "<=", line, column), nil
		}
		return l.makeTokenAt(TokenLess, "<", line, column), nil

	case '>':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return l.makeTokenAt(TokenGreaterEqual, ">=", line, column), nil
		}
		return l.makeTokenAt(TokenGreater, ">", line, column), nil

	case '!':
		l.readChar()
		if l.ch == '=' {
			l.readChar()
			return l.makeTokenAt(TokenNotEqual, "!=", line, column), nil
		}
		return nil, fmt.Errorf("unexpected character '!' at line %d, column %d", line, column)

	case '.':
		l.readChar()
		return l.makeTokenAt(TokenDot, ".", line, column), nil

	case ',':
		l.readChar()
		return l.makeTokenAt(TokenComma, ",", line, column), nil

	case ':':
		l.readChar()
		return l.makeTokenAt(TokenColon, ":", line, column), nil

	case ';':
		l.readChar()
		return l.makeTokenAt(TokenSemicolon, ";", line, column), nil

	case '(':
		l.readChar()
		return l.makeTokenAt(TokenLeftParen, "(", line, column), nil

	case ')':
		l.readChar()
		return l.makeTokenAt(TokenRightParen, ")", line, column), nil

	case '[':
		l.readChar()
		return l.makeTokenAt(TokenLeftBracket, "[", line, column), nil

	case ']':
		l.readChar()
		return l.makeTokenAt(TokenRightBracket, "]", line, column), nil

	case '{':
		l.readChar()
		return l.makeTokenAt(TokenLeftBrace, "{", line, column), nil

	case '}':
		l.readChar()
		return l.makeTokenAt(TokenRightBrace, "}", line, column), nil

	case '"', '\'':
		return l.lexString()

	default:
		if isDigit(l.ch) {
			return l.lexNumber()
		}
		if isLetter(l.ch) || l.ch == '_' {
			return l.lexIdentifier()
		}
	}

	return nil, fmt.Errorf("unexpected character %q at line %d, column %d", l.ch, line, column)
}

func (l *Lexer) lexString() (*Token, error) {
	line := l.line
	column := l.column
	quote := l.ch
	l.readChar()

	var sb strings.Builder

	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"', '\'':
				sb.WriteByte(l.ch)
			default:
				sb.WriteByte(l.ch)
			}
		} else {
			sb.WriteByte(l.ch)
		}
		l.readChar()
	}

	if l.ch != quote {
		return nil, fmt.Errorf("unterminated string at line %d, column %d", line, column)
	}

	l.readChar() // consume closing quote

	return &Token{
		Type:   TokenString,
		Value:  sb.String(),
		Line:   line,
		Column: column,
	}, nil
}

func (l *Lexer) lexNumber() (*Token, error) {
	line := l.line
	column := l.column
	startPos := l.pos

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}

		// Check for scientific notation
		if l.ch == 'e' || l.ch == 'E' {
			l.readChar()
			if l.ch == '+' || l.ch == '-' {
				l.readChar()
			}
			for isDigit(l.ch) {
				l.readChar()
			}
		}

		return &Token{
			Type:   TokenFloat,
			Value:  l.input[startPos:l.pos],
			Line:   line,
			Column: column,
		}, nil
	}

	return &Token{
		Type:   TokenInteger,
		Value:  l.input[startPos:l.pos],
		Line:   line,
		Column: column,
	}, nil
}

func (l *Lexer) lexIdentifier() (*Token, error) {
	line := l.line
	column := l.column
	startPos := l.pos

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	ident := l.input[startPos:l.pos]
	tokenType := LookupKeyword(ident)

	return &Token{
		Type:   tokenType,
		Value:  ident,
		Line:   line,
		Column: column,
	}, nil
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
		if l.ch == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) peekString(s string) bool {
	if l.pos+len(s) > len(l.input) {
		return false
	}
	return l.input[l.pos:l.pos+len(s)] == s
}

func (l *Lexer) consumeString(s string) {
	for i := 0; i < len(s); i++ {
		l.readChar()
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) makeToken(typ TokenType, value string) *Token {
	return &Token{
		Type:   typ,
		Value:  value,
		Line:   l.line,
		Column: l.column,
	}
}

func (l *Lexer) makeTokenAt(typ TokenType, value string, line, column int) *Token {
	return &Token{
		Type:   typ,
		Value:  value,
		Line:   line,
		Column: column,
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlphaNumeric(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
