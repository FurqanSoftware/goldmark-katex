package katex

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
)

// update rewrites the testdata/*.html golden files from the current output,
// for use after a KaTeX upgrade changes its markup:
//
//	go test -run TestExtenderTestExtender -update
//
// Review the diff before committing it.
var update = flag.Bool("update", false, "rewrite testdata golden files")

func TestExtenderTestExtender(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join("testdata", strings.TrimSuffix(entry.Name(), ".md")+".html")
			want, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			var link []byte
			if bytes.HasPrefix(want, []byte("<link ")) {
				n := bytes.Index(want, []byte("\n")) + 1
				link, want = want[:n], want[n:]
			}
			got := bytes.Buffer{}
			err = goldmark.New(goldmark.WithExtensions(&Extender{})).Convert(in, &got)
			if err != nil {
				t.Fatal(err)
			}
			if *update {
				if err := os.WriteFile(name, append(link, got.Bytes()...), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			if diff := cmp.Diff(want, got.Bytes()); diff != "" {
				t.Fatalf("%s:\n\nwant:\n%s\n\ngot:\n%s\n\ndiff:\n%s\n", entry.Name(), want, got.String(), diff)
			}
		})
	}
}

const displayMarker = `<span class="katex-display"`

// displaySpans returns each rendered display equation in html, delimited by
// balancing the <span> that opens it against its matching </span>.
func displaySpans(html string) []string {
	var spans []string
	for i := 0; i < len(html); {
		start := strings.Index(html[i:], displayMarker)
		if start < 0 {
			return spans
		}
		start += i
		depth, j := 0, start
		for j < len(html) {
			switch {
			case strings.HasPrefix(html[j:], "<span"):
				depth++
				j += len("<span")
			case strings.HasPrefix(html[j:], "</span>"):
				depth--
				j += len("</span>")
				if depth == 0 {
					spans = append(spans, html[start:j])
				}
			default:
				j++
			}
			if depth == 0 {
				break
			}
		}
		i = j
	}
	return spans
}

// Display math on lines of its own is a block, so it stands beside the
// surrounding paragraphs rather than inside one. Wrapping it in a <p> was
// invalid nesting, since KaTeX's display output is block-level.
func TestDisplayMathNesting(t *testing.T) {
	src := []byte("Before.\n\n$$\nx = 1\n$$\n\nAfter.\n")

	got := bytes.Buffer{}
	if err := goldmark.New(goldmark.WithExtensions(&Extender{})).Convert(src, &got); err != nil {
		t.Fatal(err)
	}
	out := got.String()

	if n := strings.Count(out, displayMarker); n != 1 {
		t.Fatalf("want 1 rendered display equation, got %d:\n%s", n, out)
	}
	// Nothing in the document opens a paragraph that is still open where the
	// equation begins.
	before := out[:strings.Index(out, displayMarker)]
	if strings.Count(before, "<p>") != strings.Count(before, "</p>") {
		t.Errorf("display math is nested inside a <p>; got:\n%s", out)
	}
	if !strings.Contains(out, "<p>Before.</p>") || !strings.Contains(out, "<p>After.</p>") {
		t.Errorf("surrounding paragraphs did not survive; got:\n%s", out)
	}
}

// Display math written in the middle of a line stays an inline node, so it does
// nest inside the paragraph around it, and the text around it survives.
func TestDisplayMathInline(t *testing.T) {
	src := []byte("text $$x = 1$$ more\n")

	got := bytes.Buffer{}
	if err := goldmark.New(goldmark.WithExtensions(&Extender{})).Convert(src, &got); err != nil {
		t.Fatal(err)
	}
	out := got.String()

	if n := strings.Count(out, displayMarker); n != 1 {
		t.Fatalf("want 1 rendered display equation, got %d:\n%s", n, out)
	}
	if !strings.HasPrefix(out, "<p>text ") || !strings.HasSuffix(out, " more</p>\n") {
		t.Errorf("text around mid-line display math did not survive; got:\n%s", out)
	}
}

// The renderer serves repeated equations from a cache. A cached rendering must
// be byte-identical to the first one, and inline and display renderings of the
// same equation must not collide in the cache.
func TestRenderCacheConsistency(t *testing.T) {
	src := []byte("$$\nx = 1\n$$\n\nBetween.\n\n$$\nx = 1\n$$\n")

	got := bytes.Buffer{}
	if err := goldmark.New(goldmark.WithExtensions(&Extender{})).Convert(src, &got); err != nil {
		t.Fatal(err)
	}
	out := got.String()

	eqs := displaySpans(out)
	if len(eqs) != 2 {
		t.Fatalf("want 2 display equations, got %d:\n%s", len(eqs), out)
	}
	if eqs[0] != eqs[1] {
		t.Errorf("cached rendering differs from the first:\nfirst:  %s\nsecond: %s", eqs[0], eqs[1])
	}

	// The same equation inline and in display mode. Sharing one cache would
	// serve the display rendering for the inline one, or the reverse.
	got.Reset()
	if err := goldmark.New(goldmark.WithExtensions(&Extender{})).Convert([]byte("$x = 1$\n\n$$\nx = 1\n$$\n"), &got); err != nil {
		t.Fatal(err)
	}
	out = got.String()
	if n := strings.Count(out, displayMarker); n != 1 {
		t.Errorf("want exactly 1 display rendering, got %d:\n%s", n, out)
	}
	if !strings.Contains(out, `<p><span class="katex">`) {
		t.Errorf("inline equation was not rendered in inline mode; got:\n%s", out)
	}
}
