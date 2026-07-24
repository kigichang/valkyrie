// Package option provides an Option[T] type, an explicit alternative to
// using a zero value or nil to represent "no value". It is inspired by
// Rust's std::option::Option and Scala's scala.Option, adapted to Go
// idiom: comma-ok returns instead of panics where Go convention expects
// them, (T, error) instead of a Result type, and generic methods (Go 1.27+)
// for transformations that change the contained type, such as Map.
package option

import (
	"fmt"

	"github.com/kigichang/valkyrie/tuple"
)

// Option[T] represents an optional value: every Option is either Some and contains a value,
// or None, and does not.
type Option[T any] struct {
	value T
	ok    bool
}

// Some creates an Option containing the provided value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, ok: true}
}

// None creates an empty Option.
func None[T any]() Option[T] {
	return Option[T]{}
}

// FromOk converts Go's standard (value, ok) multi-return pattern into an Option.
// Example: opt := option.FromOk(hashMap[key])
func FromOk[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// FromPtr converts a pointer into an Option. nil returns None.
func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}
	return Some(*ptr)
}

// FromErr converts Go's standard (value, error) pattern into an Option.
func FromErr[T any](value T, err error) Option[T] {
	if err != nil {
		return None[T]()
	}
	return Some(value)
}

// -----------------------------------------------------------------------
// Queries
// -----------------------------------------------------------------------

// IsSome reports whether the Option contains a value.
func (o Option[T]) IsSome() bool {
	return o.ok
}

// IsNone reports whether the Option is empty.
func (o Option[T]) IsNone() bool {
	return !o.ok
}

// IsSomeAnd reports whether the Option contains a value and pred returns true for it.
func (o Option[T]) IsSomeAnd(pred func(T) bool) bool {
	return o.ok && pred(o.value)
}

// IsNoneOr reports whether the Option is empty or pred returns true for its value.
func (o Option[T]) IsNoneOr(pred func(T) bool) bool {
	return !o.ok || pred(o.value)
}

// -----------------------------------------------------------------------
// Extraction
// -----------------------------------------------------------------------

// Get returns the contained value and true, or the zero value and false if empty.
func (o Option[T]) Get() (T, bool) {
	return o.value, o.ok
}

// Must returns the contained value, or panics if the Option is empty.
func (o Option[T]) Must() T {
	return o.Unwrap("option: Must called on None")
}

// MustMsg returns the contained value, or panics with msg if the Option is empty.
func (o Option[T]) Unwrap(msg string) T {
	if !o.ok {
		panic(msg)
	}
	return o.value
}

// GetOrElse returns the contained value, or fallback if the Option is empty.
func (o Option[T]) GetOrElse(fallback T) T {
	if !o.ok {
		return fallback
	}
	return o.value
}

// GetOrElseFunc returns the contained value, or the result of calling fallback if the Option is empty.
func (o Option[T]) GetOrElseFunc(fallback func() T) T {
	if !o.ok {
		return fallback()
	}
	return o.value
}

// GetOrZero returns the contained value, or the zero value of T if the Option is empty.
func (o Option[T]) GetOrZero() T {
	return o.value
}

// Ptr returns a pointer to a copy of the contained value, or nil if the Option is empty.
func (o Option[T]) Ptr() *T {
	if !o.ok {
		return nil
	}
	v := o.value
	return &v
}

// -----------------------------------------------------------------------
// Transformation
// -----------------------------------------------------------------------

// Map transforms the contained value with f, or returns None if the Option is empty.
func (o Option[T]) Map[U any](f func(T) U) Option[U] {
	if !o.ok {
		return None[U]()
	}
	return Some(f(o.value))
}

// FlatMap transforms the contained value with f, flattening the result, or returns None if the Option is empty.
func (o Option[T]) FlatMap[U any](f func(T) Option[U]) Option[U] {
	if !o.ok {
		return None[U]()
	}
	return f(o.value)
}

// MapOr transforms the contained value with f, or returns fallback if the Option is empty.
func (o Option[T]) MapOr[U any](fallback U, f func(T) U) U {
	if !o.ok {
		return fallback
	}
	return f(o.value)
}

// MapOrFunc transforms the contained value with f, or returns the result of calling fallback if the Option is empty.
func (o Option[T]) MapOrFunc[U any](fallback func() U, f func(T) U) U {
	if !o.ok {
		return fallback()
	}
	return f(o.value)
}

// And returns other if the Option contains a value, or None if it is empty.
func (o Option[T]) And[U any](other Option[U]) Option[U] {
	if !o.ok {
		return None[U]()
	}
	return other
}

