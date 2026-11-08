package functions

import (
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// while()(): 繰り返し処理(仕様書 4.5)。
// 条件式が TRUE の間、2つ目の引数のコードを繰り返し実行する。
// 条件式は毎回評価し直すため、1つ目の引数は遅延評価(ParamLazy)で受け取る。
// 条件式が smthng のときは TRUE ではないのでループを抜ける。
func whileFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "while",
		Params: []interpreter.ParamKind{interpreter.ParamLazy, interpreter.ParamCode},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			for {
				v, err := ctx.Eval(args[0].Expr)
				if err != nil {
					return nil, err
				}
				res, err := condOf("while", v)
				if err != nil {
					return nil, err
				}
				if res != interpreter.CondTrue {
					return types.NullValue, nil
				}
				if err := ctx.Run(args[1].Code); err != nil {
					return nil, err
				}
			}
		},
	}
}
