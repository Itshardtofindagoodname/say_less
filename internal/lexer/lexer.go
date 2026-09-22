package lexer

import (
	"fmt"
	"unicode"
)

// Token types
type TokenType int

const (
	// Literals
	INT TokenType = iota
	FLOAT
	STRING
	IDENT
	TRUE
	FALSE
	NONE

	// Operators
	PLUS       // +
	MINUS      // -
	STAR       // *
	SLASH      // /
	PERCENT    // %
	ASSIGN     // =
	PLUS_EQ    // +=
	MINUS_EQ   // -=
	STAR_EQ    // *=
	SLASH_EQ   // /=
	PERCENT_EQ // %=
	EQ         // ==
	NEQ        // !=
	LT         // <
	GT         // >
	LTE        // <=
	GTE        // >=
	AND        // and
	OR         // or
	NOT        // not
	ARROW      // ->
	DOT        // .
	DOTDOT     // ..

	// Delimiters
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	LBRACE   // {
	RBRACE   // }
	COMMA    // ,
	COLON    // :
	PIPE     // |

	// Keywords
	FN       // fn
	IF       // if
	ELSE     // else
	ELSEIF   // else_if
	WHILE    // while
	FOR      // for
	IN       // in
	RETURN   // return
	BREAK    // break
	CONTINUE // continue
	USE      // use
	AS       // as
	STRUCT   // struct
	TEST     // test
	ASSERT   // assert
	LET      // let
	MUT      // mut
	CONST    // const
	MATCH    // match
	WHEN     // when
	SELF     // self
	GET      // get
	POST     // post
	PUT      // put
	DELETE   // delete
	PATCH    // patch
	ON       // on
	SERVER   // server

	// Web tokens
	PAGE      // page
	COMPONENT // component
	STATE     // state
	STYLE     // style
	AWAIT     // await
	TRY       // try
	CATCH     // catch
	THROW     // throw

	NEWLINE
	INDENT
	DEDENT
	EOF
)

var keywords = map[string]TokenType{
	"fn":        FN,
	"if":        IF,
	"else":      ELSE,
	"otherwise": ELSE,
	"else_if":   ELSEIF,
	"while":     WHILE,
	"loop":      WHILE,
	"for":       FOR,
	"each":      FOR,
	"in":        IN,
	"return":    RETURN,
	"give":      RETURN,
	"break":     BREAK,
	"continue":  CONTINUE,
	"use":       USE,
	"as":        AS,
	"struct":    STRUCT,
	"test":      TEST,
	"assert":    ASSERT,
	"let":       LET,
	"mut":       MUT,
	"const":     CONST,
	"match":     MATCH,
	"when":      WHEN,
	"self":      SELF,
	"get":       GET,
	"post":      POST,
	"put":       PUT,
	"delete":    DELETE,
	"patch":     PATCH,
	"on":        ON,
	"server":    SERVER,
	"page":      PAGE,
	"component": COMPONENT,
	"state":     STATE,
	"style":     STYLE,
	"await":     AWAIT,
	"try":       TRY,
	"catch":     CATCH,
	"throw":     THROW,
	"true":      TRUE,
	"false":     FALSE,
	"none":      NONE,
	"and":       AND,
	"or":        OR,
	"not":       NOT,
}

var tokenNames = map[TokenType]string{
	INT: "INT", FLOAT: "FLOAT", STRING: "STRING", IDENT: "IDENT",
	TRUE: "TRUE", FALSE: "FALSE", NONE: "NONE",
	PLUS: "+", MINUS: "-", STAR: "*", SLASH: "/", PERCENT: "%",
	ASSIGN: "=", PLUS_EQ: "+=", MINUS_EQ: "-=", STAR_EQ: "*=", SLASH_EQ: "/=", PERCENT_EQ: "%=",
	EQ: "==", NEQ: "!=", LT: "<", GT: ">", LTE: "<=", GTE: ">=",
	AND: "and", OR: "or", NOT: "not", ARROW: "->", DOT: ".", DOTDOT: "..",
	LPAREN: "(", RPAREN: ")", LBRACKET: "[", RBRACKET: "]", LBRACE: "{", RBRACE: "}",
	COMMA: ",", COLON: ":", PIPE: "|",
	FN: "fn", IF: "if", ELSE: "else", ELSEIF: "else_if", WHILE: "while",
	FOR: "for", IN: "in", RETURN: "return", BREAK: "break", CONTINUE: "continue",
	USE: "use", AS: "as", STRUCT: "struct", TEST: "test", ASSERT: "assert",
	LET: "let", MUT: "mut", CONST: "const", MATCH: "match", WHEN: "when",
	SELF: "self", GET: "get", POST: "post", PUT: "put", DELETE: "delete",
	PATCH: "patch", ON: "on", SERVER: "server",
	PAGE: "page", COMPONENT: "component", STATE: "state", STYLE: "style",
	AWAIT: "await", TRY: "try", CATCH: "catch", THROW: "throw",
	NEWLINE: "NEWLINE", INDENT: "INDENT", DEDENT: "DEDENT", EOF: "EOF",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TOKEN(%d)", int(t))
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
	Filename string
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Value, t.Line, t.Column)
}

