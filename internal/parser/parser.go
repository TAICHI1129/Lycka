package parser

import (
	"math/big"
	"strings"

	lyerr "lycka/internal/errors"
	"lycka/internal/lexer"
)

// CommentFunc はコメント関数 c() の名前。
// 仕様書 4.7: 実行前に当該関数が削除される。パーサが構文解析直後に取り除く。
const CommentFunc = "c"

// Parser は構文解析器。
type Parser struct {
	src  string
	toks []lexer.Token
	pos  int
	// pendingComments は「文の位置で使われた」と確認できていない c() 呼び出し。
	// 解析終了時に残っていれば、式の途中で使われた不正な c() として報告する。
	pendingComments map[*Call]struct{}
}

// Parse はソースコードを構文解析して Program を返す。
// 失敗時のエラーは *errors.Error(位置情報付き)。
func Parse(src string) (*Program, error) {
	toks, err := lexer.Tokenize(src)
	if err != nil {
		return nil, err
	}
	p := &Parser{src: src, toks: toks, pendingComments: map[*Call]struct{}{}}
	return p.parseProgram()
}

func (p *Parser) cur() lexer.Token { return p.toks[p.pos] }

func (p *Parser) next() lexer.Token {
	t := p.toks[p.pos]
	if t.Kind != lexer.EOF {
		p.pos++
	}
	return t
}

func tokPos(t lexer.Token) Pos { return Pos{Line: t.Line, Col: t.Col} }

func (p *Parser) unexpected(t lexer.Token) error {
	if t.Kind == lexer.EOF {
		return lyerr.UnexpectedEOF().WithPos(t.Line, t.Col)
	}
	return lyerr.UnexpectedToken(p.src[t.Start:t.End]).WithPos(t.Line, t.Col)
}

func (p *Parser) expect(k lexer.Kind) (lexer.Token, error) {
	t := p.cur()
	if t.Kind != k {
		return t, p.unexpected(t)
	}
	return p.next(), nil
}

// ---- プログラム・文 ----

