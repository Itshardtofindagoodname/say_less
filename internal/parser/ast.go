package parser

import "sayless/internal/lexer"

type NodeType int

const (
	NodeProgram NodeType = iota
	NodeBlock
	NodeAssign
	NodeAugAssign
	NodeFnDef
	NodeStructDef
	NodeIf
	NodeWhile
	NodeFor
	NodeReturn
	NodeBreak
	NodeContinue
	NodeUse
	NodeTest
	NodeAssert
	NodeExprStmt
	NodeBinary
	NodeUnary
	NodeCall
	NodeMember
	NodeIdent
	NodeInt
	NodeFloat
	NodeString
	NodeBool
	NodeNone
	NodeList
	NodeMap
	NodeRange
	NodeFnLiteral
	NodeRouteHandler
	NodeServerDecl

	// Web nodes
	NodePage
	NodeComponent
	NodeState
	NodeHtmlElement
	NodeEventHandler
	NodeStyleBlock
	NodeAwaitExpr
	NodeTryCatch
	NodeThrow
	NodeTextInterp
)

type ASTNode interface {
	Type() NodeType
	Position() lexer.Token
}

type Program struct {
	Stmts    []ASTNode
	Pos      lexer.Token
}

func (p *Program) Type() NodeType      { return NodeProgram }
func (p *Program) Position() lexer.Token { return p.Pos }

type Block struct {
	Stmts []ASTNode
	Pos   lexer.Token
}

func (b *Block) Type() NodeType      { return NodeBlock }
func (b *Block) Position() lexer.Token { return b.Pos }

type Assign struct {
	Name  string
	Type_ string
	Value ASTNode
	Pos   lexer.Token
}

func (a *Assign) Type() NodeType      { return NodeAssign }
func (a *Assign) Position() lexer.Token { return a.Pos }

type AugAssign struct {
	Name  string
	Op    string
	Value ASTNode
	Pos   lexer.Token
}

func (a *AugAssign) Type() NodeType      { return NodeAugAssign }
func (a *AugAssign) Position() lexer.Token { return a.Pos }

type Param struct {
	Name    string
	Type_   string
	Default ASTNode
}

type FnDef struct {
	Name      string
	Params    []Param
	ReturnTyp string
	Body      *Block
	Pos       lexer.Token
}

func (f *FnDef) Type() NodeType      { return NodeFnDef }
func (f *FnDef) Position() lexer.Token { return f.Pos }

type Field struct {
	Name string
	Type_ string
}

type StructDef struct {
	Name   string
	Fields []Field
	Pos    lexer.Token
}

func (s *StructDef) Type() NodeType      { return NodeStructDef }
func (s *StructDef) Position() lexer.Token { return s.Pos }

type If struct {
	Cond   ASTNode
	Then   *Block
	ElseIf []Elif
	Else_  *Block
	Pos    lexer.Token
}

func (i *If) Type() NodeType      { return NodeIf }
func (i *If) Position() lexer.Token { return i.Pos }

type Elif struct {
	Cond ASTNode
	Then *Block
}

type While struct {
	Cond ASTNode
	Body *Block
	Pos  lexer.Token
}

func (w *While) Type() NodeType      { return NodeWhile }
func (w *While) Position() lexer.Token { return w.Pos }

type For struct {
	Var   string
	Iter  ASTNode
	Body  *Block
	Pos   lexer.Token
}

func (f *For) Type() NodeType      { return NodeFor }
func (f *For) Position() lexer.Token { return f.Pos }

type Return struct {
	Value ASTNode
	Pos   lexer.Token
}

func (r *Return) Type() NodeType      { return NodeReturn }
func (r *Return) Position() lexer.Token { return r.Pos }

type Break struct {
	Pos lexer.Token
}

func (b *Break) Type() NodeType      { return NodeBreak }
func (b *Break) Position() lexer.Token { return b.Pos }

type Continue struct {
	Pos lexer.Token
}

func (c *Continue) Type() NodeType      { return NodeContinue }
func (c *Continue) Position() lexer.Token { return c.Pos }

type Use struct {
	Path string
	As   string
	Pos  lexer.Token
}

func (u *Use) Type() NodeType      { return NodeUse }
func (u *Use) Position() lexer.Token { return u.Pos }

type TestBlock struct {
	Name string
	Body *Block
	Pos  lexer.Token
}

func (t *TestBlock) Type() NodeType      { return NodeTest }
func (t *TestBlock) Position() lexer.Token { return t.Pos }

type Assert struct {
	Expr ASTNode
	Pos  lexer.Token
}

func (a *Assert) Type() NodeType      { return NodeAssert }
func (a *Assert) Position() lexer.Token { return a.Pos }

type ExprStmt struct {
	Expr ASTNode
	Pos  lexer.Token
}

func (e *ExprStmt) Type() NodeType      { return NodeExprStmt }
func (e *ExprStmt) Position() lexer.Token { return e.Pos }

type Binary struct {
	Op   string
	Left ASTNode
	Right ASTNode
	Pos  lexer.Token
}

func (b *Binary) Type() NodeType      { return NodeBinary }
func (b *Binary) Position() lexer.Token { return b.Pos }

type Unary struct {
	Op   string
	Expr ASTNode
	Pos  lexer.Token
}

func (u *Unary) Type() NodeType      { return NodeUnary }
func (u *Unary) Position() lexer.Token { return u.Pos }

type Call struct {
	Callee ASTNode
	Args   []ASTNode
	Pos    lexer.Token
}

func (c *Call) Type() NodeType      { return NodeCall }
func (c *Call) Position() lexer.Token { return c.Pos }

type Member struct {
	Object ASTNode
	Member string
	Pos    lexer.Token
}

