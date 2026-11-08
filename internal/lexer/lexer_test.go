package lexer

import (
	"errors"
	"testing"

	lyerr "lycka/internal/errors"
)

func kinds(toks []Token) []Kind {
	out := make([]Kind, len(toks))
	for i, t := range toks {
		out[i] = t.Kind
	}
	return out
}

func TestTokenizeBasic(t *testing.T) {
	toks, err := Tokenize(`set("x") to(1 + 2 ÷ 3 - 4 = x)`)
	if err != nil {
		t.Fatal(err)
	}
	want := []Kind{
		IDENT, LPAREN, STRING, RPAREN,
		IDENT, LPAREN, INT, PLUS, INT, DIV, INT, MINUS, INT, EQ, IDENT, RPAREN,
		EOF,
	}
	got := kinds(toks)
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestStringsAndEscapes(t *testing.T) {
	cases := []struct{ src, want string }{
		{`"AB"`, "AB"},
		{`'CD'`, "CD"},
		{`"a\nb"`, "a\nb"},
		{`"tab\there"`, "tab\there"},
		{`"q\"q"`, `q"q`},
		{`'q\'q'`, `q'q`},
		{`"back\\slash"`, `back\slash`},
		{`"日本語 ÷ 😀"`, "日本語 ÷ 😀"},
		{`"it's"`, "it's"},
	}
	for _, c := range cases {
		toks, err := Tokenize(c.src)
		if err != nil {
			t.Errorf("%s: %v", c.src, err)
			continue
		}
		if toks[0].Kind != STRING || toks[0].Text != c.want {
			t.Errorf("%s: got %q, want %q", c.src, toks[0].Text, c.want)
		}
	}
}

func TestPositionsAndOffsets(t *testing.T) {
	src := "print(1)\n  print(22)"
	toks, err := Tokenize(src)
	if err != nil {
		t.Fatal(err)
	}
	// 2行目の print は 2 行 3 列。
	var second Token
	for _, tk := range toks {
		if tk.Kind == IDENT && tk.Line == 2 {
			second = tk
		}
	}
	if second.Line != 2 || second.Col != 3 {
		t.Errorf("got %d:%d, want 2:3", second.Line, second.Col)
	}
	if src[second.Start:second.End] != "print" {
		t.Errorf("offsets slice %q", src[second.Start:second.End])
	}
}

func TestFullWidthSpaceIsWhitespace(t *testing.T) {
	toks, err := Tokenize("print(1)　print(2)")
	if err != nil {
		t.Fatal(err)
	}
	if len(toks) != 9 { // print ( 1 ) print ( 2 ) EOF
		t.Errorf("got %d tokens", len(toks))
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		name, src, id string
	}{
		{"unexpected char", "print(1) @", "4/1/@"},
		{"arrow is not syntax", "x <- null", "4/1/<"},
		{"unterminated", `print("abc`, "4/2"},
		{"newline in string", "print(\"ab\ncd\")", "4/2"},
		{"bad escape", `print("a\qb")`, `4/3/\q`},
		{"digit then letter", "print(12abc)", "4/1/a"},
	}
	for _, c := range cases {
		_, err := Tokenize(c.src)
		var le *lyerr.Error
		if !errors.As(err, &le) {
			t.Errorf("%s: expected *errors.Error, got %v", c.name, err)
			continue
		}
		if le.ID() != c.id {
			t.Errorf("%s: id %q, want %q", c.name, le.ID(), c.id)
		}
		if !le.HasPos() {
			t.Errorf("%s: missing position", c.name)
		}
	}
}
