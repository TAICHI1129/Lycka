package functions

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	lyerr "lycka/internal/errors"
)

// run はプログラムを実行して標準出力とエラー番号(エラーが無ければ "")を返す。
func run(t *testing.T, src string) (out string, errID string) {
	t.Helper()
	var buf bytes.Buffer
	err := NewInterpreter(&buf).Run(src)
	if err == nil {
		return buf.String(), ""
	}
	var le *lyerr.Error
	if !errors.As(err, &le) {
		t.Fatalf("expected *errors.Error, got %T: %v", err, err)
	}
	return buf.String(), le.ID()
}

func TestPrograms(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"hello", `print("Hello, World!")`, "Hello, World!\n"},
		{"print int", `print(42)`, "42\n"},
		{"print negative", `print(-5)`, "-5\n"},
		{"print bl", `print(TRUE) print(FALSE)`, "TRUE\nFALSE\n"},
		{"print list quotes strings", `print(list[1, "a", TRUE, null])`, "list[1, \"a\", TRUE, null]\n"},
		{"print nested list", `print(list[list[1], list[]])`, "list[list[1], list[]]\n"},
		{"print str without quotes", `print('single')`, "single\n"},
		{"print escapes", `print("a\tb")`, "a\tb\n"},

		{"set to", `set("x") to(1) print(x)`, "1\n"},
		{"set to on separate lines", "set(\"x\")\nto(7)\nprint(x)", "7\n"},
		{"set keeps existing", `set("x") to(1) set("x") print(x)`, "1\n"},
		{"to after repeated set", `set("x") to(1) set("x") to(2) print(x)`, "2\n"},
		{"set without to is null", `set("x") print(x = null)`, "TRUE\n"},
		{"to without set does nothing", `to(5) print(1)`, "1\n"},
		{"to without set does not evaluate", `to(undefinedvar) print(1)`, "1\n"},
		{"to targets latest set", `set("a") to(1) set("b") to(2) to(3) print(a) print(b)`, "1\n3\n"},
		{"to with an expression", `set("x") to(2) set("y") to(x + x + x) print(y)`, "6\n"},

		{"int ops", `print(1 + 2) print(5 - 7) print(7 ÷ 2) print(-7 ÷ 2)`, "3\n-2\n3\n-3\n"},
		{"precedence", `print(1 + 4 ÷ 2) print((1 + 4) ÷ 5) print(10 - 3 - 2)`, "3\n1\n5\n"},
		{"big ints", `print(99999999999999999999 + 1)`, "100000000000000000000\n"},
		{"unary minus twice", `print(- -3)`, "3\n"},
		{"str ops", `print("ab" + "cd") print("abcabc" - "bc")`, "abcd\naa\n"},
		{"list ops", `print(list[1, 2, 2] + list[2, 3]) print(list[1, 2, 3, 2] - list[2])`, "list[1, 2, 3]\nlist[1, 3]\n"},
		{"null ops", `print(null + null = null) print(null - null = null)`, "TRUE\nTRUE\n"},

		{"if true", `if(TRUE)(print("yes"))`, "yes\n"},
		{"if false", `if(FALSE)(print("yes"))`, ""},
		{"if else true", `if(1 = 1)(print("a")) else(print("b"))`, "a\n"},
		{"if else false", `if(1 = 2)(print("a")) else(print("b"))`, "b\n"},
		{"else without if", `else(print("x")) print("end")`, "end\n"},
		{"else after true if", `if(TRUE)(print("a")) else(print("b"))`, "a\n"},
		{"else across unrelated function", `if(FALSE)(print("a")) print("mid") else(print("b"))`, "mid\nb\n"},
		{"else uses most recent if", `if(TRUE)(print("a")) if(FALSE)(print("b")) else(print("c"))`, "a\nc\n"},
		{"else pairs with outer if", `if(TRUE)(if(FALSE)(print("x"))) else(print("no"))`, ""},
		{"else after outer false if", `if(FALSE)(if(TRUE)(print("x"))) else(print("outer else"))`, "outer else\n"},
		{"multi statement body", `if(TRUE)(print(1) print(2) print(3))`, "1\n2\n3\n"},
		{"explicit cd arg", `if(TRUE)(cd[print("x")])`, "x\n"},
		{"empty code body", `if(TRUE)() print("ok")`, "ok\n"},

		{"while counts", `set("i") to(3) while(i = 0 = FALSE)(print(i) set("i") to(i - 1))`, "3\n2\n1\n"},
		{"while false", `while(FALSE)(print("x")) print("done")`, "done\n"},
		{"while list contains", `set("i") to(0) while(list[0, 1] = i)(print(i) set("i") to(i + 1))`, "0\n1\n"},

		{"cd in variable", `set("c1") to(cd[print("hi")]) if(TRUE)(c1) if(TRUE)(c1)`, "hi\nhi\n"},
		{"cd concat", `set("c1") to(cd[print(1)] + cd[print(2)]) if(TRUE)(c1)`, "1\n2\n"},
		{"cd arg is lazy", `set("x") to(cd[print(undefinedvar)]) print("not run")`, "not run\n"},
		{"cd not run by to", `set("x") to(cd[print(1)]) print(2)`, "2\n"},

		{"smthng from 0 ÷ 0", `print(isitsth(0 ÷ 0))`, "TRUE\n"},
		{"smthng x = x", `set("s") to(0 ÷ 0) print(isitsth(s = s))`, "TRUE\n"},
		{"smthng is not bl", `set("s") to(0 ÷ 0) print(isitsth(s = s = s))`, "TRUE\n"},
		{"isitsth others", `print(isitsth(1)) print(isitsth("a")) print(isitsth(null)) print(isitsth(TRUE)) print(isitsth(list[0 ÷ 0])) print(isitsth(cd[print(1)]))`, "FALSE\nFALSE\nFALSE\nFALSE\nFALSE\nFALSE\n"},
		{"smthng arithmetic", `print(isitsth((0 ÷ 0) + (0 ÷ 0))) print(isitsth((0 ÷ 0) - (0 ÷ 0))) print(isitsth((0 ÷ 0) ÷ (0 ÷ 0)))`, "TRUE\nTRUE\nTRUE\n"},
		{"smthng condition runs neither branch", `set("s") to(0 ÷ 0) if(s = s)(print("t")) else(print("f")) print("end")`, "end\n"},
		{"smthng while stops", `set("s") to(0 ÷ 0) while(s = s)(print("loop")) print("end")`, "end\n"},
		{"smthng in list equality", `set("s") to(0 ÷ 0) print(isitsth(list[s] = list[s]))`, "TRUE\n"},

		{"comments removed", `c("note") print(1) c(print("never")) print(2)`, "1\n2\n"},
		{"comment takes anything", `c(1)(list[1, 2])(cd[print("never")])("x")(TRUE)(null) print("ok")`, "ok\n"},
		{"comment inside code", `if(TRUE)(c("x") print(1) c("y"))`, "1\n"},
		{"comment with undefined names", `c(foo bar) print("ok")`, "ok\n"},

		{"concur runs all", `concur(set("a") to(1))(set("b") to(2)) print(a + b)`, "3\n"},
		{"concur keeps own set/to", `concur(set("a") to(1) set("b") to(a + 1))(set("c") to(10)) print(a + b + c)`, "13\n"},
		{"concur many", `concur(set("a") to(1))(set("b") to(2))(set("c") to(3))(set("d") to(4)) print(a + b + c + d)`, "10\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, id := run(t, c.src)
			if id != "" {
				t.Fatalf("unexpected error %s\noutput so far: %q", id, out)
			}
			if out != c.want {
				t.Errorf("output %q, want %q", out, c.want)
			}
		})
	}
}

