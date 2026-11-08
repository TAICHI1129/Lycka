// Package types は Lycka の7種類のデータ型と、その比較・集合演算を定義する。
package types

import lyerr "lycka/internal/errors"

// 型名。エラー表示や isitsth などで使う。
const (
	TypeInt    = "int"
	TypeList   = "list"
	TypeStr    = "str"
	TypeCode   = "cd"
	TypeBool   = "bl"
	TypeNull   = "null"
	TypeSmthng = "smthng"
)

// Value は Lycka のすべての値が満たす共通インターフェース。
// 値はすべて不変(イミュータブル)として扱う。
type Value interface {
	// Type は型名を返す。
	Type() string
	// String は print() で表示する形を返す(str は引用符なし)。
	String() string
	// Inspect はリスト内などで使うリテラル風の形を返す(str は引用符付き)。
	Inspect() string
}

// Tri は三値論理の値。smthng を含む比較は TriUnknown になりうる。
type Tri int

const (
	TriFalse Tri = iota
	TriTrue
	TriUnknown
)

func triFromBool(b bool) Tri {
	if b {
		return TriTrue
	}
	return TriFalse
}

func triAnd(a, b Tri) Tri {
	switch {
	case a == TriFalse || b == TriFalse:
		return TriFalse
	case a == TriUnknown || b == TriUnknown:
		return TriUnknown
	}
	return TriTrue
}

func triOr(a, b Tri) Tri {
	switch {
	case a == TriTrue || b == TriTrue:
		return TriTrue
	case a == TriUnknown || b == TriUnknown:
		return TriUnknown
	}
	return TriFalse
}

func isCode(v Value) bool   { _, ok := v.(*Code); return ok }
func isSmthng(v Value) bool { _, ok := v.(*Smthng); return ok }

// Equal は等価演算子 `x = y` の意味論(仕様書 5.3)を実装する。
//
//   - cd 型には使用できない(エラー)
//   - どちらかが smthng なら結果は不明(TriUnknown)
//   - list と非 list: list の要素に一致するものがあれば TRUE (5.3.1)
//   - list と list: 互いの全要素が相手に含まれれば TRUE、順序は無視 (5.3.2)
//   - 型が異なるその他の組み合わせは FALSE(仕様に規定がないための解釈)
//
// リストの要素に smthng が含まれる場合は Kleene の三値論理で畳み込む。
func Equal(a, b Value) (Tri, error) {
	if isCode(a) {
		return TriFalse, lyerr.OperatorType("=", TypeCode)
	}
	if isCode(b) {
		return TriFalse, lyerr.OperatorType("=", TypeCode)
	}
	if isSmthng(a) || isSmthng(b) {
		return TriUnknown, nil
	}
	la, aList := a.(*List)
	lb, bList := b.(*List)
	switch {
	case aList && bList:
		return listEqual(la, lb)
	case aList:
		return contains(la, b)
	case bList:
		return contains(lb, a)
	}
	return scalarEqual(a, b), nil
}

// elementEqual はリストの要素同士の比較に使う厳密な等価判定。
// `=` と違い、list と非 list は常に FALSE(包含判定をしない)。
func elementEqual(a, b Value) (Tri, error) {
	if isCode(a) || isCode(b) {
		return TriFalse, lyerr.OperatorType("=", TypeCode)
	}
	if isSmthng(a) || isSmthng(b) {
		return TriUnknown, nil
	}
	la, aList := a.(*List)
	lb, bList := b.(*List)
	switch {
	case aList && bList:
		return listEqual(la, lb)
	case aList || bList:
		return TriFalse, nil
	}
	return scalarEqual(a, b), nil
}

func scalarEqual(a, b Value) Tri {
	if a.Type() != b.Type() {
		return TriFalse
	}
	switch x := a.(type) {
	case *Int:
		return triFromBool(x.V.Cmp(b.(*Int).V) == 0)
	case *Str:
		return triFromBool(x.V == b.(*Str).V)
	case *Bool:
		return triFromBool(x.V == b.(*Bool).V)
	case *Null:
		return TriTrue
	}
	return TriFalse
}

// contains は l の要素に v と等しいものがあるかを返す(OR の畳み込み)。
func contains(l *List, v Value) (Tri, error) {
	result := TriFalse
	for _, e := range l.Items {
		t, err := elementEqual(e, v)
		if err != nil {
			return TriFalse, err
		}
		result = triOr(result, t)
	}
	return result, nil
}

// allIn は x の全要素が y に含まれるかを返す(AND の畳み込み)。
func allIn(x, y *List) (Tri, error) {
	result := TriTrue
	for _, e := range x.Items {
		t, err := contains(y, e)
		if err != nil {
			return TriFalse, err
		}
		result = triAnd(result, t)
	}
	return result, nil
}

func listEqual(a, b *List) (Tri, error) {
	ab, err := allIn(a, b)
	if err != nil {
		return TriFalse, err
	}
	ba, err := allIn(b, a)
	if err != nil {
		return TriFalse, err
	}
	return triAnd(ab, ba), nil
}

// looseEqual は和・差で「同じ値」とみなすかの判定。
// TRUE と断定できる場合のみ true(smthng や cd は常に別の値として扱う)。
func looseEqual(a, b Value) bool {
	t, err := elementEqual(a, b)
	return err == nil && t == TriTrue
}
