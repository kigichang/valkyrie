// Package future provides a Future[T] type representing a value of type T produced
// asynchronously on its own goroutine, inspired by Scala's scala.concurrent.Future. It is
// adapted to Go idiom: a result.Result[T] instead of Scala's Try for the completed outcome,
// comma-ok returns instead of PartialFunction where Scala convention expects them, no implicit
// ExecutionContext (every callback and combinator runs on its own goroutine), and generic
// methods (Go 1.27+) for transformations that change the contained type, such as Map.
package future

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kigichang/valkyrie/option"
	"github.com/kigichang/valkyrie/result"
	"github.com/kigichang/valkyrie/tuple"
)

// ErrFilterFalse is the error a Future fails with when Filter's predicate returns false.
var ErrFilterFalse = errors.New("future: Filter predicate was false")

// ErrReduceLeftEmpty is the error ReduceLeft fails with when called with no futures.
var ErrReduceLeftEmpty = errors.New("future: ReduceLeft called with no futures")

// Future[T] represents a value that may not yet be available, produced by a computation
// running asynchronously on its own goroutine.
type Future[T any] struct {
	mu        sync.Mutex
	done      chan struct{}
	result    result.Result[T]
	completed bool
	callbacks []func(result.Result[T])
}

func newFuture[T any]() *Future[T] {
	return &Future[T]{done: make(chan struct{})}
}

// complete resolves the Future with r, running any registered callbacks. Only the first call
// has any effect, mirroring Scala's Promise semantics where a Future completes at most once.
func (f *Future[T]) complete(r result.Result[T]) {
	f.mu.Lock()
	if f.completed {
		f.mu.Unlock()
		return
	}
	f.result = r
	f.completed = true
	callbacks := f.callbacks
	f.callbacks = nil
	close(f.done)
	f.mu.Unlock()

	for _, cb := range callbacks {
		go cb(r)
	}
}

// -----------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------

// Go starts f on a new goroutine and returns a Future for its eventual result, ala Scala's
// Future.apply, adapted to Go's (value, error) idiom. A panic in f is recovered and reported
// as an error.
func Go[T any](f func() (T, error)) *Future[T] {
	fut := newFuture[T]()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				fut.complete(result.Err[T](fmt.Errorf("future: panic: %v", p)))
			}
		}()
		fut.complete(result.From(f()))
	}()
	return fut
}

// New starts f on a new goroutine and returns a Future for its eventual value, ala Go but for
// functions that cannot fail. A panic in f is recovered and reported as an error.
func New[T any](f func() T) *Future[T] {
	return Go(func() (T, error) {
		return f(), nil
	})
}

// Successful returns a Future already completed with value, ala Scala's Future.successful.
func Successful[T any](value T) *Future[T] {
	fut := newFuture[T]()
	fut.complete(result.Ok(value))
	return fut
}

// Failed returns a Future already completed with err, ala Scala's Future.failed.
func Failed[T any](err error) *Future[T] {
	fut := newFuture[T]()
	fut.complete(result.Err[T](err))
	return fut
}

// FromResult returns a Future already completed with r, ala Scala's Future.fromTry.
func FromResult[T any](r result.Result[T]) *Future[T] {
	fut := newFuture[T]()
	fut.complete(r)
	return fut
}

// Never returns a Future that never completes, ala Scala's Future.never.
func Never[T any]() *Future[T] {
	return newFuture[T]()
}

// -----------------------------------------------------------------------
// Queries
// -----------------------------------------------------------------------

// IsCompleted reports whether the Future has completed, either successfully or with an error.
func (f *Future[T]) IsCompleted() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// Value returns the Future's result if it has completed, or option.None if it is still
// running.
func (f *Future[T]) Value() option.Option[result.Result[T]] {
	select {
	case <-f.done:
		f.mu.Lock()
		r := f.result
		f.mu.Unlock()
		return option.Some(r)
	default:
		return option.None[result.Result[T]]()
	}
}

// -----------------------------------------------------------------------
// Waiting
// -----------------------------------------------------------------------

// Done returns a channel that is closed once the Future completes, for use alongside other
// channels in a select statement.
func (f *Future[T]) Done() <-chan struct{} {
	return f.done
}

