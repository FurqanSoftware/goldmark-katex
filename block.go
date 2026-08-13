package katex

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var delimiter = []byte("$$")

// BlockParser parses display math ("$$...$$") that begins a line.
//
// Running at the block level keeps the equation opaque to Markdown, so an
// equation of any length survives intact. Parsed inline it would first be cut
// up by block parsing, most visibly by a line followed by "=" becoming a Setext
// heading, which truncated the equation at that point.
//
// A "$$" that begins a line opens a block, which runs to the next "$$" or to
// end of input, as a fenced code block does. Text before the opening delimiter
// or after the closing one is never consumed: display math written mid-line is
// left to the inline Parser, which renders it in place.
type BlockParser struct {
}

func (b *BlockParser) Trigger() []byte {
	return []byte{'$'}
}

func (b *BlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := util.TrimLeftSpaceLength(line)

	// The line has to open with "$$".
	if pos+1 >= len(line) || line[pos] != '$' || line[pos+1] != '$' {
		return nil, parser.NoChildren
	}

	start := segment.Start + pos + 2
	rest := line[pos+2:]
	node := &Block{}

	if idx := bytes.Index(rest, delimiter); idx >= 0 {
		// Both delimiters sit on this line. Only take it as a block when the
		// closing "$$" ends the line; "$$x$$ more" keeps its trailing text by
		// falling through to a paragraph, where the inline Parser renders the
		// display math in place.
		if !util.IsBlank(rest[idx+2:]) {
			return nil, parser.NoChildren
		}
		// "$$$$" is an empty equation, not display math.
		if util.IsBlank(rest[:idx]) {
			return nil, parser.NoChildren
		}
		node.Lines().Append(text.NewSegment(start, start+idx))
		node.closed = true
		reader.AdvanceToEOL()
		return node, parser.NoChildren
	}

	// Whatever follows "$$" on the opening line is part of the equation, as in
	// "$$\begin{aligned}". Dropping it would feed KaTeX a truncated equation.
	if !util.IsBlank(rest) {
		node.Lines().Append(text.NewSegment(start, segment.Stop))
	}

	reader.AdvanceToEOL()
	return node, parser.NoChildren
}

func (b *BlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	// Open read the whole equation; close before consuming anything further.
	if node.(*Block).closed {
		return parser.Close
	}

	line, segment := reader.PeekLine()

	// Like a fenced code block, the block runs to its closing delimiter or to
	// end of input. Blank lines are kept: KaTeX treats them as whitespace, and
	// ending on one would split an equation and swallow its closing "$$".
	idx := bytes.Index(line, delimiter)
	if idx < 0 {
		node.Lines().Append(segment)
		reader.AdvanceToEOL()
		return parser.Continue | parser.NoChildren
	}

	// The closing "$$" need not begin the line: "\end{aligned}$$" closes too.
	// Indentation ahead of a closing delimiter on its own line is not equation.
	if !util.IsBlank(line[:idx]) {
		node.Lines().Append(text.NewSegment(segment.Start, segment.Start+idx))
	}

	// Consume through the closing delimiter and no further. Anything after it
	// stays in the reader and is parsed as a new block instead of vanishing.
	reader.Advance(idx + 2)
	if util.IsBlank(line[idx+2:]) {
		reader.AdvanceToEOL()
	}

	return parser.Close
}

func (b *BlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// A "$$" left dangling at end of input yields an empty equation. Drop it
	// rather than rendering an empty display block.
	if node.Lines().Len() == 0 {
		if parent := node.Parent(); parent != nil {
			parent.RemoveChild(parent, node)
		}
	}
}

func (b *BlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *BlockParser) CanAcceptIndentedLine() bool {
	return false
}