type Lexer struct {
	input   string
	pos     int
	line    int
	col     int
	file    string
	tokens  []Token
	indent  []int
	pending []Token
}

func New(input, filename string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		col:    1,
		file:   filename,
		indent: []int{0},
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	afterNewline := true

	for l.pos < len(l.input) {
		ch := l.current()

		if ch == '\n' {
			l.newline()
			l.pos++
			l.col = 1
			l.addPending(Token{Type: NEWLINE, Line: l.line, Column: 1, Filename: l.file})
			afterNewline = true
			continue
		}

		if afterNewline {
			afterNewline = false
			if ch == ' ' || ch == '\t' {
				if err := l.handleIndent(); err != nil {
					return nil, err
				}
				continue
			}
			if ch == '#' {
				l.skipComment()
				afterNewline = true
				continue
			}
			if err := l.handleIndent(); err != nil {
				return nil, err
			}
		}

		if ch == ' ' || ch == '\r' || ch == '\t' {
			l.pos++
			l.col++
			continue
		}

		if ch == '#' {
			l.skipComment()
			continue
		}

		if ch == '"' || ch == '\'' {
			if err := l.readString(ch); err != nil {
				return nil, err
			}
			continue
		}

		if unicode.IsDigit(ch) {
			l.readNumber()
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			l.readIdent()
			continue
		}

		if err := l.readOperator(); err != nil {
			return nil, err
		}
	}

	l.flushIndent()

	l.addPending(Token{Type: EOF, Line: l.line, Column: l.col, Filename: l.file})

	return l.pending, nil
}

func (l *Lexer) current() rune {
	return rune(l.input[l.pos])
}

func (l *Lexer) peek(offset int) rune {
	pos := l.pos + offset
	if pos < 0 || pos >= len(l.input) {
		return 0
	}
	return rune(l.input[pos])
}

func (l *Lexer) addPending(t Token) {
	l.pending = append(l.pending, t)
}

func (l *Lexer) newline() {
	l.line++
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
		l.col++
	}
}

func (l *Lexer) handleIndent() error {
	if l.pos >= len(l.input) {
		return nil
	}

	// Skip blank lines and comment-only lines
	if l.input[l.pos] == '\n' || l.input[l.pos] == '#' {
		return nil
	}

	count := 0
	startCol := l.col
	for l.pos < len(l.input) && l.input[l.pos] == ' ' {
		count++
		l.pos++
		l.col++
	}

	if l.pos < len(l.input) && l.input[l.pos] == '#' {
		l.skipComment()
		return nil
	}

	if l.pos < len(l.input) && l.input[l.pos] == '\n' {
		return nil
	}

	if l.pos >= len(l.input) {
		// EOF - dedent all remaining
		for len(l.indent) > 1 {
			l.indent = l.indent[:len(l.indent)-1]
			l.addPending(Token{Type: DEDENT, Line: l.line, Column: startCol, Filename: l.file})
		}
		return nil
	}

	currentIndent := l.indent[len(l.indent)-1]

	if count > currentIndent {
		if l.isNextTripleQuote() {
			return nil
		}
		l.indent = append(l.indent, count)
		l.addPending(Token{Type: INDENT, Line: l.line, Column: startCol, Filename: l.file})
	} else if count < currentIndent {
		for len(l.indent) > 1 && l.indent[len(l.indent)-1] > count {
			l.indent = l.indent[:len(l.indent)-1]
			l.addPending(Token{Type: DEDENT, Line: l.line, Column: startCol, Filename: l.file})
		}
		if len(l.indent) == 0 || l.indent[len(l.indent)-1] != count {
			return fmt.Errorf("line %d: unexpected indentation level", l.line)
		}
	}

	return nil
}

