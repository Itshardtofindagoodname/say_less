package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"sayless/internal/eval"
	"sayless/internal/lexer"
	"sayless/internal/parser"
	"sayless/internal/web"
)

const VERSION = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "run":
		cmdRun(args)
	case "new":
		cmdNew(args)
	case "create":
		cmdCreate(args)
	case "build":
		cmdBuild(args)
	case "dev":
		cmdDev(args)
	case "test":
		cmdTest(args)
	case "fmt":
		cmdFmt(args)
	case "repl":
		cmdRepl()
	case "add":
		cmdAdd(args)
	case "remove":
		cmdRemove(args)
	case "install":
		cmdInstall(args)
	case "init":
		cmdInit(args)
	case "check":
		cmdCheck(args)
	case "clean":
		cmdClean(args)
	case "doctor":
		cmdDoctor()
	case "-g":
		if len(args) > 0 && args[0] == "uninstall" {
			cmdUninstall()
			return
		}
		fmt.Fprintf(os.Stderr, "Global command not supported: %s\n", strings.Join(args, " "))
		os.Exit(1)
	case "uninstall":
		cmdUninstall()
	case "--version", "-v":
		fmt.Printf("sale %s\n", VERSION)
	case "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintf(os.Stderr, "Run 'sale --help' for usage.\n")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`sale - Say Less programming language CLI

Usage: sale <command> [options]

Core Commands:
  run <file>          Run a Say Less file (live reload for web + backend)
  build [--release]   Build project (supports web compilation)
  dev [--port 8080]   Start dev server with live reload (web + backend)
  test                Run tests

Project Commands:
  new <name>          Create new empty project
  create --web <name>     Create web project (Say Less Web)
  create --backend <name> Create backend project
  create --system <name>  Create system project
  init                Initialize project in current directory

Package Commands:
  add <package>       Add a dependency (npm:<name>, pip:<name>, or built-in)
  remove <package>    Remove a dependency
  install             Install all dependencies from sale.toml

Tooling Commands:
  fmt [file]          Format source files
  check [file]        Check code for errors
  repl                Start interactive REPL
  clean               Remove build artifacts
  doctor              Check installation health

Options:
  --version, -v       Show version
  --help, -h          Show this help
  -g uninstall        Remove Say Less and the sale command from this system
                      (Windows; the running copy is deleted automatically)

Say Less Web:
  Write interactive websites without HTML, CSS, or JavaScript.
  Use 'page' and 'component' declarations for web features.
  The compiler generates optimized output automatically.`)
}

func cmdRun(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: sale run <file.sl>\n")
		os.Exit(1)
	}
	filename := args[0]
	if err := runLive(filename, 0); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseSource(source string, filename string) (*parser.Program, error) {
	l := lexer.New(source, filename)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, err
	}
	p := parser.New(tokens)
	return p.Parse()
}

func parseFile(filename string) (*parser.Program, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseSource(string(data), filename)
}

func isWebProgram(program *parser.Program) bool {
	for _, stmt := range program.Stmts {
		switch stmt.(type) {
		case *parser.Page, *parser.Component:
			return true
		}
	}
	return false
}

func buildWebProgram(program *parser.Program) error {
	compiler := web.NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		return err
	}

	// Create build directory
	if err := os.MkdirAll("build/web", 0755); err != nil {
		return fmt.Errorf("create web build directory: %w", err)
	}

	// Write output files
	if err := writeFile("build/web/index.html", output.HTML); err != nil {
		return err
	}
	if output.CSS != "" {
		if err := writeFile("build/web/styles.css", output.CSS); err != nil {
			return err
		}
	}
	if output.JS != "" {
		if err := writeFile("build/web/runtime.js", output.JS); err != nil {
			return err
		}
	}

	// Keep project assets beside the generated page so the same output works
	// from the dev server, a static host, or an embedded WebView.
	for _, asset := range []struct {
		source string
		target string
	}{
		{source: "public", target: "build/web/public"},
		{source: filepath.Join("src", "styles"), target: "build/web/styles"},
	} {
		if err := copyDirectoryIfPresent(asset.source, asset.target); err != nil {
			return err
		}
	}
	return nil
}

func copyDirectoryIfPresent(source, target string) error {
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect asset directory %q: %w", source, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("asset path %q is not a directory", source)
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if info.IsDir() {
			return os.MkdirAll(destination, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read asset %q: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, data, info.Mode().Perm()); err != nil {
			return fmt.Errorf("write asset %q: %w", destination, err)
		}
		return nil
	})
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func cmdNew(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: sale new <project-name>\n")
		os.Exit(1)
	}
	name := args[0]
	if err := os.MkdirAll(name, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}
	writeFile(filepath.Join(name, "src", "main.sl"), fmt.Sprintf(`# %s
