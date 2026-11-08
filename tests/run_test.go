// Package tests は .ly ファイルをインタプリタで実際に実行する統合テスト。
package tests

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"lycka/internal/functions"
)

func runFile(t *testing.T, path string) string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := functions.NewInterpreter(&buf).Run(string(src)); err != nil {
		t.Fatalf("%s failed:\n%v\noutput so far:\n%s", path, err, buf.String())
	}
	return buf.String()
}

// test.ly の各チェックは「ok ...」か「FAIL ...」を1行表示する。
// FAIL が無く、ソース中の `print("ok ...")` の数と同じだけ ok が表示されることを確認する。
func TestSelfCheckingProgram(t *testing.T) {
	out := runFile(t, "test.ly")

	src, err := os.ReadFile("test.ly")
	if err != nil {
		t.Fatal(err)
	}
	wantOK := len(regexp.MustCompile(`print\("ok `).FindAll(src, -1))

	gotOK := 0
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		switch {
		case strings.HasPrefix(line, "FAIL"):
			t.Errorf("check failed: %s", line)
		case strings.HasPrefix(line, "ok"):
			gotOK++
		default:
			t.Errorf("unexpected output line: %q", line)
		}
	}
	if gotOK != wantOK {
		t.Errorf("%d checks passed, want %d (some branch did not run?)", gotOK, wantOK)
	}
}

func TestExampleHello(t *testing.T) {
	if got, want := runFile(t, "../examples/hello.ly"), "Hello, World!\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExampleVariables(t *testing.T) {
	want := strings.Join([]string{
		"10",
		"15",
		"こんにちは",
		"list[1, 2, 3, 4]",
		"lycka-lang",
		"y は 15 です",
		"3", "2", "1",
		"TRUE",  // isitsth(s)
		"TRUE",  // isitsth(s = s)
		"FALSE", // isitsth(1)
		"3",
	}, "\n") + "\n"
	if got := runFile(t, "../examples/variables.ly"); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
