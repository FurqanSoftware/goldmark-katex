package katex

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser is the inline parser for math delimited by single dollars ("$...$").
//
// Display math ("$$...$$") is handled at the block level by BlockParser, so
// this parser deliberately ignores a leading "$$".
type Parser struct {
}

func (s *Parser) Trigger() []byte {
	return []byte{'$'}
}

func (s *Parser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	buf := block.Source()
	ln, pos := block.Position()

	lstart := pos.Start
	lend := pos.Stop
	line := buf[lstart:lend]

	trigger := line[0]

	// "$$" is display math, handled by BlockParser. Leave it alone.
	if len(line) > 1 && line[1] == trigger {
		return nil
	}

	start := lstart + 1

	var end, advance int
	for i := 1; i < len(line); i++ {
		c := line[i]
		if c == '\\' {
			i++
			continue
		}
		if c == trigger {
			end = lstart + i
			advance = 1
			break
		}
	}
	if end >= len(buf) || buf[end] != trigger {
		return nil
	}

	if start >= end {
		return nil
	}

	newpos := end + advance
	if newpos < lend {
		block.SetPosition(ln, text.NewSegment(newpos, lend))
	} else {
		block.Advance(newpos)
	}

	return &Inline{
		Equation: buf[start:end],
	}
}
