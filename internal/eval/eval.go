package eval

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	Default  string
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
	stdout      *os.File
	stderr      *os.File
	routes      map[string]*Function
	port        int
	forcePort   int
	httpServer  *http.Server
	liveReload  bool
	revision    int
	NpmPackages []string
	PipPackages []string
}

func New() *Interpreter {
	return &Interpreter{
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		routes:    make(map[string]*Function),
		port:      8080,
		forcePort: 0,
	}
}

// SetPort overrides the port declared by "server on <port>" in the program.
// A value of 0 means the program's declared port is used.
func (interp *Interpreter) SetPort(p int) {
	interp.forcePort = p
}

// EnableLiveReload turns on dev-mode live reload. When enabled, HTML responses
// get a reload script injected and plain-text/JSON responses are wrapped in a
// small HTML page, so editing the source file reloads the browser
// automatically. It is used by the backend live-reload runner.
func (interp *Interpreter) EnableLiveReload(enabled bool) {
	interp.liveReload = enabled
}

// SetRevision sets the revision advertised by the /__sale_reload endpoint. The
// live-reload runner increments this each time it restarts the server so
// browsers watching the page know to reload.
func (interp *Interpreter) SetRevision(revision int) {
	interp.revision = revision
}

// hasRoute reports whether any registered user route serves the given path.
func (interp *Interpreter) hasRoute(path string) bool {
	suffix := " " + path
	for key := range interp.routes {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}
	return false
}

