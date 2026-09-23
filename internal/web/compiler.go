package web

import (
	"fmt"
	"html"
	"strings"

	"sayless/internal/parser"
)

// Compiler transforms Say Less AST into web output (HTML + CSS + JS)
type Compiler struct {
	ir          *WebIR
	components  map[string]*parser.Component
	stateVars   map[string]bool
	pageCounter int
	componentID int
}

// Output represents the final compiled web output
type Output struct {
	HTML string
	CSS  string
	JS   string
}

// NewCompiler creates a new web compiler
func NewCompiler() *Compiler {
	return &Compiler{
		ir: &WebIR{
			Pages:      make([]*PageNode, 0),
			Components: make([]*ComponentNode, 0),
			Styles:     make([]*StyleRule, 0),
		},
		components:  make(map[string]*parser.Component),
		stateVars:   make(map[string]bool),
		pageCounter: 0,
		componentID: 0,
	}
}

// Compile compiles a Say Less program into web output
func (c *Compiler) Compile(program *parser.Program) (*Output, error) {
	// First pass: collect components and state
	for _, stmt := range program.Stmts {
		switch n := stmt.(type) {
		case *parser.Component:
			c.components[n.Name] = n
		case *parser.State:
			c.stateVars[n.Name] = true
		}
	}

	// Second pass: build IR
	for _, stmt := range program.Stmts {
		switch n := stmt.(type) {
		case *parser.Page:
			page, err := c.compilePage(n)
			if err != nil {
				return nil, err
			}
			c.ir.Pages = append(c.ir.Pages, page)
		case *parser.Component:
			comp, err := c.compileComponent(n)
			if err != nil {
				return nil, err
			}
			c.ir.Components = append(c.ir.Components, comp)
		case *parser.ServerDecl:
			c.ir.Server = &ServerNode{Port: 8080}
		}
	}

	// Third pass: generate output
	return c.generateOutput()
}

func (c *Compiler) compilePage(page *parser.Page) (*PageNode, error) {
	nodes, err := c.compileBlock(page.Body)
	if err != nil {
		return nil, err
	}
	return &PageNode{
		Route:    page.Route,
		Elements: nodes,
		Meta:     make(map[string]string),
	}, nil
}

func (c *Compiler) compileComponent(comp *parser.Component) (*ComponentNode, error) {
	c.componentID++
	nodes, err := c.compileBlock(comp.Body)
	if err != nil {
		return nil, err
	}
	return &ComponentNode{
		Name:       comp.Name,
		Params:     comp.Params,
		Elements:   nodes,
		IslandID:   fmt.Sprintf("island-%d", c.componentID),
		HasState:   c.componentHasState(comp),
		HasEvents:  c.componentHasEvents(comp),
	}, nil
}

func (c *Compiler) componentHasState(comp *parser.Component) bool {
	for _, stmt := range comp.Body.Stmts {
		if _, ok := stmt.(*parser.State); ok {
			return true
		}
	}
	return false
}

