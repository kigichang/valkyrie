// Package tuple provides fixed-arity Tuple2, Tuple3, and Tuple4 structs for holding
// two, three, or four values of possibly different types.
package tuple

// Tuple2 holds two values.
type Tuple2[A, B any] struct {
	v1 A
	v2 B
}

// Tuple is an alias for Tuple2.
type Tuple[A, B any] = Tuple2[A, B]

// New is an alias for New2.
func New[A, B any](v1 A, v2 B) Tuple2[A, B] {
	return New2(v1, v2)
}

// New2 returns a Tuple2 holding v1 and v2.
func New2[A, B any](v1 A, v2 B) Tuple2[A, B] {
	return Tuple2[A, B]{v1: v1, v2: v2}
}

// V1 returns the first value.
func (t Tuple2[A, B]) V1() A { return t.v1 }

// V2 returns the second value.
func (t Tuple2[A, B]) V2() B { return t.v2 }

// Tuple3 holds three values.
type Tuple3[A, B, C any] struct {
	v1 A
	v2 B
	v3 C
}

// New3 returns a Tuple3 holding v1, v2, and v3.
func New3[A, B, C any](v1 A, v2 B, v3 C) Tuple3[A, B, C] {
	return Tuple3[A, B, C]{v1: v1, v2: v2, v3: v3}
}

// V1 returns the first value.
func (t Tuple3[A, B, C]) V1() A { return t.v1 }

// V2 returns the second value.
func (t Tuple3[A, B, C]) V2() B { return t.v2 }

// V3 returns the third value.
func (t Tuple3[A, B, C]) V3() C { return t.v3 }

// Tuple4 holds four values.
type Tuple4[A, B, C, D any] struct {
	v1 A
	v2 B
	v3 C
	v4 D
}

// New4 returns a Tuple4 holding v1, v2, v3, and v4.
func New4[A, B, C, D any](v1 A, v2 B, v3 C, v4 D) Tuple4[A, B, C, D] {
	return Tuple4[A, B, C, D]{v1: v1, v2: v2, v3: v3, v4: v4}
}

// V1 returns the first value.
func (t Tuple4[A, B, C, D]) V1() A { return t.v1 }

// V2 returns the second value.
func (t Tuple4[A, B, C, D]) V2() B { return t.v2 }

// V3 returns the third value.
func (t Tuple4[A, B, C, D]) V3() C { return t.v3 }

// V4 returns the fourth value.
func (t Tuple4[A, B, C, D]) V4() D { return t.v4 }
