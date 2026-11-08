package types

import (
	"errors"
	"testing"

	lyerr "lycka/internal/errors"
)

func i(n int64) Value  { return IntFromInt64(n) }
func s(x string) Value { return &Str{V: x} }
func l(vs ...Value) *List {
	return NewList(vs)
}

func mustEqual(t *testing.T, a, b Value) Tri {
	t.Helper()
	got, err := Equal(a, b)
	if err != nil {
		t.Fatalf("Equal(%s, %s): %v", a.Inspect(), b.Inspect(), err)
	}
	return got
}

func TestEqualScalars(t *testing.T) {
	cases := []struct {
		a, b Value
		want Tri
	}{
		{i(1), i(1), TriTrue},
		{i(1), i(2), TriFalse},
		{s("a"), s("a"), TriTrue},
		{s("a"), s("b"), TriFalse},
		{True, True, TriTrue},
		{True, False, TriFalse},
		{NullValue, NullValue, TriTrue},
		{i(1), s("1"), TriFalse},
		{NullValue, i(0), TriFalse},
		{True, i(1), TriFalse},
	}
	for _, c := range cases {
		if got := mustEqual(t, c.a, c.b); got != c.want {
			t.Errorf("%s = %s: got %v, want %v", c.a.Inspect(), c.b.Inspect(), got, c.want)
		}
	}
}

func TestSmthngEqualityIsUnknown(t *testing.T) {
	x := NewSmthng()
	if got := mustEqual(t, x, x); got != TriUnknown {
		t.Errorf("x = x: got %v, want unknown", got)
	}
	if got := mustEqual(t, x, NewSmthng()); got != TriUnknown {
		t.Errorf("x = y: got %v, want unknown", got)
	}
	if got := mustEqual(t, x, i(1)); got != TriUnknown {
		t.Errorf("smthng = 1: got %v, want unknown", got)
	}
	if got := mustEqual(t, i(1), x); got != TriUnknown {
		t.Errorf("1 = smthng: got %v, want unknown", got)
	}
}

func TestListVsScalar(t *testing.T) {
	if got := mustEqual(t, l(i(1), i(2), i(3)), i(2)); got != TriTrue {
		t.Errorf("contained: %v", got)
	}
	if got := mustEqual(t, i(2), l(i(1), i(2), i(3))); got != TriTrue {
		t.Errorf("contained (reversed): %v", got)
	}
	if got := mustEqual(t, l(i(1), i(2)), i(5)); got != TriFalse {
		t.Errorf("absent: %v", got)
	}
	if got := mustEqual(t, l(), i(1)); got != TriFalse {
		t.Errorf("empty list: %v", got)
	}
	// smthng を含むリストは「含まれるかもしれない」ので不明。
	if got := mustEqual(t, l(i(1), NewSmthng()), i(5)); got != TriUnknown {
		t.Errorf("maybe contained: %v", got)
	}
	// 断定できる一致があれば、smthng が混ざっていても TRUE。
	if got := mustEqual(t, l(i(1), NewSmthng()), i(1)); got != TriTrue {
		t.Errorf("surely contained: %v", got)
	}
}

func TestListVsList(t *testing.T) {
	cases := []struct {
		a, b *List
		want Tri
	}{
		{l(i(1), i(2), i(3)), l(i(3), i(1), i(2)), TriTrue}, // 順序は無関係
		{l(i(1), i(2)), l(i(1), i(2), i(3)), TriFalse},      // 片方向だけでは不成立
		{l(i(1), i(2), i(3)), l(i(1), i(2)), TriFalse},
		{l(), l(), TriTrue},
		{l(i(1), i(1)), l(i(1)), TriTrue},             // 重複は集合として無視される
		{l(l(i(1), i(2))), l(l(i(2), i(1))), TriTrue}, // 入れ子も同じ規則
		{l(NewSmthng()), l(NewSmthng()), TriUnknown},
		{l(i(1), NewSmthng()), l(i(2)), TriFalse}, // 1 が相手に無いので断定で FALSE
	}
	for _, c := range cases {
		if got := mustEqual(t, c.a, c.b); got != c.want {
			t.Errorf("%s = %s: got %v, want %v", c.a.Inspect(), c.b.Inspect(), got, c.want)
		}
	}
}

