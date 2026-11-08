package functions

import (
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// if()(): 条件分岐(仕様書 4.2)。
// 条件式が TRUE のとき2つ目の引数のコードを実行する。
//
// 条件の評価結果は、コードの実行が終わった後に「直近の if の結果」として記録する。
// こうすると if のコードの中で別の if が実行されても、直後の else() は
// 外側の if の結果を参照する(else の位置から見て直近に完了した if が対象)。
func ifFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:   "if",
		Params: []interpreter.ParamKind{interpreter.ParamValue, interpreter.ParamCode},
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			res, err := condOf("if", args[0].Value)
			if err != nil {
				return nil, err
			}
			if res == interpreter.CondTrue {
				if err := ctx.Run(args[1].Code); err != nil {
					return nil, err
				}
			}
			ctx.SetLastCond(res)
			return types.NullValue, nil
		},
	}
}
