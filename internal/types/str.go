package types

import "strings"

// Str は str 型。
type Str struct{ V string }

func (s *Str) Type() string   { return TypeStr }
func (s *Str) String() string { return s.V }

// Inspect はダブルクォート付き・エスケープ済みの表現を返す。
func (s *Str) Inspect() string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s.V {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// Concat は文字列の結合(`+`)。
func (s *Str) Concat(o *Str) *Str { return &Str{V: s.V + o.V} }

// Remove は s から o をすべて除去した文字列(`-`)。
func (s *Str) Remove(o *Str) *Str {
	if o.V == "" {
		return s
	}
	return &Str{V: strings.ReplaceAll(s.V, o.V, "")}
}
