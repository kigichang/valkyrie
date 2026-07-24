// Package either provides an Either[L, R] type for representing a value that is one of two
// possible types, conventionally used to hold a success value (Right) or a failure/alternate
// value (Left). It is inspired by Scala's scala.util.Either, which has been right-biased since
// Scala 2.12: Map, FlatMap, and other transformations act on the Right value by default. It is
// adapted to Go idiom: comma-ok returns instead of panics where Go convention expects them,
// (T, error) conversions for Either[error, R], and generic methods (Go 1.27+) for
// transformations that change a contained type, such as Map.
package either

import "fmt"

// Either[L, R] holds a value of one of two types: Left, conventionally used for a failure or
// alternate value, or Right, conventionally used for a success value. Either is right-biased:
// Map, FlatMap, and other transformations operate on the Right value.
type Either[L, R any] struct {
	left    L
	right   R
	isRight bool
}

// Left creates an Either holding a Left value.
func Left[L, R any](value L) Either[L, R] {
	return Either[L, R]{left: value}
}

// Right creates an Either holding a Right value.
func Right[L, R any](value R) Either[L, R] {
	return Either[L, R]{right: value, isRight: true}
}

// Cond creates a Right containing right if test is true, or a Left containing left otherwise.
// Equivalent to Scala's Either.cond.
func Cond[L, R any](test bool, left L, right R) Either[L, R] {
	if test {
		return Right[L, R](right)
	}
	return Left[L, R](left)
}

// CondFunc creates a Right containing the result of calling right if test is true, or a Left
// containing the result of calling left otherwise.
func CondFunc[L, R any](test bool, left func() L, right func() R) Either[L, R] {
	if test {
		return Right[L, R](right())
	}
	return Left[L, R](left())
}

// FromErr converts Go's standard (value, error) pattern into an Either, holding err in Left if
// non-nil, or value in Right otherwise.
// Example: e := either.FromErr(os.Open(name))
func FromErr[R any](value R, err error) Either[error, R] {
	if err != nil {
		return Left[error, R](err)
	}
	return Right[error, R](value)
}

// -----------------------------------------------------------------------
// Queries
// -----------------------------------------------------------------------

// IsLeft reports whether the Either holds a Left value.
func (e Either[L, R]) IsLeft() bool {
	return !e.isRight
}

// IsRight reports whether the Either holds a Right value.
func (e Either[L, R]) IsRight() bool {
	return e.isRight
}

// IsLeftAnd reports whether the Either holds a Left value and pred returns true for it.
func (e Either[L, R]) IsLeftAnd(pred func(L) bool) bool {
	return !e.isRight && pred(e.left)
}

// IsRightAnd reports whether the Either holds a Right value and pred returns true for it.
func (e Either[L, R]) IsRightAnd(pred func(R) bool) bool {
	return e.isRight && pred(e.right)
}

// -----------------------------------------------------------------------
// Extraction
// -----------------------------------------------------------------------

// GetLeft returns the contained Left value and true, or the zero value and false if Right.
func (e Either[L, R]) GetLeft() (L, bool) {
	return e.left, !e.isRight
}

// GetRight returns the contained Right value and true, or the zero value and false if Left.
func (e Either[L, R]) GetRight() (R, bool) {
	return e.right, e.isRight
}

// Must returns the contained Right value, or panics if the Either is Left.
func (e Either[L, R]) Must() R {
	return e.MustMsg(fmt.Sprintf("either: Must called on Left(%v)", e.left))
}

// MustMsg returns the contained Right value, or panics with msg if the Either is Left.
func (e Either[L, R]) MustMsg(msg string) R {
	if !e.isRight {
		panic(msg)
	}
	return e.right
}

// MustLeft returns the contained Left value, or panics if the Either is Right.
func (e Either[L, R]) MustLeft() L {
	if e.isRight {
		panic(fmt.Sprintf("either: MustLeft called on Right(%v)", e.right))
	}
	return e.left
}

// GetOrElse returns the contained Right value, or fallback if the Either is Left.
func (e Either[L, R]) GetOrElse(fallback R) R {
	if !e.isRight {
		return fallback
	}
	return e.right
}

// GetOrElseFunc returns the contained Right value, or the result of calling fallback with the
// Left value if the Either is Left.
func (e Either[L, R]) GetOrElseFunc(fallback func(L) R) R {
	if !e.isRight {
		return fallback(e.left)
	}
	return e.right
}

// -----------------------------------------------------------------------
// Transformation
// -----------------------------------------------------------------------

// Map transforms the contained Right value with f, or returns the Either unchanged if Left.
func (e Either[L, R]) Map[R2 any](f func(R) R2) Either[L, R2] {
	if !e.isRight {
		return Left[L, R2](e.left)
	}
	return Right[L, R2](f(e.right))
}

