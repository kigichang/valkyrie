package result

import (
	"errors"
	"testing"

	"github.com/kigichang/valkyrie/option"
)

var errBoom = errors.New("boom")

func TestConstructors(t *testing.T) {
	if !Ok(1).IsOk() {
		t.Fatal("Ok should be IsOk")
	}
	if !Err[int](errBoom).IsErr() {
		t.Fatal("Err should be IsErr")
	}
	if v, _ := From(1, nil).Get(); v != 1 {
		t.Fatal("From with nil error should be Ok")
	}
	if From[int](1, errBoom).IsOk() {
		t.Fatal("From with error should be Err")
	}
}

func TestQueries(t *testing.T) {
	if !Ok(4).IsOkAnd(func(v int) bool { return v == 4 }) {
		t.Fatal("IsOkAnd should be true")
	}
	if Ok(4).IsOkAnd(func(v int) bool { return v == 5 }) {
		t.Fatal("IsOkAnd should be false")
	}
	if Err[int](errBoom).IsOkAnd(func(v int) bool { return true }) {
		t.Fatal("IsOkAnd on Err should be false")
	}
	if !Err[int](errBoom).IsErrAnd(func(err error) bool { return err == errBoom }) {
		t.Fatal("IsErrAnd should be true")
	}
	if Ok(4).IsErrAnd(func(err error) bool { return true }) {
		t.Fatal("IsErrAnd on Ok should be false")
	}
}

func TestExtraction(t *testing.T) {
	if Ok(3).Unwrap() != 3 {
		t.Fatal("Unwrap should return contained value")
	}
	if Err[int](errBoom).UnwrapErr() != errBoom {
		t.Fatal("UnwrapErr should return contained error")
	}
	if v, err := Ok(3).Std(); v != 3 || err != nil {
		t.Fatal("Std on Ok should return value with nil error")
	}
	if v, err := Err[int](errBoom).Std(); v != 0 || err != errBoom {
		t.Fatal("Std on Err should return zero value with error")
	}
	if Ok(3).UnwrapOr(9) != 3 {
		t.Fatal("UnwrapOr on Ok should return contained value")
	}
	if Err[int](errBoom).UnwrapOr(9) != 9 {
		t.Fatal("UnwrapOr on Err should return fallback")
	}
	if Err[int](errBoom).UnwrapOrElse(func(err error) int { return 9 }) != 9 {
		t.Fatal("UnwrapOrElse on Err should return computed fallback")
	}
	if Err[int](errBoom).UnwrapOrZero() != 0 {
		t.Fatal("UnwrapOrZero on Err should return zero value")
	}
}

func TestUnwrapPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Unwrap on Err should panic")
		}
	}()
	Err[int](errBoom).Unwrap()
}

func TestUnwrapErrPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("UnwrapErr on Ok should panic")
		}
	}()
	Ok(3).UnwrapErr()
}

func TestExpectPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r != "custom: boom" {
			t.Fatalf("expected panic message 'custom: boom', got %v", r)
		}
	}()
	Err[int](errBoom).Expect("custom")
}

func TestExpectErrPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r != "custom: 3" {
			t.Fatalf("expected panic message 'custom: 3', got %v", r)
		}
	}()
	Ok(3).ExpectErr("custom")
}

func TestTransformation(t *testing.T) {
	if v := Ok(3).Map(func(i int) int { return i * 2 }).Unwrap(); v != 6 {
		t.Fatal("Map should transform Ok")
	}
	if Err[int](errBoom).Map(func(i int) int { return i * 2 }).IsOk() {
		t.Fatal("Map on Err should stay Err")
	}
	if err := Err[int](errBoom).MapErr(func(err error) error { return errors.New("wrapped: " + err.Error()) }).UnwrapErr(); err.Error() != "wrapped: boom" {
		t.Fatal("MapErr should transform Err")
	}
	if !Ok(3).MapErr(func(err error) error { return errBoom }).IsOk() {
		t.Fatal("MapErr on Ok should stay Ok")
	}

	if Ok(3).MapOr(0, func(i int) int { return i * 2 }) != 6 {
		t.Fatal("MapOr on Ok should apply f")
	}
	if Err[int](errBoom).MapOr(0, func(i int) int { return i * 2 }) != 0 {
		t.Fatal("MapOr on Err should return fallback")
	}
	if Err[int](errBoom).MapOrElse(func(err error) int { return 9 }, func(i int) int { return i * 2 }) != 9 {
		t.Fatal("MapOrElse on Err should return computed fallback")
	}
	if Ok(3).MapOrElse(func(err error) int { return 9 }, func(i int) int { return i * 2 }) != 6 {
		t.Fatal("MapOrElse on Ok should apply f")
	}
}

func TestChaining(t *testing.T) {
	if v := Ok(3).And(Ok("x")).UnwrapOr("y"); v != "x" {
		t.Fatal("And on Ok should return other")
	}
	if Err[int](errBoom).And(Ok("x")).IsOk() {
		t.Fatal("And on Err should be Err")
	}
	if v := Ok(3).AndThen(func(i int) Result[int] { return Ok(i + 1) }).Unwrap(); v != 4 {
		t.Fatal("AndThen should chain Ok")
	}
	if Ok(3).AndThen(func(i int) Result[int] { return Err[int](errBoom) }).IsOk() {
		t.Fatal("AndThen returning Err should propagate Err")
	}

	if Ok(1).Or(Ok(2)).Unwrap() != 1 {
		t.Fatal("Or on Ok should return self")
	}
	if Err[int](errBoom).Or(Ok(2)).Unwrap() != 2 {
		t.Fatal("Or on Err should return other")
	}
	if Err[int](errBoom).OrElse(func(err error) Result[int] { return Ok(2) }).Unwrap() != 2 {
		t.Fatal("OrElse on Err should return computed fallback")
	}
	if Ok(1).OrElse(func(err error) Result[int] { return Ok(2) }).Unwrap() != 1 {
		t.Fatal("OrElse on Ok should return self")
	}
}

