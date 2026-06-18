package katex

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// Inline is an inline-level math node produced from "$...$".
type Inline struct {
	ast.BaseInline

	Equation []byte
}

func (n *Inline) Inline() {}

func (n *Inline) IsBlank(source []byte) bool {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		text := c.(*ast.Text).Segment
		if !util.IsBlank(text.Value(source)) {
			return false
		}
	}
	return true
}

func (n *Inline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var KindInline = ast.NewNodeKind("Inline")

func (n *Inline) Kind() ast.NodeKind {
	return KindInline
}

// Block is a block-level display-math node produced from "$$...$$".
//
// It is parsed at the block level (see BlockParser) so multi-line equations are
// never reinterpreted as Markdown. Setext headings ("=" on the next line), hard
// wraps and "\\" row separators would otherwise mangle a multi-line equation
// before KaTeX ever sees it. The raw LaTeX is held in Lines().
type Block struct {
	ast.BaseBlock

	// closed marks a single-line "$$...$$" whose body was fully read in
	// BlockParser.Open, so BlockParser.Continue closes it immediately.
	closed bool
}

// IsRaw reports that the body is raw LaTeX, so goldmark does not run inline
// Markdown parsing over it (which would turn "_" into emphasis, drop "\,", etc.).
func (n *Block) IsRaw() bool {
	return true
}

func (n *Block) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var KindBlock = ast.NewNodeKind("Block")

func (n *Block) Kind() ast.NodeKind {
	return KindBlock
}