func (m *Member) Type() NodeType      { return NodeMember }
func (m *Member) Position() lexer.Token { return m.Pos }

type Ident struct {
	Name string
	Pos  lexer.Token
}

func (i *Ident) Type() NodeType      { return NodeIdent }
func (i *Ident) Position() lexer.Token { return i.Pos }

type IntLit struct {
	Value int64
	Pos   lexer.Token
}

func (i *IntLit) Type() NodeType      { return NodeInt }
func (i *IntLit) Position() lexer.Token { return i.Pos }

type FloatLit struct {
	Value float64
	Pos   lexer.Token
}

func (f *FloatLit) Type() NodeType      { return NodeFloat }
func (f *FloatLit) Position() lexer.Token { return f.Pos }

type StringLit struct {
	Value string
	Pos   lexer.Token
}

func (s *StringLit) Type() NodeType      { return NodeString }
func (s *StringLit) Position() lexer.Token { return s.Pos }

type BoolLit struct {
	Value bool
	Pos   lexer.Token
}

func (b *BoolLit) Type() NodeType      { return NodeBool }
func (b *BoolLit) Position() lexer.Token { return b.Pos }

type NoneLit struct {
	Pos lexer.Token
}

func (n *NoneLit) Type() NodeType      { return NodeNone }
func (n *NoneLit) Position() lexer.Token { return n.Pos }

type ListLit struct {
	Elems []ASTNode
	Pos   lexer.Token
}

func (l *ListLit) Type() NodeType      { return NodeList }
func (l *ListLit) Position() lexer.Token { return l.Pos }

type MapEntry struct {
	Key   ASTNode
	Value ASTNode
}

type MapLit struct {
	Entries []MapEntry
	Pos     lexer.Token
}

func (m *MapLit) Type() NodeType      { return NodeMap }
func (m *MapLit) Position() lexer.Token { return m.Pos }

type RangeLit struct {
	From ASTNode
	To   ASTNode
	Pos  lexer.Token
}

func (r *RangeLit) Type() NodeType      { return NodeRange }
func (r *RangeLit) Position() lexer.Token { return r.Pos }

type FnLiteral struct {
	Params []Param
	Body   *Block
	Pos    lexer.Token
}

func (f *FnLiteral) Type() NodeType      { return NodeFnLiteral }
func (f *FnLiteral) Position() lexer.Token { return f.Pos }

type NamedArg struct {
	Name  string
	Value ASTNode
	Pos   lexer.Token
}

func (n *NamedArg) Type() NodeType      { return NodeExprStmt }
func (n *NamedArg) Position() lexer.Token { return n.Pos }

type RouteHandler struct {
	Method string
	Route  ASTNode
	Body   *Block
	Pos    lexer.Token
}

func (r *RouteHandler) Type() NodeType      { return NodeRouteHandler }
func (r *RouteHandler) Position() lexer.Token { return r.Pos }

type ServerDecl struct {
	Port ASTNode
	Pos  lexer.Token
}

func (s *ServerDecl) Type() NodeType      { return NodeServerDecl }
func (s *ServerDecl) Position() lexer.Token { return s.Pos }

type Page struct {
	Route string
	Body  *Block
	Pos   lexer.Token
}

func (p *Page) Type() NodeType      { return NodePage }
func (p *Page) Position() lexer.Token { return p.Pos }

type Component struct {
	Name   string
	Params []Param
	Body   *Block
	Pos    lexer.Token
}

func (c *Component) Type() NodeType      { return NodeComponent }
func (c *Component) Position() lexer.Token { return c.Pos }

type State struct {
	Name  string
	Value ASTNode
	Pos   lexer.Token
}

func (s *State) Type() NodeType      { return NodeState }
func (s *State) Position() lexer.Token { return s.Pos }

type HtmlElement struct {
	Tag        string
	Attributes []HtmlAttr
	Children   []ASTNode
	Pos        lexer.Token
}

func (h *HtmlElement) Type() NodeType      { return NodeHtmlElement }
func (h *HtmlElement) Position() lexer.Token { return h.Pos }

type HtmlAttr struct {
	Name  string
	Value ASTNode
}

type EventHandler struct {
	Event string
	Body  *Block
	Pos   lexer.Token
}

func (e *EventHandler) Type() NodeType      { return NodeEventHandler }
func (e *EventHandler) Position() lexer.Token { return e.Pos }

type StyleBlock struct {
	Properties []StyleProp
	Pos        lexer.Token
}

type StyleProp struct {
	Name  string
	Value ASTNode
}

func (s *StyleBlock) Type() NodeType      { return NodeStyleBlock }
func (s *StyleBlock) Position() lexer.Token { return s.Pos }

type AwaitExpr struct {
	Value ASTNode
	Pos   lexer.Token
}

func (a *AwaitExpr) Type() NodeType      { return NodeAwaitExpr }
func (a *AwaitExpr) Position() lexer.Token { return a.Pos }

type TryCatch struct {
	TryBody     *Block
	CatchVar    string
	CatchBody   *Block
	Pos         lexer.Token
}

func (t *TryCatch) Type() NodeType      { return NodeTryCatch }
func (t *TryCatch) Position() lexer.Token { return t.Pos }

type Throw struct {
	Value ASTNode
	Pos   lexer.Token
}

func (t *Throw) Type() NodeType      { return NodeThrow }
func (t *Throw) Position() lexer.Token { return t.Pos }

type TextInterp struct {
	Parts []ASTNode
	Pos   lexer.Token
}

func (t *TextInterp) Type() NodeType      { return NodeTextInterp }
func (t *TextInterp) Position() lexer.Token { return t.Pos }
