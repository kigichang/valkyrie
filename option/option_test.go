package option

import (
	"errors"
	"testing"

	"github.com/kigichang/valkyrie/tuple"
)

func TestConstructors(t *testing.T) {
	if !Some(1).IsSome() {
		t.Fatal("Some should be IsSome")
	}
	if !None[int]().IsNone() {
		t.Fatal("None should be IsNone")
	}
	if FromOk(1, false).IsSome() {
		t.Fatal("FromOk(_, false) should be None")
	}
	if v, _ := FromOk(1, true).Get(); v != 1 {
		t.Fatal("FromOk(1, true) should be Some(1)")
	}
	x := 5
	if v, _ := FromPtr(&x).Get(); v != 5 {
		t.Fatal("FromPtr(&x) should be Some(5)")
	}
	if FromPtr[int](nil).IsSome() {
		t.Fatal("FromPtr(nil) should be None")
	}
	if FromErr(1, errors.New("boom")).IsSome() {
		t.Fatal("FromErr with error should be None")
	}
	if v, _ := FromErr(1, nil).Get(); v != 1 {
		t.Fatal("FromErr with nil error should be Some(1)")
	}
}

func TestQueries(t *testing.T) {
	if !Some(4).IsSomeAnd(func(v int) bool { return v == 4 }) {
		t.Fatal("IsSomeAnd should be true")
	}
	if Some(4).IsSomeAnd(func(v int) bool { return v == 5 }) {
		t.Fatal("IsSomeAnd should be false")
	}
	if None[int]().IsSomeAnd(func(v int) bool { return true }) {
		t.Fatal("IsSomeAnd on None should be false")
	}
	if !None[int]().IsNoneOr(func(v int) bool { return false }) {
		t.Fatal("IsNoneOr on None should be true")
	}
	if !Some(4).IsNoneOr(func(v int) bool { return v == 4 }) {
		t.Fatal("IsNoneOr should be true")
	}
}

func TestExtraction(t *testing.T) {
	if Some(3).Must() != 3 {
		t.Fatal("Must should return contained value")
	}
	if Some(3).GetOrElse(9) != 3 {
		t.Fatal("GetOrElse on Some should return contained value")
	}
	if None[int]().GetOrElse(9) != 9 {
		t.Fatal("GetOrElse on None should return fallback")
	}
	if None[int]().GetOrElseFunc(func() int { return 9 }) != 9 {
		t.Fatal("GetOrElseFunc on None should return fallback")
	}
	if None[int]().GetOrZero() != 0 {
		t.Fatal("GetOrZero on None should return zero value")
	}
	if p := None[int]().Ptr(); p != nil {
		t.Fatal("Ptr on None should be nil")
	}
	if p := Some(3).Ptr(); p == nil || *p != 3 {
		t.Fatal("Ptr on Some should point to contained value")
	}
}

func TestMustPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Must on None should panic")
		}
	}()
	None[int]().Must()
}

func TestMustMsgPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r != "custom" {
			t.Fatalf("expected panic message 'custom', got %v", r)
		}
	}()
	None[int]().Unwrap("custom")
}

func TestTransform(t *testing.T) {
	if v, _ := Some(3).Map(func(i int) int { return i * 2 }).Get(); v != 6 {
		t.Fatal("Map should transform Some")
	}
	if None[int]().Map(func(i int) int { return i * 2 }).IsSome() {
		t.Fatal("Map on None should stay None")
	}
	if v, _ := Some(3).FlatMap(func(i int) Option[int] { return Some(i + 1) }).Get(); v != 4 {
		t.Fatal("FlatMap should chain Some")
	}
	if Some(3).FlatMap(func(i int) Option[int] { return None[int]() }).IsSome() {
		t.Fatal("FlatMap returning None should propagate None")
	}
	if Some(3).MapOr(0, func(i int) int { return i * 2 }) != 6 {
		t.Fatal("MapOr on Some should apply f")
	}
	if None[int]().MapOr(0, func(i int) int { return i * 2 }) != 0 {
		t.Fatal("MapOr on None should return fallback")
	}
	if None[int]().MapOrFunc(func() int { return 9 }, func(i int) int { return i * 2 }) != 9 {
		t.Fatal("MapOrFunc on None should return fallback")
	}
	if Some(3).And(Some("x")).GetOrZero() != "x" {
		t.Fatal("And on Some should return other")
	}
	if None[int]().And(Some("x")).IsSome() {
		t.Fatal("And on None should be None")
	}
}

