package functions

import (
	lyerr "lycka/internal/errors"
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// print(): 値の表示(仕様書 4.1)。
// 対応する型は int / str / bl / list。表示後に改行を付ける。
func printFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "print",
		Params: []interpreter.ParamKind{interpreter.ParamValue},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			v := args[0].Value
			switch v.(type) {
			case *types.Int, *types.Str, *types.Bool, *types.List:
			default:
				return nil, lyerr.FunctionType("print", v.Type())
			}
			if err := ctx.Write(v.String() + "\n"); err != nil {
				return nil, lyerr.OutputFailed(err)
			}
			return types.NullValue, nil
		},
	}
}
