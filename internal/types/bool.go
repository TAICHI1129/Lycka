package types

// Bool は bl 型。値は TRUE / FALSE の2つ。
type Bool struct{ V bool }

// 共有インスタンス。
var (
	True  = &Bool{V: true}
	False = &Bool{V: false}
)

// FromBool は Go の bool を bl 型の値に変換する。
func FromBool(b bool) *Bool {
	if b {
		return True
	}
	return False
}

func (b *Bool) Type() string { return TypeBool }

func (b *Bool) String() string {
	if b.V {
		return "TRUE"
	}
	return "FALSE"
}

func (b *Bool) Inspect() string { return b.String() }
