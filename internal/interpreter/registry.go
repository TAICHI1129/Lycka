package interpreter

import (
	"fmt"

	"lycka/internal/parser"
	"lycka/internal/types"
)

// ParamKind は関数の引数(括弧1組)をどう受け取るかを表す。
type ParamKind int

const (
	// ParamValue は呼び出し時に1つの式として評価し、値で受け取る。
	ParamValue ParamKind = iota
	// ParamLazy は式を評価せずそのまま受け取る(while の条件、to の値など)。
	// 関数側が Context.Eval で必要なときに評価する。
	ParamLazy
	// ParamCode は cd 型として受け取る。cd[] は省略でき、括弧内の文列がそのままコードになる。
	// 括弧内が cd[...] リテラル1つ、または cd 型の変数1つの場合はそれを使う。
	ParamCode
)

// Arg は関数に渡される引数1つ分。Kind に応じた欄だけが有効。
type Arg struct {
	Kind  ParamKind
	Value types.Value // ParamValue
	Expr  parser.Node // ParamLazy
	Code  *types.Code // ParamCode
}

// Func は組み込み関数の定義。
type Func struct {
	Name string
	// Params は先頭から順の引数の種類。
	Params []ParamKind
	// Rest が非 nil のとき、Params を超える引数はすべてこの種類で受け取り、
	// 引数の総数は MinArgs 以上であればよい(上限なし)。
	Rest    *ParamKind
	MinArgs int
	// Fn は本体。引数は種類に従って解決済み。
	Fn func(ctx *Context, args []Arg) (types.Value, error)
}

// RestKind は Func.Rest に渡すための補助関数。
func RestKind(k ParamKind) *ParamKind { return &k }

func (f *Func) kindAt(i int) ParamKind {
	if i < len(f.Params) {
		return f.Params[i]
	}
	return *f.Rest
}

// checkArity は引数の数 n が許容されるかを返す。許容されない場合、
// 多すぎたのか(tooMany)少なすぎたのかも返す。
func (f *Func) checkArity(n int) (ok, tooMany bool) {
	if f.Rest != nil {
		return n >= f.MinArgs, false
	}
	switch {
	case n < len(f.Params):
		return false, false
	case n > len(f.Params):
		return false, true
	}
	return true, false
}

// Registry は関数名と実際の処理を結び付ける。
// 登録は実行前に済ませる前提で、実行中は読み取り専用(並行実行しても安全)。
type Registry struct {
	funcs map[string]*Func
}

// NewRegistry は空のレジストリを返す。
func NewRegistry() *Registry {
	return &Registry{funcs: map[string]*Func{}}
}

// Register は関数を登録する。名前の重複や不正な定義はエラー。
func (r *Registry) Register(f *Func) error {
	if f == nil || f.Name == "" || f.Fn == nil {
		return fmt.Errorf("invalid function definition")
	}
	if _, dup := r.funcs[f.Name]; dup {
		return fmt.Errorf("function %q is already registered", f.Name)
	}
	r.funcs[f.Name] = f
	return nil
}

// Lookup は名前から関数を探す。
func (r *Registry) Lookup(name string) (*Func, bool) {
	f, ok := r.funcs[name]
	return f, ok
}

// Names は登録済みの関数名の一覧(順不同)を返す。
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.funcs))
	for n := range r.funcs {
		names = append(names, n)
	}
	return names
}
