package functions

import (
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// to(): 変数への代入(仕様書 4.9)。
// 直近に実行された set() の引数と同じ名前の変数に、引数の値を代入する。
// それ以前に set() が実行されていない場合は何もしない。
//
// 「何もしない」ので、その場合は引数の式を評価すらしないよう遅延評価(ParamLazy)で受ける。
// すべてのデータ型(cd 型を含む)を代入できる。
func toFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "to",
		Params: []interpreter.ParamKind{interpreter.ParamLazy},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			name, ok := ctx.LastSet()
			if !ok {
				return types.NullValue, nil
			}
			v, err := ctx.Eval(args[0].Expr)
			if err != nil {
				return nil, err
			}
			ctx.SetVar(name, v)
			return types.NullValue, nil
		},
	}
}
