package parser

import (
	"fmt"
	"strconv"
	"strings"

	"sayless/internal/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []string
}

var htmlTags = map[string]bool{
	"html": true, "head": true, "body": true, "title": true,
	"meta": true, "link": true, "script": true, "style": true,
	"div": true, "span": true, "p": true, "h1": true, "h2": true,
	"h3": true, "h4": true, "h5": true, "h6": true,
	"a": true, "img": true, "ul": true, "ol": true, "li": true,
	"table": true, "tr": true, "td": true, "th": true,
	"form": true, "input": true, "textarea": true, "select": true,
	"option": true, "button": true, "label": true,
	"header": true, "footer": true, "nav": true, "main": true,
	"section": true, "article": true, "aside": true,
	"strong": true, "em": true, "small": true, "code": true,
	"pre": true, "blockquote": true, "br": true, "hr": true,
	"canvas": true, "video": true, "audio": true, "source": true,
	"svg": true, "path": true, "iframe": true,
}

func isHtmlTag(name string) bool {
	return htmlTags[name]
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (*Program, error) {
	prog := &Program{Pos: p.current()}
	for !p.isAtEnd() {
		if p.check(lexer.NEWLINE) {
			p.advance()
			continue
		}
		if p.check(lexer.EOF) {
			break
		}
		stmt, err := p.statement()
		if err != nil {
			p.errors = append(p.errors, err.Error())
			p.skipToNextLine()
			continue
		}
		if stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
	}
	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors:\n%s", strings.Join(p.errors, "\n"))
	}
	return prog, nil
}

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	tok := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) check(tt lexer.TokenType) bool {
	return p.current().Type == tt
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.current().Type == lexer.EOF
}

func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, error) {
	if p.check(tt) {
		return p.advance(), nil
	}
	tok := p.current()
	return tok, fmt.Errorf("line %d:%d: expected %s, got %s (%q)", tok.Line, tok.Column, tt, tok.Type, tok.Value)
}

func (p *Parser) match(tt lexer.TokenType) bool {
	if p.check(tt) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) peek(offset int) lexer.Token {
	pos := p.pos + offset
	if pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[pos]
}

func (p *Parser) skipNewlines() {
	for p.check(lexer.NEWLINE) {
		p.advance()
	}
}

// skipLayout skips structural tokens (newlines, indents) that may appear
// inside bracketed expressions such as multi-line call arguments.
func (p *Parser) skipLayout() {
	for p.check(lexer.NEWLINE) || p.check(lexer.INDENT) || p.check(lexer.DEDENT) {
		p.advance()
	}
}

func (p *Parser) skipToNextLine() {
	for !p.isAtEnd() && !p.check(lexer.NEWLINE) && !p.check(lexer.EOF) {
		p.advance()
	}
	if p.check(lexer.NEWLINE) {
		p.advance()
	}
}

func (p *Parser) statement() (ASTNode, error) {
	p.skipNewlines()
	tok := p.current()
	switch tok.Type {
	case lexer.FN:
		return p.fnDef()
	case lexer.STRUCT:
		return p.structDef()
	case lexer.IF:
		return p.ifStmt()
	case lexer.WHILE:
		return p.whileStmt()
	case lexer.FOR:
		return p.forStmt()
	case lexer.RETURN:
		return p.returnStmt()
	case lexer.BREAK:
		p.advance()
		return &Break{Pos: tok}, nil
	case lexer.CONTINUE:
		p.advance()
		return &Continue{Pos: tok}, nil
	case lexer.USE:
		return p.useStmt()
	case lexer.TEST:
		return p.testBlock()
	case lexer.ASSERT:
		return p.assertStmt()
	case lexer.SERVER:
		return p.serverDecl()
	case lexer.GET, lexer.POST, lexer.PUT, lexer.DELETE, lexer.PATCH:
		return p.routeHandler()
	case lexer.PAGE:
		return p.pageDecl()
	case lexer.COMPONENT:
		return p.componentDecl()
	case lexer.STATE:
		return p.stateDecl()
	case lexer.AWAIT:
		return p.awaitStmt()
	case lexer.TRY:
		return p.tryCatch()
	case lexer.THROW:
		return p.throwStmt()
	case lexer.MUT:
		p.advance()
		return p.exprOrAssign()
	case lexer.CONST:
		p.advance()
		return p.exprOrAssign()
	case lexer.LET:
		p.advance()
		return p.exprOrAssign()
	}
	// Check for HTML elements (identifiers that are known HTML tags)
	if p.check(lexer.IDENT) && isHtmlTag(p.current().Value) {
		return p.htmlElement()
	}
	// Check for event handlers (on click, on input, etc.)
	if p.check(lexer.ON) {
		return p.eventHandler()
	}
	// Check for style blocks
	if p.check(lexer.STYLE) {
		return p.styleBlock()
	}
	return p.exprOrAssign()
}

