package katex

import (
	_ "embed"
	"fmt"
	"io"
	"runtime"

	"github.com/lithdew/quickjs"
)

//go:embed katex.min.js
var code string

func Render(w io.Writer, src []byte, display bool, throwOnError bool) error {
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
	result, err = context.Eval(fmt.Sprintf(`katex.renderToString(_EqSrc3120, {
		displayMode: %t,
		throwOnError: %t
	})`, display, throwOnError))
	if err != nil {
		return err
	}
	defer result.Free()

	_, err = io.WriteString(w, result.String())
	return err
}
