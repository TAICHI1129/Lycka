// Package lexer は Lycka のソースコードをトークン列に分解する。
package lexer

// Kind はトークンの種類。
type Kind int

const (
	EOF Kind = iota
	IDENT
	INT
	STRING
	PLUS     // +
	MINUS    // -
	EQ       // =
	DIV      // ÷
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	COMMA    // ,
)

var kindNames = [...]string{
	EOF: "EOF", IDENT: "IDENT", INT: "INT", STRING: "STRING",
	PLUS: "+", MINUS: "-", EQ: "=", DIV: "/",
	LPAREN: "(", RPAREN: ")", LBRACKET: "[", RBRACKET: "]", COMMA: ",",
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "?"
}

// Token は1つの字句。
type Token struct {
	Kind Kind
	// Text は IDENT なら名前、INT なら10進数字列、STRING ならエスケープ解除後の内容。
	Text string
	Line int // 1始まり
	Col  int // 1始まり(文字数単位)
	// Start, End はソース中のバイト範囲 [Start, End)。
	Start, End int
}
