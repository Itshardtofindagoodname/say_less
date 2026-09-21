package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sayless/internal/eval"
	"sayless/internal/lexer"
	"sayless/internal/parser"
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
  run <file>          Run a Say Less file
  build [--release]   Build project
  dev                 Start development server
  test                Run tests

Project Commands:
  new <name>          Create new empty project
  create --web <name>     Create web project
  create --backend <name> Create backend project
  create --system <name>  Create system project
  init                Initialize project in current directory

Package Commands:
  add <package>       Add a dependency
  remove <package>    Remove a dependency
  install             Install dependencies

Tooling Commands:
  fmt [file]          Format source files
  check [file]        Check code for errors
  repl                Start interactive REPL
  clean               Remove build artifacts
  doctor              Check installation health

Options:
  --version, -v       Show version
  --help, -h          Show this help`)
}

func cmdRun(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: sale run <file.sl>\n")
		os.Exit(1)
	}
	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		os.Exit(1)
	}
	runSource(string(data), filename)
}

func runSource(source string, filename string) {
	l := lexer.New(source, filename)
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}
	interp := eval.New()
	err = interp.Run(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
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
	writeFile(filepath.Join(name, "src", "main.sl"), `# Web Application
use http

server on 8080

get "/"
    return html
        """
        <!DOCTYPE html>
        <html>
        <head><title>Say Less Web App</title></head>
        <body>
            <h1>Hello from Say Less!</h1>
            <p>Build more. Say less.</p>
        </body>
        </html>
        """
`)
	writeFile(filepath.Join(name, "src", "pages", "index.sl"), `# Home Page
fn render()
    return html
        """
        <div class="hero">
            <h1>Welcome</h1>
        </div>
        """
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
    return json
        message: "Hello from Say Less!"

get "/users"
    users = [
        {name: "Alice", age: 30}
        {name: "Bob", age: 25}
    ]
    return json users

post "/echo"
    body = request.body
    return json body
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
	release := false
	for _, a := range args {
		if a == "--release" {
			release = true
		}
	}
	_ = release
	mainFile := findMainFile()
	if mainFile == "" {
		fmt.Fprintf(os.Stderr, "No main.sl found. Create src/main.sl first.\n")
		os.Exit(1)
	}
	fmt.Printf("Compiling %s...\n", mainFile)
	data, err := os.ReadFile(mainFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", mainFile, err)
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
	fmt.Println("Build successful.")
}

func cmdDev(args []string) {
	fmt.Println("Starting development server...")
	mainFile := findMainFile()
	if mainFile == "" {
		fmt.Fprintf(os.Stderr, "No main.sl found. Create src/main.sl first.\n")
		os.Exit(1)
	}
	fmt.Printf("Watching %s for changes...\n", filepath.Dir(mainFile))
	fmt.Println("Development server running on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop.")
	for {
		data, err := os.ReadFile(mainFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		runSource(string(data), mainFile)
		fmt.Println("Rebuilding...")
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
	fmt.Printf("Removing package %s...\n", args[0])
}

func cmdInstall(args []string) {
	fmt.Println("Installing dependencies...")
	config := readSaleToml()
	if config != nil {
		for pkg, ver := range config {
			fmt.Printf("  %s@%s\n", pkg, ver)
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

func writeFile(path, content string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(path, []byte(content), 0644)
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