func (p *Parser) parseProgram() (*Program, error) {
	var items []Node
	for p.cur().Kind != lexer.EOF {
		n, err := p.parseItem()
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if err := p.checkComments(); err != nil {
		return nil, err
	}
	body, err := finishStatements(items)
	if err != nil {
		return nil, err
	}
	return &Program{Body: body}, nil
}

// parseItem は括弧やプログラム直下に並ぶ要素(式)を1つ読む。
// 要素そのものが c() 呼び出しなら、文の位置での正当な使用として記録から外す。
func (p *Parser) parseItem() (Node, error) {
	n, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if c, ok := n.(*Call); ok && c.Name == CommentFunc {
		delete(p.pendingComments, c)
	}
	return n, nil
}

func (p *Parser) checkComments() error {
	var first *Call
	for c := range p.pendingComments {
		if first == nil || before(c.P, first.P) {
			first = c
		}
	}
	if first != nil {
		return lyerr.CommentInExpression().WithPos(first.P.Line, first.P.Col)
	}
	return nil
}

func before(a, b Pos) bool {
	return a.Line < b.Line || (a.Line == b.Line && a.Col < b.Col)
}

func isComment(n Node) bool {
	c, ok := n.(*Call)
	return ok && c.Name == CommentFunc
}

// stripComments は要素列から c() を取り除く。
func stripComments(items []Node) []Node {
	out := items[:0:0]
	for _, n := range items {
		if !isComment(n) {
			out = append(out, n)
		}
	}
	return out
}

// finishStatements は c() を除去し、残りがすべて関数呼び出しであることを確認する。
func finishStatements(items []Node) ([]Node, error) {
	body := stripComments(items)
	for _, n := range body {
		if _, ok := n.(*Call); !ok {
			pos := n.Position()
			return nil, lyerr.NotAFunctionCall().WithPos(pos.Line, pos.Col)
		}
	}
	return body, nil
}

// ---- 式 ----
//
//	expr    := add ('=' add)*
//	add     := mul (('+' | '-') mul)*
//	mul     := unary ('/' unary)*
//	unary   := '-' unary | primary
//	primary := INT | STRING | '(' expr ')' | TRUE | FALSE | null
//	         | list '[' [expr (',' expr)*] ']'
//	         | cd '[' item* ']'
//	         | IDENT ( '(' item* ')' )+     関数呼び出し
//	         | IDENT                         変数参照

func (p *Parser) parseExpr() (Node, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	for p.cur().Kind == lexer.EQ {
		op := p.next()
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{base: base{tokPos(op)}, Op: "=", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAdd() (Node, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.cur().Kind == lexer.PLUS || p.cur().Kind == lexer.MINUS {
		op := p.next()
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{base: base{tokPos(op)}, Op: op.Text, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMul() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur().Kind == lexer.DIV {
		op := p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{base: base{tokPos(op)}, Op: "/", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.cur().Kind == lexer.MINUS {
		op := p.next()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryMinus{base: base{tokPos(op)}, Operand: operand}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Node, error) {
	t := p.cur()
	switch t.Kind {
	case lexer.INT:
		p.next()
		v, ok := new(big.Int).SetString(t.Text, 10)
		if !ok {
			return nil, lyerr.Internal("invalid integer token "+t.Text).WithPos(t.Line, t.Col)
		}
		return &IntLit{base: base{tokPos(t)}, Value: v}, nil

	case lexer.STRING:
		p.next()
		return &StrLit{base: base{tokPos(t)}, Value: t.Text}, nil

	case lexer.LPAREN:
		// 演算の優先順位を変えるための括弧。
		p.next()
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.RPAREN); err != nil {
			return nil, err
		}
		return inner, nil

	case lexer.IDENT:
		return p.parseIdent()
	}
	return nil, p.unexpected(t)
}

func (p *Parser) parseIdent() (Node, error) {
	t := p.next()
	pos := base{tokPos(t)}
	switch t.Text {
	case "TRUE":
		return &BoolLit{base: pos, Value: true}, nil
	case "FALSE":
		return &BoolLit{base: pos, Value: false}, nil
	case "null":
		return &NullLit{base: pos}, nil
	case "list":
		if p.cur().Kind == lexer.LBRACKET {
			return p.parseListLit(pos)
		}
	case "cd":
		if p.cur().Kind == lexer.LBRACKET {
			return p.parseCodeLit(pos)
		}
	}

	if p.cur().Kind == lexer.LPAREN {
		call := &Call{base: pos, Name: t.Text}
		for p.cur().Kind == lexer.LPAREN {
			arg, err := p.parseArg()
			if err != nil {
				return nil, err
			}
			call.Args = append(call.Args, arg)
		}
		if call.Name == CommentFunc {
			p.pendingComments[call] = struct{}{}
		}
		return call, nil
	}
	return &VarRef{base: pos, Name: t.Text}, nil
}

// parseArg は '(' ... ')' を1組読む。
func (p *Parser) parseArg() (*Arg, error) {
	lp := p.next()
	var items []Node
	for {
		t := p.cur()
		if t.Kind == lexer.RPAREN {
			break
		}
		if t.Kind == lexer.EOF {
			return nil, p.unexpected(t)
		}
		n, err := p.parseItem()
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	rp := p.next()
	src := strings.TrimSpace(p.src[lp.End:rp.Start])
	return &Arg{Pos: tokPos(lp), Items: stripComments(items), Src: src}, nil
}

func (p *Parser) parseListLit(pos base) (Node, error) {
	p.next() // '['
	lit := &ListLit{base: pos}
	if p.cur().Kind == lexer.RBRACKET {
		p.next()
		return lit, nil
	}
	for {
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		lit.Elems = append(lit.Elems, e)
		switch p.cur().Kind {
		case lexer.COMMA:
			p.next()
		case lexer.RBRACKET:
			p.next()
			return lit, nil
		default:
			return nil, p.unexpected(p.cur())
		}
	}
}

func (p *Parser) parseCodeLit(pos base) (Node, error) {
	lb := p.next() // '['
	var items []Node
	for {
		t := p.cur()
		if t.Kind == lexer.RBRACKET {
			break
		}
		if t.Kind == lexer.EOF {
			return nil, p.unexpected(t)
		}
		n, err := p.parseItem()
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	rb := p.next()
	body, err := finishStatements(items)
	if err != nil {
		return nil, err
	}
	src := strings.TrimSpace(p.src[lb.End:rb.Start])
	return &CodeLit{base: pos, Body: body, Src: src}, nil
}
