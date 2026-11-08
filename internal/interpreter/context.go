package interpreter

import (
	"io"
	"sync"

	"lycka/internal/types"
)

// CondResult は直近の if() の条件式の評価結果。else() が参照する。
type CondResult int

const (
	CondNone    CondResult = iota // まだ if() が実行されていない
	CondTrue                      // TRUE
	CondFalse                     // FALSE
	CondUnknown                   // smthng(TRUE とも FALSE とも言えない)
)

// maxDepth はコードの入れ子実行(cd の中の cd、変数に入れたコードの自己参照など)の上限。
// Go のスタックを使い切って異常終了するのを防ぐ。
const maxDepth = 10000

// shared は実行全体で共有される状態。並行実行(concur)から同時に触られるため
// ロックで保護する。
type shared struct {
	reg *Registry

	mu   sync.RWMutex
	vars map[string]types.Value

	outMu sync.Mutex
	out   io.Writer
}

// Context は実行コンテキスト。変数や実行状態を保持する。
//
// 変数は全体で共有される1つの名前空間にある。一方「直近の set」「直近の if」は
// Context ごとの状態で、concur の各ブランチは分岐時点の値を引き継いだ独立した
// 状態を持つ(ブランチ内の set → to が他のブランチに邪魔されないようにするため)。
type Context struct {
	sh       *shared
	lastSet  string
	hasSet   bool
	lastCond CondResult
	depth    int
}

func newShared(reg *Registry, out io.Writer) *shared {
	return &shared{reg: reg, vars: map[string]types.Value{}, out: out}
}

func newContext(sh *shared) *Context {
	return &Context{sh: sh}
}

// Fork は並行ブランチ用の Context を作る。変数と出力は共有し、
// 直近の set / if の状態は現時点の値をコピーして独立させる。
func (c *Context) Fork() *Context {
	cp := *c
	return &cp
}

// Write は標準出力側に s を書き出す。並行実行でも1回の書き込みは分断されない。
func (c *Context) Write(s string) error {
	c.sh.outMu.Lock()
	defer c.sh.outMu.Unlock()
	_, err := io.WriteString(c.sh.out, s)
	return err
}

// DeclareVar は変数 name を null で作成する。既に存在する場合は何もせず false を返す。
func (c *Context) DeclareVar(name string) bool {
	c.sh.mu.Lock()
	defer c.sh.mu.Unlock()
	if _, exists := c.sh.vars[name]; exists {
		return false
	}
	c.sh.vars[name] = types.NullValue
	return true
}

// SetVar は変数 name に値を代入する。
func (c *Context) SetVar(name string, v types.Value) {
	c.sh.mu.Lock()
	defer c.sh.mu.Unlock()
	c.sh.vars[name] = v
}

// GetVar は変数の値を返す。未定義なら ok は false。
func (c *Context) GetVar(name string) (types.Value, bool) {
	c.sh.mu.RLock()
	defer c.sh.mu.RUnlock()
	v, ok := c.sh.vars[name]
	return v, ok
}

// SetLastSet は「直近に実行された set() の引数の文字列」を記録する。
func (c *Context) SetLastSet(name string) {
	c.lastSet, c.hasSet = name, true
}

// LastSet は直近の set() の変数名を返す。まだ実行されていなければ ok は false。
func (c *Context) LastSet() (name string, ok bool) {
	return c.lastSet, c.hasSet
}

// SetLastCond は直近の if() の条件式の評価結果を記録する。
func (c *Context) SetLastCond(r CondResult) { c.lastCond = r }

// LastCond は直近の if() の条件式の評価結果を返す。
func (c *Context) LastCond() CondResult { return c.lastCond }
