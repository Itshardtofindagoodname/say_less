package web

import (
	"fmt"
	"strings"
)

// Generator produces HTML, CSS, and JS from the Web IR
type Generator struct {
	ir           *WebIR
	componentIDs map[string]string
	stateVars    map[string]bool
	scriptParts  []string
	styleParts   []string
}

// NewGenerator creates a new output generator
func NewGenerator(ir *WebIR) *Generator {
	return &Generator{
		ir:           ir,
		componentIDs: make(map[string]string),
		stateVars:    make(map[string]bool),
		scriptParts:  make([]string, 0),
		styleParts:   make([]string, 0),
	}
}

// Generate produces the final HTML output
func (g *Generator) Generate() *Output {
	// Collect state variables
	for _, page := range g.ir.Pages {
		g.collectState(page.Elements)
	}
	for _, comp := range g.ir.Components {
		g.collectState(comp.Elements)
	}

	// Generate for first page (or index page)
	var html, css, js string
	if len(g.ir.Pages) > 0 {
		page := g.ir.Pages[0]
		body := g.generateElements(page.Elements)
		html = g.wrapHTML(page.Route, body, css, js)
	}

	// Generate CSS from components
	for _, comp := range g.ir.Components {
		g.generateComponentCSS(comp)
	}

	// Generate JS runtime
	js = g.generateRuntime()

	// Re-generate HTML with final CSS/JS
	if len(g.ir.Pages) > 0 {
		page := g.ir.Pages[0]
		body := g.generateElements(page.Elements)
		css = strings.Join(g.styleParts, "\n")
		html = g.wrapHTML(page.Route, body, css, js)
	}

	return &Output{
		HTML: html,
		CSS:  css,
		JS:   js,
	}
}

func (g *Generator) collectState(nodes []IRNode) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *IRState:
			g.stateVars[n.Name] = true
		case *IRElement:
			g.collectState(n.Children)
		case *IRConditional:
			g.collectState(n.Then)
			g.collectState(n.Else)
		case *IRLoop:
			g.collectState(n.Body)
		}
	}
}

func (g *Generator) generateElements(nodes []IRNode) string {
	var parts []string
	for _, node := range nodes {
		parts = append(parts, g.generateNode(node))
	}
	return strings.Join(parts, "\n")
}

func (g *Generator) generateNode(node IRNode) string {
	if node == nil {
		return ""
	}
	switch n := node.(type) {
	case *IRText:
		return g.generateText(n)
	case *IRInterp:
		return g.generateInterp(n)
	case *IRBinding:
		return fmt.Sprintf("{{%s}}", n.Expr)
	case *IRElement:
		return g.generateElement(n)
	case *IRComponent:
		return g.generateComponent(n)
	case *IRConditional:
		return g.generateConditional(n)
	case *IRLoop:
		return g.generateLoop(n)
	case *IRAwait:
		return g.generateAwait(n)
	case *IRFragment:
		return g.generateElements(n.Children)
	case *IRState:
		return g.generateState(n)
	case *IREvent:
		return g.generateEvent(n)
	case *StyleBlock:
		return g.generateStyleBlock(n)
	default:
		return ""
	}
}

func (g *Generator) generateText(text *IRText) string {
	return escapeHTML(text.Content)
}

func (g *Generator) generateInterp(interp *IRInterp) string {
	var parts []string
	for _, part := range interp.Parts {
		parts = append(parts, g.generateNode(part))
	}
	return strings.Join(parts, "")
}

func (g *Generator) generateElement(elem *IRElement) string {
	var attrs []string
	for _, attr := range elem.Attributes {
		if attr.IsExpr {
			attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Name, escapeHTML(attr.Value)))
		} else {
			attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Name, escapeHTML(attr.Value)))
		}
	}

	// Add data bindings for state variables
	if g.needsReactivity(elem) {
		attrs = append(attrs, `data-reactive="true"`)
	}

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}

	// Self-closing tags
	selfClosing := map[string]bool{
		"br": true, "hr": true, "img": true, "input": true,
		"meta": true, "link": true, "source": true, "path": true,
	}
	if selfClosing[elem.Tag] {
		return fmt.Sprintf("<%s%s />", elem.Tag, attrStr)
	}

	var children []string
	for _, child := range elem.Children {
		childStr := g.generateNode(child)
		if childStr != "" {
			children = append(children, childStr)
		}
	}

	// Add event handlers as attributes
	for _, child := range elem.Children {
		if evt, ok := child.(*IREvent); ok {
			attrs = append(attrs, fmt.Sprintf(`data-on-%s="%s"`, evt.Name, escapeHTML(evt.Handler)))
		}
	}

	content := strings.Join(children, "\n")
	if len(content) == 0 {
		return fmt.Sprintf("<%s%s></%s>", elem.Tag, attrStr, elem.Tag)
	}
	return fmt.Sprintf("<%s%s>\n%s\n</%s>", elem.Tag, attrStr, content, elem.Tag)
}

