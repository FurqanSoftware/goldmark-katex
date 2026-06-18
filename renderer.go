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
	cacheBlock   gcache.Cache
	throwOnError bool
}

func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInline, r.renderInline)
	reg.Register(KindBlock, r.renderBlock)
}

// renderInline renders inline math ("$...$").
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

// renderBlock renders block-level display math ("$$...$$"). The raw LaTeX is
// held in the node's Lines(), wrapped in a <div> after KaTeX rendering.
func (r *HTMLRenderer) renderBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	var eq bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		eq.Write(seg.Value(source))
	}
	equation := eq.Bytes()

	html, err := r.cacheBlock.Get(string(equation))
	if err == nil {
		w.WriteString("<div>")
		w.Write(html.([]byte))
		w.WriteString("</div>")
		return ast.WalkContinue, nil
	}

	if err == gcache.KeyNotFoundError {
		b := bytes.Buffer{}
		err = Render(&b, equation, true, r.throwOnError)
		if err != nil {
			return ast.WalkStop, err
		}
		html := b.Bytes()
		w.WriteString("<div>")
		w.Write(html)
		w.WriteString("</div>")
		r.cacheBlock.Set(string(equation), html)
		return ast.WalkContinue, nil
	}

	return ast.WalkStop, err
}
