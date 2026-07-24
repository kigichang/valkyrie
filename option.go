package valkyrie

import "github.com/kigichang/valkyrie/option"

type Option[T any] = option.Option[T]

// Some creates an Option containing the provided value.
func Some[T any](v T) Option[T] {
	return option.Some(v)
}

// None creates an empty Option.
func None[T any]() Option[T] {
	return option.None[T]()
}

// FromOk converts Go's standard (value, ok) multi-return pattern into an Option.
// Example: opt := valkyrie.FromOk(hashMap[key])
func FromOk[T any](value T, ok bool) Option[T] {
	return option.FromOk(value, ok)
}

// FromPtr converts a pointer into an Option. nil returns None.
func FromPtr[T any](ptr *T) Option[T] {
	return option.FromPtr(ptr)
}

// FromErr converts Go's standard (value, error) pattern into an Option.
func FromErr[T any](value T, err error) Option[T] {
	return option.FromErr(value, err)
}
