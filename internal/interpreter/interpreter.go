// Package interpreter は字句解析・構文解析・実行処理を連携させ、Lycka プログラムを実行する。
package interpreter

import (
	"fmt"
	"io"

	lyerr "lycka/internal/errors"
	"lycka/internal/parser"
)

// Interpreter はレジストリ・変数・出力先をまとめたインタプリタ本体。
// 組み込み関数は functions パッケージが Registry に登録する
// (interpreter が functions に依存しないようにするため)。
type Interpreter struct {
	reg *Registry
	sh  *shared
}

// New は out に出力するインタプリタを返す。関数は未登録の状態なので、
// 利用前に Registry() へ登録すること(functions.NewInterpreter が行う)。
func New(out io.Writer) *Interpreter {
	reg := NewRegistry()
	return &Interpreter{reg: reg, sh: newShared(reg, out)}
}

// Registry は関数レジストリを返す。
func (it *Interpreter) Registry() *Registry { return it.reg }

// Run はソースコードを構文解析して実行する。
// 変数は Interpreter に保持されるので、同じ Interpreter への Run 同士で共有される。
// 返るエラーは *errors.Error。
func (it *Interpreter) Run(src string) (err error) {
	prog, perr := parser.Parse(src)
	if perr != nil {
		return perr
	}
	return it.RunProgram(prog)
}

// RunProgram は構文解析済みのプログラムを実行する。
func (it *Interpreter) RunProgram(prog *parser.Program) (err error) {
	ctx := newContext(it.sh)
	defer func() {
		if r := recover(); r != nil {
			err = lyerr.Internal(fmt.Sprint("panic: ", r))
		}
	}()
	return ctx.runNodes(prog.Body)
}