// -----------------------------------------------------------------------
// Filter / side effect
// -----------------------------------------------------------------------

// Filter returns the Option unchanged if it contains a value for which pred returns true,
// or None otherwise.
func (o Option[T]) Filter(pred func(T) bool) Option[T] {
	if !o.ok || !pred(o.value) {
		return None[T]()
	}
	return o
}

// ForEach calls f with the contained value, if any.
func (o Option[T]) ForEach(f func(T)) {
	if o.ok {
		f(o.value)
	}
}

// -----------------------------------------------------------------------
// Combining
// -----------------------------------------------------------------------

// Or returns the Option if it contains a value, or other otherwise.
func (o Option[T]) Or(other Option[T]) Option[T] {
	if o.ok {
		return o
	}
	return other
}

// OrFunc returns the Option if it contains a value, or the result of calling fallback otherwise.
func (o Option[T]) OrFunc(fallback func() Option[T]) Option[T] {
	if o.ok {
		return o
	}
	return fallback()
}

// Xor returns whichever of the Option and other contains a value, if exactly one does,
// or None if both or neither do.
func (o Option[T]) Xor(other Option[T]) Option[T] {
	switch {
	case o.ok && !other.ok:
		return o
	case !o.ok && other.ok:
		return other
	default:
		return None[T]()
	}
}

// -----------------------------------------------------------------------
// Conversion
// -----------------------------------------------------------------------

// OkOr returns the contained value and a nil error, or the zero value and err if the Option is empty.
func (o Option[T]) OkOr(err error) (T, error) {
	if !o.ok {
		var zero T
		return zero, err
	}
	return o.value, nil
}

// OkOrFunc returns the contained value and a nil error, or the zero value and the result of
// calling err if the Option is empty.
func (o Option[T]) OkOrFunc(err func() error) (T, error) {
	if !o.ok {
		var zero T
		return zero, err()
	}
	return o.value, nil
}

// Slice returns a single-element slice containing the value, or nil if the Option is empty.
func (o Option[T]) Slice() []T {
	if !o.ok {
		return nil
	}
	return []T{o.value}
}

// String implements fmt.Stringer.
func (o Option[T]) String() string {
	if !o.ok {
		return "None"
	}
	return fmt.Sprintf("Some(%v)", o.value)
}

// -----------------------------------------------------------------------
// Mutation
// -----------------------------------------------------------------------

// Take moves the value out of the Option, leaving None in its place, and returns the original Option.
func (o *Option[T]) Take() Option[T] {
	old := *o
	*o = None[T]()
	return old
}

// Replace stores value in the Option, returning the Option's previous value.
func (o *Option[T]) Replace(value T) Option[T] {
	old := *o
	*o = Some(value)
	return old
}

// GetOrInsert stores value in the Option if it is empty, and returns a pointer to the contained value.
func (o *Option[T]) GetOrInsert(value T) *T {
	if !o.ok {
		*o = Some(value)
	}
	return &o.value
}

// GetOrInsertFunc stores the result of calling value in the Option if it is empty,
// and returns a pointer to the contained value.
func (o *Option[T]) GetOrInsertFunc(value func() T) *T {
	if !o.ok {
		*o = Some(value())
	}
	return &o.value
}

// -----------------------------------------------------------------------
// Free functions that don't fit the receiver's type parameter
// -----------------------------------------------------------------------

// Flatten converts an Option[Option[T]] into an Option[T].
func Flatten[T any](o Option[Option[T]]) Option[T] {
	if !o.ok {
		return None[T]()
	}
	return o.value
}

// Zip combines a and b into an Option of a tuple.Tuple, or None if either is empty.
func Zip[A, B any](a Option[A], b Option[B]) Option[tuple.Tuple[A, B]] {
	if !a.ok || !b.ok {
		return None[tuple.Tuple[A, B]]()
	}
	return Some(tuple.New(a.value, b.value))
}

// Unzip splits an Option[tuple.Tuple[A, B]] into a pair of Options.
func Unzip[A, B any](o Option[tuple.Tuple[A, B]]) (Option[A], Option[B]) {
	if !o.ok {
		return None[A](), None[B]()
	}
	return Some(o.value.V1()), Some(o.value.V2())
}

// Equal reports whether a and b are both empty, or both contain equal values.
func Equal[T comparable](a, b Option[T]) bool {
	if a.ok != b.ok {
		return false
	}
	return !a.ok || a.value == b.value
}
