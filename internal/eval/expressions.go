package eval

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"sayless/internal/parser"
)

func (interp *Interpreter) eval(node parser.ASTNode, env *Environment) (Value, error) {
	if node == nil {
		return Value{Type: "none", None: true}, nil
	}
	switch n := node.(type) {
	case *parser.IntLit:
		return Value{Type: "integer", Int: n.Value}, nil
	case *parser.FloatLit:
		return Value{Type: "float", Float: n.Value}, nil
	case *parser.StringLit:
		return Value{Type: "string", Str: n.Value}, nil
	case *parser.BoolLit:
		return Value{Type: "boolean", Bool: n.Value}, nil
	case *parser.NoneLit:
		return Value{Type: "none", None: true}, nil
	case *parser.Ident:
		val, ok := env.Get(n.Name)
		if !ok {
			return Value{}, fmt.Errorf("line %d: undefined variable %q", n.Pos.Line, n.Name)
		}
		return val, nil
	case *parser.Binary:
		return interp.evalBinary(n, env)
	case *parser.Unary:
		return interp.evalUnary(n, env)
	case *parser.Call:
		return interp.evalCall(n, env)
	case *parser.Member:
		return interp.evalMember(n, env)
	case *parser.ListLit:
		var elems []Value
		for _, e := range n.Elems {
			val, err := interp.eval(e, env)
			if err != nil {
				return Value{}, err
			}
			elems = append(elems, val)
		}
		return Value{Type: "list", List: makeList(elems...)}, nil
	case *parser.MapLit:
		entries := make(map[string]Value)
		for _, entry := range n.Entries {
			key, err := interp.eval(entry.Key, env)
			if err != nil {
				return Value{}, err
			}
			val, err := interp.eval(entry.Value, env)
			if err != nil {
				return Value{}, err
			}
			entries[key.Str] = val
		}
		return Value{Type: "map", Map: makeMap(entries)}, nil
	case *parser.FnLiteral:
		fn := &Function{Params: n.Params, Body: n.Body, Env: env}
		return Value{Type: "function", Fn: fn}, nil
	case *parser.NamedArg:
		return interp.eval(n.Value, env)
	case *parser.Assign:
		val, err := interp.eval(n.Value, env)
		if err != nil {
			return Value{}, err
		}
		env.Define(n.Name, val)
		return val, nil
	case *parser.ExprStmt:
		return interp.eval(n.Expr, env)
	case *parser.Block:
		return interp.execBlock(n, NewEnv(env))
	}
	return Value{Type: "none", None: true}, nil
}

func (interp *Interpreter) evalBinary(n *parser.Binary, env *Environment) (Value, error) {
	left, err := interp.eval(n.Left, env)
	if err != nil {
		return Value{}, err
	}
	if n.Op == "and" {
		if !isTruthy(left) {
			return left, nil
		}
		right, err := interp.eval(n.Right, env)
		if err != nil {
			return Value{}, err
		}
		return right, nil
	}
	if n.Op == "or" {
		if isTruthy(left) {
			return left, nil
		}
		right, err := interp.eval(n.Right, env)
		if err != nil {
			return Value{}, err
		}
		return right, nil
	}
	right, err := interp.eval(n.Right, env)
	if err != nil {
		return Value{}, err
	}
	switch n.Op {
	case "+":
		if left.Type == "string" || right.Type == "string" {
			return Value{Type: "string", Str: toString(left) + toString(right)}, nil
		}
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int + right.Int}, nil
		}
		if left.Type == "float" || right.Type == "float" {
			return Value{Type: "float", Float: toFloat(left) + toFloat(right)}, nil
		}
		return Value{}, fmt.Errorf("cannot add %s and %s", left.Type, right.Type)
	case "-":
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int - right.Int}, nil
		}
		return Value{Type: "float", Float: toFloat(left) - toFloat(right)}, nil
	case "*":
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int * right.Int}, nil
		}
		return Value{Type: "float", Float: toFloat(left) * toFloat(right)}, nil
	case "/":
		if right.Type == "integer" && right.Int == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int / right.Int}, nil
		}
		return Value{Type: "float", Float: toFloat(left) / toFloat(right)}, nil
	case "%":
		if left.Type == "integer" && right.Type == "integer" {
			if right.Int == 0 {
				return Value{}, fmt.Errorf("modulo by zero")
			}
			return Value{Type: "integer", Int: left.Int % right.Int}, nil
		}
		return Value{Type: "float", Float: math.Mod(toFloat(left), toFloat(right))}, nil
	case "==":
		return Value{Type: "boolean", Bool: valuesEqual(left, right)}, nil
	case "!=":
		return Value{Type: "boolean", Bool: !valuesEqual(left, right)}, nil
	case "<":
		return Value{Type: "boolean", Bool: toFloat(left) < toFloat(right)}, nil
	case ">":
		return Value{Type: "boolean", Bool: toFloat(left) > toFloat(right)}, nil
	case "<=":
		return Value{Type: "boolean", Bool: toFloat(left) <= toFloat(right)}, nil
	case ">=":
		return Value{Type: "boolean", Bool: toFloat(left) >= toFloat(right)}, nil
	case "..":
		return Value{Type: "range", Int: int64(toFloat(left)), Float: toFloat(right)}, nil
	}
	return Value{}, fmt.Errorf("unknown operator %s", n.Op)
}

