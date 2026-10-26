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
	// KeepTrailingNewline keeps the single newline at the end of the source,
	// which is otherwise removed as in Jinja2.
	KeepTrailingNewline bool
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

	// pending holds tokens already produced by lexRaw.
	pending []*Token
	// Whitespace control for the next text, set by the tag that ended before it.
	stripNext   bool // the tag ended with "-": strip all leading whitespace
	trimNewline bool // trim_blocks applies: drop one leading newline
	sawTag      bool // false until the first tag, so the first text starts a line
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

	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	if !config.KeepTrailingNewline {
		input = strings.TrimSuffix(input, "\n")
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
	if len(l.pending) > 0 {
		tok := l.pending[0]
		l.pending = l.pending[1:]
		return tok, nil
	}
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
	lineStart := l.skipLeadingWhitespace()

	startPos := l.pos
	startLine := l.line
	startColumn := l.column

	for l.ch != 0 {
		var next func() (*Token, error)
		delim, lstrip := "", false
		switch {
		case l.peekString(l.config.VarStartString):
			next, delim = l.lexVarStart, l.config.VarStartString
		case l.peekString(l.config.BlockStartString):
			next, delim, lstrip = l.lexBlockStart, l.config.BlockStartString, true
		case l.peekString(l.config.CommentStartString):
			next, delim, lstrip = l.lexCommentStart, l.config.CommentStartString, true
		}
		if next == nil {
			l.readChar()
			continue
		}
		marker := byte(0)
		if l.pos+len(delim) < len(l.input) {
			marker = l.input[l.pos+len(delim)]
		}
		text := l.trimTrailing(l.input[startPos:l.pos], marker, lstrip, lineStart)
		if text == "" {
			return next()
		}
		return &Token{Type: TokenText, Value: text, Line: startLine, Column: startColumn}, nil
	}

	if l.pos > startPos {
		return &Token{Type: TokenText, Value: l.input[startPos:l.pos], Line: startLine, Column: startColumn}, nil
	}
	return l.makeToken(TokenEOF, ""), nil
}

// skipLeadingWhitespace applies the previous tag's "-" or trim_blocks to the
// text that follows it, and reports whether that text starts a line.
func (l *Lexer) skipLeadingWhitespace() bool {
	lineStart := !l.sawTag
	if l.stripNext {
		for {
			r, size := utf8.DecodeRuneInString(l.input[l.pos:])
			if l.ch == 0 || !unicode.IsSpace(r) {
				break
			}
			lineStart = r == '\n'
			for i := 0; i < size; i++ {
				l.readChar()
			}
		}
	} else if l.trimNewline && l.ch == '\n' {
		l.readChar()
		lineStart = true
	}
	l.stripNext, l.trimNewline = false, false
	return lineStart
}

// trimTrailing applies the next tag's marker ("-", "+" or none) and
// lstrip_blocks to the text before it. lstrip is false for {{ tags, which
// lstrip_blocks does not affect.
func (l *Lexer) trimTrailing(text string, marker byte, lstrip, lineStart bool) string {
	if marker == '-' {
		return strings.TrimRightFunc(text, unicode.IsSpace)
	}
	if marker == '+' || !lstrip || !l.config.LstripBlocks {
		return text
	}
	i := strings.LastIndexByte(text, '\n') + 1
	if (i > 0 || lineStart) && strings.TrimLeftFunc(text[i:], unicode.IsSpace) == "" {
		return text[:i]
	}
	return text
}

func (l *Lexer) lexVarStart() (*Token, error) {
	line := l.line
	column := l.column
	l.sawTag = true

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
	l.sawTag = true

	if ok, err := l.lexRaw(); ok || err != nil {
		if err != nil {
			return nil, err
		}
		return l.NextToken()
	}

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
	if l.ch == '+' {
		l.readChar()
	}
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
	l.sawTag = true

	l.consumeString(l.config.CommentStartString)
	bodyStart := l.pos

	// Skip everything until comment end
	for !l.peekString(l.config.CommentEndString) && l.ch != 0 {
		l.readChar()
	}

	if l.ch == 0 {
		return nil, fmt.Errorf("unclosed comment at line %d, column %d", startLine, startColumn)
	}

	marker := byte(0)
	if l.pos > bodyStart+1 {
		marker = l.input[l.pos-1]
	}
	l.consumeString(l.config.CommentEndString)
	l.state = stateText
	l.setTagEnd(marker, true)

	// Comments are skipped, continue to next token
	return l.NextToken()
}

