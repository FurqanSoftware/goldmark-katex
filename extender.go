package katex

import (
	"github.com/bluele/gcache"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Extender struct {
	ThrowOnError bool
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		// Block math ($$...$$). Runs before Markdown block parsing so a
		// multi-line equation is never mangled. Inline math ($...$) is handled
		// by the inline Parser below.
		util.Prioritized(&BlockParser{}, 100),
	))
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&Parser{}, 0),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&HTMLRenderer{
			cacheInline:  gcache.New(5000).ARC().Build(),
			cacheBlock:   gcache.New(5000).ARC().Build(),
			throwOnError: e.ThrowOnError,
		}, 0),
	))
}
