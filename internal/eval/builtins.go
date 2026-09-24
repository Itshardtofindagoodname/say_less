package eval

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func (interp *Interpreter) registerBuiltins(env *Environment) {
	env.Define("print", Value{Type: "builtin", Callable: &Builtin{Name: "print", Fn: builtinPrint}})
	env.Define("println", Value{Type: "builtin", Callable: &Builtin{Name: "println", Fn: builtinPrintln}})
	env.Define("len", Value{Type: "builtin", Callable: &Builtin{Name: "len", Fn: builtinLen}})
	env.Define("to_string", Value{Type: "builtin", Callable: &Builtin{Name: "to_string", Fn: builtinToString}})
	env.Define("to_int", Value{Type: "builtin", Callable: &Builtin{Name: "to_int", Fn: builtinToInt}})
	env.Define("to_float", Value{Type: "builtin", Callable: &Builtin{Name: "to_float", Fn: builtinToFloat}})
	env.Define("type_of", Value{Type: "builtin", Callable: &Builtin{Name: "type_of", Fn: builtinTypeOf}})
	env.Define("exit", Value{Type: "builtin", Callable: &Builtin{Name: "exit", Fn: builtinExit}})
	env.Define("range", Value{Type: "builtin", Callable: &Builtin{Name: "range", Fn: builtinRange}})
	env.Define("push", Value{Type: "builtin", Callable: &Builtin{Name: "push", Fn: builtinPush}})
	env.Define("pop", Value{Type: "builtin", Callable: &Builtin{Name: "pop", Fn: builtinPop}})
	env.Define("input", Value{Type: "builtin", Callable: &Builtin{Name: "input", Fn: builtinInput}})
	env.Define("abs", Value{Type: "builtin", Callable: &Builtin{Name: "abs", Fn: builtinAbs}})
	env.Define("min", Value{Type: "builtin", Callable: &Builtin{Name: "min", Fn: builtinMin}})
	env.Define("max", Value{Type: "builtin", Callable: &Builtin{Name: "max", Fn: builtinMax}})
	env.Define("random", Value{Type: "builtin", Callable: &Builtin{Name: "random", Fn: builtinRandom}})
	env.Define("sleep", Value{Type: "builtin", Callable: &Builtin{Name: "sleep", Fn: builtinSleep}})
	env.Define("fail", Value{Type: "builtin", Callable: &Builtin{Name: "fail", Fn: builtinFail}})
	env.Define("ok", Value{Type: "builtin", Callable: &Builtin{Name: "ok", Fn: builtinOk}})
	env.Define("err", Value{Type: "builtin", Callable: &Builtin{Name: "err", Fn: builtinErr}})
	env.Define("keys", Value{Type: "builtin", Callable: &Builtin{Name: "keys", Fn: builtinKeys}})
	env.Define("values", Value{Type: "builtin", Callable: &Builtin{Name: "values", Fn: builtinValues}})
	env.Define("has", Value{Type: "builtin", Callable: &Builtin{Name: "has", Fn: builtinHas}})
	env.Define("slice", Value{Type: "builtin", Callable: &Builtin{Name: "slice", Fn: builtinSlice}})
	env.Define("char", Value{Type: "builtin", Callable: &Builtin{Name: "char", Fn: builtinChar}})
	env.Define("ord", Value{Type: "builtin", Callable: &Builtin{Name: "ord", Fn: builtinOrd}})
	env.Define("time", Value{Type: "builtin", Callable: &Builtin{Name: "time", Fn: builtinTime}})
	env.Define("clock", Value{Type: "builtin", Callable: &Builtin{Name: "clock", Fn: builtinClock}})

	env.Define("io", Value{Type: "module", Map: makeMap(map[string]Value{
		"read":  {Type: "builtin", Callable: &Builtin{Name: "io.read", Fn: builtinIORead}},
		"write": {Type: "builtin", Callable: &Builtin{Name: "io.write", Fn: builtinIOWrite}},
		"exists": {Type: "builtin", Callable: &Builtin{Name: "io.exists", Fn: builtinIOExists}},
	})})

	env.Define("math", Value{Type: "module", Map: makeMap(map[string]Value{
		"pi":   Value{Type: "float", Float: 3.141592653589793},
		"e":    Value{Type: "float", Float: 2.718281828459045},
		"sqrt": {Type: "builtin", Callable: &Builtin{Name: "math.sqrt", Fn: builtinMathSqrt}},
		"pow":  {Type: "builtin", Callable: &Builtin{Name: "math.pow", Fn: builtinMathPow}},
		"floor": {Type: "builtin", Callable: &Builtin{Name: "math.floor", Fn: builtinMathFloor}},
		"ceil":  {Type: "builtin", Callable: &Builtin{Name: "math.ceil", Fn: builtinMathCeil}},
		"sin":   {Type: "builtin", Callable: &Builtin{Name: "math.sin", Fn: builtinMathSin}},
		"cos":   {Type: "builtin", Callable: &Builtin{Name: "math.cos", Fn: builtinMathCos}},
		"tan":   {Type: "builtin", Callable: &Builtin{Name: "math.tan", Fn: builtinMathTan}},
		"log":   {Type: "builtin", Callable: &Builtin{Name: "math.log", Fn: builtinMathLog}},
	})})

	env.Define("strings", Value{Type: "module", Map: makeMap(map[string]Value{
		"contains": {Type: "builtin", Callable: &Builtin{Name: "strings.contains", Fn: builtinStringsContains}},
		"split":    {Type: "builtin", Callable: &Builtin{Name: "strings.split", Fn: builtinStringsSplit}},
		"join":     {Type: "builtin", Callable: &Builtin{Name: "strings.join", Fn: builtinStringsJoin}},
		"replace":  {Type: "builtin", Callable: &Builtin{Name: "strings.replace", Fn: builtinStringsReplace}},
		"has_prefix": {Type: "builtin", Callable: &Builtin{Name: "strings.has_prefix", Fn: builtinStringsHasPrefix}},
		"has_suffix": {Type: "builtin", Callable: &Builtin{Name: "strings.has_suffix", Fn: builtinStringsHasSuffix}},
	})})

	env.Define("os", Value{Type: "module", Map: makeMap(map[string]Value{
		"args": Value{Type: "list", List: func() *[]Value {
			var args []Value
			for _, a := range os.Args[1:] {
				args = append(args, Value{Type: "string", Str: a})
			}
			return &args
		}()},
		"env":   {Type: "builtin", Callable: &Builtin{Name: "os.env", Fn: builtinOSEnv}},
		"clock": {Type: "builtin", Callable: &Builtin{Name: "os.clock", Fn: builtinClock}},
	})})

	env.Define("json", Value{Type: "module", Default: "encode", Map: makeMap(map[string]Value{
		"encode": {Type: "builtin", Callable: &Builtin{Name: "json.encode", Fn: builtinJSONEncode}},
		"decode": {Type: "builtin", Callable: &Builtin{Name: "json.decode", Fn: builtinJSONDecode}},
	})})

	env.Define("assert", Value{Type: "builtin", Callable: &Builtin{Name: "assert", Fn: builtinAssert}})
}

