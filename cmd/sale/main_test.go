package main

import (
	"os"
	"path/filepath"
	"testing"

	"sayless/internal/lexer"
	"sayless/internal/parser"
)

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
