package types

import "sync/atomic"

var smthngCounter atomic.Uint64

// Smthng は smthng 型。具体的な値を特定できない「何か」を表す。
//
// 同一の smthng 同士でも等式が成立するかは判断できない(`x = x` は smthng)。
// ID は生成された個々の「何か」を区別するためのもので、等価判定には使わない。
type Smthng struct{ ID uint64 }

// NewSmthng は新しい「何か」を生成する。
func NewSmthng() *Smthng { return &Smthng{ID: smthngCounter.Add(1)} }

func (s *Smthng) Type() string    { return TypeSmthng }
func (s *Smthng) String() string  { return "smthng" }
func (s *Smthng) Inspect() string { return "smthng" }