// setTagEnd records how the tag that just ended controls the following text.
func (l *Lexer) setTagEnd(marker byte, trimmable bool) {
	l.stripNext = marker == '-'
	l.trimNewline = marker != '-' && marker != '+' && trimmable && l.config.TrimBlocks
}

func (l *Lexer) lexVariable() (*Token, error) {
	l.skipWhitespace()

	// Check for variable end
	if l.peekString("-" + l.config.VarEndString) {
		line := l.line
		column := l.column
		l.consumeString("-" + l.config.VarEndString)
		l.state = stateText
		l.setTagEnd('-', false)
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
		l.setTagEnd('-', true)
		return &Token{
			Type:     TokenBlockEndTrim,
			Value:    "-" + l.config.BlockEndString,
			Line:     line,
			Column:   column,
			TrimLeft: true,
		}, nil
	}

	if l.peekString("+" + l.config.BlockEndString) {
		line := l.line
		column := l.column
		l.consumeString("+" + l.config.BlockEndString)
		l.state = stateText
		l.setTagEnd('+', true)
		return &Token{
			Type:   TokenBlockEnd,
			Value:  l.config.BlockEndString,
			Line:   line,
			Column: column,
		}, nil
	}

	if l.peekString(l.config.BlockEndString) {
		line := l.line
		column := l.column
		l.consumeString(l.config.BlockEndString)
		l.state = stateText
		l.setTagEnd(0, true)
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
			// The matching end delimiter is handled by lexVariable/lexBlock before
			// reaching here, so this is the wrong end delimiter for the open tag.
			return nil, fmt.Errorf("unexpected '-' before end delimiter at line %d, column %d", line, column)
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
		if r, _ := utf8.DecodeRuneInString(l.input[l.pos:]); isIdentStart(r) {
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

	// 0x1F, 0o17, 0b101, with optional underscores after the prefix
	if l.ch == '0' {
		if base := baseDigits(l.peekChar()); base != nil {
			if p := l.pos + 2; p < len(l.input) && (base(l.input[p]) || l.input[p] == '_' && p+1 < len(l.input) && base(l.input[p+1])) {
				l.readChar()
				l.readChar()
				l.readDigits(base)
				return &Token{Type: TokenInteger, Value: l.input[startPos:l.pos], Line: line, Column: column}, nil
			}
		}
	}

	l.readDigits(isDigit)
	tokenType := TokenInteger
	// After a "." (L.0.1) a number is an integer index, never a float, as
	// in Jinja2's lexer (its float pattern starts with (?<!\.))
	afterDot := startPos > 0 && l.input[startPos-1] == '.'
	if !afterDot && l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		l.readDigits(isDigit)
		tokenType = TokenFloat
	}
	if !afterDot && (l.ch == 'e' || l.ch == 'E') {
		p := l.pos + 1
		if p < len(l.input) && (l.input[p] == '+' || l.input[p] == '-') {
			p++
		}
		if p < len(l.input) && isDigit(l.input[p]) {
			for l.pos < p {
				l.readChar()
			}
			l.readDigits(isDigit)
			tokenType = TokenFloat
		}
	}

	return &Token{Type: tokenType, Value: l.input[startPos:l.pos], Line: line, Column: column}, nil
}

// readDigits consumes digits accepted by digit, allowing a single "_"
// between two digits, as Jinja2 does.
func (l *Lexer) readDigits(digit func(byte) bool) {
	for digit(l.ch) || l.ch == '_' && digit(l.peekChar()) {
		l.readChar()
	}
}

// baseDigits returns the digit test for the base prefix letter after a "0"
// (x, o or b, any case), or nil.
func baseDigits(c byte) func(byte) bool {
	switch c {
	case 'x', 'X':
		return func(b byte) bool { return isDigit(b) || 'a' <= b && b <= 'f' || 'A' <= b && b <= 'F' }
	case 'o', 'O':
		return func(b byte) bool { return '0' <= b && b <= '7' }
	case 'b', 'B':
		return func(b byte) bool { return b == '0' || b == '1' }
	}
	return nil
}

func (l *Lexer) lexIdentifier() (*Token, error) {
	line := l.line
	column := l.column
	startPos := l.pos

	for {
		r, size := utf8.DecodeRuneInString(l.input[l.pos:])
		if l.ch == 0 || !isIdentPart(r) {
			break
		}
		for i := 0; i < size; i++ {
			l.readChar()
		}
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

// isIdentStart and isIdentPart approximate Python identifiers, which Jinja2
// uses: a letter or "_", then letters, digits, "_" and combining marks.
// Unlike Python, names are not NFKC-normalized, and letter numbers (Ⅻ) and
// "·" are rejected.
func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r) || unicode.In(r, unicode.Mn, unicode.Mc, unicode.Pc)
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

// lexRaw lexes {% raw %}...{% endraw %} when the block starting at l.pos is a
// raw tag. The content becomes one TokenText, verbatim apart from whitespace
// control, and the tokens are queued in l.pending.
func (l *Lexer) lexRaw() (bool, error) {
	bs, be := l.config.BlockStartString, l.config.BlockEndString
	i := l.pos + len(bs)
	if i < len(l.input) && (l.input[i] == '-' || l.input[i] == '+') {
		i++ // the opening marker applies to the text before the tag, handled by lexText
	}
	j := skipSpaces(l.input, i)
	if !strings.HasPrefix(l.input[j:], "raw") {
		return false, nil
	}
	j = skipSpaces(l.input, j+len("raw"))
	if strings.HasPrefix(l.input[j:], "+"+be) {
		// Jinja2 rejects {% raw +%}: the raw content's leading whitespace is never lstripped
		return false, fmt.Errorf("unexpected '+' before %q in raw tag at line %d, column %d", be, l.line, l.column)
	}
	closeMarker := byte(0)
	if strings.HasPrefix(l.input[j:], "-"+be) {
		closeMarker = '-'
		j++
	} else if !strings.HasPrefix(l.input[j:], be) {
		return false, nil
	}
	line, column := l.line, l.column
	for l.pos < j+len(be) {
		l.readChar()
	}
	l.pending = append(l.pending,
		&Token{Type: TokenBlockStart, Value: bs, Line: line, Column: column},
		&Token{Type: TokenRaw, Value: "raw", Line: line, Column: column},
		&Token{Type: TokenBlockEnd, Value: be, Line: line, Column: column})

	// Find {% endraw %}
	end, endTag, endMarkerL, endMarkerR := -1, 0, byte(0), byte(0)
	for k := l.pos; k < len(l.input); k++ {
		if !strings.HasPrefix(l.input[k:], bs) {
			continue
		}
		m := k + len(bs)
		ml := byte(0)
		if m < len(l.input) && (l.input[m] == '-' || l.input[m] == '+') {
			ml = l.input[m]
			m++
		}
		m = skipSpaces(l.input, m)
		if !strings.HasPrefix(l.input[m:], "endraw") {
			continue
		}
		m = skipSpaces(l.input, m+len("endraw"))
		mr := byte(0)
		if m < len(l.input) && (l.input[m] == '-' || l.input[m] == '+') {
			mr = l.input[m]
			m++
		}
		if !strings.HasPrefix(l.input[m:], be) {
			continue
		}
		end, endTag, endMarkerL, endMarkerR = k, m+len(be), ml, mr
		break
	}
	if end < 0 {
		return false, fmt.Errorf("unclosed raw block at line %d, column %d", line, column)
	}

	l.stripNext, l.trimNewline = closeMarker == '-', false
	lineStart := l.skipLeadingWhitespace()
	contentLine, contentColumn := l.line, l.column
	content := l.trimTrailing(l.input[l.pos:end], endMarkerL, true, lineStart)
	if content != "" {
		l.pending = append(l.pending, &Token{Type: TokenText, Value: content, Line: contentLine, Column: contentColumn})
	}
	for l.pos < end {
		l.readChar()
	}
	endLine, endColumn := l.line, l.column
	for l.pos < endTag {
		l.readChar()
	}
	l.pending = append(l.pending,
		&Token{Type: TokenBlockStart, Value: bs, Line: endLine, Column: endColumn},
		&Token{Type: TokenEndraw, Value: "endraw", Line: endLine, Column: endColumn},
		&Token{Type: TokenBlockEnd, Value: be, Line: endLine, Column: endColumn})
	l.state = stateText
	l.setTagEnd(endMarkerR, true)
	return true, nil
}

func skipSpaces(s string, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	return i
}
