package lexer

import "fmt"

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF

	// Literals
	TokenText
	TokenInteger
	TokenFloat
	TokenString
	TokenBoolean
	TokenNone

	// Delimiters
	TokenVarStart     // {{
	TokenVarEnd       // }}
	TokenBlockStart   // {%
	TokenBlockEnd     // %}
	TokenCommentStart // {#
	TokenCommentEnd   // #}

	// Whitespace control
	TokenVarStartTrim   // {{-
	TokenVarEndTrim     // -}}
	TokenBlockStartTrim // {%-
	TokenBlockEndTrim   // -%}

	// Identifiers and Keywords
	TokenIdentifier
	TokenIf
	TokenElif
	TokenElse
	TokenEndif
	TokenFor
	TokenIn
	TokenEndfor
	TokenBlock
	TokenEndblock
	TokenExtends
	TokenInclude
	TokenMacro
	TokenEndmacro
	TokenCall
	TokenEndcall
	TokenImport
	TokenFrom
	TokenAs
	TokenSet
	TokenEndSet
	TokenWith
	TokenEndwith
	TokenFilter
	TokenEndfilter
	TokenRaw
	TokenEndraw
	TokenSuper
	TokenBreak
	TokenContinue
	TokenRecursive
	TokenPass
	TokenDo
	TokenIgnore
	TokenMissing
	TokenTrue
	TokenFalse
	TokenNoneKeyword
	TokenAutoescape
	TokenEndautoescape

	// Operators
	TokenAssign      // =
	TokenPlus        // +
	TokenMinus       // -
	TokenMultiply    // *
	TokenDivide      // /
	TokenFloorDivide // //
	TokenModulo      // %
	TokenPower       // **
	TokenPipe        // |
	TokenTilde       // ~
	TokenConcat      // ~

	// Comparison
	TokenEqual        // ==
	TokenNotEqual     // !=
	TokenLess         // <
	TokenLessEqual    // <=
	TokenGreater      // >
	TokenGreaterEqual // >=

	// Logical
	TokenAnd // and
	TokenOr  // or
	TokenNot // not
	TokenIs  // is

	// Punctuation
	TokenDot          // .
	TokenComma        // ,
	TokenColon        // :
	TokenSemicolon    // ;
	TokenLeftParen    // (
	TokenRightParen   // )
	TokenLeftBracket  // [
	TokenRightBracket // ]
	TokenLeftBrace    // {
	TokenRightBrace   // }
)