func TestErrorIDs(t *testing.T) {
	cases := []struct {
		name, src, id string
	}{
		// 仕様書 6 章の例そのもの。
		{"print cd (spec example)", `print(cd[print(0)])`, "1/1/1/cd"},
		{"print null", `print(null)`, "1/1/1/null"},
		{"print smthng", `print(0 ÷ 0)`, "1/1/1/smthng"},
		{"print with empty parens", `print()`, "1/1/3/expr"},
		{"print too many args", `print(1)(2)`, "1/1/2/many"},
		{"print arg not single expression", `print(1 2)`, "1/1/3/expr"},
		{"if condition must be bl", `if(1)(print(1))`, "1/2/1/int"},
		{"if needs two args", `if(TRUE)`, "1/2/2/few"},
		{"else too many", `else(print(1))(print(2))`, "1/3/2/many"},
		{"while condition must be bl", `while("x")(print(1))`, "1/5/1/str"},
		{"concur needs two", `concur(print(1))`, "1/6/2/few"},
		{"set needs str", `set(1)`, "1/8/1/int"},
		{"set arg count", `set("a")("b")`, "1/8/2/many"},
		{"cd param with non-cd variable", `set("x") to(1) if(TRUE)(x)`, "1/2/1/int"},

		{"add mixed", `print(1 + "a")`, "2/1/2/int,str"},
		{"add bl", `print(TRUE + TRUE)`, "2/1/1/bl"},
		{"sub bl", `print(TRUE - TRUE)`, "2/2/1/bl"},
		{"sub cd", `print(cd[print(1)] - cd[print(1)])`, "2/2/1/cd"},
		{"sub mixed", `print(1 - null)`, "2/2/2/int,null"},
		{"eq cd", `print(cd[print(1)] = cd[print(1)])`, "2/3/1/cd"},
		{"eq cd in list", `print(list[cd[print(1)]] = 1)`, "2/3/1/cd"},
		{"div by zero", `print(1 ÷ 0)`, "2/4/3"},
		{"div str", `print("a" ÷ "b")`, "2/4/1/str"},
		{"div mixed", `print(1 ÷ null)`, "2/4/2/int,null"},
		{"neg str", `print(-"a")`, "2/5/1/str"},

		{"undefined variable", `print(nope)`, "3/1/nope"},
		{"undefined function", `nope(1)`, "3/2/nope"},
		{"undefined variable as cd arg", `if(TRUE)(nope)`, "3/1/nope"},

		{"not a call at top level", `1 + 1`, "4/7"},
		{"not a call in cd arg", `if(TRUE)(1)`, "4/7"},
		{"comment in expression", `print(1 + c(2))`, "4/6"},
		{"unterminated string", `print("a`, "4/2"},
		{"unexpected token", `print(1))`, "4/4/)"},

		{"infinite cd recursion", `set("x") to(cd[if(TRUE)(x)]) if(TRUE)(x)`, "5/1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, id := run(t, c.src)
			if id != c.id {
				t.Errorf("error id %q, want %q", id, c.id)
			}
		})
	}
}

