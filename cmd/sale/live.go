package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sayless/internal/eval"
	"sayless/internal/parser"
	"sayless/internal/web"
)

const watchInterval = 400 * time.Millisecond

// runLive starts a live development server for the given source file. For web
// programs it uses the Say Less Web dev server with auto-rebuild and browser
// reload. For backend API programs it runs an interpreter-backed HTTP server
// that restarts with updated routes whenever .sl files change. Plain scripts
// are run once to completion. A port of 0 means "use the value declared in the
// program" (for backend) or the default 8080 (for web).
func runLive(filename string, port int) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}
	program, err := parseSource(string(data), filename)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if isWebProgram(program) {
		return runWebLive(filename, port)
	}
	if isServerProgram(program) {
		return runBackendLive(filename, port)
	}
	interp := eval.New()
	if port > 0 {
		interp.SetPort(port)
	}
	return interp.Run(program)
}

// runWebLive builds the web program and serves it with live reload.
func runWebLive(filename string, port int) error {
	if port == 0 {
		port = 8080
	}
	port = resolvePort(port)
	rootDir := filepath.Dir(filename)
	buildDir := filepath.Join("build", "web")
	build := func() error {
		program, err := parseFile(filename)
		if err != nil {
			return err
		}
		return buildWebProgram(program)
	}
	if err := build(); err != nil {
		return err
	}
	fmt.Printf("Watching %s for changes...\n", rootDir)
	server := web.NewDevServer(buildDir, rootDir, port, build)
	return server.Run()
}

// backendRunner owns one interpreter-backed server and coordinates graceful
// restarts so the port is always released before the next server binds it.
type backendRunner struct {
	filename string
	port     int
	revision int
	interp   *eval.Interpreter
	done     chan struct{}
	running  bool
}

func (b *backendRunner) start(program *parser.Program) error {
	interp := eval.New()
	if b.port > 0 {
		interp.SetPort(b.port)
	}
	interp.EnableLiveReload(true)
	interp.SetRevision(b.revision)
	done := make(chan struct{})
	b.interp = interp
	b.done = done
	b.running = true
	go func() {
		defer close(done)
		if err := interp.Run(program); err != nil {
			fmt.Fprintf(os.Stderr, "[sale] Runtime error: %v\n", err)
		}
	}()
	return nil
}

func (b *backendRunner) stop() {
	if !b.running {
		return
	}
	b.running = false
	if b.interp != nil {
		b.interp.Shutdown()
	}
	if b.done != nil {
		<-b.done
	}
	b.interp = nil
	b.done = nil
}

func (b *backendRunner) reload() {
	program, err := parseFile(b.filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[sale] Update failed: %v\n", err)
		return
	}
	b.revision++
	b.stop()
	if err := b.start(program); err != nil {
		fmt.Fprintf(os.Stderr, "[sale] Update failed: %v\n", err)
	}
}

// runBackendLive runs an API server and watches .sl files under the source
// directory. On change it prints "Updating changes...", restarts the server
// with the new program, and the new routes become live immediately. If the
// requested port is in use it falls back to the next available port and prints
// how to free the original one.
func runBackendLive(filename string, port int) error {
	rootDir := filepath.Dir(filename)

	program, err := parseFile(filename)
	if err != nil {
		return err
	}
	requested := port
	if requested == 0 {
		requested = programPort(program)
	}
	if requested == 0 {
		requested = 8080
	}
	activePort := resolvePort(requested)

	runner := &backendRunner{filename: filename, port: activePort}

	if err := runner.start(program); err != nil {
		return err
	}
	fmt.Printf("Watching %s for changes...\n", rootDir)

	last := slSnapshot(rootDir)
	for {
		time.Sleep(watchInterval)
		cur := slSnapshot(rootDir)
		if cur == last {
			continue
		}
		last = cur
		fmt.Println("[sale] Updating changes...")
		runner.reload()
	}
}

// programPort returns the port declared by the program's "server on <port>"
// statement, or 0 if none is declared.
func programPort(program *parser.Program) int {
	for _, stmt := range program.Stmts {
		if server, ok := stmt.(*parser.ServerDecl); ok {
			if lit, ok := server.Port.(*parser.IntLit); ok {
				return int(lit.Value)
			}
		}
	}
	return 0
}

// isServerProgram reports whether the program declares an HTTP server or any
// route handlers, i.e. whether it is a backend API program.
func isServerProgram(program *parser.Program) bool {
	for _, stmt := range program.Stmts {
		switch stmt.(type) {
		case *parser.ServerDecl, *parser.RouteHandler:
			return true
		}
	}
	return false
}

// slSnapshot returns a signature of all .sl files under rootDir so the watcher
// can detect changes by comparing consecutive snapshots.
func slSnapshot(rootDir string) string {
	var sb strings.Builder
	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			switch info.Name() {
			case "build", ".git", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".sl") {
			fmt.Fprintf(&sb, "%s|%d|%d\n", path, info.Size(), info.ModTime().UnixNano())
		}
		return nil
	})
	return sb.String()
}