func builtinPrint(args []Value) Value {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = toString(arg)
	}
	fmt.Println(strings.Join(parts, " "))
	return Value{Type: "none", None: true}
}

func builtinPrintln(args []Value) Value {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = toString(arg)
	}
	fmt.Println(strings.Join(parts, " "))
	return Value{Type: "none", None: true}
}

func builtinLen(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "integer", Int: 0}
	}
	arg := args[0]
	switch arg.Type {
	case "string":
		return Value{Type: "integer", Int: int64(len(arg.Str))}
	case "list":
		return Value{Type: "integer", Int: int64(listLen(arg))}
	case "map":
		return Value{Type: "integer", Int: int64(mapLen(arg))}
	}
	return Value{Type: "integer", Int: 0}
}

func builtinToString(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "string", Str: ""}
	}
	return Value{Type: "string", Str: toString(args[0])}
}

func builtinToInt(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "integer", Int: 0}
	}
	arg := args[0]
	if arg.Type == "integer" {
		return arg
	}
	if arg.Type == "float" {
		return Value{Type: "integer", Int: int64(arg.Float)}
	}
	if arg.Type == "string" {
		val, err := strconv.ParseInt(arg.Str, 0, 64)
		if err != nil {
			return Value{Type: "integer", Int: 0}
		}
		return Value{Type: "integer", Int: val}
	}
	return Value{Type: "integer", Int: 0}
}

func builtinToFloat(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	arg := args[0]
	if arg.Type == "float" {
		return arg
	}
	if arg.Type == "integer" {
		return Value{Type: "float", Float: float64(arg.Int)}
	}
	if arg.Type == "string" {
		val, err := strconv.ParseFloat(arg.Str, 64)
		if err != nil {
			return Value{Type: "float", Float: 0}
		}
		return Value{Type: "float", Float: val}
	}
	return Value{Type: "float", Float: 0}
}

