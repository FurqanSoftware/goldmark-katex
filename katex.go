package katex

import (
	_ "embed"
	"io"
	"runtime"

	"github.com/lithdew/quickjs"
)

//go:embed katex.min.js
var code string

func Render(w io.Writer, src []byte, display bool) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	runtime := quickjs.NewRuntime()
	defer runtime.Free()

	context := runtime.NewContext()
	defer context.Free()

	globals := context.Globals()

	result, err := context.Eval(code)
	if err != nil {
		return err
	}
	defer result.Free()

	globals.Set("_EqSrc3120", context.String(string(src)))
	if display {
		result, err = context.Eval("katex.renderToString(_EqSrc3120, { displayMode: true })")
	} else {
		result, err = context.Eval("katex.renderToString(_EqSrc3120)")
	}
	defer result.Free()

	_, err = io.WriteString(w, result.String())
	return err
}
