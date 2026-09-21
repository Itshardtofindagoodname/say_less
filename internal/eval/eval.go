package eval

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"sayless/internal/parser"
)

type Value struct {
	Type     string
	Int      int64
	Float    float64
	Str      string
	Bool     bool
	List     *[]Value
	Map      *map[string]Value
	None     bool
	Fn       *Function
	Callable *Builtin
	Return   *Value
	Break    bool
	Continue bool
	Struct   *StructVal
	Err      *ErrorVal
	Bytes    []byte
}

type Function struct {
	Params []parser.Param
	Body   *parser.Block
	Env    *Environment
	Name   string
}

type StructVal struct {
	Name   string
	Fields map[string]Value
}

type ErrorVal struct {
	Message string
}

type Builtin struct {
	Name string
	Fn   func(args []Value) Value
}

type Environment struct {
	vars  map[string]Value
	outer *Environment
}

func NewEnv(outer *Environment) *Environment {
	return &Environment{vars: make(map[string]Value), outer: outer}
}

func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.vars[name]; ok {
		return val, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return Value{}, false
}

func (e *Environment) Set(name string, val Value) {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = val
		return
	}
	if e.outer != nil {
		if _, ok := e.outer.Get(name); ok {
			e.outer.Set(name, val)
			return
		}
	}
	e.vars[name] = val
}

func (e *Environment) Define(name string, val Value) {
	e.vars[name] = val
}

type Interpreter struct {
	stdout *os.File
	stderr *os.File
	routes map[string]*Function
	port   int
}

func New() *Interpreter {
	return &Interpreter{
		stdout: os.Stdout,
		stderr: os.Stderr,
		routes: make(map[string]*Function),
		port:   8080,
	}
}

func makeList(items ...Value) *[]Value {
	return &items
}

func makeMap(m map[string]Value) *map[string]Value {
	return &m
}

func listLen(v Value) int {
	if v.List == nil {
		return 0
	}
	return len(*v.List)
}

func mapLen(v Value) int {
	if v.Map == nil {
		return 0
	}
	return len(*v.Map)
}

func (interp *Interpreter) Run(program *parser.Program) error {
	env := NewEnv(nil)
	interp.registerBuiltins(env)
	for _, stmt := range program.Stmts {
		_, err := interp.exec(stmt, env)
		if err != nil {
			return err
		}
	}
	if len(interp.routes) > 0 {
		return interp.startServer()
	}
	return nil
}

func (interp *Interpreter) RunWithEnv(program *parser.Program, env *Environment) error {
	for _, stmt := range program.Stmts {
		_, err := interp.exec(stmt, env)
		if err != nil {
			return err
		}
	}
	return nil
}

func (interp *Interpreter) RegisterBuiltinEnv(env *Environment) {
	interp.registerBuiltins(env)
}

func (interp *Interpreter) ExecStmt(node parser.ASTNode, env *Environment) (Value, error) {
	return interp.exec(node, env)
}

func (interp *Interpreter) exec(node parser.ASTNode, env *Environment) (Value, error) {
	if node == nil {
		return Value{Type: "none", None: true}, nil
	}
	switch n := node.(type) {
	case *parser.Assign:
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return Value{}, err
		}
		_, exists := env.Get(n.Name)
		if exists {
			env.Set(n.Name, val)
		} else {
			env.Define(n.Name, val)
		}
		return val, nil
	case *parser.AugAssign:
		val, ok := env.Get(n.Name)
		if !ok {
			return Value{}, fmt.Errorf("undefined variable %s", n.Name)
		}
		rhs, err := interp.eval(n.Value, env)
		if err != nil {
			return Value{}, err
		}
		result, err := applyAugAssign(val, n.Op, rhs)
		if err != nil {
			return Value{}, err
		}
		env.Set(n.Name, result)
		return result, nil
	case *parser.FnDef:
		fn := &Function{Params: n.Params, Body: n.Body, Env: env, Name: n.Name}
		env.Define(n.Name, Value{Type: "function", Fn: fn})
		return Value{Type: "function", Fn: fn}, nil
	case *parser.StructDef:
		env.Define(n.Name, Value{Type: "struct_type", Str: n.Name})
		return Value{Type: "none", None: true}, nil
	case *parser.If:
		return interp.execIf(n, env)
	case *parser.While:
		return interp.execWhile(n, env)
	case *parser.For:
		return interp.execFor(n, env)
	case *parser.Return:
		if n.Value != nil {
			val, err := interp.eval(n.Value, env)
			if err != nil {
				return Value{}, err
			}
			return Value{Type: "return", Return: &val}, nil
		}
		return Value{Type: "return"}, nil
	case *parser.Break:
		return Value{Type: "break", Break: true}, nil
	case *parser.Continue:
		return Value{Type: "continue", Continue: true}, nil
	case *parser.ExprStmt:
		return interp.eval(n.Expr, env)
	case *parser.Block:
		return interp.execBlock(n, NewEnv(env))
	case *parser.TestBlock:
		return interp.execTest(n, env)
	case *parser.Assert:
		return interp.execAssert(n, env)
	case *parser.Use:
		return Value{Type: "none", None: true}, nil
	case *parser.RouteHandler:
		routeVal, err := interp.eval(n.Route, env)
		if err != nil {
			return Value{}, err
		}
		fn := &Function{Params: []parser.Param{{Name: "request"}}, Body: n.Body, Env: env, Name: n.Method}
		key := strings.ToUpper(n.Method) + " " + routeVal.Str
		interp.routes[key] = fn
		fmt.Fprintf(interp.stdout, "Route registered: %s %s\n", strings.ToUpper(n.Method), routeVal.Str)
		return Value{Type: "none", None: true}, nil
	case *parser.ServerDecl:
		portVal, err := interp.eval(n.Port, env)
		if err != nil {
			return Value{}, err
		}
		interp.port = int(portVal.Int)
		return Value{Type: "none", None: true}, nil
	}
	return Value{Type: "none", None: true}, nil
}

