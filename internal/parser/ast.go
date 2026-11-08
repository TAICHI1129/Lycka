// Package parser はトークン列から AST を構築する。
package parser

import "math/big"

// Pos はソース上の位置(1始まり)。
type Pos struct{ Line, Col int }

// Node は AST ノード。
type Node interface {
	Position() Pos
}

type base struct{ P Pos }

func (b base) Position() Pos { return b.P }

// IntLit は int 型リテラル。
type IntLit struct {
	base
	Value *big.Int
}

// StrLit は str 型リテラル。
type StrLit struct {
	base
	Value string
}

// BoolLit は TRUE / FALSE。
type BoolLit struct {
	base
	Value bool
}

// NullLit は null。
type NullLit struct{ base }

// ListLit は list[...]。要素は式。
type ListLit struct {
	base
	Elems []Node
}

// CodeLit は cd[...]。Body は文(関数呼び出し)の列、Src は括弧内の原文。
type CodeLit struct {
	base
	Body []Node
	Src  string
}

// VarRef は変数参照。
type VarRef struct {
	base
	Name string
}

// Arg は関数呼び出しの括弧1組分の引数。
// Items は括弧内に並んだ式(または文)。値として評価される引数では1個、
// cd 型の引数ではコードとして全体が使われる。
type Arg struct {
	Pos   Pos
	Items []Node
	Src   string // 括弧内の原文
}

// Call は関数呼び出し。括弧の組の数だけ Args を持つ。
type Call struct {
	base
	Name string
	Args []*Arg
}

// BinaryOp は二項演算(+ - = ÷)。
type BinaryOp struct {
	base
	Op          string
	Left, Right Node
}

// UnaryMinus は単項マイナス。
type UnaryMinus struct {
	base
	Operand Node
}

// Program はプログラム全体。Body は関数呼び出しの列(c() は除去済み)。
type Program struct {
	Body []Node
}