func (interp *Interpreter) evalUnary(n *parser.Unary, env *Environment) (Value, error) {
	val, err := interp.eval(n.Expr, env)
	if err != nil {
		return Value{}, err
	}
	switch n.Op {
	case "-":
		if val.Type == "integer" {
			return Value{Type: "integer", Int: -val.Int}, nil
		}
		return Value{Type: "float", Float: -toFloat(val)}, nil
	case "+":
		return val, nil
	case "not":
		return Value{Type: "boolean", Bool: !isTruthy(val)}, nil
	}
	return Value{}, fmt.Errorf("unknown unary operator %s", n.Op)
}

func (interp *Interpreter) evalCall(n *parser.Call, env *Environment) (Value, error) {
	callee, err := interp.eval(n.Callee, env)
	if err != nil {
		return Value{}, err
	}
	var args []Value
	for _, arg := range n.Args {
		val, err := interp.eval(arg, env)
		if err != nil {
			return Value{}, err
		}
		args = append(args, val)
	}
	if callee.Callable != nil {
		return callee.Callable.Fn(args), nil
	}
	if callee.Type == "function" {
		fn := callee.Fn
		fnEnv := NewEnv(fn.Env)
		argIdx := 0
		for _, param := range fn.Params {
			// Check if a named arg matches this param
			found := false
			for i, a := range n.Args {
				if na, ok := a.(*parser.NamedArg); ok && na.Name == param.Name {
					val, err := interp.eval(na.Value, env)
					if err != nil {
						return Value{}, err
					}
					fnEnv.Define(param.Name, val)
					found = true
					_ = i
					break
				}
			}
			if found {
				continue
			}
			if argIdx < len(args) {
				fnEnv.Define(param.Name, args[argIdx])
				argIdx++
			} else if param.Default != nil {
				defVal, err := interp.eval(param.Default, fn.Env)
				if err != nil {
					return Value{}, err
				}
				fnEnv.Define(param.Name, defVal)
			}
		}
		val, err := interp.execBlock(fn.Body, fnEnv)
		if err != nil {
			return Value{}, err
		}
		if val.Type == "return" && val.Return != nil {
			return *val.Return, nil
		}
		return val, nil
	}
	if callee.Type == "struct_type" {
		fields := make(map[string]Value)
		for i, arg := range n.Args {
			if na, ok := arg.(*parser.NamedArg); ok {
				val, err := interp.eval(na.Value, env)
				if err != nil {
					return Value{}, err
				}
				fields[na.Name] = val
			} else if i < len(args) {
				if ident, ok := arg.(*parser.Ident); ok {
					fields[ident.Name] = args[i]
				}
			}
		}
		return Value{Type: "struct", Struct: &StructVal{Name: callee.Str, Fields: fields}}, nil
	}
	if callee.Type == "module" {
		if callee.Default != "" && callee.Map != nil {
			if member, ok := (*callee.Map)[callee.Default]; ok && member.Callable != nil {
				return member.Callable.Fn(args), nil
			}
		}
		return Value{}, fmt.Errorf("module %s cannot be called directly; use %s.<member>(...)", callee.Default, callee.Default)
	}
	return Value{}, fmt.Errorf("cannot call %s", callee.Type)
}

