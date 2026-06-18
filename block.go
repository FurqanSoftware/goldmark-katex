package katex

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// BlockParser parses display math ($$...$$) as a block-level element.
//
// Parsing at the block level (rather than inline) keeps the equation body
// completely opaque to Markdown, so display math of any size renders correctly.
// At the inline level a multi-line equation would first be chopped up by block
// parsing: a line such as "}_{A \in \mathbb{R}^{r \times d}}" followed by "="
// becomes a Setext heading, "\\" row separators collide with hard-wrap <br/>,
// and so on. A "$$" must begin a line; inline math ($...$) is handled by Parser.
type BlockParser struct {
}

func (b *BlockParser) Trigger() []byte {
	return []byte{'$'}
}

func (b *BlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := util.TrimLeftSpaceLength(line)

	// The line must open with "$$".
	if pos+1 >= len(line) || line[pos] != '$' || line[pos+1] != '$' {
		return nil, parser.NoChildren
	}

	node := &Block{}
	rest := line[pos+2:]

	// A closing "$$" on the same line means a single-line display block. We
	// record the body and mark the node closed; goldmark ignores a Close state
	// returned from Open, so the actual close happens on the next Continue.
	if idx := bytes.Index(rest, []byte("$$")); idx >= 0 {
		start := segment.Start + pos + 2
		node.Lines().Append(text.NewSegment(start, start+idx))
		node.closed = true
	}

	reader.AdvanceToEOL()
	return node, parser.NoChildren
}

func (b *BlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	// A single-line block was fully read in Open; close without consuming the
	// current line so it is parsed normally.
	if node.(*Block).closed {
		return parser.Close
	}

	line, segment := reader.PeekLine()
	pos := util.TrimLeftSpaceLength(line)

	// A line that opens with "$$" closes the block.
	if pos+1 < len(line) && line[pos] == '$' && line[pos+1] == '$' {
		reader.AdvanceToEOL()
		return parser.Close
	}

	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (b *BlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (b *BlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *BlockParser) CanAcceptIndentedLine() bool {
	return false
}
