package functions

import (
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// c()...: コメント(仕様書 4.7)。
//
// 「プログラムの実行前に当該関数が削除される」ため、c() の呼び出しは構文解析の直後に
// パーサ(parser.CommentFunc)がプログラムから取り除く。したがって実行時にこの関数が
// 呼ばれることは通常ない。ここでの登録は、c が組み込み関数名として予約されていることを
// レジストリ上でも明示するためのもので、万一呼ばれても引数を評価せず何もしない
// (すべての引数を ParamLazy で受けるので、あらゆるデータ型・式を受け付ける)。
func commentFunc() *interpreter.Func {
	return &interpreter.Func{
		Name: "c",
		Rest: interpreter.RestKind(interpreter.ParamLazy),
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			return types.NullValue, nil
		},
	}
}
