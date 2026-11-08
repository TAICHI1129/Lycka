package types

import (
	"strings"

	"lycka/internal/parser"
)

// Code は cd 型。実行可能な文(関数呼び出し)の列を値として保持する。
type Code struct {
	Body []parser.Node
	// Src は cd[...] の括弧内の原文。表示用。
	Src string
}

func (c *Code) Type() string    { return TypeCode }
func (c *Code) String() string  { return c.Inspect() }
func (c *Code) Inspect() string { return "cd[" + c.Src + "]" }

// ConcatCode は `+`: a のコードの後に b のコードを続けた新しいコード。
func ConcatCode(a, b *Code) *Code {
	body := make([]parser.Node, 0, len(a.Body)+len(b.Body))
	body = append(body, a.Body...)
	body = append(body, b.Body...)
	return &Code{Body: body, Src: strings.TrimSpace(a.Src + " " + b.Src)}
}