func builtinTypeOf(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "string", Str: "none"}
	}
	return Value{Type: "string", Str: args[0].Type}
}

func builtinExit(args []Value) Value {
	code := 0
	if len(args) > 0 {
		code = int(toFloat(args[0]))
	}
	os.Exit(code)
	return Value{Type: "none", None: true}
}

func builtinRange(args []Value) Value {
	if len(args) == 1 {
		return Value{Type: "range", Int: 0, Float: toFloat(args[0])}
	}
	if len(args) >= 2 {
		return Value{Type: "range", Int: int64(toFloat(args[0])), Float: toFloat(args[1])}
	}
	return Value{Type: "range", Int: 0, Float: 0}
}

func builtinPush(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "none", None: true}
	}
	list := args[0]
	item := args[1]
	if list.Type == "list" && list.List != nil {
		*list.List = append(*list.List, item)
		return list
	}
	return Value{Type: "none", None: true}
}

func builtinPop(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "none", None: true}
	}
	list := args[0]
	if list.Type == "list" && list.List != nil && len(*list.List) > 0 {
		items := *list.List
		item := items[len(items)-1]
		*list.List = items[:len(items)-1]
		return item
	}
	return Value{Type: "none", None: true}
}

func builtinInput(args []Value) Value {
	prompt := ""
	if len(args) > 0 {
		prompt = toString(args[0])
	}
	if prompt != "" {
		fmt.Print(prompt)
	}
	var line string
	fmt.Scanln(&line)
	return Value{Type: "string", Str: line}
}

func builtinAbs(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	v := toFloat(args[0])
	if v < 0 {
		return Value{Type: "float", Float: -v}
	}
	return Value{Type: "float", Float: v}
}

func builtinMin(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	min := toFloat(args[0])
	for _, arg := range args[1:] {
		v := toFloat(arg)
		if v < min {
			min = v
		}
	}
	return Value{Type: "float", Float: min}
}

func builtinMax(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	max := toFloat(args[0])
	for _, arg := range args[1:] {
		v := toFloat(arg)
		if v > max {
			max = v
		}
	}
	return Value{Type: "float", Float: max}
}

func builtinRandom(args []Value) Value {
	return Value{Type: "float", Float: rand.Float64()}
}

func builtinSleep(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "none", None: true}
	}
	secs := toFloat(args[0])
	time.Sleep(time.Duration(secs * float64(time.Second)))
	return Value{Type: "none", None: true}
}

func builtinFail(args []Value) Value {
	msg := "error"
	if len(args) > 0 {
		msg = toString(args[0])
	}
	return Value{Type: "error", Err: &ErrorVal{Message: msg}}
}

func builtinOk(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "none", None: true}
	}
	return args[0]
}

func builtinErr(args []Value) Value {
	msg := "error"
	if len(args) > 0 {
		msg = toString(args[0])
	}
	return Value{Type: "error", Err: &ErrorVal{Message: msg}}
}

func builtinKeys(args []Value) Value {
	if len(args) < 1 || args[0].Type != "map" || args[0].Map == nil {
		return Value{Type: "list", List: makeList()}
	}
	var keys []Value
	for k := range *args[0].Map {
		keys = append(keys, Value{Type: "string", Str: k})
	}
	return Value{Type: "list", List: makeList(keys...)}
}

func builtinValues(args []Value) Value {
	if len(args) < 1 || args[0].Type != "map" || args[0].Map == nil {
		return Value{Type: "list", List: makeList()}
	}
	var vals []Value
	for _, v := range *args[0].Map {
		vals = append(vals, v)
	}
	return Value{Type: "list", List: makeList(vals...)}
}

func builtinHas(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "boolean", Bool: false}
	}
	obj := args[0]
	key := toString(args[1])
	if obj.Type == "map" && obj.Map != nil {
		_, ok := (*obj.Map)[key]
		return Value{Type: "boolean", Bool: ok}
	}
	if obj.Type == "string" {
		return Value{Type: "boolean", Bool: strings.Contains(obj.Str, key)}
	}
	return Value{Type: "boolean", Bool: false}
}

