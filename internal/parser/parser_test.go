package parser

import (
	"errors"
	"testing"

	lyerr "lycka/internal/errors"
)

func mustParse(t *testing.T, src string) *Program {
	t.Helper()
	p, err := Parse(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return p
}

func errID(t *testing.T, src string) string {
	t.Helper()
	_, err := Parse(src)
	var le *lyerr.Error
	if !errors.As(err, &le) {
		t.Fatalf("parse %q: expected *errors.Error, got %v", src, err)
	}
	return le.ID()
}

func TestStatementsAreCalls(t *testing.T) {
	p := mustParse(t, "set(\"x\") to(1)\nprint(x)")
	if len(p.Body) != 3 {
		t.Fatalf("got %d statements, want 3", len(p.Body))
	}
	names := []string{"set", "to", "print"}
	for i, n := range p.Body {
		c, ok := n.(*Call)
		if !ok || c.Name != names[i] {
			t.Errorf("statement %d: %#v", i, n)
		}
	}
}

func TestMultipleParenGroupsAreArguments(t *testing.T) {
	p := mustParse(t, "concur(print(0))(print(1))(print(2))")
	c := p.Body[0].(*Call)
	if len(c.Args) != 3 {
		t.Fatalf("got %d args, want 3", len(c.Args))
	}
}

func TestPrecedence(t *testing.T) {
	// 1 + 2 ÷ 3 = 4 は ((1 + (2 ÷ 3)) = 4)
	p := mustParse(t, "print(1 + 2 ÷ 3 = 4)")
	expr := p.Body[0].(*Call).Args[0].Items[0].(*BinaryOp)
	if expr.Op != "=" {
		t.Fatalf("top op %q, want =", expr.Op)
	}
	add := expr.Left.(*BinaryOp)
	if add.Op != "+" {
		t.Fatalf("left op %q, want +", add.Op)
	}
	if div := add.Right.(*BinaryOp); div.Op != "÷" {
		t.Fatalf("right op %q, want ÷", div.Op)
	}
}

func TestSubtractionIsLeftAssociative(t *testing.T) {
	p := mustParse(t, "print(10 - 3 - 2)")
	top := p.Body[0].(*Call).Args[0].Items[0].(*BinaryOp)
	if _, ok := top.Left.(*BinaryOp); !ok {
		t.Fatalf("expected (10-3)-2, got %#v", top)
	}
}

func TestLiterals(t *testing.T) {
	p := mustParse(t, `print(list[1, "a", TRUE, null, list[2], cd[print(0)]])`)
	lit := p.Body[0].(*Call).Args[0].Items[0].(*ListLit)
	if len(lit.Elems) != 6 {
		t.Fatalf("got %d elements, want 6", len(lit.Elems))
	}
	if cd := lit.Elems[5].(*CodeLit); cd.Src != "print(0)" || len(cd.Body) != 1 {
		t.Errorf("cd literal: %#v", cd)
	}
}

func TestEmptyListAndCode(t *testing.T) {
	mustParse(t, "print(list[])")
	mustParse(t, "set(\"c\") to(cd[])")
}

func TestCommentsAreRemoved(t *testing.T) {
	p := mustParse(t, `c("hello") print(1) c(whatever 1 list[1]) print(2)`)
	if len(p.Body) != 2 {
		t.Fatalf("got %d statements, want 2", len(p.Body))
	}
}

func TestCommentsRemovedInsideCodeAndArgs(t *testing.T) {
	p := mustParse(t, `if(TRUE)(c("x") print(1) c("y"))`)
	arg := p.Body[0].(*Call).Args[1]
	if len(arg.Items) != 1 {
		t.Fatalf("code arg has %d items, want 1", len(arg.Items))
	}
	p = mustParse(t, `set("v") to(cd[c("note") print(1)])`)
	cd := p.Body[1].(*Call).Args[0].Items[0].(*CodeLit)
	if len(cd.Body) != 1 {
		t.Fatalf("cd body has %d statements, want 1", len(cd.Body))
	}
}

func TestVariableNamedLikeComment(t *testing.T) {
	// c が括弧を伴わなければ、ただの変数参照。
	p := mustParse(t, `if(c = 1)(print(c))`)
	cond := p.Body[0].(*Call).Args[0].Items[0].(*BinaryOp)
	if v, ok := cond.Left.(*VarRef); !ok || v.Name != "c" {
		t.Fatalf("got %#v", cond.Left)
	}
}

func TestSourceTextOfArgs(t *testing.T) {
	p := mustParse(t, `if(TRUE)(  print(1)  print(2) )`)
	if got := p.Body[0].(*Call).Args[1].Src; got != "print(1)  print(2)" {
		t.Errorf("src %q", got)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct{ name, src, id string }{
		{"stray close paren", "print(1))", "4/4/)"},
		{"unclosed paren", "print(1", "4/5"},
		{"unclosed cd", "set(\"x\") to(cd[print(1)", "4/5"},
		{"unclosed list", "print(list[1, 2", "4/5"},
		{"comma outside list", "print(1, 2)", "4/4/,"},
		{"non-call statement", "5", "4/7"},
		{"non-call in cd", "set(\"x\") to(cd[1 2])", "4/7"},
		{"comment in expression", "print(1 + c(2))", "4/6"},
		{"comment as list element", "print(list[c(1)])", "4/6"},
		{"operator without operand", "print(1 +)", "4/4/)"},
	}
	for _, c := range cases {
		if got := errID(t, c.src); got != c.id {
			t.Errorf("%s: id %q, want %q", c.name, got, c.id)
		}
	}
}
