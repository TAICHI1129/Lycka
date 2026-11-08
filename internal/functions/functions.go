// Package functions は Lycka の組み込み関数(仕様書 4)を実装し、レジストリに登録する。
package functions

import (
	"fmt"
	"io"

	lyerr "lycka/internal/errors"
	"lycka/internal/interpreter"
	"lycka/internal/types"
)

// RegisterAll は仕様書 4.1〜4.9 のすべての組み込み関数を reg に登録する。
func RegisterAll(reg *interpreter.Registry) error {
	defs := []*interpreter.Func{
		printFunc(),
		ifFunc(),
		elseFunc(),
		isitsthFunc(),
		whileFunc(),
		concurFunc(),
		commentFunc(),
		setFunc(),
		toFunc(),
	}
	for _, d := range defs {
		if err := reg.Register(d); err != nil {
			return fmt.Errorf("registering %q: %w", d.Name, err)
		}
	}
	return nil
}

// NewInterpreter は組み込み関数を登録済みのインタプリタを返す。
func NewInterpreter(out io.Writer) *interpreter.Interpreter {
	it := interpreter.New(out)
	if err := RegisterAll(it.Registry()); err != nil {
		// 組み込み定義の不整合はプログラミングミスなので即座に落とす。
		panic(err)
	}
	return it
}

// condOf は if / while の条件式の値を三値に変換する。
// bl 以外では smthng のみ許される(結果は CondUnknown)。
func condOf(fn string, v types.Value) (interpreter.CondResult, error) {
	switch x := v.(type) {
	case *types.Bool:
		if x.V {
			return interpreter.CondTrue, nil
		}
		return interpreter.CondFalse, nil
	case *types.Smthng:
		return interpreter.CondUnknown, nil
	}
	return interpreter.CondNone, lyerr.FunctionType(fn, v.Type())
}