// 仕様書 6 章の例の表記と完全に一致すること(位置情報の行を除く)。
func TestErrorFormatMatchesSpecExample(t *testing.T) {
	err := NewInterpreter(&bytes.Buffer{}).Run(`print(cd[print(0)])`)
	if err == nil {
		t.Fatal("expected an error")
	}
	want := strings.Join([]string{
		`ERROR: Violated The Function Rule (1)`,
		`    Violated The "print" Function Rule (1/1)`,
		`        Types That Cannot Be Used (1/1/1)`,
		`            "cd" Type Error (1/1/1/cd)`,
	}, "\n")
	if got := err.Error(); !strings.HasPrefix(got, want) {
		t.Errorf("got:\n%s\nwant prefix:\n%s", got, want)
	}
}

func TestErrorPositions(t *testing.T) {
	err := NewInterpreter(&bytes.Buffer{}).Run("print(1)\nprint(1 + \"a\")")
	var le *lyerr.Error
	if !errors.As(err, &le) {
		t.Fatal(err)
	}
	// 位置は原因の演算子 '+' (2 行 9 列)。
	if le.Line != 2 || le.Col != 9 {
		t.Errorf("position %d:%d, want 2:9", le.Line, le.Col)
	}
}

func TestOutputBeforeErrorIsKept(t *testing.T) {
	out, id := run(t, `print(1) print(2) print(nope) print(3)`)
	if out != "1\n2\n" || id != "3/1/nope" {
		t.Errorf("out %q id %q", out, id)
	}
}