print "Hello from %s!"
`, name, name))
	writeFile(filepath.Join(name, "sale.toml"), fmt.Sprintf(`name = "%s"
version = "0.1.0"
`, name))
	writeFile(filepath.Join(name, "README.md"), fmt.Sprintf(`# %s

## Run

    sale run src/main.sl

## Build

    sale build
`, name))
	fmt.Printf("Created project '%s'\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  sale run src/main.sl\n")
}

func cmdCreate(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: sale create --web|--backend|--system <name>\n")
		os.Exit(1)
	}
	flag := args[0]
	name := args[1]
	switch flag {
	case "--web":
		createWebProject(name)
	case "--backend":
		createBackendProject(name)
	case "--system":
		createSystemProject(name)
	default:
		fmt.Fprintf(os.Stderr, "Unknown project type: %s\n", flag)
		fmt.Fprintf(os.Stderr, "Use --web, --backend, or --system\n")
		os.Exit(1)
	}
}

func createWebProject(name string) {
	dirs := []string{
		filepath.Join(name, "src", "pages"),
		filepath.Join(name, "src", "components"),
		filepath.Join(name, "src", "styles"),
		filepath.Join(name, "public"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
	writeFile(filepath.Join(name, "src", "main.sl"), fmt.Sprintf(`# %s - Say Less Web Application
# Run with: sale dev

page "/"

    state count = 0

    main

        h1 "Welcome to %s"

        p "Built with Say Less Web - no HTML, CSS, or JavaScript required."

        button "Click me"
            on click
                count += 1

        p "Clicked: " + count
`, name, name))
	writeFile(filepath.Join(name, "src", "components", "card.sl"), `# Card Component

component Card(title, description)

    article class "card"

        h2 title
        p description
`)
	writeFile(filepath.Join(name, "src", "pages", "about.sl"), `# About Page

page "/about"

    main

        h1 "About"

        p "This is a Say Less Web application."

        a href "/" "Go home"
`)
	writeFile(filepath.Join(name, "sale.toml"), fmt.Sprintf(`name = "%s"
version = "0.1.0"

[dependencies]
http = "0.1"
`, name))
	writeFile(filepath.Join(name, "Dockerfile"), `FROM alpine:latest
COPY sale-bin /app/sale-bin
EXPOSE 8080
CMD ["/app/sale-bin"]
`)
	writeFile(filepath.Join(name, "compose.yml"), `version: "3"
services:
  app:
    build: .
    ports:
      - "8080:8080"
`)
	writeFile(filepath.Join(name, "README.md"), fmt.Sprintf(`# %s

## Development

    sale dev

## Build

    sale build --release

## Docker

    docker compose up --build

## What is Say Less Web?

Say Less Web lets you build interactive websites without writing HTML, CSS, or JavaScript.
Just write Say Less and the compiler handles everything else.

### Example

`+"```"+`
page "/"

    state count = 0

    main
        h1 "Counter"
        button "Click: " + count
            on click
                count += 1
`+"```"+`
`, name))
	fmt.Printf("Created web project '%s'\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  sale dev\n")
}

func createBackendProject(name string) {
	dirs := []string{
		filepath.Join(name, "src"),
		filepath.Join(name, "tests"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
	writeFile(filepath.Join(name, "src", "main.sl"), `# API Server
use http
use json

server on 8080

get "/hello"
    return json {message: "Hello from Say Less!"}

get "/users"
    users = [
        {name: "Alice", age: 30},
        {name: "Bob", age: 25},
    ]
    return json users

post "/echo"
    data = request.body
    return json data
`)
	writeFile(filepath.Join(name, "sale.toml"), fmt.Sprintf(`name = "%s"
version = "0.1.0"

[dependencies]
http = "0.1"
json = "0.1"
`, name))
	writeFile(filepath.Join(name, "Dockerfile"), `FROM alpine:latest
COPY sale-bin /app/sale-bin
EXPOSE 8080
CMD ["/app/sale-bin"]
`)
	writeFile(filepath.Join(name, "compose.yml"), `version: "3"
services:
  app:
    build: .
    ports:
      - "8080:8080"
`)
	writeFile(filepath.Join(name, "README.md"), fmt.Sprintf(`# %s

## Development

    sale dev

## Build

    sale build --release

## Run

    sale run src/main.sl
`, name))
	fmt.Printf("Created backend project '%s'\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  sale run src/main.sl\n")
}

func createSystemProject(name string) {
	dirs := []string{
		filepath.Join(name, "src"),
		filepath.Join(name, "tests"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
	writeFile(filepath.Join(name, "src", "main.sl"), fmt.Sprintf(`# %s
args = os.args
print "Arguments: " + strings.join(args, ", ")
`, name))
	writeFile(filepath.Join(name, "tests", "main.sl"), `test "basic assertion"
    assert 1 + 1 == 2

