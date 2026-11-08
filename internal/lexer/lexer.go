package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	lyerr "lycka/internal/errors"
)

// Lexer はソース文字列を先頭から読み進める。
type Lexer struct {
	src  string
	pos  int
	line int
	col  int
}

// New は src を対象とする Lexer を返す。
func New(src string) *Lexer {
	return &Lexer{src: src, line: 1, col: 1}
}

// Tokenize は src 全体をトークン列にする。末尾には必ず EOF が付く。
// 失敗時のエラーは *errors.Error(位置情報付き)。
func Tokenize(src string) ([]Token, error) {
	return New(src).Tokenize()
}

// Tokenize は残りの入力をすべてトークン化する。
func (l *Lexer) Tokenize() ([]Token, error) {
	var toks []Token
	for {
		l.skipSpace()
		r, size := l.peek()
		start, line, col := l.pos, l.line, l.col
		if size == 0 {
			toks = append(toks, Token{Kind: EOF, Line: line, Col: col, Start: start, End: start})
			return toks, nil
		}

		switch {
		case r == '_' || unicode.IsLetter(r):
			for {
				c, n := l.peek()
				if n == 0 || !(c == '_' || unicode.IsLetter(c) || isDigit(c)) {
					break
				}
				l.advance()
			}
			toks = append(toks, Token{Kind: IDENT, Text: l.src[start:l.pos], Line: line, Col: col, Start: start, End: l.pos})

		case isDigit(r):
			for {
				c, n := l.peek()
				if n == 0 || !isDigit(c) {
					break
				}
				l.advance()
			}
			// 数字の直後に識別子文字が続くもの(例: 12abc)は不正。
			if c, n := l.peek(); n != 0 && (c == '_' || unicode.IsLetter(c)) {
				return nil, lyerr.UnexpectedChar(string(c)).WithPos(l.line, l.col)
			}
			toks = append(toks, Token{Kind: INT, Text: l.src[start:l.pos], Line: line, Col: col, Start: start, End: l.pos})

		case r == '"' || r == '\'':
			text, err := l.readString(r)
			if err != nil {
				return nil, err.WithPos(line, col)
			}
			toks = append(toks, Token{Kind: STRING, Text: text, Line: line, Col: col, Start: start, End: l.pos})

		default:
			kind, ok := singleKind(r)
			if !ok {
				return nil, lyerr.UnexpectedChar(string(r)).WithPos(line, col)
			}
			l.advance()
			toks = append(toks, Token{Kind: kind, Text: string(r), Line: line, Col: col, Start: start, End: l.pos})
		}
	}
}

func singleKind(r rune) (Kind, bool) {
	switch r {
	case '+':
		return PLUS, true
	case '-':
		return MINUS, true
	case '=':
		return EQ, true
	case '÷':
		return DIV, true
	case '(':
		return LPAREN, true
	case ')':
		return RPAREN, true
	case '[':
		return LBRACKET, true
	case ']':
		return RBRACKET, true
	case ',':
		return COMMA, true
	}
	return 0, false
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func (l *Lexer) peek() (rune, int) {
	if l.pos >= len(l.src) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(l.src[l.pos:])
}

func (l *Lexer) advance() {
	r, size := l.peek()
	if size == 0 {
		return
	}
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
}

// skipSpace は全角空白を含む空白文字を読み飛ばす。
func (l *Lexer) skipSpace() {
	for {
		r, size := l.peek()
		if size == 0 || !unicode.IsSpace(r) {
			return
		}
		l.advance()
	}
}

// readString は開き引用符 quote から始まる文字列を読み、エスケープを解除した内容を返す。
// 対応するエスケープ: \n \t \r \0 \\ \" \'
func (l *Lexer) readString(quote rune) (string, *lyerr.Error) {
	l.advance() // 開き引用符
	var b strings.Builder
	for {
		r, size := l.peek()
		if size == 0 || r == '\n' {
			return "", lyerr.UnterminatedString()
		}
		l.advance()
		switch r {
		case quote:
			return b.String(), nil
		case '\\':
			e, esize := l.peek()
			if esize == 0 || e == '\n' {
				return "", lyerr.UnterminatedString()
			}
			escLine, escCol := l.line, l.col
			l.advance()
			switch e {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '0':
				b.WriteByte(0)
			case '\\', '"', '\'':
				b.WriteRune(e)
			default:
				return "", lyerr.InvalidEscape("\\"+string(e)).WithPos(escLine, escCol-1)
			}
		default:
			b.WriteRune(r)
		}
	}
}