// Wait blocks until the Future completes and returns its result as Go's standard
// (value, error) pattern, ala Scala's Await.result with an unbounded timeout.
func (f *Future[T]) Wait() (T, error) {
	<-f.done
	return f.result.Std()
}

// WaitTimeout blocks until the Future completes or timeout elapses, ala Scala's
// Await.result(f, atMost). ok is false if timeout elapsed first.
func (f *Future[T]) WaitTimeout(timeout time.Duration) (value T, err error, ok bool) {
	select {
	case <-f.done:
		v, e := f.result.Std()
		return v, e, true
	case <-time.After(timeout):
		var zero T
		return zero, nil, false
	}
}

// Ready blocks until the Future completes or timeout elapses, reporting whether it completed
// in time, ala Scala's Await.ready.
func (f *Future[T]) Ready(timeout time.Duration) bool {
	select {
	case <-f.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// -----------------------------------------------------------------------
// Callbacks (side effect)
// -----------------------------------------------------------------------

// OnComplete registers fn to be called, on its own goroutine, with the Future's result once it
// completes. fn runs immediately, on its own goroutine, if the Future has already completed.
func (f *Future[T]) OnComplete(fn func(result.Result[T])) {
	f.mu.Lock()
	if f.completed {
		r := f.result
		f.mu.Unlock()
		go fn(r)
		return
	}
	f.callbacks = append(f.callbacks, fn)
	f.mu.Unlock()
}

// Foreach calls fn with the Future's value, on its own goroutine, if it completes
// successfully.
func (f *Future[T]) Foreach(fn func(T)) {
	f.OnComplete(func(r result.Result[T]) {
		if v, ok := r.Get(); ok {
			fn(v)
		}
	})
}

// AndThen registers fn to run with the Future's result once it completes, returning a new
// Future that completes with the same result, letting side effects be chained by method call
// order, ala Scala's Future.andThen.
func (f *Future[T]) AndThen(fn func(result.Result[T])) *Future[T] {
	out := newFuture[T]()
	f.OnComplete(func(r result.Result[T]) {
		fn(r)
		out.complete(r)
	})
	return out
}

// -----------------------------------------------------------------------
// Transformation
// -----------------------------------------------------------------------

// Map transforms the Future's value with fn once it completes successfully, or propagates the
// error unchanged.
func (f *Future[T]) Map[U any](fn func(T) U) *Future[U] {
	out := newFuture[U]()
	f.OnComplete(func(r result.Result[T]) {
		out.complete(r.Map(fn))
	})
	return out
}

// FlatMap transforms the Future's value with fn once it completes successfully, flattening the
// resulting Future, or propagates the error unchanged.
func (f *Future[T]) FlatMap[U any](fn func(T) *Future[U]) *Future[U] {
	out := newFuture[U]()
	f.OnComplete(func(r result.Result[T]) {
		v, ok := r.Get()
		if !ok {
			out.complete(result.Err[U](r.UnwrapErr()))
			return
		}
		fn(v).OnComplete(out.complete)
	})
	return out
}

// Transform transforms the Future's completed result with fn into a new result, ala Scala's
// Future.transform(Try[T] => Try[S]).
func (f *Future[T]) Transform[U any](fn func(result.Result[T]) result.Result[U]) *Future[U] {
	out := newFuture[U]()
	f.OnComplete(func(r result.Result[T]) {
		out.complete(fn(r))
	})
	return out
}

// TransformWith transforms the Future's completed result with fn into a new Future, flattening
// the result, ala Scala's Future.transformWith.
func (f *Future[T]) TransformWith[U any](fn func(result.Result[T]) *Future[U]) *Future[U] {
	out := newFuture[U]()
	f.OnComplete(func(r result.Result[T]) {
		fn(r).OnComplete(out.complete)
	})
	return out
}

// -----------------------------------------------------------------------
// Filter
// -----------------------------------------------------------------------

// Filter keeps the Future's value if it completes successfully and pred returns true for it,
// or fails the Future with ErrFilterFalse otherwise. An already-failed Future is unaffected.
func (f *Future[T]) Filter(pred func(T) bool) *Future[T] {
	out := newFuture[T]()
	f.OnComplete(func(r result.Result[T]) {
		if v, ok := r.Get(); ok && !pred(v) {
			out.complete(result.Err[T](ErrFilterFalse))
			return
		}
		out.complete(r)
	})
	return out
}

// -----------------------------------------------------------------------
// Recovery
// -----------------------------------------------------------------------

// Recover replaces a failed Future's error with the value returned by fn, if fn reports ok, or
// leaves a failed Future unchanged otherwise. A successful Future is unaffected. Mirrors
// Scala's Future.recover, adapted to Go's comma-ok idiom instead of PartialFunction.
func (f *Future[T]) Recover(fn func(error) (T, bool)) *Future[T] {
	out := newFuture[T]()
	f.OnComplete(func(r result.Result[T]) {
		if r.IsOk() {
			out.complete(r)
			return
		}
		if v, ok := fn(r.UnwrapErr()); ok {
			out.complete(result.Ok(v))
			return
		}
		out.complete(r)
	})
	return out
}

// RecoverWith replaces a failed Future with the Future returned by fn, if fn reports ok, or
// leaves a failed Future unchanged otherwise. A successful Future is unaffected. Mirrors
// Scala's Future.recoverWith.
func (f *Future[T]) RecoverWith(fn func(error) (*Future[T], bool)) *Future[T] {
	out := newFuture[T]()
	f.OnComplete(func(r result.Result[T]) {
		if r.IsOk() {
			out.complete(r)
			return
		}
		if next, ok := fn(r.UnwrapErr()); ok {
			next.OnComplete(out.complete)
			return
		}
		out.complete(r)
	})
	return out
}

// FallbackTo returns the Future's value if it succeeds, or other's value if the Future fails
// and other succeeds. If both fail, the Future's own error is kept, ala Scala's
// Future.fallbackTo.
func (f *Future[T]) FallbackTo(other *Future[T]) *Future[T] {
	out := newFuture[T]()
	f.OnComplete(func(r result.Result[T]) {
		if r.IsOk() {
			out.complete(r)
			return
		}
		other.OnComplete(func(r2 result.Result[T]) {
			if r2.IsOk() {
				out.complete(r2)
				return
			}
			out.complete(r)
		})
	})
	return out
}

// -----------------------------------------------------------------------
// Conversion
// -----------------------------------------------------------------------

// String implements fmt.Stringer, reporting the Future's current state.
func (f *Future[T]) String() string {
	if r, ok := f.Value().Get(); ok {
		return fmt.Sprintf("Future(%v)", r)
	}
	return "Future(<pending>)"
}

// -----------------------------------------------------------------------
// Free functions that don't fit the receiver's type parameters
// -----------------------------------------------------------------------

// Flatten converts a Future[*Future[T]] into a Future[T], ala Scala's Future.flatten.
func Flatten[T any](f *Future[*Future[T]]) *Future[T] {
	out := newFuture[T]()
	f.OnComplete(func(r result.Result[*Future[T]]) {
		inner, ok := r.Get()
		if !ok {
			out.complete(result.Err[T](r.UnwrapErr()))
			return
		}
		inner.OnComplete(out.complete)
	})
	return out
}

// Zip combines a's value and b's value into a tuple.Tuple once both complete successfully, or
// fails with whichever error is observed first. A free function rather than a method: a
// method's own type parameter can't be instantiated with a type built from the receiver's type
// parameter (here, Tuple[T, U] from T), so combinators shaped like this live at package level.
func Zip[A, B any](a *Future[A], b *Future[B]) *Future[tuple.Tuple[A, B]] {
	return ZipWith(a, b, tuple.New)
}

// ZipWith combines a's value and b's value using fn once both complete successfully, or fails
// with whichever error is observed first.
func ZipWith[A, B, R any](a *Future[A], b *Future[B], fn func(A, B) R) *Future[R] {
	out := newFuture[R]()
	var once sync.Once
	tryComplete := func() {
		ar, aok := a.Value().Get()
		br, bok := b.Value().Get()
		if !aok || !bok {
			return
		}
		once.Do(func() {
			v1, ok1 := ar.Get()
			if !ok1 {
				out.complete(result.Err[R](ar.UnwrapErr()))
				return
			}
			v2, ok2 := br.Get()
			if !ok2 {
				out.complete(result.Err[R](br.UnwrapErr()))
				return
			}
			out.complete(result.Ok(fn(v1, v2)))
		})
	}
	a.OnComplete(func(result.Result[A]) { tryComplete() })
	b.OnComplete(func(result.Result[B]) { tryComplete() })
	return out
}

// Sequence combines futures into a single Future holding a slice of every value, in the same
// order as futures, or fails with the first error observed, ala Scala's Future.sequence.
func Sequence[T any](futures []*Future[T]) *Future[[]T] {
	out := newFuture[[]T]()
	if len(futures) == 0 {
		out.complete(result.Ok([]T{}))
		return out
	}

	var (
		mu        sync.Mutex
		remaining = len(futures)
		values    = make([]T, len(futures))
		failed    bool
	)
	for i, fut := range futures {
		fut.OnComplete(func(r result.Result[T]) {
			mu.Lock()
			defer mu.Unlock()
			if failed {
				return
			}
			v, ok := r.Get()
			if !ok {
				failed = true
				out.complete(result.Err[[]T](r.UnwrapErr()))
				return
			}
			values[i] = v
			remaining--
			if remaining == 0 {
				out.complete(result.Ok(values))
			}
		})
	}
	return out
}

// Traverse applies fn to each item, then combines the resulting Futures as Sequence does,
// ala Scala's Future.traverse.
func Traverse[T, U any](items []T, fn func(T) *Future[U]) *Future[[]U] {
	futures := make([]*Future[U], len(items))
	for i, item := range items {
		futures[i] = fn(item)
	}
	return Sequence(futures)
}

// FirstCompletedOf returns a Future that completes with the result of whichever of futures
// completes first, ala Scala's Future.firstCompletedOf.
func FirstCompletedOf[T any](futures []*Future[T]) *Future[T] {
	out := newFuture[T]()
	var once sync.Once
	for _, fut := range futures {
		fut.OnComplete(func(r result.Result[T]) {
			once.Do(func() {
				out.complete(r)
			})
		})
	}
	return out
}

// Find returns a Future holding the first value among futures, in completion order, for which
// pred returns true, or option.None if none do, ala Scala's Future.find. A future that fails
// is treated as a non-match rather than failing the overall search.
func Find[T any](futures []*Future[T], pred func(T) bool) *Future[option.Option[T]] {
	out := newFuture[option.Option[T]]()
	if len(futures) == 0 {
		out.complete(result.Ok(option.None[T]()))
		return out
	}

	var (
		mu        sync.Mutex
		remaining = len(futures)
	)
	for _, fut := range futures {
		fut.OnComplete(func(r result.Result[T]) {
			mu.Lock()
			defer mu.Unlock()
			if out.IsCompleted() {
				return
			}
			remaining--
			if v, ok := r.Get(); ok && pred(v) {
				out.complete(result.Ok(option.Some(v)))
				return
			}
			if remaining == 0 {
				out.complete(result.Ok(option.None[T]()))
			}
		})
	}
	return out
}

// FoldLeft combines the values of futures in list order, starting from zero and applying op
// cumulatively, ala Scala's Future.foldLeft. Fails with the first error encountered among
// futures, in list order.
func FoldLeft[T, R any](futures []*Future[T], zero R, op func(R, T) R) *Future[R] {
	out := newFuture[R]()
	if len(futures) == 0 {
		out.complete(result.Ok(zero))
		return out
	}
	go func() {
		acc := zero
		for _, fut := range futures {
			v, err := fut.Wait()
			if err != nil {
				out.complete(result.Err[R](err))
				return
			}
			acc = op(acc, v)
		}
		out.complete(result.Ok(acc))
	}()
	return out
}

// ReduceLeft combines the values of futures in list order using op, starting from the first
// future's value, ala Scala's Future.reduceLeft. Fails with ErrReduceLeftEmpty if futures is
// empty, or the first error encountered among futures, in list order.
func ReduceLeft[T any](futures []*Future[T], op func(T, T) T) *Future[T] {
	out := newFuture[T]()
	if len(futures) == 0 {
		out.complete(result.Err[T](ErrReduceLeftEmpty))
		return out
	}
	go func() {
		acc, err := futures[0].Wait()
		if err != nil {
			out.complete(result.Err[T](err))
			return
		}
		for _, fut := range futures[1:] {
			v, err := fut.Wait()
			if err != nil {
				out.complete(result.Err[T](err))
				return
			}
			acc = op(acc, v)
		}
		out.complete(result.Ok(acc))
	}()
	return out
}