func builtinSlice(args []Value) Value {
	if len(args) < 3 {
		return Value{Type: "none", None: true}
	}
	obj := args[0]
	start := int(toFloat(args[1]))
	end := int(toFloat(args[2]))
	if obj.Type == "string" {
		if start < 0 { start = 0 }
		if end > len(obj.Str) { end = len(obj.Str) }
		if start > end { start = end }
		return Value{Type: "string", Str: obj.Str[start:end]}
	}
	if obj.Type == "list" && obj.List != nil {
		if start < 0 { start = 0 }
		if end > len(*obj.List) { end = len(*obj.List) }
		if start > end { start = end }
		sliced := (*obj.List)[start:end]
		return Value{Type: "list", List: makeList(sliced...)}
	}
	return Value{Type: "none", None: true}
}

func builtinChar(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "string", Str: ""}
	}
	n := int(toFloat(args[0]))
	return Value{Type: "string", Str: string(rune(n))}
}

func builtinOrd(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "integer", Int: 0}
	}
	s := toString(args[0])
	if len(s) > 0 {
		return Value{Type: "integer", Int: int64(s[0])}
	}
	return Value{Type: "integer", Int: 0}
}

func builtinTime(args []Value) Value {
	return Value{Type: "integer", Int: time.Now().Unix()}
}

func builtinClock(args []Value) Value {
	return Value{Type: "float", Float: float64(time.Now().UnixNano()) / float64(time.Second)}
}

func builtinAssert(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "none", None: true}
	}
	if !isTruthy(args[0]) {
		return Value{Type: "error", Err: &ErrorVal{Message: "assertion failed"}}
	}
	return Value{Type: "none", None: true}
}

func builtinIORead(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "error", Err: &ErrorVal{Message: "io.read requires a filename"}}
	}
	path := toString(args[0])
	data, err := os.ReadFile(path)
	if err != nil {
		return Value{Type: "error", Err: &ErrorVal{Message: err.Error()}}
	}
	return Value{Type: "string", Str: string(data)}
}

func builtinIOWrite(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "error", Err: &ErrorVal{Message: "io.write requires path and content"}}
	}
	path := toString(args[0])
	content := toString(args[1])
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return Value{Type: "error", Err: &ErrorVal{Message: err.Error()}}
	}
	return Value{Type: "none", None: true}
}

func builtinIOExists(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "boolean", Bool: false}
	}
	path := toString(args[0])
	_, err := os.Stat(path)
	return Value{Type: "boolean", Bool: err == nil}
}

func builtinMathSqrt(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Sqrt(toFloat(args[0]))}
}

func builtinMathPow(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Pow(toFloat(args[0]), toFloat(args[1]))}
}

func builtinMathFloor(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Floor(toFloat(args[0]))}
}

func builtinMathCeil(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Ceil(toFloat(args[0]))}
}

func builtinMathSin(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Sin(toFloat(args[0]))}
}

func builtinMathCos(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Cos(toFloat(args[0]))}
}

func builtinMathTan(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Tan(toFloat(args[0]))}
}

func builtinMathLog(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "float", Float: 0}
	}
	return Value{Type: "float", Float: math.Log(toFloat(args[0]))}
}

func builtinStringsContains(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "boolean", Bool: false}
	}
	return Value{Type: "boolean", Bool: strings.Contains(toString(args[0]), toString(args[1]))}
}

func builtinStringsSplit(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "list", List: makeList()}
	}
	parts := strings.Split(toString(args[0]), toString(args[1]))
	var result []Value
	for _, p := range parts {
		result = append(result, Value{Type: "string", Str: p})
	}
	return Value{Type: "list", List: makeList(result...)}
}

func builtinStringsJoin(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "string", Str: ""}
	}
	list := args[0]
	sep := toString(args[1])
	if list.Type == "list" && list.List != nil {
		parts := make([]string, len(*list.List))
		for i, item := range *list.List {
			parts[i] = toString(item)
		}
		return Value{Type: "string", Str: strings.Join(parts, sep)}
	}
	return Value{Type: "string", Str: ""}
}

func builtinStringsReplace(args []Value) Value {
	if len(args) < 3 {
		return Value{Type: "string", Str: ""}
	}
	s := toString(args[0])
	old := toString(args[1])
	new := toString(args[2])
	return Value{Type: "string", Str: strings.Replace(s, old, new, -1)}
}

func builtinStringsHasPrefix(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "boolean", Bool: false}
	}
	return Value{Type: "boolean", Bool: strings.HasPrefix(toString(args[0]), toString(args[1]))}
}

func builtinStringsHasSuffix(args []Value) Value {
	if len(args) < 2 {
		return Value{Type: "boolean", Bool: false}
	}
	return Value{Type: "boolean", Bool: strings.HasSuffix(toString(args[0]), toString(args[1]))}
}

