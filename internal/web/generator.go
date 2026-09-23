package web

import (
	"fmt"
	"sort"
	"strings"
)

// Generator produces HTML, CSS, and JS from the Web IR
type Generator struct {
	ir           *WebIR
	componentIDs map[string]string
	stateVars    map[string]bool
	stateInit    map[string]string
	scriptParts  []string
	styleParts   []string
}

// NewGenerator creates a new output generator
func NewGenerator(ir *WebIR) *Generator {
	return &Generator{
		ir:           ir,
		componentIDs: make(map[string]string),
		stateVars:    make(map[string]bool),
		stateInit:    make(map[string]string),
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
			g.stateInit[n.Name] = n.InitialValue
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
		return fmt.Sprintf(`<span data-sl-expr="%s"></span>`, escapeHTML(n.Expr))
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
	// Self-closing tags (no children, no event handlers)
	selfClosing := map[string]bool{
		"br": true, "hr": true, "img": true, "input": true,
		"meta": true, "link": true, "source": true, "path": true,
	}
	if selfClosing[elem.Tag] {
		var attrs []string
		for _, attr := range elem.Attributes {
			attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Name, escapeHTML(attr.Value)))
		}
		attrStr := ""
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return fmt.Sprintf("<%s%s />", elem.Tag, attrStr)
	}

	// Children first (events are rendered as attributes, not content)
	var children []string
	var events []*IREvent
	for _, child := range elem.Children {
		if evt, ok := child.(*IREvent); ok {
			events = append(events, evt)
			continue
		}
		childStr := g.generateNode(child)
		if childStr != "" {
			children = append(children, childStr)
		}
	}

	// Then attributes: static attrs, reactivity marker, event handlers
	var attrs []string
	for _, attr := range elem.Attributes {
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, attr.Name, escapeHTML(attr.Value)))
	}
	if g.needsReactivity(elem) {
		attrs = append(attrs, `data-reactive="true"`)
	}
	for _, evt := range events {
		attrs = append(attrs, fmt.Sprintf(`data-on-%s="%s"`, evt.Name, escapeHTML(evt.Handler)))
	}
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
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
	g.stateInit[state.Name] = state.InitialValue
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
	for _, child := range elem.Children {
		if g.nodeIsReactive(child) {
			return true
		}
	}
	return false
}

func (g *Generator) nodeIsReactive(node IRNode) bool {
	switch n := node.(type) {
	case *IRBinding:
		return true
	case *IRInterp:
		for _, part := range n.Parts {
			if g.nodeIsReactive(part) {
				return true
			}
		}
	case *IRElement:
		for _, child := range n.Children {
			if g.nodeIsReactive(child) {
				return true
			}
		}
	}
	return false
}

func (g *Generator) generateRuntime() string {
	names := make([]string, 0, len(g.stateInit))
	for name := range g.stateInit {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%q:%s", name, g.stateInit[name]))
	}
	stateObj := "{" + strings.Join(parts, ",") + "}"

	return fmt.Sprintf(`<script>
(function() {
  'use strict';

  const SL = {
    state: %s,

    setState(name, value) {
      this.state[name] = value;
      this.render();
    },

    getState(name) {
      return this.state[name];
    },

    evalExpr(expr) {
      const keys = Object.keys(this.state);
      const vals = keys.map(k => this.state[k]);
      try {
        return new Function(keys.join(','), 'return (' + expr + ');').apply(null, vals);
      } catch (e) {
        console.error('Expression error:', expr, e);
        return '';
      }
    },

    runHandler(code) {
      const keys = Object.keys(this.state);
      const vals = keys.map(k => this.state[k]);
      let out = null;
      try {
        out = new Function(keys.join(','), code + '\nreturn [' + keys.join(',') + '];').apply(null, vals);
      } catch (e) {
        console.error('Handler error:', code, e);
        return;
      }
      if (out) {
        let changed = false;
        keys.forEach((k, i) => {
          if (this.state[k] !== out[i]) {
            this.state[k] = out[i];
            changed = true;
          }
        });
        if (changed) this.render();
      }
    },

    render() {
      const self = this;
      document.querySelectorAll('[data-sl-expr]').forEach(el => {
        el.textContent = self.evalExpr(el.getAttribute('data-sl-expr'));
      });
      self.renderConditionals();
    },

    renderConditionals() {
      const self = this;
      const comments = [];
      const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_COMMENT);
      let commentNode;
      while ((commentNode = walker.nextNode())) {
        const text = commentNode.data.trim();
        if (text.indexOf('if ') === 0 || text === 'else' || text === '/if') {
          comments.push(commentNode);
        }
      }
      const stack = [];
      comments.forEach(c => {
        const text = c.data.trim();
        if (text.indexOf('if ') === 0) {
          stack.push({ cond: text.slice(3).trim(), start: c, elseComment: null });
        } else if (text === 'else') {
          if (stack.length) stack[stack.length - 1].elseComment = c;
        } else if (text === '/if') {
          if (!stack.length) return;
          const frame = stack.pop();
          const nodes = [];
          let cur = frame.start.nextSibling;
          while (cur && cur !== c) {
            nodes.push(cur);
            cur = cur.nextSibling;
          }
          const thenNodes = [];
          const elseNodes = [];
          let inElse = false;
          for (let k = 0; k < nodes.length; k++) {
            const nd = nodes[k];
            if (nd === frame.elseComment) { inElse = true; continue; }
            if (nd.nodeType === Node.ELEMENT_NODE) {
              if (inElse) elseNodes.push(nd); else thenNodes.push(nd);
            }
          }
          const showThen = !!self.evalExpr(frame.cond);
          thenNodes.forEach(nd => { nd.style.display = showThen ? '' : 'none'; });
          elseNodes.forEach(nd => { nd.style.display = showThen ? 'none' : ''; });
        }
      });
    },

    bindEvents() {
      const self = this;
      document.querySelectorAll('*').forEach(el => {
        Array.prototype.slice.call(el.attributes).forEach(attr => {
          if (attr.name.indexOf('data-on-') === 0) {
            const evtName = attr.name.slice('data-on-'.length);
            const code = attr.value;
            el.addEventListener(evtName, e => {
              if (evtName === 'submit') e.preventDefault();
              self.runHandler(code);
            });
          }
        });
      });
    },

    init() {
      this.render();
      this.bindEvents();
    }
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => SL.init());
  } else {
    SL.init();
  }

  window.SL = SL;
})();
</script>`, stateObj)
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
