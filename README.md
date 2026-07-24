# valkyrie

Monads and functional data types for Go, built on **generic methods**
(`func (r Recv[T]) Map[U any](f func(T) U) Recv[U]`), a feature introduced in
Go 1.27. Prior to generic methods, a `Map` that changes a contained type had
to be a free function (`Map[T, U any](o Option[T], f func(T) U) Option[U]`)
because a method's type parameters were fixed to its receiver. valkyrie uses
generic methods throughout so transformations read as `opt.Map(f)` instead of
`option.Map(opt, f)`.

The library is deliberately idiomatic-Go where the borrowed languages are
not: comma-ok returns instead of panics where Go convention expects them,
`(T, error)` conversions at the edges, and no forced adoption — each type
also converts cleanly to and from plain Go values.

## Install

```sh
go get github.com/kigichang/valkyrie
```

Requires Go 1.27 or later (generic methods).

## Types

| Type | Package | Inspired by | Purpose |
|---|---|---|---|
| `Option[T]` | [`option`](./option) | Rust `Option`, Scala `Option` | An explicit alternative to a zero value or `nil` for "no value". |
| `Result[T]` | [`result`](./result) | Rust `Result<T, E>` | Wraps Go's `(value, error)` pattern as a value; error side fixed to `error`. |
| `Either[L, R]` | [`either`](./either) | Scala `Either` | A value that is one of two types; right-biased, for a success/alternate pair not tied to `error`. |
| `Tuple2`/`Tuple3`/`Tuple4` | [`tuple`](./tuple) | — | Fixed-arity generic structs for grouping 2–4 values without a named type. |
| `Pair[K, V]` | [`pair`](./pair) | — | A key-value element, e.g. the element type of a hash map. |
| `Future[T]` | [`future`](./future) | Scala `Future` | A value produced asynchronously on its own goroutine, with `Map`/`FlatMap`/`Recover`/`Zip`-style combinators. |

Every type is immutable, comparison-friendly where its contents allow, and
implements `fmt.Stringer` — except `Future[T]`, which is inherently mutable
(it starts pending and completes once) and is always used through a
`*Future[T]` pointer rather than copied.

The root package (`github.com/kigichang/valkyrie`) re-exports `Option`,
`Result`, `Either`, and `Future` as type aliases with their constructors, so
you can `import "github.com/kigichang/valkyrie"` and write `valkyrie.Some(v)`
instead of importing each subpackage individually. `Tuple` and `Pair` are
imported from their own subpackages (`tuple`, `pair`).

## Example

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/kigichang/valkyrie/option"
	"github.com/kigichang/valkyrie/result"
)

func main() {
	// Option: explicit absence instead of a zero value.
	m := map[string]int{"a": 1}
	v, ok := m["a"]
	opt := option.FromOk(v, ok)
	doubled := opt.Map(func(v int) int { return v * 2 })
	fmt.Println(doubled) // Some(2)

	// Result: (value, error) as a value, chained with generic Map/AndThen.
	r := result.From(strconv.Atoi("42")).
		Map(func(v int) int { return v + 1 }).
		AndThen(func(v int) result.Result[string] {
			return result.Ok(strconv.Itoa(v))
		})
	fmt.Println(r.UnwrapOr("fallback")) // 43

	// Convert back to idiomatic Go at the boundary.
	value, err := r.Std()
	_ = value
	_ = err
}
```

## Design notes

- **Right/success-biased.** `Either`'s `Map`/`FlatMap`/etc. act on `Right`;
  `Result`'s act on the `Ok` value. Left/error-specific operations are
  suffixed (`MapLeft`, `MapErr`) rather than requiring a `Swap()` first.
- **Private fields, method access.** Every type stores its state in private
  fields and exposes it only through methods (`Get`, `Unwrap`, `V1()`,
  `Key()`, ...), so a type's invariants (e.g. `Result` never has both a value
  and an error) can't be broken by direct field mutation.
- **Interop at the edges.** Each type converts to and from the Go pattern it
  replaces: `Result.Std()` → `(T, error)`, `Either[error, R]` ↔ `(R, error)`
  via `either.ToErr`/`either.FromErr`, `Option.Get()` → `(T, bool)`,
  `Result.ToOption()`/`ToEither()` cross between types.
- **No hidden panics.** Panicking accessors (`Must`, `Unwrap`, `Expect`, ...)
  are opt-in and explicitly named; the default extraction path is always
  comma-ok or `-Or`/`-OrElse` with a fallback.

## Status

Early stage — API may still change. See each package's godoc comments for
the full method list.

## License

[MIT](./LICENSE)
