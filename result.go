package valkyrie

import "github.com/kigichang/valkyrie/result"

type Result[T any] = result.Result[T]

// Ok creates a Result holding a successful value.
func Ok[T any](value T) Result[T] {
	return result.Ok(value)
}

// Err creates a Result holding an error.
func Err[T any](err error) Result[T] {
	return result.Err[T](err)
}

// ResultFrom converts Go's standard (value, error) pattern into a Result, holding err if
// non-nil, or value otherwise.
func ResultFrom[T any](value T, err error) Result[T] {
	return result.From(value, err)
}