func (p *Parser) fnDef() (ASTNode, error) {
	tok := p.advance()
	name := ""
	if p.check(lexer.IDENT) {
		name = p.advance().Value
	}
	params, err := p.paramList()
	if err != nil {
		return nil, err
	}
	retType := ""
	if p.match(lexer.COLON) {
		retType = p.advance().Value
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &FnDef{Name: name, Params: params, ReturnTyp: retType, Body: body, Pos: tok}, nil
}

func (p *Parser) paramList() ([]Param, error) {
	if !p.match(lexer.LPAREN) {
		return nil, nil
	}
	var params []Param
	if !p.check(lexer.RPAREN) {
		for {
			param, err := p.param()
			if err != nil {
				return nil, err
			}
			params = append(params, param)
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *Parser) param() (Param, error) {
	name := p.advance().Value
	param := Param{Name: name}
	if p.match(lexer.COLON) {
		param.Type_ = p.advance().Value
	}
	if p.match(lexer.ASSIGN) {
		defaultVal, err := p.expression()
		if err != nil {
			return param, err
		}
		param.Default = defaultVal
	}
	return param, nil
}

func (p *Parser) block() (*Block, error) {
	if p.check(lexer.LBRACE) {
		return p.braceBlock()
	}
	if p.check(lexer.COLON) {
		p.advance()
	}
	// Consume all newlines (including blank lines) before the indented block
	p.skipNewlines()
	hadIndent := p.check(lexer.INDENT)
	if hadIndent {
		p.advance()
	}
	block := &Block{Pos: p.current()}
	for !p.isAtEnd() {
		if p.check(lexer.DEDENT) {
			if hadIndent {
				p.advance()
			}
			break
		}
		if p.check(lexer.RBRACE) {
			break
		}
		if p.check(lexer.NEWLINE) {
			p.advance()
			continue
		}
		if p.check(lexer.INDENT) {
			p.advance()
			continue
		}
		stmt, err := p.statement()
		if err != nil {
			p.errors = append(p.errors, err.Error())
			p.skipToNextLine()
			continue
		}
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	return block, nil
}

func (p *Parser) braceBlock() (*Block, error) {
	tok := p.advance()
	block := &Block{Pos: tok}
	for !p.check(lexer.RBRACE) && !p.isAtEnd() {
		if p.check(lexer.NEWLINE) || p.check(lexer.INDENT) || p.check(lexer.DEDENT) {
			p.advance()
			continue
		}
		stmt, err := p.statement()
		if err != nil {
			p.errors = append(p.errors, err.Error())
			p.skipToNextLine()
			continue
		}
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, fmt.Errorf("line %d: expected '}}'", p.current().Line)
	}
	return block, nil
}

func (p *Parser) blockNoDedent() (*Block, error) {
	if p.check(lexer.LBRACE) {
		return p.braceBlock()
	}
	if _, err := p.expect(lexer.COLON); err != nil {
		return nil, err
	}
	p.skipNewlines()
	if _, err := p.expect(lexer.INDENT); err != nil {
		return nil, fmt.Errorf("line %d: expected indented block after colon", p.current().Line)
	}
	block := &Block{Pos: p.current()}
	for !p.check(lexer.DEDENT) && !p.isAtEnd() {
		if p.check(lexer.NEWLINE) {
			p.advance()
			continue
		}
		stmt, err := p.statement()
		if err != nil {
			p.errors = append(p.errors, err.Error())
			p.skipToNextLine()
			continue
		}
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	return block, nil
}

// webBlock parses a body block for page/component/event-handler declarations.
// Unlike block(), it does NOT skip nested INDENT tokens — they are left for
// htmlElement() to consume when parsing nested children.
func (p *Parser) webBlock() (*Block, error) {
	if p.check(lexer.LBRACE) {
		return p.braceBlock()
	}
	if p.check(lexer.COLON) {
		p.advance()
	}
	// Skip all newlines before the indented block
	for p.check(lexer.NEWLINE) {
		p.advance()
	}
	hadIndent := p.check(lexer.INDENT)
	if hadIndent {
		p.advance()
	}
	block := &Block{Pos: p.current()}
	for !p.isAtEnd() {
		if p.check(lexer.DEDENT) {
			if hadIndent {
				p.advance()
			}
			break
		}
		if p.check(lexer.RBRACE) {
			break
		}
		if p.check(lexer.NEWLINE) {
			p.advance()
			continue
		}
		// Do NOT skip INDENT — leave for htmlElement() to consume
		stmt, err := p.statement()
		if err != nil {
			p.errors = append(p.errors, err.Error())
			p.skipToNextLine()
			continue
		}
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	return block, nil
}

func (p *Parser) structDef() (ASTNode, error) {
	tok := p.advance()
	name := p.advance().Value
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	var fields []Field
	for _, stmt := range body.Stmts {
		if assign, ok := stmt.(*Assign); ok {
			fields = append(fields, Field{Name: assign.Name, Type_: assign.Type_})
		}
	}
	return &StructDef{Name: name, Fields: fields, Pos: tok}, nil
}

func (p *Parser) ifStmt() (ASTNode, error) {
	tok := p.advance()
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	then, err := p.block()
	if err != nil {
		return nil, err
	}
	node := &If{Cond: cond, Then: then, Pos: tok}
	p.skipNewlines()
	for p.check(lexer.ELSEIF) {
		p.advance()
		elifCond, err := p.expression()
		if err != nil {
			return nil, err
		}
		elifThen, err := p.block()
		if err != nil {
			return nil, err
		}
		node.ElseIf = append(node.ElseIf, Elif{Cond: elifCond, Then: elifThen})
		p.skipNewlines()
	}
	p.skipNewlines()
	if p.check(lexer.ELSE) {
		p.advance()
		elseBlock, err := p.block()
		if err != nil {
			return nil, err
		}
		node.Else_ = elseBlock
	}
	return node, nil
}

func (p *Parser) whileStmt() (ASTNode, error) {
	tok := p.advance()
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &While{Cond: cond, Body: body, Pos: tok}, nil
}

func (p *Parser) forStmt() (ASTNode, error) {
	tok := p.advance()
	varName := p.advance().Value
	if _, err := p.expect(lexer.IN); err != nil {
		return nil, err
	}
	iter, err := p.expression()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &For{Var: varName, Iter: iter, Body: body, Pos: tok}, nil
}

func (p *Parser) returnStmt() (ASTNode, error) {
	tok := p.advance()
	if p.check(lexer.NEWLINE) || p.check(lexer.EOF) || p.check(lexer.DEDENT) {
		return &Return{Value: nil, Pos: tok}, nil
	}
	val, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &Return{Value: val, Pos: tok}, nil
}

func (p *Parser) useStmt() (ASTNode, error) {
	tok := p.advance()
	path := p.advance().Value
	// Handle npm:foo, pip:foo style paths
	for p.check(lexer.COLON) && p.peek(1).Type == lexer.IDENT {
		p.advance() // consume colon
		path += ":" + p.advance().Value
	}
	asName := ""
	if p.match(lexer.AS) {
		asName = p.advance().Value
	}
	return &Use{Path: path, As: asName, Pos: tok}, nil
}

func (p *Parser) testBlock() (ASTNode, error) {
	tok := p.advance()
	name := ""
	if p.check(lexer.STRING) {
		name = p.advance().Value
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &TestBlock{Name: name, Body: body, Pos: tok}, nil
}

func (p *Parser) assertStmt() (ASTNode, error) {
	tok := p.advance()
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &Assert{Expr: expr, Pos: tok}, nil
}

func (p *Parser) exprOrAssign() (ASTNode, error) {
	tok := p.current()

	// Check for IDENT = ... or IDENT += ... etc before parsing full expression
	if p.check(lexer.IDENT) || p.check(lexer.MUT) || p.check(lexer.CONST) || p.check(lexer.LET) {
		isMut := false
		if p.check(lexer.MUT) || p.check(lexer.CONST) || p.check(lexer.LET) {
			isMut = true
			p.advance()
		}
		if p.check(lexer.IDENT) {
			name := p.advance().Value
			if p.check(lexer.ASSIGN) {
				p.advance()
				val, err := p.expression()
				if err != nil {
					return nil, err
				}
				_ = isMut
				return &Assign{Name: name, Type_: "", Value: val, Pos: tok}, nil
			}
			augOps := map[lexer.TokenType]string{
				lexer.PLUS_EQ: "+=", lexer.MINUS_EQ: "-=", lexer.STAR_EQ: "*=",
				lexer.SLASH_EQ: "/=", lexer.PERCENT_EQ: "%=",
			}
			if op, ok := augOps[p.current().Type]; ok {
				p.advance()
				val, err := p.expression()
				if err != nil {
					return nil, err
				}
				return &AugAssign{Name: name, Op: op, Value: val, Pos: tok}, nil
			}
			// Not an assignment - put the name back as an expression and continue
			p.pos--
			if isMut {
				p.pos--
			}
		} else if isMut {
			p.pos--
		}
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &ExprStmt{Expr: expr, Pos: tok}, nil
}

func (p *Parser) expression() (ASTNode, error) {
	return p.orExpr()
}

func (p *Parser) orExpr() (ASTNode, error) {
	left, err := p.andExpr()
	if err != nil {
		return nil, err
	}
	for p.check(lexer.OR) {
		tok := p.advance()
		right, err := p.andExpr()
		if err != nil {
			return nil, err
		}
		left = &Binary{Op: "or", Left: left, Right: right, Pos: tok}
	}
	return left, nil
}

func (p *Parser) andExpr() (ASTNode, error) {
	left, err := p.notExpr()
	if err != nil {
		return nil, err
	}
	for p.check(lexer.AND) {
		tok := p.advance()
		right, err := p.notExpr()
		if err != nil {
			return nil, err
		}
		left = &Binary{Op: "and", Left: left, Right: right, Pos: tok}
	}
	return left, nil
}

func (p *Parser) notExpr() (ASTNode, error) {
	if p.check(lexer.NOT) {
		tok := p.advance()
		expr, err := p.notExpr()
		if err != nil {
			return nil, err
		}
		return &Unary{Op: "not", Expr: expr, Pos: tok}, nil
	}
	return p.comparison()
}

func (p *Parser) comparison() (ASTNode, error) {
	left, err := p.addition()
	if err != nil {
		return nil, err
	}
	for p.check(lexer.EQ) || p.check(lexer.NEQ) || p.check(lexer.LT) || p.check(lexer.GT) || p.check(lexer.LTE) || p.check(lexer.GTE) {
		tok := p.advance()
		right, err := p.addition()
		if err != nil {
			return nil, err
		}
		left = &Binary{Op: tok.Value, Left: left, Right: right, Pos: tok}
	}
	if p.check(lexer.DOTDOT) {
		tok := p.advance()
		right, err := p.addition()
		if err != nil {
			return nil, err
		}
		left = &Binary{Op: "..", Left: left, Right: right, Pos: tok}
	}
	return left, nil
}

func (p *Parser) addition() (ASTNode, error) {
	left, err := p.multiplication()
	if err != nil {
		return nil, err
	}
	for p.check(lexer.PLUS) || p.check(lexer.MINUS) {
		tok := p.advance()
		right, err := p.multiplication()
		if err != nil {
			return nil, err
		}
		left = &Binary{Op: tok.Value, Left: left, Right: right, Pos: tok}
	}
	return left, nil
}

func (p *Parser) multiplication() (ASTNode, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.check(lexer.STAR) || p.check(lexer.SLASH) || p.check(lexer.PERCENT) {
		tok := p.advance()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		left = &Binary{Op: tok.Value, Left: left, Right: right, Pos: tok}
	}
	return left, nil
}

func (p *Parser) unary() (ASTNode, error) {
	if p.check(lexer.MINUS) || p.check(lexer.PLUS) {
		tok := p.advance()
		expr, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &Unary{Op: tok.Value, Expr: expr, Pos: tok}, nil
	}
	return p.postfix()
}

func (p *Parser) postfix() (ASTNode, error) {
	left, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(lexer.DOT) {
			member := p.advance().Value
			left = &Member{Object: left, Member: member, Pos: p.current()}
		} else if p.check(lexer.LPAREN) {
			tok := p.advance()
			var args []ASTNode
			p.skipLayout()
			for !p.check(lexer.RPAREN) && !p.isAtEnd() {
				if p.check(lexer.IDENT) && p.peek(1).Type == lexer.COLON {
					nameTok := p.advance()
					p.advance() // colon
					val, err := p.expression()
					if err != nil {
						return nil, err
					}
					args = append(args, &NamedArg{Name: nameTok.Value, Value: val, Pos: nameTok})
				} else {
					arg, err := p.expression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
				if !p.match(lexer.COMMA) {
					break
				}
				p.skipLayout()
			}
			p.skipLayout()
			if _, err := p.expect(lexer.RPAREN); err != nil {
				return nil, err
			}
			left = &Call{Callee: left, Args: args, Pos: tok}
		} else if _, ok := left.(*Ident); ok && p.isImplicitArg() {
			tok := p.current()
			arg, err := p.expression()
			if err != nil {
				break
			}
			left = &Call{Callee: left, Args: []ASTNode{arg}, Pos: tok}
			break
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) isImplicitArg() bool {
	tt := p.current().Type
	return tt == lexer.STRING || tt == lexer.INT || tt == lexer.FLOAT ||
		tt == lexer.TRUE || tt == lexer.FALSE || tt == lexer.NONE ||
		tt == lexer.IDENT || tt == lexer.LPAREN || tt == lexer.LBRACKET ||
		tt == lexer.LBRACE
}

func (p *Parser) primary() (ASTNode, error) {
	tok := p.current()
	switch tok.Type {
	case lexer.INT:
		p.advance()
		val, err := strconv.ParseInt(strings.Replace(tok.Value, "_", "", -1), 0, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid integer %q", tok.Line, tok.Value)
		}
		return &IntLit{Value: val, Pos: tok}, nil
	case lexer.FLOAT:
		p.advance()
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid float %q", tok.Line, tok.Value)
		}
		return &FloatLit{Value: val, Pos: tok}, nil
	case lexer.STRING:
		p.advance()
		return &StringLit{Value: tok.Value, Pos: tok}, nil
	case lexer.TRUE:
		p.advance()
		return &BoolLit{Value: true, Pos: tok}, nil
	case lexer.FALSE:
		p.advance()
		return &BoolLit{Value: false, Pos: tok}, nil
	case lexer.NONE:
		p.advance()
		return &NoneLit{Pos: tok}, nil
	case lexer.IDENT:
		p.advance()
		return &Ident{Name: tok.Value, Pos: tok}, nil
	case lexer.LPAREN:
		p.advance()
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.RPAREN); err != nil {
			return nil, err
		}
		return expr, nil
	case lexer.LBRACKET:
		return p.listLiteral()
	case lexer.LBRACE:
		return p.mapLiteral()
	case lexer.FN:
		return p.fnLiteral()
	}
	return nil, fmt.Errorf("line %d:%d: unexpected token %s (%q)", tok.Line, tok.Column, tok.Type, tok.Value)
}

func (p *Parser) listLiteral() (ASTNode, error) {
	tok := p.advance()
	var elems []ASTNode
	p.skipLayout()
	for !p.check(lexer.RBRACKET) && !p.isAtEnd() {
		elem, err := p.expression()
		if err != nil {
			return nil, err
		}
		elems = append(elems, elem)
		if !p.match(lexer.COMMA) {
			break
		}
		p.skipLayout()
	}
	p.skipLayout()
	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}
	return &ListLit{Elems: elems, Pos: tok}, nil
}

func (p *Parser) mapLiteral() (ASTNode, error) {
	tok := p.advance()
	var entries []MapEntry
	p.skipLayout()
	for !p.check(lexer.RBRACE) && !p.isAtEnd() {
		var key ASTNode
		if p.check(lexer.IDENT) && p.peek(1).Type == lexer.COLON {
			keyTok := p.advance()
			key = &StringLit{Value: keyTok.Value, Pos: keyTok}
		} else {
			var err error
			key, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
		if _, err := p.expect(lexer.COLON); err != nil {
			return nil, err
		}
		val, err := p.expression()
		if err != nil {
			return nil, err
		}
		entries = append(entries, MapEntry{Key: key, Value: val})
		if !p.match(lexer.COMMA) {
			break
		}
		p.skipLayout()
	}
	p.skipLayout()
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &MapLit{Entries: entries, Pos: tok}, nil
}

func (p *Parser) fnLiteral() (ASTNode, error) {
	tok := p.advance()
	params, err := p.paramList()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &FnLiteral{Params: params, Body: body, Pos: tok}, nil
}

func (p *Parser) routeHandler() (ASTNode, error) {
	tok := p.advance()
	method := tok.Value
	route, err := p.expression()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &RouteHandler{Method: method, Route: route, Body: body, Pos: tok}, nil
}

func (p *Parser) serverDecl() (ASTNode, error) {
	tok := p.advance()
	if !p.match(lexer.ON) {
		return nil, fmt.Errorf("line %d: expected 'on' after 'server'", tok.Line)
	}
	port, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &ServerDecl{Port: port, Pos: tok}, nil
}

func (p *Parser) pageDecl() (ASTNode, error) {
	tok := p.advance()
	route := ""
	if p.check(lexer.STRING) {
		route = p.advance().Value
	}
	body, err := p.webBlock()
	if err != nil {
		return nil, err
	}
	return &Page{Route: route, Body: body, Pos: tok}, nil
}

func (p *Parser) componentDecl() (ASTNode, error) {
	tok := p.advance()
	name := p.advance().Value
	params, err := p.paramList()
	if err != nil {
		return nil, err
	}
	body, err := p.webBlock()
	if err != nil {
		return nil, err
	}
	return &Component{Name: name, Params: params, Body: body, Pos: tok}, nil
}

func (p *Parser) stateDecl() (ASTNode, error) {
	tok := p.advance()
	name := p.advance().Value
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	val, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &State{Name: name, Value: val, Pos: tok}, nil
}

func (p *Parser) htmlElement() (ASTNode, error) {
	tok := p.advance()
	tag := tok.Value
	var attrs []HtmlAttr
	var children []ASTNode

	// Parse attributes: class "foo", id "bar", etc.
	for p.check(lexer.IDENT) && !p.check(lexer.NEWLINE) {
		attrName := p.current().Value
		if attrName == "on" || attrName == "style" {
			break
		}
		if p.peek(1).Type == lexer.STRING || p.peek(1).Type == lexer.INT || p.peek(1).Type == lexer.IDENT || p.peek(1).Type == lexer.LBRACE {
			p.advance()
			val, err := p.expression()
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, HtmlAttr{Name: attrName, Value: val})
		} else {
			break
		}
	}

	// Check for inline text content on same line
	if !p.check(lexer.NEWLINE) && !p.check(lexer.EOF) && !p.check(lexer.INDENT) {
		textExpr, err := p.expression()
		if err == nil {
			children = append(children, &TextInterp{Parts: []ASTNode{textExpr}, Pos: p.current()})
		}
	}

	// If no newline follows, the element is complete (self-closing)
	if !p.check(lexer.NEWLINE) {
		return &HtmlElement{Tag: tag, Attributes: attrs, Children: children, Pos: tok}, nil
	}

	// Consume newlines and blank lines, then check for children block
	p.advance()
	for p.check(lexer.NEWLINE) {
		p.advance()
	}
	if !p.check(lexer.INDENT) {
		return &HtmlElement{Tag: tag, Attributes: attrs, Children: children, Pos: tok}, nil
	}

	// Consume INDENT and parse children until DEDENT
	p.advance()
	for !p.check(lexer.DEDENT) && !p.isAtEnd() {
		if p.check(lexer.NEWLINE) {
			p.advance()
			continue
		}
		if p.check(lexer.ON) {
			evt, err := p.eventHandler()
			if err != nil {
				p.errors = append(p.errors, err.Error())
				p.skipToNextLine()
				continue
			}
			children = append(children, evt)
			continue
		}
		if p.check(lexer.STYLE) {
			style, err := p.styleBlock()
			if err != nil {
				p.errors = append(p.errors, err.Error())
				p.skipToNextLine()
				continue
			}
			children = append(children, style)
			continue
		}
		if p.check(lexer.IDENT) && isHtmlTag(p.current().Value) {
			child, err := p.htmlElement()
			if err != nil {
				p.errors = append(p.errors, err.Error())
				p.skipToNextLine()
				continue
			}
			children = append(children, child)
			continue
		}
		stmt, err := p.statement()
		if err != nil {
			p.errors = append(p.errors, err.Error())
			p.skipToNextLine()
			continue
		}
		if stmt != nil {
			children = append(children, stmt)
		}
	}
	if p.check(lexer.DEDENT) {
		p.advance()
	}

	return &HtmlElement{Tag: tag, Attributes: attrs, Children: children, Pos: tok}, nil
}

func (p *Parser) eventHandler() (ASTNode, error) {
	tok := p.advance() // consume 'on'
	event := p.advance().Value
	body, err := p.webBlock()
	if err != nil {
		return nil, err
	}
	return &EventHandler{Event: event, Body: body, Pos: tok}, nil
}

func (p *Parser) styleBlock() (ASTNode, error) {
	tok := p.advance() // consume 'style'
	var props []StyleProp

	// Parse on same line
	if !p.check(lexer.NEWLINE) && !p.check(lexer.EOF) {
		// Inline style properties
		for {
			if p.check(lexer.NEWLINE) || p.check(lexer.EOF) || p.check(lexer.DEDENT) {
				break
			}
			propName := ""
			if p.check(lexer.IDENT) {
				propName = p.advance().Value
			} else {
				break
			}
			if _, err := p.expect(lexer.COLON); err != nil {
				break
			}
			propVal, err := p.expression()
			if err != nil {
				break
			}
			props = append(props, StyleProp{Name: propName, Value: propVal})
			if !p.match(lexer.COMMA) {
				break
			}
		}
		return &StyleBlock{Properties: props, Pos: tok}, nil
	}

	// Parse indented block
	if p.check(lexer.NEWLINE) {
		p.advance()
		if p.check(lexer.INDENT) {
			p.advance()
			for !p.check(lexer.DEDENT) && !p.isAtEnd() {
				if p.check(lexer.NEWLINE) {
					p.advance()
					continue
				}
				if p.check(lexer.IDENT) {
					propName := p.advance().Value
					if p.check(lexer.COLON) {
						p.advance() // consume colon
						propVal, err := p.expression()
						if err != nil {
							p.skipToNextLine()
							continue
						}
						props = append(props, StyleProp{Name: propName, Value: propVal})
					}
				} else {
					p.skipToNextLine()
				}
			}
			if p.check(lexer.DEDENT) {
				p.advance()
			}
		}
	}

	return &StyleBlock{Properties: props, Pos: tok}, nil
}

func (p *Parser) awaitStmt() (ASTNode, error) {
	tok := p.advance()
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &AwaitExpr{Value: expr, Pos: tok}, nil
}

func (p *Parser) tryCatch() (ASTNode, error) {
	tok := p.advance()
	tryBody, err := p.block()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	catchVar := "error"
	if p.check(lexer.CATCH) {
		p.advance()
		if p.check(lexer.IDENT) {
			catchVar = p.advance().Value
		}
	} else {
		return nil, fmt.Errorf("line %d: expected 'catch' after 'try' block", tok.Line)
	}
	catchBody, err := p.block()
	if err != nil {
		return nil, err
	}
	return &TryCatch{TryBody: tryBody, CatchVar: catchVar, CatchBody: catchBody, Pos: tok}, nil
}

func (p *Parser) throwStmt() (ASTNode, error) {
	tok := p.advance()
	val, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &Throw{Value: val, Pos: tok}, nil
}
