package katex

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// Inline is math that sits within a line of text.
//
// It covers inline math ("$...$") and display math written mid-line
// ("text $$x$$ more"), which stays an inline node so it nests validly inside
// the surrounding paragraph. Display math on a line of its own is a Block.
type Inline struct {
	ast.BaseInline

	Equation []byte

	// Display selects KaTeX display mode, set for "$$...$$".
	Display bool
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

// Block is display math occupying whole lines of its own ("$$...$$").
//
// BlockParser builds it during block parsing, before Markdown ever looks at the
// equation. Parsing it inline instead let block constructs mangle multi-line
// equations first: a line followed by "=" became a Setext heading, and the
// equation was silently truncated at the heading. The equation body is held in
// Lines(), and IsRaw keeps Markdown away from it.
//
// Display math written in the middle of a line stays inline math; see Parser.
type Block struct {
	ast.BaseBlock

	// closed marks a "$$...$$" whose body was read in full by BlockParser.Open.
	// Open cannot end a block itself, so Continue closes it on the next line.
	closed bool
}

// IsRaw keeps goldmark from running inline Markdown parsing over the equation,
// which would otherwise read "_" as emphasis, "*" as a list bullet, and so on.
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

// Equation returns the raw LaTeX between the delimiters.
func (n *Block) Equation(source []byte) []byte {
	lines := n.Lines()
	var b []byte
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b = append(b, seg.Value(source)...)
	}
	return b
}
