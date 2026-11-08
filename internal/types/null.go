package types

// Null は null 型。「無」を表す。
type Null struct{}

// NullValue は唯一の null 値。
var NullValue = &Null{}

func (n *Null) Type() string    { return TypeNull }
func (n *Null) String() string  { return "null" }
func (n *Null) Inspect() string { return "null" }