// MapLeft transforms the contained Left value with f, or returns the Either unchanged if Right.
func (e Either[L, R]) MapLeft[L2 any](f func(L) L2) Either[L2, R] {
	if e.isRight {
		return Right[L2, R](e.right)
	}
	return Left[L2, R](f(e.left))
}

// FlatMap transforms the contained Right value with f, flattening the result, or returns the
// Either unchanged if Left.
func (e Either[L, R]) FlatMap[R2 any](f func(R) Either[L, R2]) Either[L, R2] {
	if !e.isRight {
		return Left[L, R2](e.left)
	}
	return f(e.right)
}

// FlatMapLeft transforms the contained Left value with f, flattening the result, or returns the
// Either unchanged if Right.
func (e Either[L, R]) FlatMapLeft[L2 any](f func(L) Either[L2, R]) Either[L2, R] {
	if e.isRight {
		return Right[L2, R](e.right)
	}
	return f(e.left)
}

// Fold collapses the Either to a single value by applying onLeft or onRight to the contained
// value, whichever is present.
func (e Either[L, R]) Fold[U any](onLeft func(L) U, onRight func(R) U) U {
	if e.isRight {
		return onRight(e.right)
	}
	return onLeft(e.left)
}

// Swap exchanges Left and Right: a Left becomes a Right and vice versa.
func (e Either[L, R]) Swap() Either[R, L] {
	if e.isRight {
		return Left[R, L](e.right)
	}
	return Right[R, L](e.left)
}

// -----------------------------------------------------------------------
// Filter / side effect
// -----------------------------------------------------------------------

// FilterOrElse returns the Either unchanged if it holds a Right value for which pred returns
// true, or a Left containing the result of calling zero otherwise. An Either that is already
// Left is returned unchanged.
func (e Either[L, R]) FilterOrElse(pred func(R) bool, zero func() L) Either[L, R] {
	if e.isRight && !pred(e.right) {
		return Left[L, R](zero())
	}
	return e
}

// ForEach calls f with the contained Right value, if any.
func (e Either[L, R]) ForEach(f func(R)) {
	if e.isRight {
		f(e.right)
	}
}

// ForEachLeft calls f with the contained Left value, if any.
func (e Either[L, R]) ForEachLeft(f func(L)) {
	if !e.isRight {
		f(e.left)
	}
}

// -----------------------------------------------------------------------
// Combining
// -----------------------------------------------------------------------

// OrElse returns the Either if it holds a Right value, or other otherwise.
func (e Either[L, R]) OrElse(other Either[L, R]) Either[L, R] {
	if e.isRight {
		return e
	}
	return other
}

// OrElseFunc returns the Either if it holds a Right value, or the result of calling fallback
// with the Left value otherwise.
func (e Either[L, R]) OrElseFunc(fallback func(L) Either[L, R]) Either[L, R] {
	if e.isRight {
		return e
	}
	return fallback(e.left)
}

// -----------------------------------------------------------------------
// Conversion
// -----------------------------------------------------------------------

// Slice returns a single-element slice containing the Right value, or nil if Left.
func (e Either[L, R]) Slice() []R {
	if !e.isRight {
		return nil
	}
	return []R{e.right}
}

// LeftSlice returns a single-element slice containing the Left value, or nil if Right.
func (e Either[L, R]) LeftSlice() []L {
	if e.isRight {
		return nil
	}
	return []L{e.left}
}

// String implements fmt.Stringer.
func (e Either[L, R]) String() string {
	if !e.isRight {
		return fmt.Sprintf("Left(%v)", e.left)
	}
	return fmt.Sprintf("Right(%v)", e.right)
}

// -----------------------------------------------------------------------
// Free functions that don't fit the receiver's type parameters
// -----------------------------------------------------------------------

// ToErr converts an Either[error, R] into Go's standard (value, error) pattern.
func ToErr[R any](e Either[error, R]) (R, error) {
	if v, ok := e.GetRight(); ok {
		return v, nil
	}
	left, _ := e.GetLeft()
	var zero R
	return zero, left
}

// ContainsRight reports whether the Either holds a Right value equal to elem.
func ContainsRight[L any, R comparable](e Either[L, R], elem R) bool {
	v, ok := e.GetRight()
	return ok && v == elem
}

// ContainsLeft reports whether the Either holds a Left value equal to elem.
func ContainsLeft[L comparable, R any](e Either[L, R], elem L) bool {
	v, ok := e.GetLeft()
	return ok && v == elem
}

// Equal reports whether a and b hold the same side (Left or Right) with equal values.
func Equal[L, R comparable](a, b Either[L, R]) bool {
	if a.IsRight() != b.IsRight() {
		return false
	}
	if a.IsRight() {
		av, _ := a.GetRight()
		bv, _ := b.GetRight()
		return av == bv
	}
	av, _ := a.GetLeft()
	bv, _ := b.GetLeft()
	return av == bv
}