func (l *Lexer) flushIndent() {
	for len(l.indent) > 1 {
		l.indent = l.indent[:len(l.indent)-1]
		l.addPending(Token{Type: DEDENT, Line: l.line, Column: l.col, Filename: l.file})
	}
}

func (l *Lexer) isNextTripleQuote() bool {
	return l.pos+2 < len(l.input) && l.input[l.pos] == '"' && l.input[l.pos+1] == '"' && l.input[l.pos+2] == '"'
}

func (l *Lexer) readString(quote rune) error {
	start := l.pos
	l.pos++
	l.col++

	var value []rune
	for l.pos < len(l.input) {
		ch := l.current()
		if ch == quote {
			l.pos++
			l.col++
			if ch == '"' && l.pos < len(l.input) && l.input[l.pos] == '"' {
				l.pos++
				l.col++
				for l.pos < len(l.input) {
					if l.current() == '"' && l.pos+2 < len(l.input) && l.input[l.pos+1] == '"' && l.input[l.pos+2] == '"' {
						l.pos += 3
						l.col += 3
						l.addPending(Token{Type: STRING, Value: string(value), Line: l.line, Column: start + 1, Filename: l.file})
						return nil
					}
					if l.current() == '\n' {
						l.line++
						l.col = 1
					}
					value = append(value, l.current())
					l.pos++
					l.col++
				}
				return fmt.Errorf("line %d: unterminated multi-line string", l.line)
			}
			l.addPending(Token{Type: STRING, Value: string(value), Line: l.line, Column: start + 1, Filename: l.file})
			return nil
		}
		if ch == '\\' {
			l.pos++
			l.col++
			if l.pos >= len(l.input) {
				return fmt.Errorf("line %d: unterminated string escape", l.line)
			}
			esc := l.current()
			switch esc {
			case 'n':
				value = append(value, '\n')
			case 't':
				value = append(value, '\t')
			case 'r':
				value = append(value, '\r')
			case '\\':
				value = append(value, '\\')
			case '\'':
				value = append(value, '\'')
			case '"':
				value = append(value, '"')
			default:
				value = append(value, '\\', esc)
			}
			l.pos++
			l.col++
			continue
		}
		if ch == '\n' {
			return fmt.Errorf("line %d: unterminated string", l.line)
		}
		value = append(value, ch)
		l.pos++
		l.col++
	}
	return fmt.Errorf("line %d: unterminated string", l.line)
}

func (l *Lexer) readNumber() {
	start := l.pos
	isFloat := false

	if l.input[l.pos] == '0' && l.pos+1 < len(l.input) {
		next := l.input[l.pos+1]
		if next == 'x' || next == 'X' {
			l.pos += 2
			l.col += 2
			for l.pos < len(l.input) && isHexDigit(l.current()) {
				l.pos++
				l.col++
			}
			l.addPending(Token{Type: INT, Value: l.input[start:l.pos], Line: l.line, Column: start + 1, Filename: l.file})
			return
		}
		if next == 'b' || next == 'B' {
			l.pos += 2
			l.col += 2
			for l.pos < len(l.input) && (l.current() == '0' || l.current() == '1') {
				l.pos++
				l.col++
			}
			l.addPending(Token{Type: INT, Value: l.input[start:l.pos], Line: l.line, Column: start + 1, Filename: l.file})
			return
		}
		if next == 'o' || next == 'O' {
			l.pos += 2
			l.col += 2
			for l.pos < len(l.input) && l.current() >= '0' && l.current() <= '7' {
				l.pos++
				l.col++
			}
			l.addPending(Token{Type: INT, Value: l.input[start:l.pos], Line: l.line, Column: start + 1, Filename: l.file})
			return
		}
	}

	for l.pos < len(l.input) && unicode.IsDigit(l.current()) {
		l.pos++
		l.col++
	}

	if l.pos < len(l.input) && l.current() == '.' && l.pos+1 < len(l.input) && l.input[l.pos+1] != '.' {
		isFloat = true
		l.pos++
		l.col++
		for l.pos < len(l.input) && unicode.IsDigit(l.current()) {
			l.pos++
			l.col++
		}
	}

	if l.pos < len(l.input) && (l.current() == 'e' || l.current() == 'E') {
		isFloat = true
		l.pos++
		l.col++
		if l.pos < len(l.input) && (l.current() == '+' || l.current() == '-') {
			l.pos++
			l.col++
		}
		for l.pos < len(l.input) && unicode.IsDigit(l.current()) {
			l.pos++
			l.col++
		}
	}

	for l.pos < len(l.input) && l.current() == '_' {
		l.pos++
		l.col++
	}

	tt := INT
	if isFloat {
		tt = FLOAT
	}

	l.addPending(Token{Type: tt, Value: l.input[start:l.pos], Line: l.line, Column: start + 1, Filename: l.file})
}

