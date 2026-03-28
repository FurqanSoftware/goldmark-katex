package katex

import (
	_ "embed"
	"fmt"
	"io"

	"modernc.org/quickjs"
)

//go:embed katex.min.js
var code string

func Render(w io.Writer, src []byte, display bool, throwOnError bool) error {
	vm, err := quickjs.NewVM()
	if err != nil {
		return err
	}
	defer vm.Close()

	_, err = vm.Eval(code, quickjs.EvalGlobal)
	if err != nil {
		return err
	}

	expr := fmt.Sprintf(`katex.renderToString(%q, {
		displayMode: %t,
		throwOnError: %t
	})`, string(src), display, throwOnError)

	result, err := vm.Eval(expr, quickjs.EvalGlobal)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, result.(string))
	return err
}
