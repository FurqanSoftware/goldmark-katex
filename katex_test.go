package katex

import (
	"bytes"
	"fmt"
	"testing"
)

func ExampleRender() {
	b := bytes.Buffer{}
	Render(&b, []byte(`Y = A \dot X^2 + B \dot X + C`), false, false)
	fmt.Println(b.String())

	// Output:
	// <span class="katex"><span class="katex-mathml"><math xmlns="http://www.w3.org/1998/Math/MathML"><semantics><mrow><mi>Y</mi><mo>=</mo><mi>A</mi><msup><mover accent="true"><mi>X</mi><mo>˙</mo></mover><mn>2</mn></msup><mo>+</mo><mi>B</mi><mover accent="true"><mi>X</mi><mo>˙</mo></mover><mo>+</mo><mi>C</mi></mrow><annotation encoding="application/x-tex">Y = A \dot X^2 + B \dot X + C</annotation></semantics></math></span><span class="katex-html" aria-hidden="true"><span class="base"><span class="strut" style="height:0.6833em;"></span><span class="mord mathnormal" style="margin-right:0.22222em;">Y</span><span class="mspace" style="margin-right:0.2778em;"></span><span class="mrel">=</span><span class="mspace" style="margin-right:0.2778em;"></span></span><span class="base"><span class="strut" style="height:1.0035em;vertical-align:-0.0833em;"></span><span class="mord mathnormal">A</span><span class="mord"><span class="mord accent"><span class="vlist-t"><span class="vlist-r"><span class="vlist" style="height:0.9202em;"><span style="top:-3em;"><span class="pstrut" style="height:3em;"></span><span class="mord mathnormal" style="margin-right:0.07847em;">X</span></span><span style="top:-3.2523em;"><span class="pstrut" style="height:3em;"></span><span class="accent-body" style="left:-0.0556em;"><span class="mord">˙</span></span></span></span></span></span></span><span class="msupsub"><span class="vlist-t"><span class="vlist-r"><span class="vlist" style="height:0.8141em;"><span style="top:-3.063em;margin-right:0.05em;"><span class="pstrut" style="height:2.7em;"></span><span class="sizing reset-size6 size3 mtight"><span class="mord mtight">2</span></span></span></span></span></span></span></span><span class="mspace" style="margin-right:0.2222em;"></span><span class="mbin">+</span><span class="mspace" style="margin-right:0.2222em;"></span></span><span class="base"><span class="strut" style="height:1.0035em;vertical-align:-0.0833em;"></span><span class="mord mathnormal" style="margin-right:0.05017em;">B</span><span class="mord accent"><span class="vlist-t"><span class="vlist-r"><span class="vlist" style="height:0.9202em;"><span style="top:-3em;"><span class="pstrut" style="height:3em;"></span><span class="mord mathnormal" style="margin-right:0.07847em;">X</span></span><span style="top:-3.2523em;"><span class="pstrut" style="height:3em;"></span><span class="accent-body" style="left:-0.0556em;"><span class="mord">˙</span></span></span></span></span></span></span><span class="mspace" style="margin-right:0.2222em;"></span><span class="mbin">+</span><span class="mspace" style="margin-right:0.2222em;"></span></span><span class="base"><span class="strut" style="height:0.6833em;"></span><span class="mord mathnormal" style="margin-right:0.07153em;">C</span></span></span></span>
}

func BenchmarkRender(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b := bytes.Buffer{}
		Render(&b, []byte(`Y = A \dot X^2 + B \dot X + C`), false, false)
	}
}

func ErrorRender(t *testing.T) {
	// with throwOnError = true
	b := bytes.Buffer{}
	err := Render(&b, []byte(`\invalidcommand`), false, true)
	if err == nil {
		t.Error("Expected error for invalid KaTeX with throwOnError=true, got nil")
	}

	// with throwOnError = false
	err = Render(&b, []byte(`\invalidcommand`), false, false)
	if err != nil {
		t.Errorf("Expected no error for invalid KaTeX with throwOnError=false, got: %v", err)
	}
	if !bytes.Contains(b.Bytes(), []byte("color:#cc0000")) {
		t.Error("Couldn't find a red error message when rendering invalid KaTeX with throwOnError=false.")
	}
}