func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func (l *Lexer) readIdent() {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.current()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			l.pos++
			l.col++
		} else {
			break
		}
	}

	word := l.input[start:l.pos]

	// Handle underscored numbers like 1_000
	if unicode.IsDigit(rune(word[0])) {
		l.addPending(Token{Type: INT, Value: word, Line: l.line, Column: start + 1, Filename: l.file})
		return
	}

	tt, ok := keywords[word]
	if !ok {
		tt = IDENT
	}

	l.addPending(Token{Type: tt, Value: word, Line: l.line, Column: start + 1, Filename: l.file})
}

func (l *Lexer) readOperator() error {
	ch := l.current()
	start := l.pos

	switch ch {
	case '+':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: PLUS_EQ, Value: "+=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: PLUS, Value: "+", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '-':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: MINUS_EQ, Value: "-=", Line: l.line, Column: start + 1, Filename: l.file})
		} else if l.pos < len(l.input) && l.input[l.pos] == '>' {
			l.pos++
			l.col++
			l.addPending(Token{Type: ARROW, Value: "->", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: MINUS, Value: "-", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '*':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: STAR_EQ, Value: "*=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: STAR, Value: "*", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '/':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: SLASH_EQ, Value: "/=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: SLASH, Value: "/", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '%':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: PERCENT_EQ, Value: "%=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: PERCENT, Value: "%", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '=':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: EQ, Value: "==", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: ASSIGN, Value: "=", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '!':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: NEQ, Value: "!=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			return fmt.Errorf("line %d: unexpected character '!'", l.line)
		}
	case '<':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: LTE, Value: "<=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: LT, Value: "<", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '>':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			l.col++
			l.addPending(Token{Type: GTE, Value: ">=", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: GT, Value: ">", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '(':
		l.pos++
		l.col++
		l.addPending(Token{Type: LPAREN, Value: "(", Line: l.line, Column: start + 1, Filename: l.file})
	case ')':
		l.pos++
		l.col++
		l.addPending(Token{Type: RPAREN, Value: ")", Line: l.line, Column: start + 1, Filename: l.file})
	case '[':
		l.pos++
		l.col++
		l.addPending(Token{Type: LBRACKET, Value: "[", Line: l.line, Column: start + 1, Filename: l.file})
	case ']':
		l.pos++
		l.col++
		l.addPending(Token{Type: RBRACKET, Value: "]", Line: l.line, Column: start + 1, Filename: l.file})
	case '{':
		l.pos++
		l.col++
		l.addPending(Token{Type: LBRACE, Value: "{", Line: l.line, Column: start + 1, Filename: l.file})
	case '}':
		l.pos++
		l.col++
		l.addPending(Token{Type: RBRACE, Value: "}", Line: l.line, Column: start + 1, Filename: l.file})
	case ',':
		l.pos++
		l.col++
		l.addPending(Token{Type: COMMA, Value: ",", Line: l.line, Column: start + 1, Filename: l.file})
	case ':':
		l.pos++
		l.col++
		l.addPending(Token{Type: COLON, Value: ":", Line: l.line, Column: start + 1, Filename: l.file})
	case '.':
		l.pos++
		l.col++
		if l.pos < len(l.input) && l.input[l.pos] == '.' {
			l.pos++
			l.col++
			l.addPending(Token{Type: DOTDOT, Value: "..", Line: l.line, Column: start + 1, Filename: l.file})
		} else {
			l.addPending(Token{Type: DOT, Value: ".", Line: l.line, Column: start + 1, Filename: l.file})
		}
	case '|':
		l.pos++
		l.col++
		l.addPending(Token{Type: PIPE, Value: "|", Line: l.line, Column: start + 1, Filename: l.file})
	default:
		return fmt.Errorf("line %d:%d: unexpected character '%c'", l.line, l.col, ch)
	}
	return nil
}