func TestZipFlattenUnzip(t *testing.T) {
	z := Zip(Some(1), Some("a"))
	p, ok := z.Get()
	if !ok || p.V1() != 1 || p.V2() != "a" {
		t.Fatal("Zip should combine two Somes")
	}
	if Zip(None[int](), Some("a")).IsSome() {
		t.Fatal("Zip with a None should be None")
	}

	if v, _ := Flatten(Some(Some(7))).Get(); v != 7 {
		t.Fatal("Flatten should unwrap nested Some")
	}
	if Flatten(None[Option[int]]()).IsSome() {
		t.Fatal("Flatten of None should be None")
	}

	a, b := Unzip(Some(tuple.New(1, "a")))
	if av, _ := a.Get(); av != 1 {
		t.Fatal("Unzip should split V1")
	}
	if bv, _ := b.Get(); bv != "a" {
		t.Fatal("Unzip should split V2")
	}
	a, b = Unzip(None[tuple.Tuple[int, string]]())
	if a.IsSome() || b.IsSome() {
		t.Fatal("Unzip of None should produce two Nones")
	}
}

func TestFilterForEach(t *testing.T) {
	if Some(3).Filter(func(i int) bool { return i > 5 }).IsSome() {
		t.Fatal("Filter should reject non-matching value")
	}
	if v, _ := Some(3).Filter(func(i int) bool { return i > 1 }).Get(); v != 3 {
		t.Fatal("Filter should keep matching value")
	}

	called := false
	Some(3).ForEach(func(i int) { called = true })
	if !called {
		t.Fatal("ForEach should call f on Some")
	}
	called = false
	None[int]().ForEach(func(i int) { called = true })
	if called {
		t.Fatal("ForEach should not call f on None")
	}
}

func TestCombining(t *testing.T) {
	if v, _ := Some(1).Or(Some(2)).Get(); v != 1 {
		t.Fatal("Or on Some should return self")
	}
	if v, _ := None[int]().Or(Some(2)).Get(); v != 2 {
		t.Fatal("Or on None should return other")
	}
	if v, _ := None[int]().OrFunc(func() Option[int] { return Some(2) }).Get(); v != 2 {
		t.Fatal("OrFunc on None should return fallback")
	}

	if v, _ := Some(1).Xor(None[int]()).Get(); v != 1 {
		t.Fatal("Xor(Some, None) should return the Some")
	}
	if v, _ := None[int]().Xor(Some(2)).Get(); v != 2 {
		t.Fatal("Xor(None, Some) should return the Some")
	}
	if Some(1).Xor(Some(2)).IsSome() {
		t.Fatal("Xor(Some, Some) should be None")
	}
	if None[int]().Xor(None[int]()).IsSome() {
		t.Fatal("Xor(None, None) should be None")
	}
}

func TestConversion(t *testing.T) {
	errBoom := errors.New("boom")
	if v, err := Some(3).OkOr(errBoom); v != 3 || err != nil {
		t.Fatal("OkOr on Some should return value with nil error")
	}
	if _, err := None[int]().OkOr(errBoom); err != errBoom {
		t.Fatal("OkOr on None should return err")
	}
	if _, err := None[int]().OkOrFunc(func() error { return errBoom }); err != errBoom {
		t.Fatal("OkOrFunc on None should return computed err")
	}

	if s := Some(3).Slice(); len(s) != 1 || s[0] != 3 {
		t.Fatal("Slice on Some should return single-element slice")
	}
	if s := None[int]().Slice(); s != nil {
		t.Fatal("Slice on None should return nil")
	}

	if Some(3).String() != "Some(3)" {
		t.Fatalf("unexpected String() output: %s", Some(3).String())
	}
	if None[int]().String() != "None" {
		t.Fatalf("unexpected String() output: %s", None[int]().String())
	}
}

func TestMutation(t *testing.T) {
	o := Some(3)
	old := o.Take()
	if v, _ := old.Get(); v != 3 {
		t.Fatal("Take should return the original value")
	}
	if o.IsSome() {
		t.Fatal("Take should leave None behind")
	}

	o = Some(3)
	old = o.Replace(5)
	if v, _ := old.Get(); v != 3 {
		t.Fatal("Replace should return the previous value")
	}
	if v, _ := o.Get(); v != 5 {
		t.Fatal("Replace should install the new value")
	}

	var n Option[int]
	p := n.GetOrInsert(7)
	*p = 42
	if v, _ := n.Get(); v != 42 {
		t.Fatal("GetOrInsert should insert and return a pointer to the stored value")
	}
	p2 := n.GetOrInsert(100)
	if *p2 != 42 {
		t.Fatal("GetOrInsert on an already-Some Option should not overwrite")
	}

	var n2 Option[int]
	p3 := n2.GetOrInsertFunc(func() int { return 9 })
	if *p3 != 9 {
		t.Fatal("GetOrInsertFunc should insert computed value")
	}
}

func TestEqual(t *testing.T) {
	if !Equal(Some(1), Some(1)) {
		t.Fatal("Equal(Some(1), Some(1)) should be true")
	}
	if Equal(Some(1), Some(2)) {
		t.Fatal("Equal(Some(1), Some(2)) should be false")
	}
	if !Equal(None[int](), None[int]()) {
		t.Fatal("Equal(None, None) should be true")
	}
	if Equal(Some(1), None[int]()) {
		t.Fatal("Equal(Some, None) should be false")
	}
}