func (interp *Interpreter) evalMember(n *parser.Member, env *Environment) (Value, error) {
	obj, err := interp.eval(n.Object, env)
	if err != nil {
		return Value{}, err
	}
	if obj.Type == "struct" {
		if val, ok := obj.Struct.Fields[n.Member]; ok {
			return val, nil
		}
		return Value{}, fmt.Errorf("no field %s in struct %s", n.Member, obj.Struct.Name)
	}
	if obj.Type == "map" {
		if obj.Map != nil {
			if val, ok := (*obj.Map)[n.Member]; ok {
				return val, nil
			}
		}
		return Value{Type: "none", None: true}, nil
	}
	if obj.Type == "string" {
		return interp.evalStringMember(obj, n.Member)
	}
	if obj.Type == "list" {
		return interp.evalListMember(obj, n.Member)
	}
	if obj.Type == "module" {
		if obj.Map != nil {
			if val, ok := (*obj.Map)[n.Member]; ok {
				return val, nil
			}
		}
		return Value{}, fmt.Errorf("module has no member %s", n.Member)
	}
	if obj.Type == "error" {
		if n.Member == "message" || n.Member == "error" {
			return Value{Type: "string", Str: obj.Err.Message}, nil
		}
		if n.Member == "err" {
			return Value{Type: "boolean", Bool: true}, nil
		}
		if n.Member == "ok" {
			return Value{Type: "boolean", Bool: false}, nil
		}
		return Value{}, fmt.Errorf("error has no member %s", n.Member)
	}
	return Value{}, fmt.Errorf("cannot access member %s on %s", n.Member, obj.Type)
}