func (interp *Interpreter) execBlock(block *parser.Block, env *Environment) (Value, error) {
	var last Value
	for _, stmt := range block.Stmts {
		val, err := interp.exec(stmt, env)
		if err != nil {
			return Value{}, err
		}
		last = val
		if val.Type == "return" || val.Break || val.Continue {
			return val, nil
		}
	}
	return last, nil
}

func (interp *Interpreter) startServer() error {
	mux := http.NewServeMux()
	for key, fn := range interp.routes {
		parts := strings.SplitN(key, " ", 2)
		method := parts[0]
		path := parts[1]
		capturedFn := fn
		capturedMethod := method
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if capturedMethod != "ANY" && r.Method != capturedMethod {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			reqEnv := NewEnv(capturedFn.Env)
			reqEnv.Define("request", Value{Type: "map", Map: makeMap(map[string]Value{
				"method": Value{Type: "string", Str: r.Method},
				"path":   Value{Type: "string", Str: r.URL.Path},
			})})
			result, err := interp.execBlock(capturedFn.Body, reqEnv)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, toString(result))
		})
	}
	addr := fmt.Sprintf(":%d", interp.port)
	fmt.Fprintf(interp.stdout, "Server running on http://localhost%s\n", addr)
	fmt.Fprintf(interp.stdout, "Press Ctrl+C to stop.\n")
	return http.ListenAndServe(addr, mux)
}

func (interp *Interpreter) execIf(n *parser.If, env *Environment) (Value, error) {
	cond, err := interp.eval(n.Cond, env)
	if err != nil {
		return Value{}, err
	}
	if isTruthy(cond) {
		return interp.execBlock(n.Then, NewEnv(env))
	}
	for _, elif := range n.ElseIf {
		cond, err := interp.eval(elif.Cond, env)
		if err != nil {
			return Value{}, err
		}
		if isTruthy(cond) {
			return interp.execBlock(elif.Then, NewEnv(env))
		}
	}
	if n.Else_ != nil {
		return interp.execBlock(n.Else_, NewEnv(env))
	}
	return Value{Type: "none", None: true}, nil
}

func (interp *Interpreter) execWhile(n *parser.While, env *Environment) (Value, error) {
	for {
		cond, err := interp.eval(n.Cond, env)
		if err != nil {
			return Value{}, err
		}
		if !isTruthy(cond) {
			break
		}
		val, err := interp.execBlock(n.Body, NewEnv(env))
		if err != nil {
			return Value{}, err
		}
		if val.Break {
			break
		}
		if val.Continue {
			continue
		}
	}
	return Value{Type: "none", None: true}, nil
}

func (interp *Interpreter) execFor(n *parser.For, env *Environment) (Value, error) {
	iterVal, err := interp.eval(n.Iter, env)
	if err != nil {
		return Value{}, err
	}
	if iterVal.Type == "range" {
		from := iterVal.Int
		to := int64(iterVal.Float)
		for i := from; i < to; i++ {
			loopEnv := NewEnv(env)
			loopEnv.Define(n.Var, Value{Type: "integer", Int: i})
			val, err := interp.execBlock(n.Body, loopEnv)
			if err != nil {
				return Value{}, err
			}
			if val.Break {
				break
			}
		}
	} else if iterVal.Type == "list" && iterVal.List != nil {
		for _, item := range *iterVal.List {
			loopEnv := NewEnv(env)
			loopEnv.Define(n.Var, item)
			val, err := interp.execBlock(n.Body, loopEnv)
			if err != nil {
				return Value{}, err
			}
			if val.Break {
				break
			}
		}
	} else if iterVal.Type == "map" && iterVal.Map != nil {
		for key, val := range *iterVal.Map {
			loopEnv := NewEnv(env)
			loopEnv.Define(n.Var, Value{Type: "string", Str: key})
			loopEnv.Define("_value", val)
			v, err := interp.execBlock(n.Body, loopEnv)
			if err != nil {
				return Value{}, err
			}
			if v.Break {
				break
			}
		}
	}
	return Value{Type: "none", None: true}, nil
}

func (interp *Interpreter) execTest(n *parser.TestBlock, env *Environment) (Value, error) {
	fmt.Fprintf(interp.stdout, "Running test: %s\n", n.Name)
	val, err := interp.execBlock(n.Body, NewEnv(env))
	if err != nil {
		fmt.Fprintf(interp.stderr, "  FAIL: %s - %v\n", n.Name, err)
		return Value{}, nil
	}
	if val.Err != nil {
		fmt.Fprintf(interp.stderr, "  FAIL: %s - %s\n", n.Name, val.Err.Message)
		return Value{}, nil
	}
	fmt.Fprintf(interp.stdout, "  PASS: %s\n", n.Name)
	return Value{Type: "none", None: true}, nil
}

func (interp *Interpreter) execAssert(n *parser.Assert, env *Environment) (Value, error) {
	val, err := interp.eval(n.Expr, env)
	if err != nil {
		return Value{Type: "error", Err: &ErrorVal{Message: err.Error()}}, nil
	}
	if !isTruthy(val) {
		return Value{Type: "error", Err: &ErrorVal{Message: "assertion failed"}}, nil
	}
	return Value{Type: "none", None: true}, nil
}
