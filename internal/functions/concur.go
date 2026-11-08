package functions

import (
	"fmt"
	"sync"

	lyerr "lycka/internal/errors"
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// concur()()...: 並行・並列実行(仕様書 4.6)。
// 2つ以上のコードをそれぞれ goroutine で実行し、すべての完了を待つ。
// goroutine は並列実行が可能な環境(GOMAXPROCS > 1)では実際に並列に走るので、
// 「並行実行、可能なら並列実行」の仕様をそのまま満たす。
//
// 各ブランチは Context.Fork() で独立した「直近の set / if」の状態を持つ。
// 変数と出力は共有される(変数はロックで保護、print は1回ずつ分断されない)。
// 複数のブランチがエラーになった場合は、引数の並び順で最初のエラーを返す。
func concurFunc() *interpreter.Func {
	return &interpreter.Func{
		Name:    "concur",
		Rest:    interpreter.RestKind(interpreter.ParamCode),
		MinArgs: 2,
		Fn: func(ctx *interpreter.Context, args []interpreter.Arg) (types.Value, error) {
			errs := make([]error, len(args))
			var wg sync.WaitGroup
			for i := range args {
				child := ctx.Fork()
				code := args[i].Code
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					defer func() {
						if r := recover(); r != nil {
							errs[i] = lyerr.Internal(fmt.Sprint("panic: ", r))
						}
					}()
					errs[i] = child.Run(code)
				}(i)
			}
			wg.Wait()
			for _, err := range errs {
				if err != nil {
					return nil, err
				}
			}
			return types.NullValue, nil
		},
	}
}