func builtinOSEnv(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "string", Str: ""}
	}
	return Value{Type: "string", Str: os.Getenv(toString(args[0]))}
}

func builtinJSONEncode(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "string", Str: "{}"}
	}
	s := jsonEncodeValue(args[0])
	return Value{Type: "string", Str: s}
}

func builtinJSONDecode(args []Value) Value {
	if len(args) < 1 {
		return Value{Type: "error", Err: &ErrorVal{Message: "json.decode requires a string"}}
	}
	s := toString(args[0])
	val, err := jsonDecodeString(s)
	if err != nil {
		return Value{Type: "error", Err: &ErrorVal{Message: err.Error()}}
	}
	return val
}

func jsonEncodeValue(v Value) string {
	switch v.Type {
	case "none":
		return "null"
	case "boolean":
		if v.Bool {
			return "true"
		}
		return "false"
	case "integer":
		return strconv.FormatInt(v.Int, 10)
	case "float":
		return strconv.FormatFloat(v.Float, 'f', -1, 64)
	case "string":
		return "\"" + jsonEscapeString(v.Str) + "\""
	case "list":
		parts := make([]string, listLen(v))
		for i, item := range *v.List {
			parts[i] = jsonEncodeValue(item)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case "map":
		parts := make([]string, 0, mapLen(v))
		for key, val := range *v.Map {
			parts = append(parts, "\""+jsonEscapeString(key)+"\":"+jsonEncodeValue(val))
		}
		return "{" + strings.Join(parts, ",") + "}"
	case "struct":
		parts := make([]string, 0, len(v.Struct.Fields))
		for key, val := range v.Struct.Fields {
			parts = append(parts, "\""+jsonEscapeString(key)+"\":"+jsonEncodeValue(val))
		}
		return "{" + strings.Join(parts, ",") + "}"
	}
	return "null"
}

func jsonEscapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func jsonDecodeString(s string) (Value, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return Value{}, fmt.Errorf("empty JSON string")
	}
	if s[0] == '"' {
		s = strings.TrimPrefix(s, "\"")
		s = strings.TrimSuffix(s, "\"")
		return Value{Type: "string", Str: s}, nil
	}
	if s == "true" {
		return Value{Type: "boolean", Bool: true}, nil
	}
	if s == "false" {
		return Value{Type: "boolean", Bool: false}, nil
	}
	if s == "null" {
		return Value{Type: "none", None: true}, nil
	}
	if s[0] >= '0' && s[0] <= '9' || s[0] == '-' {
		if strings.Contains(s, ".") {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				return Value{Type: "float", Float: f}, nil
			}
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return Value{Type: "integer", Int: n}, nil
		}
	}
	if s[0] == '[' {
		return jsonDecodeList(s)
	}
	if s[0] == '{' {
		return jsonDecodeMap(s)
	}
	return Value{Type: "string", Str: s}, nil
}

func jsonDecodeList(s string) (Value, error) {
	s = strings.TrimSpace(s[1:])
	s = strings.TrimSuffix(s, "]")
	if len(strings.TrimSpace(s)) == 0 {
		return Value{Type: "list", List: makeList()}, nil
	}
	parts := jsonSplitTop(s)
	var items []Value
	for _, part := range parts {
		val, err := jsonDecodeString(strings.TrimSpace(part))
		if err != nil {
			return Value{}, err
		}
		items = append(items, val)
	}
	return Value{Type: "list", List: makeList(items...)}, nil
}

func jsonDecodeMap(s string) (Value, error) {
	s = strings.TrimSpace(s[1:])
	s = strings.TrimSuffix(s, "}")
	if len(strings.TrimSpace(s)) == 0 {
		return Value{Type: "map", Map: makeMap(make(map[string]Value))}, nil
	}
	pairs := jsonSplitTop(s)
	m := make(map[string]Value)
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		colonIdx := strings.Index(pair, ":")
		if colonIdx < 0 {
			continue
		}
		key := strings.TrimSpace(pair[:colonIdx])
		key = strings.Trim(key, "\"")
		val, err := jsonDecodeString(strings.TrimSpace(pair[colonIdx+1:]))
		if err != nil {
			return Value{}, err
		}
		m[key] = val
	}
	return Value{Type: "map", Map: makeMap(m)}, nil
}

func jsonSplitTop(s string) []string {
	var parts []string
	depth := 0
	inString := false
	escaped := false
	start := 0
	for i, ch := range s {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '[' || ch == '{' {
			depth++
		} else if ch == ']' || ch == '}' {
			depth--
		} else if ch == ',' && depth == 0 {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}


