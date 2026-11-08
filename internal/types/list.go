package types

import "strings"

// List は list 型。要素の型に制限はない。
type List struct{ Items []Value }

// NewList は items を要素とするリストを返す。items は以後変更しないこと。
func NewList(items []Value) *List { return &List{Items: items} }

func (l *List) Type() string   { return TypeList }
func (l *List) String() string { return l.Inspect() }

// Inspect は list[1, "a", TRUE] の形式。
func (l *List) Inspect() string {
	parts := make([]string, len(l.Items))
	for i, it := range l.Items {
		parts[i] = it.Inspect()
	}
	return "list[" + strings.Join(parts, ", ") + "]"
}

// Union は `+`: 両方の全要素を含む新しいリスト。重複する値は1つにまとめる。
// 順序は「a の要素、続いて b の要素」の初出順。
// smthng と cd は同一かどうか断定できない/比較不能なので、まとめずそのまま残す。
func Union(a, b *List) *List {
	out := make([]Value, 0, len(a.Items)+len(b.Items))
	add := func(v Value) {
		for _, e := range out {
			if looseEqual(e, v) {
				return
			}
		}
		out = append(out, v)
	}
	for _, v := range a.Items {
		add(v)
	}
	for _, v := range b.Items {
		add(v)
	}
	return &List{Items: out}
}

// Difference は `-`: a の要素から b に含まれる値を除いた新しいリスト。
// 同一と断定できない要素(smthng など)は除去されない。
func Difference(a, b *List) *List {
	out := make([]Value, 0, len(a.Items))
	for _, v := range a.Items {
		removed := false
		for _, e := range b.Items {
			if looseEqual(v, e) {
				removed = true
				break
			}
		}
		if !removed {
			out = append(out, v)
		}
	}
	return &List{Items: out}
}