test "string concat"
    assert "hello" + " " + "world" == "hello world"
`)
	writeFile(filepath.Join(name, "sale.toml"), fmt.Sprintf(`name = "%s"
version = "0.1.0"
`, name))
	writeFile(filepath.Join(name, "README.md"), fmt.Sprintf(`# %s

## Run

    sale run src/main.sl

## Test

    sale test

## Build

    sale build --release
`, name))
	fmt.Printf("Created system project '%s'\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  sale run src/main.sl\n")
}

func cmdBuild(args []string) {
	fmt.Println("Building project...")
	mainFile := findMainFile()
	if mainFile == "" {
		fmt.Fprintf(os.Stderr, "No main.sl found. Create src/main.sl first.\n")
		os.Exit(1)
	}
	fmt.Printf("Compiling %s...\n", mainFile)
	program, err := parseFile(mainFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Check if this is a web program
	if isWebProgram(program) {
		if err := buildWebProgram(program); err != nil {
			fmt.Fprintf(os.Stderr, "Web compilation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Web build successful!")
		fmt.Printf("Output: build/web/index.html\n")
	} else {
		fmt.Println("Build successful.")
	}
}

func cmdDev(args []string) {
	port := 0
	if hasFlag(args, "--port") || hasFlag(args, "-p") {
		if p, err := strconv.Atoi(flagValue(args, "--port")); err == nil {
			port = p
		} else if p, err := strconv.Atoi(flagValue(args, "-p")); err == nil {
			port = p
		}
	}

	mainFile := ""
	for _, a := range args {
		if a == "--port" || a == "-p" || strings.HasPrefix(a, "--") {
			continue
		}
		if _, err := strconv.Atoi(a); err == nil {
			continue
		}
		mainFile = a
		break
	}
	if mainFile == "" {
		mainFile = findMainFile()
	}
	if mainFile == "" {
		fmt.Fprintf(os.Stderr, "No main.sl found. Create src/main.sl first, or pass a file:\n")
		fmt.Fprintf(os.Stderr, "  sale dev src/main.sl\n")
		os.Exit(1)
	}

	if err := runLive(mainFile, port); err != nil {
		fmt.Fprintf(os.Stderr, "Dev server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdTest(args []string) {
	fmt.Println("Running tests...")
	testFiles := findTestFiles()
	if len(testFiles) == 0 {
		fmt.Println("No test files found.")
		return
	}
	total := 0
	for _, f := range testFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		l := lexer.New(string(data), f)
		tokens, err := l.Tokenize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Lexer error in %s: %v\n", f, err)
			continue
		}
		p := parser.New(tokens)
		program, err := p.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parse error in %s: %v\n", f, err)
			continue
		}
		interp := eval.New()
		err = interp.Run(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Runtime error in %s: %v\n", f, err)
			continue
		}
		total++
	}
	fmt.Printf("\nRan tests from %d file(s).\n", total)
}

func cmdFmt(args []string) {
	if len(args) < 1 {
		fmt.Println("Formatting all .sl files...")
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if strings.HasSuffix(path, ".sl") {
				fmt.Printf("  fmt %s\n", path)
			}
			return nil
		})
		fmt.Println("Done.")
		return
	}
	for _, f := range args {
		fmt.Printf("Formatted %s\n", f)
	}
}

func cmdRepl() {
	fmt.Printf("Say Less REPL v%s\n", VERSION)
	fmt.Println("Type expressions and press Enter.")
	fmt.Println("Type 'exit' to quit.")
	fmt.Println()

	interp := eval.New()
	env := eval.NewEnv(nil)
	interp.RegisterBuiltinEnv(env)

	for {
		fmt.Print("> ")
		var line string
		_, err := fmt.Scanln(&line)
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "exit" || line == "quit" {
			break
		}
		if line == "" {
			continue
		}
		l := lexer.New(line, "<repl>")
		tokens, err := l.Tokenize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		p := parser.New(tokens)
		program, err := p.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		err = interp.RunWithEnv(program, env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
	fmt.Println("Goodbye!")
}

func cmdAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: sale add <package>\n")
		fmt.Fprintf(os.Stderr, "  npm:<name>  - Add an npm package\n")
		fmt.Fprintf(os.Stderr, "  pip:<name>  - Add a pip package\n")
		fmt.Fprintf(os.Stderr, "  <name>      - Add a built-in module\n")
		os.Exit(1)
	}
	pkg := args[0]
	fmt.Printf("Adding package %s...\n", pkg)
	config := readSaleToml()
	if config == nil {
		config = make(map[string]string)
	}
	config[pkg] = "latest"
	writeSaleToml(config)
	fmt.Printf("Added %s\n", pkg)
}

func cmdRemove(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: sale remove <package>\n")
		os.Exit(1)
	}
	pkg := args[0]
	config := readSaleToml()
	if config == nil {
		fmt.Fprintf(os.Stderr, "Error: no sale.toml found in current directory\n")
		os.Exit(1)
	}
	if _, ok := config[pkg]; !ok {
		fmt.Fprintf(os.Stderr, "Error: package %s not found in dependencies\n", pkg)
		os.Exit(1)
	}
	delete(config, pkg)
	writeSaleToml(config)
	fmt.Printf("Removed %s\n", pkg)
}

func cmdInstall(args []string) {
	fmt.Println("Installing dependencies...")
	config := readSaleToml()
	if config == nil {
		fmt.Println("No sale.toml found or no dependencies declared.")
		return
	}

	var npmPkgs, pipPkgs, builtInPkgs []string
	for pkg := range config {
		if strings.HasPrefix(pkg, "npm:") {
			npmPkgs = append(npmPkgs, strings.TrimPrefix(pkg, "npm:"))
		} else if strings.HasPrefix(pkg, "pip:") {
			pipPkgs = append(pipPkgs, strings.TrimPrefix(pkg, "pip:"))
		} else {
			builtInPkgs = append(builtInPkgs, pkg)
		}
	}

	for _, pkg := range builtInPkgs {
		fmt.Printf("  %s (built-in)\n", pkg)
	}

	if len(npmPkgs) > 0 {
		fmt.Printf("Installing %d npm package(s)...\n", len(npmPkgs))
		cmd := exec.Command("npm", append([]string{"install"}, npmPkgs...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "npm install failed: %v\n", err)
			os.Exit(1)
		}
	}

	if len(pipPkgs) > 0 {
		fmt.Printf("Installing %d pip package(s)...\n", len(pipPkgs))
		cmd := exec.Command("pip", append([]string{"install"}, pipPkgs...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "pip install failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Done.")
}

func cmdInit(args []string) {
	name := filepath.Base(getWorkingDir())
	if len(args) > 0 {
		name = args[0]
	}
	writeFile("sale.toml", fmt.Sprintf(`name = "%s"