func (g *Generator) generateComponent(comp *IRComponent) string {
	attrs := []string{
		fmt.Sprintf(`data-component="%s"`, comp.Name),
		fmt.Sprintf(`data-instance="%s"`, comp.InstanceID),
	}
	attrStr := strings.Join(attrs, " ")
	children := g.generateElements(comp.Children)
	return fmt.Sprintf(`<div %s>%s</div>`, attrStr, children)
}

func (g *Generator) generateConditional(cond *IRConditional) string {
	then := g.generateElements(cond.Then)
	else_ := g.generateElements(cond.Else)
	return fmt.Sprintf(`<!-- if %s -->%s<!-- else -->%s<!-- /if -->`,
		cond.Condition, then, else_)
}

func (g *Generator) generateLoop(loop *IRLoop) string {
	body := g.generateElements(loop.Body)
	return fmt.Sprintf(`<!-- for %s in %s -->%s<!-- /for -->`,
		loop.Variable, loop.Iterable, body)
}

func (g *Generator) generateAwait(await *IRAwait) string {
	loading := g.generateElements(await.Loading)
	error_ := g.generateElements(await.Error)
	success := g.generateElements(await.Success)
	return fmt.Sprintf(`<!-- await %s --><!-- loading -->%s<!-- error -->%s<!-- success -->%s<!-- /await -->`,
		await.Expr, loading, error_, success)
}

func (g *Generator) generateState(state *IRState) string {
	g.stateVars[state.Name] = true
	return fmt.Sprintf(`<!-- state %s = %s -->`, state.Name, state.InitialValue)
}

func (g *Generator) generateEvent(evt *IREvent) string {
	return fmt.Sprintf(`<!-- on %s: %s -->`, evt.Name, evt.Handler)
}

func (g *Generator) generateStyleBlock(style *StyleBlock) string {
	var props []string
	for _, prop := range style.Properties {
		props = append(props, fmt.Sprintf("  %s: %s;", prop.Name, prop.Value))
	}
	return fmt.Sprintf("<!-- style -->\n%s\n<!-- /style -->", strings.Join(props, "\n"))
}

func (g *Generator) generateComponentCSS(comp *ComponentNode) {
	if !comp.HasState && !comp.HasEvents {
		return
	}
	css := fmt.Sprintf(`[data-component="%s"] {
  /* scoped styles */
}`, comp.Name)
	g.styleParts = append(g.styleParts, css)
}

func (g *Generator) needsReactivity(elem *IRElement) bool {
	// Check if element contains state variable references
	for _, child := range elem.Children {
		if text, ok := child.(*IRText); ok {
			for stateVar := range g.stateVars {
				if strings.Contains(text.Content, stateVar) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) generateRuntime() string {
	return `<script>
(function() {
  'use strict';
  
  // Say Less Web Runtime
  const SL = {
    state: {},
    listeners: {},
    islands: new Map(),
    
    setState(name, value) {
      this.state[name] = value;
      this.updateDependents(name);
    },
    
    getState(name) {
      return this.state[name];
    },
    
    updateDependents(name) {
      const selector = '[data-state-' + name + ']';
      const els = document.querySelectorAll(selector);
      els.forEach(el => {
        const template = el.getAttribute('data-state-' + name);
        if (template) {
          el.textContent = this.evaluateTemplate(template);
        }
      });
    },
    
    evaluateTemplate(template) {
      return template.replace(/\$\{([^}]+)\}/g, (match, expr) => {
        try {
          return eval(expr);
        } catch(e) {
          return '';
        }
      });
    },
    
    init() {
      document.querySelectorAll('[data-reactive="true"]').forEach(el => {
        this.makeReactive(el);
      });
      
      document.querySelectorAll('[data-on-click]').forEach(el => {
        const handler = el.getAttribute('data-on-click');
        el.addEventListener('click', () => {
          try {
            new Function(handler)();
          } catch(e) {
            console.error('Event handler error:', e);
          }
        });
      });
      
      document.querySelectorAll('[data-on-input]').forEach(el => {
        const handler = el.getAttribute('data-on-input');
        el.addEventListener('input', () => {
          try {
            new Function(handler)();
          } catch(e) {
            console.error('Event handler error:', e);
          }
        });
      });
      
      document.querySelectorAll('[data-on-submit]').forEach(el => {
        const handler = el.getAttribute('data-on-submit');
        el.addEventListener('submit', (e) => {
          e.preventDefault();
          try {
            new Function(handler)();
          } catch(e) {
            console.error('Event handler error:', e);
          }
        });
      });
    },
    
    makeReactive(el) {
      const text = el.textContent;
      for (const name in this.state) {
        if (text.includes(name)) {
          el.setAttribute('data-state-' + name, text);
        }
      }
    }
  };
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => SL.init());
  } else {
    SL.init();
  }
  
  window.SL = SL;
})();
</script>`
}

func (g *Generator) wrapHTML(route, body, css, js string) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Say Less Web App</title>`)
	if css != "" {
		sb.WriteString(fmt.Sprintf("\n<style>\n%s\n</style>", css))
	}
	sb.WriteString("\n</head>\n<body>\n")
	sb.WriteString(body)
	sb.WriteString("\n")
	if js != "" {
		sb.WriteString(js)
	}
	sb.WriteString("\n</body>\n</html>")
	return sb.String()
}
