// Package result provides a Result[T] type wrapping Go's ubiquitous (value, error) pattern,
// inspired by Rust's std::result::Result<T, E>. Unlike Rust, where E is generic, Result[T] fixes
// the error side to Go's built-in error interface, since that is the idiomatic Go representation
// of failure. It is adapted to Go idiom: comma-ok returns instead of panics where Go convention
// expects them, a Std method to convert back to (T, error), and generic methods (Go 1.27+) for
// transformations that change the contained type, such as Map.
package result

import (
	"fmt"

	"github.com/kigichang/valkyrie/either"
	"github.com/kigichang/valkyrie/option"
)

// Result[T] holds either a successful value of type T, or an error.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a Result holding a successful value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a Result holding an error. err should be non-nil; a nil err is indistinguishable
// from Ok's zero value.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// From converts Go's standard (value, error) pattern into a Result, holding err if non-nil, or
// value otherwise.
// Example: r := result.From(os.Open(name))
func From[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// -----------------------------------------------------------------------
// Queries
// -----------------------------------------------------------------------

// IsOk reports whether the Result holds a successful value.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr reports whether the Result holds an error.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// IsOkAnd reports whether the Result holds a value and pred returns true for it.
func (r Result[T]) IsOkAnd(pred func(T) bool) bool {
	return r.err == nil && pred(r.value)
}

// IsErrAnd reports whether the Result holds an error and pred returns true for it.
func (r Result[T]) IsErrAnd(pred func(error) bool) bool {
	return r.err != nil && pred(r.err)
}

// -----------------------------------------------------------------------
// Extraction
// -----------------------------------------------------------------------

// Get returns the contained value and true, or the zero value and false if the Result is an
// error.
func (r Result[T]) Get() (T, bool) {
	return r.value, r.err == nil
}

// Std returns the Result as Go's standard (value, error) pattern.
func (r Result[T]) Std() (T, error) {
	return r.value, r.err
}

// Unwrap returns the contained value, or panics with the error if the Result is an error.
func (r Result[T]) Unwrap() T {
	return r.Expect("result: Unwrap called on Err")
}

// Expect returns the contained value, or panics with msg and the error if the Result is an
// error.
func (r Result[T]) Expect(msg string) T {
	if r.err != nil {
		panic(fmt.Sprintf("%s: %v", msg, r.err))
	}
	return r.value
}

// UnwrapErr returns the contained error, or panics if the Result holds a value.
func (r Result[T]) UnwrapErr() error {
	return r.ExpectErr("result: UnwrapErr called on Ok")
}

// ExpectErr returns the contained error, or panics with msg and the value if the Result holds a
// value.
func (r Result[T]) ExpectErr(msg string) error {
	if r.err == nil {
		panic(fmt.Sprintf("%s: %v", msg, r.value))
	}
	return r.err
}

// UnwrapOr returns the contained value, or fallback if the Result is an error.
func (r Result[T]) UnwrapOr(fallback T) T {
	if r.err != nil {
		return fallback
	}
	return r.value
}

// UnwrapOrElse returns the contained value, or the result of calling fallback with the error if
// the Result is an error.
func (r Result[T]) UnwrapOrElse(fallback func(error) T) T {
	if r.err != nil {
		return fallback(r.err)
	}
	return r.value
}

// UnwrapOrZero returns the contained value, or the zero value of T if the Result is an error.
func (r Result[T]) UnwrapOrZero() T {
	return r.value
}

// -----------------------------------------------------------------------
// Transformation
// -----------------------------------------------------------------------

// Map transforms the contained value with f, or returns the error unchanged if the Result is an
// error.
func (r Result[T]) Map[U any](f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.value))
}

// MapErr transforms the contained error with f, or returns the value unchanged if the Result
// holds a value.
func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if r.err == nil {
		return r
	}
	return Err[T](f(r.err))
}

// MapOr transforms the contained value with f, or returns fallback if the Result is an error.
func (r Result[T]) MapOr[U any](fallback U, f func(T) U) U {
	if r.err != nil {
		return fallback
	}
	return f(r.value)
}

// MapOrElse transforms the contained value with f, or returns the result of calling fallback
// with the error if the Result is an error.
func (r Result[T]) MapOrElse[U any](fallback func(error) U, f func(T) U) U {
	if r.err != nil {
		return fallback(r.err)
	}
	return f(r.value)
}

