package functions

import (
	lyerr "lycka/internal/errors"
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// set(): 変数の宣言(仕様書 4.8)。
// 引数(str 型のみ)と同じ名前の変数を作る(初期値は null)。
// 同名の変数が既にあれば何も作らず、既存の値も変えない。
// どちらの場合も「直近の set」は更新される(直後の to() の対象になる)。
func setFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "set",
		Params: []interpreter.ParamKind{interpreter.ParamValue},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			s, ok := args[0].Value.(*types.Str)
			if !ok {
				return nil, lyerr.FunctionType("set", args[0].Value.Type())
			}
			ctx.DeclareVar(s.V)
			ctx.SetLastSet(s.V)
			return types.NullValue, nil
		},
	}
}
