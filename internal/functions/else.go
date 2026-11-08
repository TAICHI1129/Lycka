package functions

import (
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// else(): 条件分岐(仕様書 4.3)。
// 直近の if() の条件式が FALSE のときだけ引数のコードを実行する。
// if() が未実行の場合や、結果が TRUE の場合は何もしない。
// 結果が smthng(TRUE とも FALSE とも言えない)の場合も、FALSE ではないので何もしない。
// 間に無関係な関数が挟まっていても参照は有効(仕様書 7.7)。
func elseFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "else",
		Params: []interpreter.ParamKind{interpreter.ParamCode},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			if ctx.LastCond() == interpreter.CondFalse {
				if err := ctx.Run(args[0].Code); err != nil {
					return nil, err
				}
			}
			return types.NullValue, nil
		},
	}
}
