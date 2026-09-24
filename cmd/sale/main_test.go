package main

import (
	"os"
	"path/filepath"
	"testing"

	"sayless/internal/lexer"
	"sayless/internal/parser"
)

func TestProgramTypeDetection(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		web    bool
		server bool
	}{
		{"plain script", "print \"hello\"\n", false, false},
		{"web page", "page \"/\"\n    h1 \"Hi\"\n", true, false},
		{"backend server", "server on 8080\n", false, true},
		{"backend route", "get \"/hello\"\n    return \"hi\"\n", false, true},
	}
	for _, tc := range tests {
		program, err := parseSource(tc.src, "<test>")
		if err != nil {
			t.Fatalf("%s: parse: %v", tc.name, err)
		}
		if got := isWebProgram(program); got != tc.web {
			t.Errorf("%s: isWebProgram = %v, want %v", tc.name, got, tc.web)
		}
		if got := isServerProgram(program); got != tc.server {
			t.Errorf("%s: isServerProgram = %v, want %v", tc.name, got, tc.server)
		}
	}
}

func TestProgramPort(t *testing.T) {
	serverSrc := "server on 8080\nget \"/a\"\n    return \"x\"\n"
	program, err := parseSource(serverSrc, "<test>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := programPort(program); got != 8080 {
		t.Errorf("programPort = %d, want 8080", got)
	}

	script, err := parseSource("print \"hi\"\n", "<test>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := programPort(script); got != 0 {
		t.Errorf("programPort for plain script = %d, want 0", got)
	}
}

func TestBuildWebProgramCopiesProjectAssets(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, "src", "styles"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "public", "images"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "styles", "base.css"), []byte("body { color: red; }"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "public", "images", "logo.txt"), []byte("logo"), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

	l := lexer.New("page \"/\"\n    h1 \"Hello\"\n", "main.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if err := buildWebProgram(program); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join("build", "web", "styles", "base.css"),
		filepath.Join("build", "web", "public", "images", "logo.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected copied asset %q: %v", path, err)
		}
	}
}
