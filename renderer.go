package katex

import (
	"bytes"

	"github.com/bluele/gcache"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type HTMLRenderer struct {
	html.Config

	cacheInline  gcache.Cache
	cacheDisplay gcache.Cache
	throwOnError bool
}

func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInline, r.renderInline)
	reg.Register(KindBlock, r.renderBlock)
}

// renderInline renders math that sits within a line of text.
func (r *HTMLRenderer) renderInline(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	node := n.(*Inline)
	cache := r.cacheInline
	if node.Display {
		cache = r.cacheDisplay
	}
	return r.render(w, node.Equation, node.Display, cache)
}

// renderBlock renders display math occupying lines of its own. The equation is
// held in the node's Lines() rather than on the node itself, because the block
// parser collects it a line at a time.
func (r *HTMLRenderer) renderBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	return r.render(w, n.(*Block).Equation(source), true, r.cacheDisplay)
}

// render writes equation as HTML, going through cache. KaTeX's display output
// is already a block-level element, so nothing wraps it.
func (r *HTMLRenderer) render(w util.BufWriter, equation []byte, display bool, cache gcache.Cache) (ast.WalkStatus, error) {
	key := string(equation)

	html, err := cache.Get(key)
	if err == nil {
		w.Write(html.([]byte))
		return ast.WalkContinue, nil
	}
	if err != gcache.KeyNotFoundError {
		return ast.WalkStop, err
	}

	b := bytes.Buffer{}
	err = Render(&b, equation, display, r.throwOnError)
	if err != nil {
		return ast.WalkStop, err
	}
	rendered := b.Bytes()
	w.Write(rendered)
	cache.Set(key, rendered)

	return ast.WalkContinue, nil
}