func (interp *Interpreter) evalStringMember(obj Value, member string) (Value, error) {
	switch member {
	case "length":
		return Value{Type: "integer", Int: int64(len(obj.Str))}, nil
	case "upper":
		return Value{Type: "builtin", Callable: &Builtin{Name: "upper", Fn: func(args []Value) Value {
			result := ""
			for _, ch := range obj.Str {
				if ch >= 'a' && ch <= 'z' {
					result += string(ch - 32)
				} else {
					result += string(ch)
				}
			}
			return Value{Type: "string", Str: result}
		}}}, nil
	case "lower":
		return Value{Type: "builtin", Callable: &Builtin{Name: "lower", Fn: func(args []Value) Value {
			result := ""
			for _, ch := range obj.Str {
				if ch >= 'A' && ch <= 'Z' {
					result += string(ch + 32)
				} else {
					result += string(ch)
				}
			}
			return Value{Type: "string", Str: result}
		}}}, nil
	case "trim":
		return Value{Type: "builtin", Callable: &Builtin{Name: "trim", Fn: func(args []Value) Value {
			s := obj.Str
			for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
				s = s[1:]
			}
			for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
				s = s[:len(s)-1]
			}
			return Value{Type: "string", Str: s}
		}}}, nil
	case "contains":
		return Value{Type: "builtin", Callable: &Builtin{Name: "contains", Fn: func(args []Value) Value {
			if len(args) < 1 { return Value{Type: "boolean", Bool: false} }
			return Value{Type: "boolean", Bool: strings.Contains(obj.Str, toString(args[0]))}
		}}}, nil
	case "split":
		return Value{Type: "builtin", Callable: &Builtin{Name: "split", Fn: func(args []Value) Value {
			sep := " "
			if len(args) > 0 { sep = toString(args[0]) }
			parts := strings.Split(obj.Str, sep)
			var result []Value
			for _, p := range parts { result = append(result, Value{Type: "string", Str: p}) }
			return Value{Type: "list", List: makeList(result...)}
		}}}, nil
	case "replace":
		return Value{Type: "builtin", Callable: &Builtin{Name: "replace", Fn: func(args []Value) Value {
			if len(args) < 2 { return obj }
			return Value{Type: "string", Str: strings.Replace(obj.Str, toString(args[0]), toString(args[1]), -1)}
		}}}, nil
	case "starts_with":
		return Value{Type: "builtin", Callable: &Builtin{Name: "starts_with", Fn: func(args []Value) Value {
			if len(args) < 1 { return Value{Type: "boolean", Bool: false} }
			return Value{Type: "boolean", Bool: strings.HasPrefix(obj.Str, toString(args[0]))}
		}}}, nil
	case "ends_with":
		return Value{Type: "builtin", Callable: &Builtin{Name: "ends_with", Fn: func(args []Value) Value {
			if len(args) < 1 { return Value{Type: "boolean", Bool: false} }
			return Value{Type: "boolean", Bool: strings.HasSuffix(obj.Str, toString(args[0]))}
		}}}, nil
	case "index_of":
		return Value{Type: "builtin", Callable: &Builtin{Name: "index_of", Fn: func(args []Value) Value {
			if len(args) < 1 { return Value{Type: "integer", Int: -1} }
			idx := strings.Index(obj.Str, toString(args[0]))
			return Value{Type: "integer", Int: int64(idx)}
		}}}, nil
	case "substring":
		return Value{Type: "builtin", Callable: &Builtin{Name: "substring", Fn: func(args []Value) Value {
			if len(args) < 2 { return obj }
			start := int(toFloat(args[0]))
			end := int(toFloat(args[1]))
			if start < 0 { start = 0 }
			if end > len(obj.Str) { end = len(obj.Str) }
			return Value{Type: "string", Str: obj.Str[start:end]}
		}}}, nil
	case "chars":
		return Value{Type: "builtin", Callable: &Builtin{Name: "chars", Fn: func(args []Value) Value {
			var chars []Value
			for _, ch := range obj.Str {
				chars = append(chars, Value{Type: "string", Str: string(ch)})
			}
			return Value{Type: "list", List: makeList(chars...)}
		}}}, nil
	case "to_int":
		return Value{Type: "builtin", Callable: &Builtin{Name: "to_int", Fn: func(args []Value) Value {
			n, _ := strconv.ParseInt(obj.Str, 0, 64)
			return Value{Type: "integer", Int: n}
		}}}, nil
	}
	return Value{}, fmt.Errorf("no method '%s' on string", member)
}

func (interp *Interpreter) evalListMember(obj Value, member string) (Value, error) {
	switch member {
	case "length":
		return Value{Type: "integer", Int: int64(len(*obj.List))}, nil
	case "push":
		return Value{Type: "builtin", Callable: &Builtin{Name: "push", Fn: func(args []Value) Value {
			if len(args) < 1 { return obj }
			*obj.List = append(*obj.List, args[0])
			return obj
		}}}, nil
	case "pop":
		return Value{Type: "builtin", Callable: &Builtin{Name: "pop", Fn: func(args []Value) Value {
			if len(*obj.List) == 0 { return Value{Type: "none", None: true} }
			items := *obj.List
			item := items[len(items)-1]
			*obj.List = items[:len(items)-1]
			return item
		}}}, nil
	case "join":
		return Value{Type: "builtin", Callable: &Builtin{Name: "join", Fn: func(args []Value) Value {
			sep := ", "
			if len(args) > 0 { sep = toString(args[0]) }
			parts := make([]string, len(*obj.List))
			for i, item := range *obj.List { parts[i] = toString(item) }
			return Value{Type: "string", Str: strings.Join(parts, sep)}
		}}}, nil
	case "map":
		return Value{Type: "builtin", Callable: &Builtin{Name: "map", Fn: func(args []Value) Value {
			if len(args) < 1 || args[0].Type != "function" { return obj }
			fn := args[0].Fn
			var result []Value
			for _, item := range *obj.List {
				fnEnv := NewEnv(fn.Env)
				fnEnv.Define(fn.Params[0].Name, item)
				val, err := interp.execBlock(fn.Body, fnEnv)
				if err == nil {
					result = append(result, val)
				}
			}
			return Value{Type: "list", List: makeList(result...)}
		}}}, nil
	case "filter":
		return Value{Type: "builtin", Callable: &Builtin{Name: "filter", Fn: func(args []Value) Value {
			if len(args) < 1 || args[0].Type != "function" { return obj }
			fn := args[0].Fn
			var result []Value
			for _, item := range *obj.List {
				fnEnv := NewEnv(fn.Env)
				fnEnv.Define(fn.Params[0].Name, item)
				val, err := interp.execBlock(fn.Body, fnEnv)
				if err == nil && isTruthy(val) {
					result = append(result, item)
				}
			}
			return Value{Type: "list", List: makeList(result...)}
		}}}, nil
	case "find":
		return Value{Type: "builtin", Callable: &Builtin{Name: "find", Fn: func(args []Value) Value {
			if len(args) < 1 { return Value{Type: "none", None: true} }
			for _, item := range *obj.List {
				if valuesEqual(item, args[0]) {
					return item
				}
			}
			return Value{Type: "none", None: true}
		}}}, nil
	}
	return Value{}, fmt.Errorf("no method '%s' on list", member)
}

