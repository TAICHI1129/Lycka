package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, strings.NewReader(stdin), &out, &errOut)
	return code, out.String(), errOut.String()
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "prog.ly")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunFile(t *testing.T) {
	code, out, errOut := runCLI(t, "", writeTemp(t, `print("hi")`))
	if code != 0 || out != "hi\n" || errOut != "" {
		t.Errorf("code %d out %q err %q", code, out, errOut)
	}
}

func TestRunStdin(t *testing.T) {
	code, out, _ := runCLI(t, `print(1 + 1)`, "-")
	if code != 0 || out != "2\n" {
		t.Errorf("code %d out %q", code, out)
	}
}

func TestBOMIsIgnored(t *testing.T) {
	code, out, errOut := runCLI(t, "", writeTemp(t, utf8BOM+`print("bom")`))
	if code != 0 || out != "bom\n" {
		t.Errorf("code %d out %q err %q", code, out, errOut)
	}
}

func TestRuntimeErrorGoesToStderrAfterOutput(t *testing.T) {
	code, out, errOut := runCLI(t, "", writeTemp(t, `print(1) print(nope)`))
	if code != 1 {
		t.Errorf("code %d, want 1", code)
	}
	if out != "1\n" {
		t.Errorf("stdout %q", out)
	}
	if !strings.HasPrefix(errOut, "ERROR: Violated The Reference Rule (3)") {
		t.Errorf("stderr %q", errOut)
	}
}

func TestSyntaxError(t *testing.T) {
	code, out, errOut := runCLI(t, "", writeTemp(t, `print("unterminated`))
	if code != 1 || out != "" || !strings.HasPrefix(errOut, "ERROR: Syntax Error (4)") {
		t.Errorf("code %d out %q err %q", code, out, errOut)
	}
}

func TestMissingFile(t *testing.T) {
	code, _, errOut := runCLI(t, "", filepath.Join(t.TempDir(), "missing.ly"))
	if code != 1 || !strings.HasPrefix(errOut, "ERROR: Input Output Error (6)") {
		t.Errorf("code %d err %q", code, errOut)
	}
}

func TestUsage(t *testing.T) {
	if code, _, errOut := runCLI(t, ""); code != 2 || !strings.Contains(errOut, "Usage") {
		t.Errorf("no args: code %d err %q", code, errOut)
	}
	if code, _, _ := runCLI(t, "", "a.ly", "b.ly"); code != 2 {
		t.Errorf("two args: code %d", code)
	}
	if code, out, _ := runCLI(t, "", "--help"); code != 0 || !strings.Contains(out, "Usage") {
		t.Errorf("--help: code %d out %q", code, out)
	}
}