func (c *Compiler) componentHasEvents(comp *parser.Component) bool {
	for _, stmt := range comp.Body.Stmts {
		if elem, ok := stmt.(*parser.HtmlElement); ok {
			if len(elem.Children) > 0 {
				for _, child := range elem.Children {
					if _, ok := child.(*parser.EventHandler); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

func (c *Compiler) compileBlock(block *parser.Block) ([]IRNode, error) {
	var nodes []IRNode
	for _, stmt := range block.Stmts {
		node, err := c.compileStmt(stmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (c *Compiler) compileStmt(stmt parser.ASTNode) (IRNode, error) {
	switch n := stmt.(type) {
	case *parser.HtmlElement:
		return c.compileElement(n)
	case *parser.State:
		return c.compileState(n)
	case *parser.If:
		return c.compileIf(n)
	case *parser.For:
		return c.compileFor(n)
	case *parser.ExprStmt:
		return c.compileExprStmt(n)
	case *parser.EventHandler:
		return c.compileEventHandler(n)
	case *parser.StyleBlock:
		return c.compileStyleBlock(n)
	case *parser.Component:
		return c.compileComponentUsage(n)
	case *parser.AwaitExpr:
		return c.compileAwait(n)
	case *parser.Assign:
		return c.compileAssign(n)
	case *parser.TextInterp:
		return c.compileTextInterp(n)
	case *parser.StringLit:
		return &IRText{Content: n.Value}, nil
	case *parser.IntLit:
		return &IRText{Content: fmt.Sprintf("%d", n.Value)}, nil
	case *parser.FloatLit:
		return &IRText{Content: fmt.Sprintf("%g", n.Value)}, nil
	case *parser.BoolLit:
		if n.Value {
			return &IRText{Content: "true"}, nil
		}
		return &IRText{Content: "false"}, nil
	case *parser.Binary:
		return &IRBinding{Expr: c.exprToString(n)}, nil
	case *parser.Unary:
		return &IRBinding{Expr: c.exprToString(n)}, nil
	case *parser.Ident:
		return &IRBinding{Expr: n.Name}, nil
	case *parser.Member:
		return &IRBinding{Expr: c.exprToString(n)}, nil
	case *parser.Call:
		// Check if this is a component call
		if ident, ok := n.Callee.(*parser.Ident); ok {
			if _, isComp := c.components[ident.Name]; isComp {
				return c.compileComponentCall(n, ident.Name)
			}
		}
		return &IRBinding{Expr: c.exprToString(n)}, nil
	default:
		return nil, nil
	}
}

func (c *Compiler) compileTextInterp(interp *parser.TextInterp) (IRNode, error) {
	var parts []IRNode
	for _, part := range interp.Parts {
		node, err := c.compileStmt(part)
		if err != nil {
			return nil, err
		}
		if node != nil {
			parts = append(parts, node)
		}
	}
	return &IRInterp{Parts: parts}, nil
}

func (c *Compiler) compileElement(elem *parser.HtmlElement) (IRNode, error) {
	irElem := &IRElement{
		Tag:      elem.Tag,
		Attributes: make([]*IRAttribute, 0),
		Children:   make([]IRNode, 0),
		Events:     make([]*IREvent, 0),
	}

	// Compile attributes
	for _, attr := range elem.Attributes {
		irAttr := &IRAttribute{
			Name: attr.Name,
		}
		if s, ok := attr.Value.(*parser.StringLit); ok {
			irAttr.Value = s.Value
			irAttr.IsExpr = false
		} else {
			irAttr.Value = c.exprToString(attr.Value)
			irAttr.IsExpr = true
		}
		irElem.Attributes = append(irElem.Attributes, irAttr)
	}

	// Compile children
	for _, child := range elem.Children {
		node, err := c.compileStmt(child)
		if err != nil {
			return nil, err
		}
		if node != nil {
			irElem.Children = append(irElem.Children, node)
		}
	}

	return irElem, nil
}

func (c *Compiler) compileState(state *parser.State) (IRNode, error) {
	return &IRState{
		Name:         state.Name,
		InitialValue: c.exprToString(state.Value),
		Dependencies: make([]string, 0),
		AffectedDOM:  make([]string, 0),
	}, nil
}

func (c *Compiler) compileIf(ifStmt *parser.If) (IRNode, error) {
	cond := c.exprToString(ifStmt.Cond)
	thenNodes, err := c.compileBlock(ifStmt.Then)
	if err != nil {
		return nil, err
	}
	var elseNodes []IRNode
	if ifStmt.Else_ != nil {
		elseNodes, err = c.compileBlock(ifStmt.Else_)
		if err != nil {
			return nil, err
		}
	}
	return &IRConditional{
		Condition: cond,
		Then:      thenNodes,
		Else:      elseNodes,
	}, nil
}

func (c *Compiler) compileFor(forStmt *parser.For) (IRNode, error) {
	iterable := c.exprToString(forStmt.Iter)
	bodyNodes, err := c.compileBlock(forStmt.Body)
	if err != nil {
		return nil, err
	}
	return &IRLoop{
		Variable: forStmt.Var,
		Iterable: iterable,
		Body:     bodyNodes,
	}, nil
}

func (c *Compiler) compileExprStmt(expr *parser.ExprStmt) (IRNode, error) {
	// Check if the expression is a component call
	if call, ok := expr.Expr.(*parser.Call); ok {
		if ident, ok := call.Callee.(*parser.Ident); ok {
			if _, isComp := c.components[ident.Name]; isComp {
				return c.compileComponentCall(call, ident.Name)
			}
		}
	}
	return c.compileStmt(expr.Expr)
}

func (c *Compiler) compileEventHandler(evt *parser.EventHandler) (IRNode, error) {
	handler := c.blockToString(evt.Body)
	return &IREvent{
		Name:    evt.Event,
		Handler: handler,
		IsExpr:  true,
	}, nil
}

func (c *Compiler) compileStyleBlock(style *parser.StyleBlock) (IRNode, error) {
	props := make([]*StyleProperty, 0)
	for _, prop := range style.Properties {
		props = append(props, &StyleProperty{
			Name:  prop.Name,
			Value: c.exprToString(prop.Value),
		})
	}
	return &StyleBlock{Properties: props}, nil
}

func (c *Compiler) compileComponentUsage(comp *parser.Component) (IRNode, error) {
	return &IRComponent{
		Name:       comp.Name,
		Props:      make([]*IRPropBinding, 0),
		Children:   make([]IRNode, 0),
		InstanceID: fmt.Sprintf("comp-%d", c.componentID),
	}, nil
}

func (c *Compiler) compileComponentCall(call *parser.Call, name string) (IRNode, error) {
	c.componentID++
	props := make([]*IRPropBinding, 0)
	for _, arg := range call.Args {
		if na, ok := arg.(*parser.NamedArg); ok {
			props = append(props, &IRPropBinding{
				Name:  na.Name,
				Value: c.exprToString(na.Value),
				IsExpr: true,
			})
		}
	}
	return &IRComponent{
		Name:       name,
		Props:      props,
		Children:   make([]IRNode, 0),
		InstanceID: fmt.Sprintf("comp-%d", c.componentID),
	}, nil
}

func (c *Compiler) compileAwait(await *parser.AwaitExpr) (IRNode, error) {
	return &IRAwait{
		Expr:    c.exprToString(await.Value),
		Loading: make([]IRNode, 0),
		Error:   make([]IRNode, 0),
		Success: make([]IRNode, 0),
	}, nil
}

func (c *Compiler) compileAssign(assign *parser.Assign) (IRNode, error) {
	return &IRText{
		Content: fmt.Sprintf("let %s = %s", assign.Name, c.exprToString(assign.Value)),
	}, nil
}

func (c *Compiler) exprToString(expr parser.ASTNode) string {
	if expr == nil {
		return ""
	}
	switch n := expr.(type) {
	case *parser.StringLit:
		return fmt.Sprintf("%q", n.Value)
	case *parser.IntLit:
		return fmt.Sprintf("%d", n.Value)
	case *parser.FloatLit:
		return fmt.Sprintf("%g", n.Value)
	case *parser.BoolLit:
		if n.Value {
			return "true"
		}
		return "false"
	case *parser.NoneLit:
		return "null"
	case *parser.Ident:
		return n.Name
	case *parser.Binary:
		op := n.Op
		switch op {
		case "and":
			op = "&&"
		case "or":
			op = "||"
		}
		return fmt.Sprintf("%s %s %s", c.exprToString(n.Left), op, c.exprToString(n.Right))
	case *parser.Unary:
		op := n.Op
		if op == "not" {
			op = "!"
		}
		return fmt.Sprintf("%s%s", op, c.exprToString(n.Expr))
	case *parser.Call:
		args := make([]string, len(n.Args))
		for i, arg := range n.Args {
			args[i] = c.exprToString(arg)
		}
		return fmt.Sprintf("%s(%s)", c.exprToString(n.Callee), strings.Join(args, ", "))
	case *parser.Member:
		return fmt.Sprintf("%s.%s", c.exprToString(n.Object), n.Member)
	case *parser.ListLit:
		elems := make([]string, len(n.Elems))
		for i, elem := range n.Elems {
			elems[i] = c.exprToString(elem)
		}
		return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	case *parser.MapLit:
		entries := make([]string, len(n.Entries))
		for i, entry := range n.Entries {
			entries[i] = fmt.Sprintf("%s: %s", c.exprToString(entry.Key), c.exprToString(entry.Value))
		}
		return fmt.Sprintf("{%s}", strings.Join(entries, ", "))
	default:
		return "undefined"
	}
}

func (c *Compiler) blockToString(block *parser.Block) string {
	var parts []string
	for _, stmt := range block.Stmts {
		parts = append(parts, c.stmtToString(stmt))
	}
	return strings.Join(parts, "\n")
}

func (c *Compiler) stmtToString(stmt parser.ASTNode) string {
	if stmt == nil {
		return ""
	}
	switch n := stmt.(type) {
	case *parser.Assign:
		return fmt.Sprintf("%s = %s", n.Name, c.exprToString(n.Value))
	case *parser.AugAssign:
		return fmt.Sprintf("%s %s %s", n.Name, n.Op, c.exprToString(n.Value))
	case *parser.ExprStmt:
		return c.exprToString(n.Expr)
	case *parser.Return:
		if n.Value != nil {
			return fmt.Sprintf("return %s", c.exprToString(n.Value))
		}
		return "return"
	case *parser.If:
		cond := c.exprToString(n.Cond)
		then := c.blockToString(n.Then)
		if n.Else_ != nil {
			else_ := c.blockToString(n.Else_)
			return fmt.Sprintf("if (%s) {\n%s\n} else {\n%s\n}", cond, then, else_)
		}
		return fmt.Sprintf("if (%s) {\n%s\n}", cond, then)
	case *parser.Block:
		return c.blockToString(n)
	default:
		return ""
	}
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// generateOutput produces the final HTML, CSS, and JS output
func (c *Compiler) generateOutput() (*Output, error) {
	gen := NewGenerator(c.ir)
	return gen.Generate(), nil
}
