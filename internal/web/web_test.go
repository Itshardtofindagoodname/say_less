package web

import (
	"strings"
	"testing"

	"sayless/internal/lexer"
	"sayless/internal/parser"
)

func TestLexerWebTokens(t *testing.T) {
	source := `page "/"
component Card(title)
state count = 0
on click
style
await fetch "/api"
try
catch error
throw "error"
`
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	// Check that web tokens are recognized
	foundPage := false
	foundComponent := false
	foundState := false
	foundOn := false
	foundStyle := false
	foundAwait := false
	foundTry := false
	foundCatch := false
	foundThrow := false

	for _, tok := range tokens {
		switch tok.Type.String() {
		case "page":
			foundPage = true
		case "component":
			foundComponent = true
		case "state":
			foundState = true
		case "on":
			foundOn = true
		case "style":
			foundStyle = true
		case "await":
			foundAwait = true
		case "try":
			foundTry = true
		case "catch":
			foundCatch = true
		case "throw":
			foundThrow = true
		}
	}

	if !foundPage {
		t.Error("Expected PAGE token")
	}
	if !foundComponent {
		t.Error("Expected COMPONENT token")
	}
	if !foundState {
		t.Error("Expected STATE token")
	}
	if !foundOn {
		t.Error("Expected ON token")
	}
	if !foundStyle {
		t.Error("Expected STYLE token")
	}
	if !foundAwait {
		t.Error("Expected AWAIT token")
	}
	if !foundTry {
		t.Error("Expected TRY token")
	}
	if !foundCatch {
		t.Error("Expected CATCH token")
	}
	if !foundThrow {
		t.Error("Expected THROW token")
	}
}

func TestParserPageDecl(t *testing.T) {
	source := "page \"/\"\n    h1 \"Hello\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Stmts) == 0 {
		t.Fatal("Expected at least one statement")
	}

	page, ok := program.Stmts[0].(*parser.Page)
	if !ok {
		t.Fatalf("Expected Page node, got %T", program.Stmts[0])
	}

	if page.Route != "/" {
		t.Errorf("Expected route '/', got '%s'", page.Route)
	}

	if page.Body == nil || len(page.Body.Stmts) == 0 {
		t.Error("Expected page body with statements")
	}
}

func TestParserComponentDecl(t *testing.T) {
	source := "component Card(title)\n    h2 title\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Stmts) == 0 {
		t.Fatal("Expected at least one statement")
	}

	comp, ok := program.Stmts[0].(*parser.Component)
	if !ok {
		t.Fatalf("Expected Component node, got %T", program.Stmts[0])
	}

	if comp.Name != "Card" {
		t.Errorf("Expected component name 'Card', got '%s'", comp.Name)
	}

	if len(comp.Params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(comp.Params))
	}
}

func TestParserStateDecl(t *testing.T) {
	source := `state count = 0`
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Stmts) == 0 {
		t.Fatal("Expected at least one statement")
	}

	state, ok := program.Stmts[0].(*parser.State)
	if !ok {
		t.Fatalf("Expected State node, got %T", program.Stmts[0])
	}

	if state.Name != "count" {
		t.Errorf("Expected state name 'count', got '%s'", state.Name)
	}
}

func TestParserHtmlElement(t *testing.T) {
	source := `h1 "Hello World"`
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Stmts) == 0 {
		t.Fatal("Expected at least one statement")
	}

	elem, ok := program.Stmts[0].(*parser.HtmlElement)
	if !ok {
		t.Fatalf("Expected HtmlElement node, got %T", program.Stmts[0])
	}

	if elem.Tag != "h1" {
		t.Errorf("Expected tag 'h1', got '%s'", elem.Tag)
	}
}

func TestCompilerBasicPage(t *testing.T) {
	source := "page \"/\"\n    h1 \"Hello World\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	compiler := NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	t.Logf("HTML output:\n%s", output.HTML)

	if output.HTML == "" {
		t.Error("Expected non-empty HTML output")
	}

	if !strings.Contains(output.HTML, "<!DOCTYPE html>") {
		t.Error("Expected DOCTYPE in HTML output")
	}

	if !strings.Contains(output.HTML, "<h1>") {
		t.Error("Expected <h1> element in HTML output")
	}

	if !strings.Contains(output.HTML, "Hello World") {
		t.Error("Expected 'Hello World' in HTML output")
	}
}

func TestCompilerWithState(t *testing.T) {
	source := "page \"/\"\n    state count = 0\n    h1 \"Counter\"\n    button \"Click me\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	compiler := NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	if output.HTML == "" {
		t.Error("Expected non-empty HTML output")
	}

	if output.JS == "" {
		t.Error("Expected non-empty JS output (runtime)")
	}
}

func TestCompilerWithComponent(t *testing.T) {
	source := "component Card(title)\n    h2 title\n\npage \"/\"\n    Card(title: \"Hello\")\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	compiler := NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	t.Logf("HTML output:\n%s", output.HTML)

	if output.HTML == "" {
		t.Error("Expected non-empty HTML output")
	}

	if !strings.Contains(output.HTML, "data-component") {
		t.Error("Expected component in HTML output")
	}
}