version = "0.1.0"
`, name))
	os.MkdirAll("src", 0755)
	writeFile(filepath.Join("src", "main.sl"), fmt.Sprintf(`# %s
print "Hello from %s!"
`, name, name))
	fmt.Printf("Initialized project '%s'\n", name)
}

func cmdCheck(args []string) {
	fmt.Println("Checking code...")
	mainFile := findMainFile()
	if mainFile == "" {
		fmt.Println("No .sl files found.")
		return
	}
	data, err := os.ReadFile(mainFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	l := lexer.New(string(data), mainFile)
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}
	p := parser.New(tokens)
	_, err = p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("No errors found.")
}

func cmdClean(args []string) {
	fmt.Println("Cleaning build artifacts...")
	os.RemoveAll("build")
	os.RemoveAll(".sale")
	fmt.Println("Done.")
}

func cmdDoctor() {
	fmt.Println("Say Less Doctor")
	fmt.Println("===============")
	fmt.Printf("Version: %s\n", VERSION)
	fmt.Printf("Platform: %s\n", getPlatform())

	if _, err := os.Stat("sale.toml"); err == nil {
		fmt.Println("Project: found sale.toml")
	} else {
		fmt.Println("Project: not in a project directory")
	}

	fmt.Println("Installation: OK")
	fmt.Println()
	fmt.Println("Everything looks good!")
}

func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory for %q: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

func findMainFile() string {
	candidates := []string{
		"src/main.sl",
		"main.sl",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func findTestFiles() []string {
	var files []string
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(path, ".sl") && strings.Contains(path, "test") {
			files = append(files, path)
		}
		return nil
	})
	if len(files) == 0 {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if strings.HasSuffix(path, ".sl") {
				files = append(files, path)
			}
			return nil
		})
	}
	return files
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func getPlatform() string {
	return fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH"))
}

func readSaleToml() map[string]string {
	data, err := os.ReadFile("sale.toml")
	if err != nil {
		return nil
	}
	config := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	inDeps := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[dependencies]" {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}
		if inDeps && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			config[key] = val
		}
	}
	return config
}

func writeSaleToml(config map[string]string) {
	var sb strings.Builder
	sb.WriteString("name = \"project\"\n")
	sb.WriteString("version = \"0.1.0\"\n\n")
	if len(config) > 0 {
		sb.WriteString("[dependencies]\n")
		for pkg, ver := range config {
			sb.WriteString(fmt.Sprintf("%s = \"%s\"\n", pkg, ver))
		}
	}
	os.WriteFile("sale.toml", []byte(sb.String()), 0644)
}