var tokenNames = map[TokenType]string{
	TokenError:          "ERROR",
	TokenEOF:            "EOF",
	TokenText:           "TEXT",
	TokenInteger:        "INTEGER",
	TokenFloat:          "FLOAT",
	TokenString:         "STRING",
	TokenBoolean:        "BOOLEAN",
	TokenNone:           "NONE",
	TokenVarStart:       "VAR_START",
	TokenVarEnd:         "VAR_END",
	TokenBlockStart:     "BLOCK_START",
	TokenBlockEnd:       "BLOCK_END",
	TokenCommentStart:   "COMMENT_START",
	TokenCommentEnd:     "COMMENT_END",
	TokenVarStartTrim:   "VAR_START_TRIM",
	TokenVarEndTrim:     "VAR_END_TRIM",
	TokenBlockStartTrim: "BLOCK_START_TRIM",
	TokenBlockEndTrim:   "BLOCK_END_TRIM",
	TokenIdentifier:     "IDENTIFIER",
	TokenIf:             "IF",
	TokenElif:           "ELIF",
	TokenElse:           "ELSE",
	TokenEndif:          "ENDIF",
	TokenFor:            "FOR",
	TokenIn:             "IN",
	TokenEndfor:         "ENDFOR",
	TokenBlock:          "BLOCK",
	TokenEndblock:       "ENDBLOCK",
	TokenExtends:        "EXTENDS",
	TokenInclude:        "INCLUDE",
	TokenMacro:          "MACRO",
	TokenEndmacro:       "ENDMACRO",
	TokenCall:           "CALL",
	TokenEndcall:        "ENDCALL",
	TokenImport:         "IMPORT",
	TokenFrom:           "FROM",
	TokenAs:             "AS",
	TokenSet:            "SET",
	TokenEndSet:         "ENDSET",
	TokenWith:           "WITH",
	TokenEndwith:        "ENDWITH",
	TokenFilter:         "FILTER",
	TokenEndfilter:      "ENDFILTER",
	TokenRaw:            "RAW",
	TokenEndraw:         "ENDRAW",
	TokenSuper:          "SUPER",
	TokenBreak:          "BREAK",
	TokenContinue:       "CONTINUE",
	TokenRecursive:      "RECURSIVE",
	TokenPass:           "PASS",
	TokenDo:             "DO",
	TokenIgnore:         "IGNORE",
	TokenMissing:        "MISSING",
	TokenTrue:           "TRUE",
	TokenFalse:          "FALSE",
	TokenNoneKeyword:    "NONE_KEYWORD",
	TokenAutoescape:     "AUTOESCAPE",
	TokenEndautoescape:  "ENDAUTOESCAPE",
	TokenAssign:         "ASSIGN",
	TokenPlus:           "PLUS",
	TokenMinus:          "MINUS",
	TokenMultiply:       "MULTIPLY",
	TokenDivide:         "DIVIDE",
	TokenFloorDivide:    "FLOOR_DIVIDE",
	TokenModulo:         "MODULO",
	TokenPower:          "POWER",
	TokenPipe:           "PIPE",
	TokenTilde:          "TILDE",
	TokenConcat:         "CONCAT",
	TokenEqual:          "EQUAL",
	TokenNotEqual:       "NOT_EQUAL",
	TokenLess:           "LESS",
	TokenLessEqual:      "LESS_EQUAL",
	TokenGreater:        "GREATER",
	TokenGreaterEqual:   "GREATER_EQUAL",
	TokenAnd:            "AND",
	TokenOr:             "OR",
	TokenNot:            "NOT",
	TokenIs:             "IS",
	TokenDot:            "DOT",
	TokenComma:          "COMMA",
	TokenColon:          "COLON",
	TokenSemicolon:      "SEMICOLON",
	TokenLeftParen:      "LEFT_PAREN",
	TokenRightParen:     "RIGHT_PAREN",
	TokenLeftBracket:    "LEFT_BRACKET",
	TokenRightBracket:   "RIGHT_BRACKET",
	TokenLeftBrace:      "LEFT_BRACE",
	TokenRightBrace:     "RIGHT_BRACE",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", t)
}

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int

	// For whitespace control
	TrimLeft  bool
	TrimRight bool
}

func NewToken(typ TokenType, value string, line, column int) *Token {
	return &Token{
		Type:   typ,
		Value:  value,
		Line:   line,
		Column: column,
	}
}

func (t *Token) String() string {
	if t.Value != "" {
		return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Value, t.Line, t.Column)
	}
	return fmt.Sprintf("%s at %d:%d", t.Type, t.Line, t.Column)
}

var keywords = map[string]TokenType{
	"if":            TokenIf,
	"elif":          TokenElif,
	"else":          TokenElse,
	"endif":         TokenEndif,
	"for":           TokenFor,
	"in":            TokenIn,
	"endfor":        TokenEndfor,
	"block":         TokenBlock,
	"endblock":      TokenEndblock,
	"extends":       TokenExtends,
	"include":       TokenInclude,
	"macro":         TokenMacro,
	"endmacro":      TokenEndmacro,
	"call":          TokenCall,
	"endcall":       TokenEndcall,
	"import":        TokenImport,
	"from":          TokenFrom,
	"as":            TokenAs,
	"set":           TokenSet,
	"endset":        TokenEndSet,
	"with":          TokenWith,
	"endwith":       TokenEndwith,
	"filter":        TokenFilter,
	"endfilter":     TokenEndfilter,
	"raw":           TokenRaw,
	"endraw":        TokenEndraw,
	"autoescape":    TokenAutoescape,
	"endautoescape": TokenEndautoescape,
	"super":         TokenSuper,
	"break":         TokenBreak,
	"continue":      TokenContinue,
	"recursive":     TokenRecursive,
	"pass":          TokenPass,
	"do":            TokenDo,
	"ignore":        TokenIgnore,
	"missing":       TokenMissing,
	"and":           TokenAnd,
	"or":            TokenOr,
	"not":           TokenNot,
	"is":            TokenIs,
	"true":          TokenTrue,
	"True":          TokenTrue,
	"false":         TokenFalse,
	"False":         TokenFalse,
	"none":          TokenNoneKeyword,
	"None":          TokenNoneKeyword,
}

func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdentifier
}