func isTruthy(v Value) bool {
	if v.Type == "none" {
		return false
	}
	if v.Type == "boolean" {
		return v.Bool
	}
	if v.Type == "integer" {
		return v.Int != 0
	}
	if v.Type == "float" {
		return v.Float != 0
	}
	if v.Type == "string" {
		return len(v.Str) > 0
	}
	if v.Type == "list" {
		return listLen(v) > 0
	}
	return true
}

func valuesEqual(a, b Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case "none":
		return true
	case "boolean":
		return a.Bool == b.Bool
	case "integer":
		return a.Int == b.Int
	case "float":
		return a.Float == b.Float
	case "string":
		return a.Str == b.Str
	}
	return false
}

func toString(v Value) string {
	switch v.Type {
	case "integer":
		return strconv.FormatInt(v.Int, 10)
	case "float":
		return strconv.FormatFloat(v.Float, 'f', -1, 64)
	case "boolean":
		if v.Bool {
			return "true"
		}
		return "false"
	case "string":
		return v.Str
	case "none":
		return "none"
	}
	return "<value>"
}

func toFloat(v Value) float64 {
	switch v.Type {
	case "integer":
		return float64(v.Int)
	case "float":
		return v.Float
	case "boolean":
		if v.Bool {
			return 1
		}
		return 0
	}
	return 0
}

func applyAugAssign(left Value, op string, right Value) (Value, error) {
	switch op {
	case "+=":
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int + right.Int}, nil
		}
		if left.Type == "float" || right.Type == "float" {
			return Value{Type: "float", Float: toFloat(left) + toFloat(right)}, nil
		}
		if left.Type == "string" {
			return Value{Type: "string", Str: left.Str + toString(right)}, nil
		}
	case "-=":
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int - right.Int}, nil
		}
		return Value{Type: "float", Float: toFloat(left) - toFloat(right)}, nil
	case "*=":
		if left.Type == "integer" && right.Type == "integer" {
			return Value{Type: "integer", Int: left.Int * right.Int}, nil
		}
		return Value{Type: "float", Float: toFloat(left) * toFloat(right)}, nil
	case "/=":
		if left.Type == "integer" && right.Type == "integer" {
			if right.Int == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return Value{Type: "integer", Int: left.Int / right.Int}, nil
		}
		return Value{Type: "float", Float: toFloat(left) / toFloat(right)}, nil
	case "%=":
		if left.Type == "integer" && right.Type == "integer" {
			if right.Int == 0 {
				return Value{}, fmt.Errorf("modulo by zero")
			}
			return Value{Type: "integer", Int: left.Int % right.Int}, nil
		}
		return Value{Type: "float", Float: math.Mod(toFloat(left), toFloat(right))}, nil
	}
	return Value{}, fmt.Errorf("unknown augmented assignment operator %s", op)
}