func TestGeneratorProducesValidHTML(t *testing.T) {
	source := "page \"/\"\n    main\n        h1 \"My Website\"\n        p \"Welcome to Say Less Web.\"\n        button \"Click me\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	compiler := NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	// Check HTML structure
	if !strings.Contains(output.HTML, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE")
	}
	if !strings.Contains(output.HTML, "<html") {
		t.Error("Missing <html> tag")
	}
	if !strings.Contains(output.HTML, "<head>") {
		t.Error("Missing <head> tag")
	}
	if !strings.Contains(output.HTML, "<body>") {
		t.Error("Missing <body> tag")
	}
	if !strings.Contains(output.HTML, "<h1>") {
		t.Error("Missing <h1> element")
	}
	if !strings.Contains(output.HTML, "<p>") {
		t.Error("Missing <p> element")
	}
	if !strings.Contains(output.HTML, "<button>") {
		t.Error("Missing <button> element")
	}
}

func TestParserNestedHtmlWithEventHandler(t *testing.T) {
	source := "page \"/\"\n    button \"Click me\"\n        on click\n            count += 1\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Stmts) == 0 {
		t.Fatal("Expected at least one statement")
	}

	page, ok := program.Stmts[0].(*parser.Page)
	if !ok {
		t.Fatalf("Expected Page node, got %T", program.Stmts[0])
	}

	// Find the button element in page body
	var button *parser.HtmlElement
	for _, stmt := range page.Body.Stmts {
		if elem, ok := stmt.(*parser.HtmlElement); ok && elem.Tag == "button" {
			button = elem
			break
		}
	}
	if button == nil {
		t.Fatal("Expected button element in page body")
	}

	// button should have an EventHandler child
	if len(button.Children) == 0 {
		t.Fatal("Expected button to have children (event handler)")
	}

	var evt *parser.EventHandler
	for _, child := range button.Children {
		if e, ok := child.(*parser.EventHandler); ok {
			evt = e
			break
		}
	}
	if evt == nil {
		t.Fatal("Expected EventHandler child in button")
	}

	if evt.Event != "click" {
		t.Errorf("Expected event 'click', got '%s'", evt.Event)
	}

	if evt.Body == nil || len(evt.Body.Stmts) == 0 {
		t.Error("Expected event handler body with statements")
	}
}

func TestParserDeeplyNestedElements(t *testing.T) {
	source := "page \"/\"\n    main\n        section\n            h1 \"Hello\"\n            p \"World\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	page, ok := program.Stmts[0].(*parser.Page)
	if !ok {
		t.Fatalf("Expected Page node, got %T", program.Stmts[0])
	}

	// Find main element
	var main *parser.HtmlElement
	for _, stmt := range page.Body.Stmts {
		if elem, ok := stmt.(*parser.HtmlElement); ok && elem.Tag == "main" {
			main = elem
			break
		}
	}
	if main == nil {
		t.Fatal("Expected main element")
	}

	// main should have section child
	if len(main.Children) == 0 {
		t.Fatal("Expected main to have children")
	}

	section, ok := main.Children[0].(*parser.HtmlElement)
	if !ok {
		t.Fatalf("Expected HtmlElement child (section), got %T", main.Children[0])
	}

	if section.Tag != "section" {
		t.Errorf("Expected section tag, got '%s'", section.Tag)
	}

	// section should have h1 and p children
	if len(section.Children) < 2 {
		t.Fatalf("Expected section to have 2 children, got %d", len(section.Children))
	}

	h1, ok := section.Children[0].(*parser.HtmlElement)
	if !ok || h1.Tag != "h1" {
		t.Errorf("Expected first child to be h1, got %T", section.Children[0])
	}

	pTag, ok := section.Children[1].(*parser.HtmlElement)
	if !ok || pTag.Tag != "p" {
		t.Errorf("Expected second child to be p, got %T", section.Children[1])
	}
}

func TestCompilerNestedWithEvents(t *testing.T) {
	source := "page \"/\"\n    state count = 0\n    button \"Click me\"\n        on click\n            count += 1\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	compiler := NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	if output.HTML == "" {
		t.Error("Expected non-empty HTML output")
	}

	if !strings.Contains(output.HTML, "<button") {
		t.Error("Expected <button> element")
	}

	if !strings.Contains(output.HTML, "Click me") {
		t.Error("Expected button text 'Click me'")
	}
}

func TestMultiplePages(t *testing.T) {
	source := "page \"/\"\n    h1 \"Home\"\n\npage \"/about\"\n    h1 \"About\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Stmts) != 2 {
		t.Fatalf("Expected 2 page statements, got %d", len(program.Stmts))
	}

	page1, ok := program.Stmts[0].(*parser.Page)
	if !ok {
		t.Fatalf("Expected first statement to be Page, got %T", program.Stmts[0])
	}
	if page1.Route != "/" {
		t.Errorf("Expected first page route '/', got '%s'", page1.Route)
	}

	page2, ok := program.Stmts[1].(*parser.Page)
	if !ok {
		t.Fatalf("Expected second statement to be Page, got %T", program.Stmts[1])
	}
	if page2.Route != "/about" {
		t.Errorf("Expected second page route '/about', got '%s'", page2.Route)
	}
}

func TestCompilerConditionalRuntime(t *testing.T) {
	source := "page \"/\"\n    state showDetails = false\n    section\n        h2 \"Conditional Content\"\n        button \"Toggle Details\"\n            on click\n                showDetails = not showDetails\n        if showDetails\n            p \"shown when on\"\n        else\n            p \"hidden until toggled\"\n"
	l := lexer.New(source, "test.sl")
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	compiler := NewCompiler()
	output, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	if !strings.Contains(output.HTML, "if showDetails") {
		t.Error("Expected conditional marker in HTML output")
	}

	if !strings.Contains(output.JS, "renderConditionals") {
		t.Error("Expected renderConditionals in runtime JS")
	}

	if !strings.Contains(output.JS, `"showDetails":false`) {
		t.Error("Expected showDetails state init in runtime JS")
	}
}
