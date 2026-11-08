// Package errors は Lycka の階層構造エラーを定義する。
//
// 表記形式は仕様書「6. エラー」に従う:
//
//	ERROR: Violated The Function Rule (1)
//	    Violated The "print" Function Rule (1/1)
//	        Types That Cannot Be Used (1/1/1)
//	            "cd" Type Error (1/1/1/cd)
//
// エラー番号は仕様上「仮のもの」であり、ここでの採番も暫定である。
package errors

import (
	"fmt"
	"strings"
)

// Level はエラー階層の1段。
type Level struct {
	Name string // 表示名
	Seg  string // この段のエラー番号(上位との連結で完全な番号になる)
}

// Error は階層構造を持つ Lycka のエラー。
type Error struct {
	Levels []Level
	Line   int // 0 のときは位置不明
	Col    int
}

// Error は仕様書形式の複数行文字列を返す。
func (e *Error) Error() string {
	var b strings.Builder
	var path []string
	for i, lv := range e.Levels {
		path = append(path, lv.Seg)
		id := strings.Join(path, "/")
		if i == 0 {
			fmt.Fprintf(&b, "ERROR: %s (%s)", lv.Name, id)
			continue
		}
		fmt.Fprintf(&b, "\n%s%s (%s)", strings.Repeat("    ", i), lv.Name, id)
	}
	if e.Line > 0 {
		fmt.Fprintf(&b, "\n    at line %d, column %d", e.Line, e.Col)
	}
	return b.String()
}

// ID は最下層まで連結したエラー番号(例: "1/1/1/cd")を返す。
func (e *Error) ID() string {
	segs := make([]string, len(e.Levels))
	for i, lv := range e.Levels {
		segs[i] = lv.Seg
	}
	return strings.Join(segs, "/")
}

// HasPos は位置情報が設定済みかどうかを返す。
func (e *Error) HasPos() bool { return e.Line > 0 }

// WithPos は位置が未設定の場合のみ設定する(最も内側の位置が優先される)。
func (e *Error) WithPos(line, col int) *Error {
	if e.Line == 0 && line > 0 {
		e.Line, e.Col = line, col
	}
	return e
}

// 最上位カテゴリ番号。
const (
	catFunction  = "1"
	catOperator  = "2"
	catReference = "3"
	catSyntax    = "4"
	catLimit     = "5"
	catIO        = "6"
	catInternal  = "9"
)

// 関数ごとの番号。仕様書 4.1〜4.9 の順序に一致させている
// (仕様書の例で print が 1/1 であるため)。
var functionIDs = map[string]string{
	"print":   "1",
	"if":      "2",
	"else":    "3",
	"isitsth": "4",
	"while":   "5",
	"concur":  "6",
	"c":       "7",
	"set":     "8",
	"to":      "9",
}

type opInfo struct{ display, id string }

var operatorInfos = map[string]opInfo{
	"+":   {"+", "1"},
	"-":   {"-", "2"},
	"=":   {"=", "3"},
	"÷":   {"÷", "4"},
	"neg": {"-(unary)", "5"},
}

func lv(name, seg string) Level { return Level{Name: name, Seg: seg} }

func functionLevels(fn string) []Level {
	id, ok := functionIDs[fn]
	if !ok {
		id = "0"
	}
	return []Level{
		lv("Violated The Function Rule", catFunction),
		lv(fmt.Sprintf("Violated The %q Function Rule", fn), id),
	}
}

func operatorLevels(op string) []Level {
	info, ok := operatorInfos[op]
	if !ok {
		info = opInfo{op, "0"}
	}
	return []Level{
		lv("Violated The Operator Rule", catOperator),
		lv(fmt.Sprintf("Violated The %q Operator Rule", info.display), info.id),
	}
}

// ---- 関数規則違反 (1/...) ----

// FunctionType は関数に使用できないデータ型が渡されたことを表す。
func FunctionType(fn, typ string) *Error {
	return &Error{Levels: append(functionLevels(fn),
		lv("Types That Cannot Be Used", "1"),
		lv(fmt.Sprintf("%q Type Error", typ), typ))}
}

// FunctionArgCount は引数の数が不正であることを表す。
func FunctionArgCount(fn string, tooMany bool) *Error {
	name, seg := "Too Few Arguments", "few"
	if tooMany {
		name, seg = "Too Many Arguments", "many"
	}
	return &Error{Levels: append(functionLevels(fn),
		lv("Invalid Number Of Arguments", "2"),
		lv(name, seg))}
}

