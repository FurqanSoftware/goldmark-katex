package katex

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// stubRenderer stands in for HTMLRenderer, writing the equation goldmark hands
// to KaTeX instead of KaTeX's output. It keeps these tests readable and
// independent of the bundled KaTeX version, which is what the parser tests are
// actually about. Rendering itself is covered by TestExtenderTestExtender.
type stubRenderer struct{}

func (r *stubRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInline, r.renderInline)
	reg.Register(KindBlock, r.renderBlock)
}

func (r *stubRenderer) renderInline(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		node := n.(*Inline)
		tag := "inline"
		if node.Display {
			tag = "display"
		}
		fmt.Fprintf(w, "<%s>%s</%s>", tag, node.Equation, tag)
	}
	return ast.WalkContinue, nil
}

func (r *stubRenderer) renderBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		fmt.Fprintf(w, "<block>%s</block>", n.(*Block).Equation(source))
	}
	return ast.WalkContinue, nil
}

// convertStub runs the real Extender, so parser wiring is under test, but with
// stubRenderer overriding the KaTeX renderer. goldmark registers node renderers
// from the highest priority number down, so the lowest number wins; Extender
// registers HTMLRenderer at 0.
func convertStub(t *testing.T, src string) string {
	t.Helper()
	md := goldmark.New(
		goldmark.WithExtensions(&Extender{}),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(util.Prioritized(&stubRenderer{}, -1)),
		),
	)
	var b bytes.Buffer
	if err := md.Convert([]byte(src), &b); err != nil {
		t.Fatalf("Convert(%q): %v", src, err)
	}
	return b.String()
}

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		// Display math on lines of its own becomes a Block, and reaches KaTeX
		// byte for byte.
		{
			name: "block",
			src:  "$$\nx = 1\n$$\n",
			want: "<block>x = 1\n</block>",
		}, {
			name: "block multiple lines",
			src:  "$$\na\nb\nc\n$$\n",
			want: "<block>a\nb\nc\n</block>",
		}, {
			// A line followed by "=" used to become a Setext heading, which
			// truncated the equation there.
			name: "block with equals line",
			src:  "$$\n\\min_{A}\n=\nB\n$$\n",
			want: "<block>\\min_{A}\n=\nB\n</block>",
		}, {
			// "\\" row separators and "&" alignment markers must survive.
			name: "block with matrix",
			src:  "$$\n\\begin{bmatrix}\na & b \\\\\nc & d\n\\end{bmatrix}\n$$\n",
			want: "<block>\\begin{bmatrix}\na & b \\\\\nc & d\n\\end{bmatrix}\n</block>",
		}, {
			// The inline parser used to give up after 20 lines.
			name: "block over twenty lines",
			src:  "$$\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20\n21\n22\n$$\n",
			want: "<block>1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20\n21\n22\n</block>",
		}, {
			// Markdown syntax inside the equation is left alone.
			name: "block with markdown characters",
			src:  "$$\n\\alpha_1 * \\beta_2 [x](y)\n$$\n",
			want: "<block>\\alpha_1 * \\beta_2 [x](y)\n</block>",
		},

		// Delimiters that hug the equation. Everything between them belongs to
		// the equation, on the opening and closing lines too.
		{
			name: "block content on opening line",
			src:  "$$\\begin{aligned}\na &= b\n\\end{aligned}\n$$\n",
			want: "<block>\\begin{aligned}\na &= b\n\\end{aligned}\n</block>",
		}, {
			name: "block content on closing line",
			src:  "$$\n\\begin{aligned}\na &= b\n\\end{aligned}$$\n",
			want: "<block>\\begin{aligned}\na &= b\n\\end{aligned}</block>",
		}, {
			name: "block content on both delimiter lines",
			src:  "$$\\begin{aligned}\na &= b\n\\end{aligned}$$\n",
			want: "<block>\\begin{aligned}\na &= b\n\\end{aligned}</block>",
		}, {
			name: "block on one line",
			src:  "$$E = mc^2$$\n",
			want: "<block>E = mc^2</block>",
		}, {
			name: "block on one line without trailing newline",
			src:  "$$E = mc^2$$",
			want: "<block>E = mc^2</block>",
		}, {
			name: "block without trailing newline",
			src:  "$$\nx = 1\n$$",
			want: "<block>x = 1\n</block>",
		}, {
			name: "block with trailing spaces after closing delimiter",
			src:  "$$\nx\n$$   \n",
			want: "<block>x\n</block>",
		},

		// Text around a block is never swallowed.
		{
			name: "text after closing delimiter",
			src:  "$$\nx\n$$ trailing\n",
			want: "<block>x\n</block><p>trailing</p>\n",
		}, {
			name: "two blocks on the closing line",
			src:  "$$\nx\n$$ $$y$$\n",
			want: "<block>x\n</block><block>y</block>",
		}, {
			name: "paragraph before block",
			src:  "Some text\n$$\nx\n$$\n",
			want: "<p>Some text</p>\n<block>x\n</block>",
		}, {
			name: "paragraph after block",
			src:  "$$\nx\n$$\nSome text\n",
			want: "<block>x\n</block><p>Some text</p>\n",
		}, {
			name: "paragraphs around block",
			src:  "Before\n\n$$\nx\n$$\n\nAfter\n",
			want: "<p>Before</p>\n<block>x\n</block><p>After</p>\n",
		}, {
			name: "consecutive blocks",
			src:  "$$\na\n$$\n$$\nb\n$$\n",
			want: "<block>a\n</block><block>b\n</block>",
		},

		// Display math written mid-line stays inline, so it nests validly in
		// the paragraph around it.
		{
			name: "display math mid line",
			src:  "text $$x$$ more\n",
			want: "<p>text <display>x</display> more</p>\n",
		}, {
			name: "display math with text after it",
			src:  "$$E = mc^2$$ and more text\n",
			want: "<p><display>E = mc^2</display> and more text</p>\n",
		}, {
			name: "display math with text before it",
			src:  "prefix $$E = mc^2$$\n",
			want: "<p>prefix <display>E = mc^2</display></p>\n",
		},

		// Inline math is untouched by any of this.
		{
			name: "inline math",
			src:  "Inline $a^2$ here.\n",
			want: "<p>Inline <inline>a^2</inline> here.</p>\n",
		}, {
			name: "inline math twice",
			src:  "$a$ and $b$\n",
			want: "<p><inline>a</inline> and <inline>b</inline></p>\n",
		}, {
			name: "inline math with escaped delimiter",
			src:  "$a \\$ b$\n",
			want: "<p><inline>a \\$ b</inline></p>\n",
		}, {
			name: "escaped dollar",
			src:  "Costs \\$5.\n",
			want: "<p>Costs $5.</p>\n",
		}, {
			name: "inline math beside a block",
			src:  "Area $A$ is:\n\n$$\nA = \\pi r^2\n$$\n",
			want: "<p>Area <inline>A</inline> is:</p>\n<block>A = \\pi r^2\n</block>",
		},

		// A block runs to its closing delimiter or to end of input, as a fenced
		// code block does. Blank lines stay part of the equation.
		{
			name: "unterminated block",
			src:  "$$\nx = 1\n",
			want: "<block>x = 1\n</block>",
		}, {
			name: "unterminated block runs to end of input",
			src:  "Hello\n$$\nx = 1\n\nmore prose\n",
			want: "<p>Hello</p>\n<block>x = 1\n\nmore prose\n</block>",
		}, {
			name: "blank line inside block",
			src:  "$$\na\n\nb\n$$\n",
			want: "<block>a\n\nb\n</block>",
		}, {
			name: "blank line inside aligned environment",
			src:  "$$\n\\begin{aligned}\na &= b \\\\\n\nc &= d\n\\end{aligned}\n$$\n",
			want: "<block>\\begin{aligned}\na &= b \\\\\n\nc &= d\n\\end{aligned}\n</block>",
		},

		// Degenerate delimiters are not display math.
		{
			name: "empty one line block",
			src:  "$$$$\n",
			want: "<p>$$$$</p>\n",
		}, {
			name: "empty block",
			src:  "$$\n$$\n",
			want: "",
		}, {
			name: "dangling opening delimiter",
			src:  "$$\n",
			want: "",
		}, {
			name: "dangling opening delimiter in blockquote",
			src:  "> $$\n",
			want: "<blockquote>\n</blockquote>\n",
		}, {
			name: "whitespace only equation",
			src:  "$$  $$\n",
			want: "<p><display>  </display></p>\n",
		},

		// Container and code blocks behave as they do for any other block.
		{
			name: "block in blockquote",
			src:  "> $$\n> x = 1\n> $$\n",
			want: "<blockquote>\n<block>x = 1\n</block></blockquote>\n",
		}, {
			name: "block in nested blockquote",
			src:  "> > $$\n> > x\n> > $$\n",
			want: "<blockquote>\n<blockquote>\n<block>x\n</block></blockquote>\n</blockquote>\n",
		}, {
			name: "block in list item",
			src:  "- item\n\n  $$\n  x = 1\n  $$\n",
			want: "<ul>\n<li>\n<p>item</p>\n<block>x = 1\n</block></li>\n</ul>\n",
		}, {
			name: "block in fenced code",
			src:  "```\n$$\nx = 1\n$$\n```\n",
			want: "<pre><code>$$\nx = 1\n$$\n</code></pre>\n",
		}, {
			name: "block in indented code",
			src:  "    $$\n    x = 1\n    $$\n",
			want: "<pre><code>$$\nx = 1\n$$\n</code></pre>\n",
		}, {
			name: "block in code span",
			src:  "Use `$$x$$` literally.\n",
			want: "<p>Use <code>$$x$$</code> literally.</p>\n",
		}, {
			// goldmark strips the list item's indentation before the block
			// parser sees the line.
			name: "block in list item with text after closing delimiter",
			src:  "- item\n\n  $$\n  x\n  $$ after\n",
			want: "<ul>\n<li>\n<p>item</p>\n<block>x\n</block><p>after</p>\n</li>\n</ul>\n",
		}, {
			// Indentation up to three spaces opens a block. It stays on the
			// equation's own lines, where KaTeX treats it as whitespace, but
			// not on the line holding the closing delimiter.
			name: "indented block",
			src:  "   $$\n   x\n   $$\n",
			want: "<block>   x\n</block>",
		}, {
			name: "carriage returns",
			src:  "Before\r\n\r\n$$\r\nx = 1\r\n$$\r\n\r\nAfter\r\n",
			want: "<p>Before</p>\n<block>x = 1\r\n</block><p>After</p>\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := convertStub(t, c.src)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("src:\n%s\nwant: %q\ngot:  %q\ndiff:\n%s", c.src, c.want, got, diff)
			}
		})
	}
}
