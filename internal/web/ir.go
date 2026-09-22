package web

import "sayless/internal/parser"

// WebIR is the intermediate representation for web compilation
type WebIR struct {
	Pages      []*PageNode
	Components []*ComponentNode
	Styles     []*StyleRule
	Server     *ServerNode
}

type PageNode struct {
	Route    string
	Elements []IRNode
	Meta     map[string]string
}

type ComponentNode struct {
	Name       string
	Params     []parser.Param
	Elements   []IRNode
	IslandID   string
	HasState   bool
	HasEvents  bool
}

type ServerNode struct {
	Port   int
	Routes []*RouteNode
}

type RouteNode struct {
	Method string
	Path   string
}

type StyleRule struct {
	Selector   string
	Properties []StyleProperty
	IsScoped   bool
	ComponentID string
}

type StyleProperty struct {
	Name  string
	Value string
}

// IRNode is the interface for all IR nodes
type IRNode interface {
	IRType() string
}

type IRText struct {
	Content string
}

func (t *IRText) IRType() string { return "text" }

type IRInterp struct {
	Parts []IRNode
}

func (i *IRInterp) IRType() string { return "interp" }

type IRBinding struct {
	Expr string
}

func (b *IRBinding) IRType() string { return "binding" }

type IRElement struct {
	Tag        string
	Attributes []*IRAttribute
	Children   []IRNode
	Events     []*IREvent
	Styles     *StyleBlock
	IsIsland   bool
	ComponentID string
}

func (e *IRElement) IRType() string { return "element" }

type IRAttribute struct {
	Name  string
	Value string
	IsExpr bool
}

type IREvent struct {
	Name    string
	Handler string
	IsExpr  bool
}

func (e *IREvent) IRType() string { return "event" }

type StyleBlock struct {
	Properties []*StyleProperty
}

func (s *StyleBlock) IRType() string { return "style" }

type IRComponent struct {
	Name       string
	Props      []*IRPropBinding
	Children   []IRNode
	InstanceID string
}

func (c *IRComponent) IRType() string { return "component" }

type IRPropBinding struct {
	Name  string
	Value string
	IsExpr bool
}

type IRState struct {
	Name         string
	InitialValue string
	Dependencies []string
	AffectedDOM  []string
}

func (s *IRState) IRType() string { return "state" }

type IRConditional struct {
	Condition string
	Then      []IRNode
	Else      []IRNode
}

func (c *IRConditional) IRType() string { return "conditional" }

type IRLoop struct {
	Variable string
	Iterable string
	Body     []IRNode
}

func (l *IRLoop) IRType() string { return "loop" }

type IRAwait struct {
	Expr       string
	Loading    []IRNode
	Error      []IRNode
	Success    []IRNode
	VarName    string
}

func (a *IRAwait) IRType() string { return "await" }

type IRFragment struct {
	Children []IRNode
}

func (f *IRFragment) IRType() string { return "fragment" }
