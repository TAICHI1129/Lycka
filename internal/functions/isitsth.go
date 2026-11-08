package functions

import (
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// isitsth(): smthng 型の判定(仕様書 4.4)。
// 引数が smthng 型なら TRUE、それ以外なら FALSE。すべてのデータ型に対応する。
func isitsthFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "isitsth",
		Params: []interpreter.ParamKind{interpreter.ParamValue},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			_, ok := args[0].Value.(*types.Smthng)
			return types.FromBool(ok), nil
		},
	}
}
