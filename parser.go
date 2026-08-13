package katex

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser is the inline parser for math delimited by dollar signs.
//
// It handles inline math ("$...$") and display math written in the middle of a
// line ("text $$x$$ more"). Display math that begins a line is claimed earlier
// by BlockParser, so it never reaches this parser; what is left here is always
// confined to a single line.
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

	var start, end, advance int

	display := len(line) > 1 && line[1] == trigger

	if display {
		start = lstart + 2

		for j := 2; j+1 < len(line); j++ {
			if line[j] == trigger && line[j+1] == trigger {
				end = lstart + j
				advance = 2
				break
			}
		}

	} else {
		start = lstart + 1

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
	}

	if end == 0 || end >= len(buf) || buf[end] != trigger {
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
		Display:  display,
	}
}