func TestEqualRejectsCode(t *testing.T) {
	code := &Code{Src: "print(0)"}
	for _, pair := range [][2]Value{{code, code}, {code, i(1)}, {i(1), code}, {l(code), i(1)}} {
		_, err := Equal(pair[0], pair[1])
		var le *lyerr.Error
		if !errors.As(err, &le) || le.ID() != "2/3/1/cd" {
			t.Errorf("Equal(%s, %s): err %v", pair[0].Inspect(), pair[1].Inspect(), err)
		}
	}
}

func TestUnionMergesDuplicates(t *testing.T) {
	got := Union(l(i(1), i(2), i(2)), l(i(2), i(3), s("a"), s("a")))
	if want := `list[1, 2, 3, "a"]`; got.Inspect() != want {
		t.Errorf("got %s, want %s", got.Inspect(), want)
	}
	// 1 と "1" は別の値。
	if got := Union(l(i(1)), l(s("1"))); len(got.Items) != 2 {
		t.Errorf("1 and \"1\" must stay distinct: %s", got.Inspect())
	}
}

func TestUnionKeepsUndeterminableValues(t *testing.T) {
	x := NewSmthng()
	if got := Union(l(x), l(x)); len(got.Items) != 2 {
		t.Errorf("smthng cannot be merged: %s", got.Inspect())
	}
	code := &Code{Src: "print(0)"}
	if got := Union(l(code), l(code)); len(got.Items) != 2 {
		t.Errorf("cd is not comparable, must not be merged: %s", got.Inspect())
	}
}

func TestDifference(t *testing.T) {
	got := Difference(l(i(1), i(2), i(3), i(2)), l(i(2), i(9)))
	if want := "list[1, 3]"; got.Inspect() != want {
		t.Errorf("got %s, want %s", got.Inspect(), want)
	}
	x := NewSmthng()
	if got := Difference(l(x), l(x)); len(got.Items) != 1 {
		t.Errorf("smthng cannot be removed: %s", got.Inspect())
	}
}

func TestStrRemove(t *testing.T) {
	if got := (&Str{V: "abcabc"}).Remove(&Str{V: "bc"}); got.V != "aa" {
		t.Errorf("got %q", got.V)
	}
	if got := (&Str{V: "abc"}).Remove(&Str{V: ""}); got.V != "abc" {
		t.Errorf("empty removal: %q", got.V)
	}
}

func TestIntArithmetic(t *testing.T) {
	a := IntFromInt64(-7)
	b := IntFromInt64(2)
	if got := a.Quo(b).String(); got != "-3" {
		t.Errorf("-7 ÷ 2 = %s, want -3 (truncation toward zero)", got)
	}
	if got := a.Neg().String(); got != "7" {
		t.Errorf("neg: %s", got)
	}
}

func TestInspectAndString(t *testing.T) {
	v := l(i(1), s(`a"b`), True, NullValue, NewSmthng(), l(i(2)))
	if want := `list[1, "a\"b", TRUE, null, smthng, list[2]]`; v.String() != want {
		t.Errorf("got %s, want %s", v.String(), want)
	}
	if s("plain").String() != "plain" || s("plain").Inspect() != `"plain"` {
		t.Error("str String/Inspect")
	}
}

func TestConcatCode(t *testing.T) {
	got := ConcatCode(&Code{Src: "print(1)"}, &Code{Src: "print(2)"})
	if got.Src != "print(1) print(2)" {
		t.Errorf("src %q", got.Src)
	}
	if got := ConcatCode(&Code{}, &Code{Src: "print(2)"}); got.Src != "print(2)" {
		t.Errorf("empty left: %q", got.Src)
	}
}
