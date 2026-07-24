package valkyrie

import "github.com/kigichang/valkyrie/future"

type Future[T any] = future.Future[T]

// Go starts f on a new goroutine and returns a Future for its eventual result.
func Go[T any](f func() (T, error)) *Future[T] {
	return future.Go(f)
}

// Successful returns a Future already completed with value.
func Successful[T any](value T) *Future[T] {
	return future.Successful(value)
}

// Failed returns a Future already completed with err.
func Failed[T any](err error) *Future[T] {
	return future.Failed[T](err)
}
