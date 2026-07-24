package either

import (
	"errors"
	"testing"
)

func TestConstructors(t *testing.T) {
	if !Right[string, int](1).IsRight() {
		t.Fatal("Right should be IsRight")
	}
	if !Left[string, int]("boom").IsLeft() {
		t.Fatal("Left should be IsLeft")
	}
	if v, _ := Cond(true, "boom", 1).GetRight(); v != 1 {
		t.Fatal("Cond(true, ...) should be Right")
	}
	if v, _ := Cond(false, "boom", 1).GetLeft(); v != "boom" {
		t.Fatal("Cond(false, ...) should be Left")
	}
	if v, _ := CondFunc(true, func() string { return "boom" }, func() int { return 1 }).GetRight(); v != 1 {
		t.Fatal("CondFunc(true, ...) should be Right")
	}
	if v, _ := CondFunc(false, func() string { return "boom" }, func() int { return 1 }).GetLeft(); v != "boom" {
		t.Fatal("CondFunc(false, ...) should be Left")
	}
	if v, _ := FromErr(1, nil).GetRight(); v != 1 {
		t.Fatal("FromErr with nil error should be Right")
	}
	errBoom := errors.New("boom")
	if v, _ := FromErr(1, errBoom).GetLeft(); v != errBoom {
		t.Fatal("FromErr with error should be Left")
	}
}

func TestQueries(t *testing.T) {
	if !Right[string, int](4).IsRightAnd(func(v int) bool { return v == 4 }) {
		t.Fatal("IsRightAnd should be true")
	}
	if Right[string, int](4).IsRightAnd(func(v int) bool { return v == 5 }) {
		t.Fatal("IsRightAnd should be false")
	}
	if Left[string, int]("boom").IsRightAnd(func(v int) bool { return true }) {
		t.Fatal("IsRightAnd on Left should be false")
	}
	if !Left[string, int]("boom").IsLeftAnd(func(v string) bool { return v == "boom" }) {
		t.Fatal("IsLeftAnd should be true")
	}
	if Right[string, int](4).IsLeftAnd(func(v string) bool { return true }) {
		t.Fatal("IsLeftAnd on Right should be false")
	}
}

func TestExtraction(t *testing.T) {
	if Right[string, int](3).Must() != 3 {
		t.Fatal("Must should return contained Right value")
	}
	if Left[string, int]("boom").MustLeft() != "boom" {
		t.Fatal("MustLeft should return contained Left value")
	}
	if Right[string, int](3).GetOrElse(9) != 3 {
		t.Fatal("GetOrElse on Right should return contained value")
	}
	if Left[string, int]("boom").GetOrElse(9) != 9 {
		t.Fatal("GetOrElse on Left should return fallback")
	}
	if Left[string, int]("boom").GetOrElseFunc(func(l string) int { return len(l) }) != 4 {
		t.Fatal("GetOrElseFunc on Left should return computed fallback")
	}
	if v, ok := Right[string, int](3).GetLeft(); ok || v != "" {
		t.Fatal("GetLeft on Right should be zero value, false")
	}
	if v, ok := Left[string, int]("boom").GetRight(); ok || v != 0 {
		t.Fatal("GetRight on Left should be zero value, false")
	}
}

func TestMustPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Must on Left should panic")
		}
	}()
	Left[string, int]("boom").Must()
}

func TestMustLeftPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustLeft on Right should panic")
		}
	}()
	Right[string, int](3).MustLeft()
}

func TestMustMsgPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r != "custom" {
			t.Fatalf("expected panic message 'custom', got %v", r)
		}
	}()
	Left[string, int]("boom").MustMsg("custom")
}

func TestTransformation(t *testing.T) {
	if v, _ := Right[string, int](3).Map(func(i int) int { return i * 2 }).GetRight(); v != 6 {
		t.Fatal("Map should transform Right")
	}
	if Left[string, int]("boom").Map(func(i int) int { return i * 2 }).IsRight() {
		t.Fatal("Map on Left should stay Left")
	}
	if v, _ := Left[string, int]("boom").MapLeft(func(s string) int { return len(s) }).GetLeft(); v != 4 {
		t.Fatal("MapLeft should transform Left")
	}
	if !Right[string, int](3).MapLeft(func(s string) int { return len(s) }).IsRight() {
		t.Fatal("MapLeft on Right should stay Right")
	}

	if v, _ := Right[string, int](3).FlatMap(func(i int) Either[string, int] { return Right[string, int](i + 1) }).GetRight(); v != 4 {
		t.Fatal("FlatMap should chain Right")
	}
	if Right[string, int](3).FlatMap(func(i int) Either[string, int] { return Left[string, int]("boom") }).IsRight() {
		t.Fatal("FlatMap returning Left should propagate Left")
	}
	if v, _ := Left[string, int]("boom").FlatMapLeft(func(s string) Either[int, int] { return Left[int, int](len(s)) }).GetLeft(); v != 4 {
		t.Fatal("FlatMapLeft should chain Left")
	}
	if !Right[string, int](3).FlatMapLeft(func(s string) Either[int, int] { return Left[int, int](len(s)) }).IsRight() {
		t.Fatal("FlatMapLeft on Right should stay Right")
	}

	if Right[string, int](3).Fold(func(s string) int { return -1 }, func(i int) int { return i * 2 }) != 6 {
		t.Fatal("Fold on Right should apply onRight")
	}
	if Left[string, int]("boom").Fold(func(s string) int { return len(s) }, func(i int) int { return i * 2 }) != 4 {
		t.Fatal("Fold on Left should apply onLeft")
	}

	if v, _ := Right[string, int](3).Swap().GetLeft(); v != 3 {
		t.Fatal("Swap on Right should produce Left with the same value")
	}
	if v, _ := Left[string, int]("boom").Swap().GetRight(); v != "boom" {
		t.Fatal("Swap on Left should produce Right with the same value")
	}
}

