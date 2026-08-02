package katex

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
)

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
			want, err := os.ReadFile(filepath.Join("testdata", strings.TrimSuffix(entry.Name(), ".md")+".html"))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.HasPrefix(want, []byte("<link ")) {
				want = want[bytes.Index(want, []byte("\n"))+1:]
			}
			got := bytes.Buffer{}
			err = goldmark.New(goldmark.WithExtensions(&Extender{})).Convert(in, &got)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want, got.Bytes()); diff != "" {
				t.Fatalf("%s:\n\nwant:\n%s\n\ngot:\n%s\n\ndiff:\n%s\n", entry.Name(), want, got.String(), diff)
			}
		})
	}
}

// The display-math renderer wrapped its output in a <div>, but only on the
// cache-miss path, so a repeated equation rendered differently from its first
// occurrence. The wrapper was also invalid nesting: Block embeds
// ast.BaseInline, so goldmark emits it inside the enclosing <p>.
func TestDisplayMathNestingAndCacheConsistency(t *testing.T) {
	src := []byte("$$\nx = 1\n$$\n\nBetween.\n\n$$\nx = 1\n$$\n")

	got := bytes.Buffer{}
	if err := goldmark.New(goldmark.WithExtensions(&Extender{})).Convert(src, &got); err != nil {
		t.Fatal(err)
	}
	out := got.String()

	if strings.Contains(out, "<div") {
		t.Errorf("display math must not emit a <div> inside <p>; got:\n%s", out)
	}
	if n := strings.Count(out, `<span class="katex-display"`); n != 2 {
		t.Errorf("want 2 rendered display equations, got %d:\n%s", n, out)
	}
	// The two equations are identical, so the second is served from cache. Both
	// renderings must be byte-identical.
	marker := `<span class="katex-display"`
	parts := strings.Split(out, marker)
	if len(parts) != 3 {
		t.Fatalf("expected exactly 2 display equations, split gave %d parts", len(parts)-1)
	}
	first := marker + parts[1][:strings.Index(parts[1], "</p>")]
	second := marker + parts[2][:strings.Index(parts[2], "</p>")]
	if first != second {
		t.Errorf("cached rendering differs from the first:\nfirst:  %s\nsecond: %s", first, second)
	}
}