func TestInspect(t *testing.T) {
	called := false
	Ok(3).Inspect(func(i int) { called = true })
	if !called {
		t.Fatal("Inspect should call f on Ok")
	}
	called = false
	Err[int](errBoom).Inspect(func(i int) { called = true })
	if called {
		t.Fatal("Inspect should not call f on Err")
	}

	called = false
	Err[int](errBoom).InspectErr(func(err error) { called = true })
	if !called {
		t.Fatal("InspectErr should call f on Err")
	}
	called = false
	Ok(3).InspectErr(func(err error) { called = true })
	if called {
		t.Fatal("InspectErr should not call f on Ok")
	}
}

func TestConversion(t *testing.T) {
	if v, ok := Ok(3).ToOption().Get(); !ok || v != 3 {
		t.Fatal("ToOption on Ok should be Some")
	}
	if Err[int](errBoom).ToOption().IsSome() {
		t.Fatal("ToOption on Err should be None")
	}
	if v, ok := Err[int](errBoom).ErrOption().Get(); !ok || v != errBoom {
		t.Fatal("ErrOption on Err should be Some(err)")
	}
	if Ok(3).ErrOption().IsSome() {
		t.Fatal("ErrOption on Ok should be None")
	}

	if v, ok := Ok(3).ToEither().GetRight(); !ok || v != 3 {
		t.Fatal("ToEither on Ok should be Right")
	}
	if v, ok := Err[int](errBoom).ToEither().GetLeft(); !ok || v != errBoom {
		t.Fatal("ToEither on Err should be Left")
	}

	if s := Ok(3).Slice(); len(s) != 1 || s[0] != 3 {
		t.Fatal("Slice on Ok should return single-element slice")
	}
	if s := Err[int](errBoom).Slice(); s != nil {
		t.Fatal("Slice on Err should return nil")
	}

	if Ok(3).String() != "Ok(3)" {
		t.Fatalf("unexpected String() output: %s", Ok(3).String())
	}
	if Err[int](errBoom).String() != "Err(boom)" {
		t.Fatalf("unexpected String() output: %s", Err[int](errBoom).String())
	}
}

func TestFlatten(t *testing.T) {
	if v := Flatten(Ok(Ok(7))).Unwrap(); v != 7 {
		t.Fatal("Flatten should unwrap nested Ok")
	}
	if Flatten(Ok(Err[int](errBoom))).IsOk() {
		t.Fatal("Flatten should propagate an inner Err")
	}
	if Flatten(Err[Result[int]](errBoom)).IsOk() {
		t.Fatal("Flatten should propagate an outer Err")
	}
}

func TestTranspose(t *testing.T) {
	v, ok := Transpose(Ok(option.Some(3))).Get()
	if !ok || v.Unwrap() != 3 {
		t.Fatal("Transpose of Ok(Some(v)) should be Some(Ok(v))")
	}
	if Transpose(Ok(option.None[int]())).IsSome() {
		t.Fatal("Transpose of Ok(None) should be None")
	}
	r, ok := Transpose(Err[option.Option[int]](errBoom)).Get()
	if !ok || r.UnwrapErr() != errBoom {
		t.Fatal("Transpose of Err should be Some(Err)")
	}
}

func TestCollect(t *testing.T) {
	vs, err := Collect([]Result[int]{Ok(1), Ok(2), Ok(3)}).Std()
	if err != nil || len(vs) != 3 || vs[0] != 1 || vs[1] != 2 || vs[2] != 3 {
		t.Fatal("Collect of all Ok should return the values")
	}
	if Collect([]Result[int]{Ok(1), Err[int](errBoom), Ok(3)}).IsOk() {
		t.Fatal("Collect should short-circuit on the first Err")
	}
}

func TestContainsEqual(t *testing.T) {
	if !Contains(Ok(3), 3) {
		t.Fatal("Contains should be true for matching Ok")
	}
	if Contains(Ok(3), 4) {
		t.Fatal("Contains should be false for non-matching Ok")
	}
	if Contains(Err[int](errBoom), 3) {
		t.Fatal("Contains should be false for Err")
	}

	if !Equal(Ok(1), Ok(1)) {
		t.Fatal("Equal(Ok(1), Ok(1)) should be true")
	}
	if Equal(Ok(1), Ok(2)) {
		t.Fatal("Equal(Ok(1), Ok(2)) should be false")
	}
	if !Equal(Err[int](errBoom), Err[int](errBoom)) {
		t.Fatal("Equal(Err, Err) with same error should be true")
	}
	if Equal(Err[int](errBoom), Err[int](errors.New("boom"))) {
		t.Fatal("Equal(Err, Err) with different error values should be false")
	}
	if Equal(Ok(1), Err[int](errBoom)) {
		t.Fatal("Equal(Ok, Err) should be false")
	}
}