// FunctionArgForm は値として評価される引数が単一の式でないことを表す。
func FunctionArgForm(fn string) *Error {
	return &Error{Levels: append(functionLevels(fn),
		lv("Invalid Argument Form", "3"),
		lv("Expected A Single Expression", "expr"))}
}

// ---- 演算規則違反 (2/...) ----

// OperatorType は演算子に使用できないデータ型が渡されたことを表す。
func OperatorType(op, typ string) *Error {
	return &Error{Levels: append(operatorLevels(op),
		lv("Types That Cannot Be Used", "1"),
		lv(fmt.Sprintf("%q Type Error", typ), typ))}
}

// OperatorMixed は異なるデータ型が同一の演算式に使われたことを表す。
func OperatorMixed(op, a, b string) *Error {
	return &Error{Levels: append(operatorLevels(op),
		lv("Mixed Data Types", "2"),
		lv(fmt.Sprintf("%q And %q", a, b), a+","+b))}
}

// DivisionByZero は 0 以外を 0 で割ったことを表す。
func DivisionByZero() *Error {
	return &Error{Levels: append(operatorLevels("÷"),
		lv("Division By Zero", "3"))}
}

// ---- 参照規則違反 (3/...) ----

// UndefinedVariable は未定義の変数を参照したことを表す。
func UndefinedVariable(name string) *Error {
	return &Error{Levels: []Level{
		lv("Violated The Reference Rule", catReference),
		lv("Undefined Variable", "1"),
		lv(fmt.Sprintf("%q", name), name)}}
}

// UndefinedFunction は未定義の関数を呼び出したことを表す。
func UndefinedFunction(name string) *Error {
	return &Error{Levels: []Level{
		lv("Violated The Reference Rule", catReference),
		lv("Undefined Function", "2"),
		lv(fmt.Sprintf("%q", name), name)}}
}

// ---- 構文エラー (4/...) ----

func syntax(sub Level, rest ...Level) *Error {
	return &Error{Levels: append([]Level{lv("Syntax Error", catSyntax), sub}, rest...)}
}

// UnexpectedChar は字句解析で認識できない文字を検出したことを表す。
func UnexpectedChar(ch string) *Error {
	return syntax(lv("Unexpected Character", "1"), lv(fmt.Sprintf("%q", ch), ch))
}

// UnterminatedString は文字列が閉じられていないことを表す。
func UnterminatedString() *Error {
	return syntax(lv("Unterminated String", "2"))
}

// InvalidEscape は文字列中の不正なエスケープシーケンスを表す。
func InvalidEscape(seq string) *Error {
	return syntax(lv("Invalid Escape Sequence", "3"), lv(fmt.Sprintf("%q", seq), seq))
}

// UnexpectedToken は文法上ここに現れてはならないトークンを表す。
func UnexpectedToken(tok string) *Error {
	return syntax(lv("Unexpected Token", "4"), lv(fmt.Sprintf("%q", tok), tok))
}

// UnexpectedEOF は入力が途中で終わったことを表す。
func UnexpectedEOF() *Error {
	return syntax(lv("Unexpected End Of Input", "5"))
}

// CommentInExpression は c() が文の位置以外で使われたことを表す。
func CommentInExpression() *Error {
	return syntax(lv("Comment Function Used Inside An Expression", "6"))
}

// NotAFunctionCall はコードの各文が関数呼び出しでないことを表す。
func NotAFunctionCall() *Error {
	return syntax(lv("Statement Is Not A Function Call", "7"))
}

// ---- 実行上限 (5/...) ----

// RecursionLimit はコードの入れ子実行が上限を超えたことを表す。
func RecursionLimit() *Error {
	return &Error{Levels: []Level{
		lv("Execution Limit Exceeded", catLimit),
		lv("Recursion Limit Exceeded", "1")}}
}

// ---- 入出力 (6/...) ----

// CannotReadFile はソースファイルを読めないことを表す。
func CannotReadFile(path string, cause error) *Error {
	return &Error{Levels: []Level{
		lv("Input Output Error", catIO),
		lv("Cannot Read Source", "1"),
		lv(fmt.Sprintf("%q: %v", path, cause), "file")}}
}

// OutputFailed は標準出力への書き込みに失敗したことを表す。
func OutputFailed(cause error) *Error {
	return &Error{Levels: []Level{
		lv("Input Output Error", catIO),
		lv("Cannot Write Output", "2"),
		lv(fmt.Sprint(cause), "write")}}
}

// ---- 内部エラー (9/...) ----

// Internal はインタプリタ自身の不具合・想定外の状態を表す。
func Internal(msg string) *Error {
	return &Error{Levels: []Level{
		lv("Internal Error", catInternal),
		lv(msg, "1")}}
}