func TestFilterForEach(t *testing.T) {
	if Right[string, int](3).FilterOrElse(func(i int) bool { return i > 5 }, func() string { return "too small" }).IsRight() {
		t.Fatal("FilterOrElse should reject non-matching value")
	}
	if v, _ := Right[string, int](3).FilterOrElse(func(i int) bool { return i > 1 }, func() string { return "too small" }).GetRight(); v != 3 {
		t.Fatal("FilterOrElse should keep matching value")
	}
	if v, _ := Left[string, int]("boom").FilterOrElse(func(i int) bool { return false }, func() string { return "too small" }).GetLeft(); v != "boom" {
		t.Fatal("FilterOrElse on Left should leave it unchanged")
	}

	called := false
	Right[string, int](3).ForEach(func(i int) { called = true })
	if !called {
		t.Fatal("ForEach should call f on Right")
	}
	called = false
	Left[string, int]("boom").ForEach(func(i int) { called = true })
	if called {
		t.Fatal("ForEach should not call f on Left")
	}

	called = false
	Left[string, int]("boom").ForEachLeft(func(s string) { called = true })
	if !called {
		t.Fatal("ForEachLeft should call f on Left")
	}
	called = false
	Right[string, int](3).ForEachLeft(func(s string) { called = true })
	if called {
		t.Fatal("ForEachLeft should not call f on Right")
	}
}

func TestCombining(t *testing.T) {
	if v, _ := Right[string, int](1).OrElse(Right[string, int](2)).GetRight(); v != 1 {
		t.Fatal("OrElse on Right should return self")
	}
	if v, _ := Left[string, int]("boom").OrElse(Right[string, int](2)).GetRight(); v != 2 {
		t.Fatal("OrElse on Left should return other")
	}
	if v, _ := Left[string, int]("boom").OrElseFunc(func(s string) Either[string, int] { return Right[string, int](len(s)) }).GetRight(); v != 4 {
		t.Fatal("OrElseFunc on Left should return computed fallback")
	}
	if v, _ := Right[string, int](1).OrElseFunc(func(s string) Either[string, int] { return Right[string, int](2) }).GetRight(); v != 1 {
		t.Fatal("OrElseFunc on Right should return self")
	}
}

func TestSliceString(t *testing.T) {
	if s := Right[string, int](3).Slice(); len(s) != 1 || s[0] != 3 {
		t.Fatal("Slice on Right should return single-element slice")
	}
	if s := Left[string, int]("boom").Slice(); s != nil {
		t.Fatal("Slice on Left should return nil")
	}
	if s := Left[string, int]("boom").LeftSlice(); len(s) != 1 || s[0] != "boom" {
		t.Fatal("LeftSlice on Left should return single-element slice")
	}
	if s := Right[string, int](3).LeftSlice(); s != nil {
		t.Fatal("LeftSlice on Right should return nil")
	}

	if Right[string, int](3).String() != "Right(3)" {
		t.Fatalf("unexpected String() output: %s", Right[string, int](3).String())
	}
	if Left[string, int]("boom").String() != "Left(boom)" {
		t.Fatalf("unexpected String() output: %s", Left[string, int]("boom").String())
	}
}

func TestFreeFunctions(t *testing.T) {
	errBoom := errors.New("boom")
	if v, err := ToErr(Right[error, int](3)); v != 3 || err != nil {
		t.Fatal("ToErr on Right should return value with nil error")
	}
	if v, err := ToErr(Left[error, int](errBoom)); v != 0 || err != errBoom {
		t.Fatal("ToErr on Left should return zero value with err")
	}

	if !ContainsRight[string](Right[string, int](3), 3) {
		t.Fatal("ContainsRight should be true for matching Right")
	}
	if ContainsRight[string](Right[string, int](3), 4) {
		t.Fatal("ContainsRight should be false for non-matching Right")
	}
	if ContainsRight[string](Left[string, int]("boom"), 3) {
		t.Fatal("ContainsRight should be false for Left")
	}

	if !ContainsLeft[string, int](Left[string, int]("boom"), "boom") {
		t.Fatal("ContainsLeft should be true for matching Left")
	}
	if ContainsLeft[string, int](Left[string, int]("boom"), "bang") {
		t.Fatal("ContainsLeft should be false for non-matching Left")
	}
	if ContainsLeft[string, int](Right[string, int](3), "boom") {
		t.Fatal("ContainsLeft should be false for Right")
	}
}

func TestEqual(t *testing.T) {
	if !Equal(Right[string, int](1), Right[string, int](1)) {
		t.Fatal("Equal(Right(1), Right(1)) should be true")
	}
	if Equal(Right[string, int](1), Right[string, int](2)) {
		t.Fatal("Equal(Right(1), Right(2)) should be false")
	}
	if !Equal(Left[string, int]("boom"), Left[string, int]("boom")) {
		t.Fatal("Equal(Left, Left) with same value should be true")
	}
	if Equal(Left[string, int]("boom"), Left[string, int]("bang")) {
		t.Fatal("Equal(Left, Left) with different values should be false")
	}
	if Equal(Right[string, int](1), Left[string, int]("boom")) {
		t.Fatal("Equal(Right, Left) should be false")
	}
}