// Shutdown stops the running HTTP server, releasing the port. It is safe to
// call multiple times and is used by the live-reload runner to restart the
// server with updated routes.
func (interp *Interpreter) Shutdown() {
	if interp.httpServer != nil {
		srv := interp.httpServer
		interp.httpServer = nil
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
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
		return interp.usePackage(n, env)
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
		if interp.forcePort > 0 {
			interp.port = interp.forcePort
		} else {
			portVal, err := interp.eval(n.Port, env)
			if err != nil {
				return Value{}, err
			}
			interp.port = int(portVal.Int)
		}
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

func (interp *Interpreter) usePackage(n *parser.Use, env *Environment) (Value, error) {
	pkg := n.Path

	// Check if it's a built-in module first (http, json, io, math, strings, os)
	if isBuiltinModule(pkg) {
		return Value{Type: "none", None: true}, nil
	}

	// Check for npm: prefix
	if strings.HasPrefix(pkg, "npm:") {
		pkgName := strings.TrimPrefix(pkg, "npm:")
		if !isPackageInstalled(pkgName, "npm") {
			interp.Fprintf(interp.stdout, "Installing npm package: %s\n", pkgName)
			if err := installPackage(pkgName, "npm"); err != nil {
				return Value{}, fmt.Errorf("failed to install npm package %s: %v", pkgName, err)
			}
		}
		// Track the npm package for HTML injection
		for _, p := range interp.NpmPackages {
			if p == pkgName {
				return Value{Type: "none", None: true}, nil
			}
		}
		interp.NpmPackages = append(interp.NpmPackages, pkgName)
		fmt.Fprintf(interp.stdout, "Loaded npm package: %s\n", pkgName)
		return Value{Type: "none", None: true}, nil
	}

	// Check for pip: prefix
	if strings.HasPrefix(pkg, "pip:") {
		pkgName := strings.TrimPrefix(pkg, "pip:")
		if !isPackageInstalled(pkgName, "pip") {
			interp.Fprintf(interp.stdout, "Installing pip package: %s\n", pkgName)
			if err := installPackage(pkgName, "pip"); err != nil {
				return Value{}, fmt.Errorf("failed to install pip package %s: %v", pkgName, err)
			}
		}
		// Track the pip package and expose as a callable
		for _, p := range interp.PipPackages {
			if p == pkgName {
				return Value{Type: "none", None: true}, nil
			}
		}
		interp.PipPackages = append(interp.PipPackages, pkgName)

		// Expose pip package as a module with a call method
		pkg := pkgName
		env.Define(pkgName, Value{Type: "module", Map: makeMap(map[string]Value{
			"call": {Type: "builtin", Callable: &Builtin{Name: pkgName + ".call", Fn: func(args []Value) Value {
				return callPipPackage(pkg, args)
			}}},
			"exec": {Type: "builtin", Callable: &Builtin{Name: pkgName + ".exec", Fn: func(args []Value) Value {
				return execPipPackage(pkg, args)
			}}},
		})})
		fmt.Fprintf(interp.stdout, "Loaded pip package: %s\n", pkgName)
		return Value{Type: "none", None: true}, nil
	}

	// Built-in modules (http, json, io, etc.) - no-op, already registered
	return Value{Type: "none", None: true}, nil
}

func isBuiltinModule(name string) bool {
	switch name {
	case "http", "json", "io", "math", "strings", "os", "env":
		return true
	}
	return false
}

func isPackageInstalled(name, manager string) bool {
	switch manager {
	case "npm":
		_, err := os.Stat(filepath.Join("node_modules", name))
		return err == nil
	case "pip":
		cmd := exec.Command("python", "-c", "import "+name)
		return cmd.Run() == nil
	}
	return false
}

func installPackage(name, manager string) error {
	switch manager {
	case "npm":
		cmd := exec.Command("npm", "install", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "pip":
		cmd := exec.Command("pip", "install", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return fmt.Errorf("unknown package manager: %s", manager)
}

func callPipPackage(pkg string, args []Value) Value {
	// Build Python expression: import pkg; result = pkg.func(*args)
	// For simplicity, we pass args as a JSON string and let the package handle it
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = "'" + toString(arg) + "'"
	}
	pyCode := fmt.Sprintf("import %s; print(%s(%s))", pkg, pkg, strings.Join(parts, ", "))
	cmd := exec.Command("python", "-c", pyCode)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Value{Type: "error", Err: &ErrorVal{Message: fmt.Sprintf("pip package %s error: %s", pkg, string(out))}}
	}
	return Value{Type: "string", Str: strings.TrimSpace(string(out))}
}

func execPipPackage(pkg string, args []Value) Value {
	// Execute arbitrary Python code with the package imported
	if len(args) < 1 {
		return Value{Type: "error", Err: &ErrorVal{Message: "exec requires a Python code string"}}
	}
	code := toString(args[0])
	pyCode := fmt.Sprintf("import %s\n%s", pkg, code)
	cmd := exec.Command("python", "-c", pyCode)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Value{Type: "error", Err: &ErrorVal{Message: fmt.Sprintf("pip package %s exec error: %s", pkg, string(out))}}
	}
	return Value{Type: "string", Str: strings.TrimSpace(string(out))}
}

func (interp *Interpreter) Fprintf(w *os.File, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
}

func (interp *Interpreter) startServer() error {
	mux := http.NewServeMux()

	// Live-reload endpoint used by the injected reload script. Individual
	// generated routes need a mux pattern to serve, but there is no real
	// handler for this path in the language, so it is registered directly.
	if interp.liveReload && !interp.hasRoute("/__sale_reload") {
		mux.HandleFunc("/__sale_reload", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fmt.Fprintf(w, "%d", interp.revision)
		})
	}

	// Serve src/styles/
	staticDir := filepath.Join("src", "styles")
	if _, err := os.Stat(staticDir); err == nil {
		mux.HandleFunc("/styles/", func(w http.ResponseWriter, r *http.Request) {
			fileName := strings.TrimPrefix(r.URL.Path, "/styles/")
			filePath := filepath.Join(staticDir, fileName)
			data, err := os.ReadFile(filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			ext := filepath.Ext(fileName)
			switch ext {
			case ".css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			case ".js":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			case ".json":
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
			default:
				w.Header().Set("Content-Type", "application/octet-stream")
			}
			w.Write(data)
		})
	}

	// Serve node_modules/ for npm packages
	nodeModulesDir := "node_modules"
	if _, err := os.Stat(nodeModulesDir); err == nil {
		mux.HandleFunc("/node_modules/", func(w http.ResponseWriter, r *http.Request) {
			fileName := strings.TrimPrefix(r.URL.Path, "/node_modules/")
			filePath := filepath.Join(nodeModulesDir, fileName)
			data, err := os.ReadFile(filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			ext := filepath.Ext(fileName)
			switch ext {
			case ".js":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".mjs":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			case ".json":
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
			default:
				w.Header().Set("Content-Type", "application/octet-stream")
			}
			w.Write(data)
		})
	}

	// Serve public/
	publicDir := "public"
	if _, err := os.Stat(publicDir); err == nil {
		mux.HandleFunc("/public/", func(w http.ResponseWriter, r *http.Request) {
			fileName := strings.TrimPrefix(r.URL.Path, "/public/")
			filePath := filepath.Join(publicDir, fileName)
			data, err := os.ReadFile(filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			ext := filepath.Ext(fileName)
			switch ext {
			case ".js":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			default:
				w.Header().Set("Content-Type", "application/octet-stream")
			}
			w.Write(data)
		})
	}

	for key, fn := range interp.routes {
		parts := strings.SplitN(key, " ", 2)
		method := parts[0]
		path := parts[1]
		capturedFn := fn
		capturedMethod := method
		interpRef := interp
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if capturedMethod != "ANY" && r.Method != capturedMethod {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			reqEnv := NewEnv(capturedFn.Env)
			bodyVal := Value{Type: "string", Str: ""}
			if r.Body != nil {
				if raw, err := io.ReadAll(r.Body); err == nil {
					text := strings.TrimSpace(string(raw))
					if text != "" {
						jsonVal, jerr := jsonDecodeString(text)
						if jerr != nil {
							bodyVal = Value{Type: "string", Str: text}
						} else {
							bodyVal = jsonVal
						}
					}
				}
			}
			reqEnv.Define("request", Value{Type: "map", Map: makeMap(map[string]Value{
				"method": Value{Type: "string", Str: r.Method},
				"path":   Value{Type: "string", Str: r.URL.Path},
				"body":   bodyVal,
			})})
			result, err := interpRef.execBlock(capturedFn.Body, reqEnv)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if result.Type == "return" && result.Return != nil {
				result = *result.Return
			}
			body := toString(result)
			// Auto-inject npm script tags into HTML responses
			if len(interpRef.NpmPackages) > 0 && strings.Contains(body, "<html") {
				body = interpRef.injectNpmScripts(body)
			}
			// In dev mode, make the response self-reloading in the browser.
			if interpRef.liveReload {
				body = interpRef.wrapForLiveReload(body)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, body)
		})
	}
	addr := fmt.Sprintf(":%d", interp.port)
	fmt.Fprintf(interp.stdout, "Server running on http://localhost%s\n", addr)
	fmt.Fprintf(interp.stdout, "Press Ctrl+C to stop.\n")
	srv := &http.Server{Addr: addr, Handler: mux}
	interp.httpServer = srv
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (interp *Interpreter) injectNpmScripts(html string) string {
	var tags []string
	for _, pkg := range interp.NpmPackages {
		// Try common package entry points
		candidates := []string{
			filepath.Join("node_modules", pkg, "dist", pkg+".min.js"),
			filepath.Join("node_modules", pkg, "dist", pkg+".js"),
			filepath.Join("node_modules", pkg, "dist", "index.min.js"),
			filepath.Join("node_modules", pkg, "dist", "index.js"),
			filepath.Join("node_modules", pkg, "build", pkg+".min.js"),
			filepath.Join("node_modules", pkg, "build", pkg+".js"),
			filepath.Join("node_modules", pkg, "lib", pkg+".js"),
			filepath.Join("node_modules", pkg, "index.js"),
			filepath.Join("node_modules", pkg, "package.js"),
		}
		injected := false
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				// Use forward slashes for URLs
				urlPath := strings.ReplaceAll(candidate, "\\", "/")
				tags = append(tags, fmt.Sprintf(`        <script src="/%s"></script>`, urlPath))
				injected = true
				break
			}
		}
		if !injected {
			// Fallback: try the package name as a UMD/global script
			tags = append(tags, fmt.Sprintf(`        <script src="/node_modules/%s/dist/%s.min.js"></script>`, pkg, pkg))
		}
	}
	if len(tags) == 0 {
		return html
	}
	scriptBlock := strings.Join(tags, "\n")
	// Inject before </head>
	if idx := strings.Index(html, "</head>"); idx != -1 {
		return html[:idx] + "\n" + scriptBlock + "\n    " + html[idx:]
	}
	// Inject before <body>
	if idx := strings.Index(html, "<body>"); idx != -1 {
		return html[:idx] + scriptBlock + "\n" + html[idx:]
	}
	// Fallback: prepend
	return scriptBlock + "\n" + html
}

// wrapForLiveReload makes a response auto-reload in the browser while running a
// backend dev server. HTML responses get the reload script injected before
// </body>; plain text and JSON responses (which cannot run scripts) are wrapped
// in a minimal HTML page showing the raw content.
func (interp *Interpreter) wrapForLiveReload(body string) string {
	lower := strings.ToLower(body)
	if strings.Contains(lower, "</body>") {
		return strings.Replace(body, "</body>", liveReloadScript+"\n</body>", 1)
	}
	if strings.Contains(lower, "<html") {
		return body + "\n" + liveReloadScript + "\n"
	}
	escaped := html.EscapeString(body)
	return "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>Say Less Response</title>\n</head>\n<body>\n<pre>" + escaped + "</pre>\n" + liveReloadScript + "\n</body>\n</html>"
}

// liveReloadScript polls the dev server for a new revision and reloads the page
// when one is detected. It mirrors the script injected by the web dev server.
const liveReloadScript = `<script>
(function() {
  var rev = 0;
  function check() {
    fetch('/__sale_reload', {cache: 'no-store'})
      .then(function(r) { return r.text(); })
      .then(function(t) {
        var n = parseInt(t, 10) || 0;
        if (rev === 0) { rev = n; return; }
        if (n !== rev) { window.location.reload(); }
      })
      .catch(function() {});
  }
  check();
  setInterval(check, 1000);
})();
</script>`

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