func TestStatePersistsAcrossRuns(t *testing.T) {
	var buf bytes.Buffer
	it := NewInterpreter(&buf)
	if err := it.Run(`set("x") to(5)`); err != nil {
		t.Fatal(err)
	}
	if err := it.Run(`print(x)`); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "5\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestConcurErrorIsReportedAfterAllBranchesFinish(t *testing.T) {
	var buf bytes.Buffer
	err := NewInterpreter(&buf).Run(`concur(print(nope))(print("other branch"))`)
	var le *lyerr.Error
	if !errors.As(err, &le) || le.ID() != "3/1/nope" {
		t.Fatalf("err %v", err)
	}
	if buf.String() != "other branch\n" {
		t.Errorf("the healthy branch must still finish: %q", buf.String())
	}
}

// concur の各ブランチが同時に出力しても、行が混ざらず、全行が出ること。
// -race 付きで実行するとデータ競合も検出できる。
func TestConcurOutputIsNotInterleaved(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("concur")
	const branches = 8
	for b := 0; b < branches; b++ {
		// カウンタはブランチごとに別名にする(同じ変数を複数ブランチで更新すると
		// 読み取り→書き込みが競合して、終了条件を飛び越えうるため)。
		v := fmt.Sprintf("i%d", b)
		fmt.Fprintf(&sb, `(set("%[1]s") to(0) while(%[1]s = 50 = FALSE)(print("line") set("%[1]s") to(%[1]s + 1)))`, v)
	}
	out, id := run(t, sb.String())
	if id != "" {
		t.Fatalf("error %s", id)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != branches*50 {
		t.Fatalf("got %d lines, want %d", len(lines), branches*50)
	}
	for _, ln := range lines {
		if ln != "line" {
			t.Fatalf("corrupted line %q", ln)
		}
	}
}

// concur のブランチ同士が同じ変数を触っても落ちない(競合しても値は壊れない)。
func TestConcurSharedVariableAccess(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`set("shared") to(0) concur`)
	for b := 0; b < 6; b++ {
		v := fmt.Sprintf("n%d", b)
		fmt.Fprintf(&sb, `(set("%[1]s") to(0) while(%[1]s = 30 = FALSE)(set("shared") to(%[1]s) set("%[1]s") to(%[1]s + 1)))`, v)
	}
	sb.WriteString(` print(isitsth(shared))`)
	out, id := run(t, sb.String())
	if id != "" || out != "FALSE\n" {
		t.Errorf("out %q id %q", out, id)
	}
}

func TestConcurBranchesOverlapInTime(t *testing.T) {
	// 実際に並行実行されていること: 2つのブランチを順番にしか実行できない実装だと、
	// 先に終わるはずのブランチの出力が後になる。ここでは出力の集合だけを確認する。
	out, id := run(t, `concur(print("a") print("b"))(print("c") print("d"))`)
	if id != "" {
		t.Fatalf("error %s", id)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	sort.Strings(lines)
	if strings.Join(lines, ",") != "a,b,c,d" {
		t.Errorf("got %v", lines)
	}
	// 各ブランチ内の順序は保たれる。
	if strings.Index(out, "a\n") > strings.Index(out, "b\n") || strings.Index(out, "c\n") > strings.Index(out, "d\n") {
		t.Errorf("order inside a branch broken: %q", out)
	}
}

func TestRegistryContainsAllSpecFunctions(t *testing.T) {
	it := NewInterpreter(&bytes.Buffer{})
	for _, name := range []string{"print", "if", "else", "isitsth", "while", "concur", "c", "set", "to"} {
		if _, ok := it.Registry().Lookup(name); !ok {
			t.Errorf("function %q is not registered", name)
		}
	}
	if got := len(it.Registry().Names()); got != 9 {
		t.Errorf("registered %d functions, want 9", got)
	}
}