// -----------------------------------------------------------------------
// Chaining
// -----------------------------------------------------------------------

// And returns other if the Result holds a value, or the error unchanged otherwise.
func (r Result[T]) And[U any](other Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return other
}

// AndThen transforms the contained value with f, flattening the result, or returns the error
// unchanged if the Result is an error.
func (r Result[T]) AndThen[U any](f func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return f(r.value)
}

// Or returns the Result if it holds a value, or other otherwise.
func (r Result[T]) Or(other Result[T]) Result[T] {
	if r.err != nil {
		return other
	}
	return r
}

// OrElse returns the Result if it holds a value, or the result of calling f with the error
// otherwise.
func (r Result[T]) OrElse(f func(error) Result[T]) Result[T] {
	if r.err != nil {
		return f(r.err)
	}
	return r
}

// -----------------------------------------------------------------------
// Inspection (side effect)
// -----------------------------------------------------------------------

// Inspect calls f with the contained value, if any, and returns the Result unchanged.
func (r Result[T]) Inspect(f func(T)) Result[T] {
	if r.err == nil {
		f(r.value)
	}
	return r
}

// InspectErr calls f with the contained error, if any, and returns the Result unchanged.
func (r Result[T]) InspectErr(f func(error)) Result[T] {
	if r.err != nil {
		f(r.err)
	}
	return r
}

// -----------------------------------------------------------------------
// Conversion
// -----------------------------------------------------------------------

// ToOption converts the Result into an option.Option, containing the value if Ok, or None if
// Err.
func (r Result[T]) ToOption() option.Option[T] {
	if r.err != nil {
		return option.None[T]()
	}
	return option.Some(r.value)
}

// ErrOption converts the Result into an option.Option, containing the error if Err, or None if
// Ok.
func (r Result[T]) ErrOption() option.Option[error] {
	if r.err == nil {
		return option.None[error]()
	}
	return option.Some(r.err)
}

// ToEither converts the Result into an either.Either, holding the error in Left, or the value in
// Right.
func (r Result[T]) ToEither() either.Either[error, T] {
	if r.err != nil {
		return either.Left[error, T](r.err)
	}
	return either.Right[error, T](r.value)
}

// Slice returns a single-element slice containing the value, or nil if the Result is an error.
func (r Result[T]) Slice() []T {
	if r.err != nil {
		return nil
	}
	return []T{r.value}
}

// String implements fmt.Stringer.
func (r Result[T]) String() string {
	if r.err != nil {
		return fmt.Sprintf("Err(%v)", r.err)
	}
	return fmt.Sprintf("Ok(%v)", r.value)
}

// -----------------------------------------------------------------------
// Free functions that don't fit the receiver's type parameters
// -----------------------------------------------------------------------

// Flatten converts a Result[Result[T]] into a Result[T].
func Flatten[T any](r Result[Result[T]]) Result[T] {
	if r.err != nil {
		return Err[T](r.err)
	}
	return r.value
}

// Transpose converts a Result[option.Option[T]] into an option.Option[Result[T]]: an error
// becomes Some(Err), a value of None becomes None, and a value of Some(v) becomes Some(Ok(v)).
func Transpose[T any](r Result[option.Option[T]]) option.Option[Result[T]] {
	if r.err != nil {
		return option.Some(Err[T](r.err))
	}
	v, ok := r.value.Get()
	if !ok {
		return option.None[Result[T]]()
	}
	return option.Some(Ok(v))
}

// Collect gathers rs into a Result holding a slice of every value, or the first error
// encountered.
func Collect[T any](rs []Result[T]) Result[[]T] {
	values := make([]T, 0, len(rs))
	for _, r := range rs {
		if r.err != nil {
			return Err[[]T](r.err)
		}
		values = append(values, r.value)
	}
	return Ok(values)
}

// Contains reports whether the Result holds a value equal to elem.
func Contains[T comparable](r Result[T], elem T) bool {
	v, ok := r.Get()
	return ok && v == elem
}

// Equal reports whether a and b are both Ok with equal values, or both Err with equal errors.
func Equal[T comparable](a, b Result[T]) bool {
	if (a.err == nil) != (b.err == nil) {
		return false
	}
	if a.err != nil {
		return a.err == b.err
	}
	return a.value == b.value
}
