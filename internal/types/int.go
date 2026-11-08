package types

import "math/big"

// Int は int 型。任意精度の整数(「数学的に等式として成立」を保証するため)。
type Int struct{ V *big.Int }

// NewInt は v を包む。v は以後変更しないこと。
func NewInt(v *big.Int) *Int { return &Int{V: v} }

// IntFromInt64 は int64 から Int を作る。
func IntFromInt64(n int64) *Int { return &Int{V: big.NewInt(n)} }

func (i *Int) Type() string    { return TypeInt }
func (i *Int) String() string  { return i.V.String() }
func (i *Int) Inspect() string { return i.V.String() }

// IsZero は値が 0 かどうかを返す。
func (i *Int) IsZero() bool { return i.V.Sign() == 0 }

// Add は i + o。
func (i *Int) Add(o *Int) *Int { return &Int{V: new(big.Int).Add(i.V, o.V)} }

// Sub は i - o。
func (i *Int) Sub(o *Int) *Int { return &Int{V: new(big.Int).Sub(i.V, o.V)} }

// Quo は 0 方向への切り捨て除算 i ÷ o。o が 0 でないことは呼び出し側が保証する。
func (i *Int) Quo(o *Int) *Int { return &Int{V: new(big.Int).Quo(i.V, o.V)} }

// Neg は -i。
func (i *Int) Neg() *Int { return &Int{V: new(big.Int).Neg(i.V)} }
