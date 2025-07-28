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

func (r *HTMLRenderer) renderInline(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		node := n.(*Inline)

		html, err := r.cacheInline.Get(string(node.Equation))

		if err == nil {
			w.Write(html.([]byte))
			return ast.WalkContinue, nil
		}

		if err == gcache.KeyNotFoundError {
			b := bytes.Buffer{}
			err = Render(&b, node.Equation, false, r.throwOnError)
			if err != nil {
				return ast.WalkStop, err
			}
			html := b.Bytes()
			w.Write(html)
			r.cacheInline.Set(string(node.Equation), html)
			return ast.WalkContinue, nil
		}

		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

func (r *HTMLRenderer) renderBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		node := n.(*Block)

		html, err := r.cacheDisplay.Get(string(node.Equation))

		if err == nil {
			w.Write(html.([]byte))
			return ast.WalkContinue, nil
		}

		if err == gcache.KeyNotFoundError {
			b := bytes.Buffer{}
			err = Render(&b, node.Equation, true, r.throwOnError)
			if err != nil {
				return ast.WalkStop, err
			}
			html := b.Bytes()
			w.WriteString("<div>")
			w.Write(html)
			w.WriteString("</div>")
			r.cacheDisplay.Set(string(node.Equation), html)
			return ast.WalkContinue, nil
		}

		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}
