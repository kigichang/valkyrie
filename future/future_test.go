package future

import (
	"errors"
	"testing"
	"time"

	"github.com/kigichang/valkyrie/result"
)

var errBoom = errors.New("boom")

const testTimeout = 2 * time.Second

func TestConstructors(t *testing.T) {
	if v, err := Successful(1).Wait(); v != 1 || err != nil {
		t.Fatal("Successful should complete with the value")
	}
	if _, err := Failed[int](errBoom).Wait(); err != errBoom {
		t.Fatal("Failed should complete with the error")
	}
	if v, err := New(func() int { return 5 }).Wait(); v != 5 || err != nil {
		t.Fatal("New should complete with the function's return value")
	}
	if _, err := New[int](func() int { panic("boom") }).Wait(); err == nil {
		t.Fatal("New should recover a panic as an error")
	}
	if v, err := Go(func() (int, error) { return 7, nil }).Wait(); v != 7 || err != nil {
		t.Fatal("Go should complete with the function's value")
	}
	if _, err := Go(func() (int, error) { return 0, errBoom }).Wait(); err != errBoom {
		t.Fatal("Go should complete with the function's error")
	}
	if v, err := FromResult(result.Ok(9)).Wait(); v != 9 || err != nil {
		t.Fatal("FromResult should complete with the Result's value")
	}
	if Never[int]().Ready(10 * time.Millisecond) {
		t.Fatal("Never should not complete")
	}
}

func TestQueriesAndWaiting(t *testing.T) {
	f := Successful(3)
	if !f.IsCompleted() {
		t.Fatal("Successful should be IsCompleted")
	}
	v, ok := f.Value().Get()
	if !ok {
		t.Fatal("Value should be Some once completed")
	}
	if got, _ := v.Get(); got != 3 {
		t.Fatal("Value should hold the completed Result")
	}

	pending := Go(func() (int, error) {
		time.Sleep(50 * time.Millisecond)
		return 1, nil
	})
	if pending.IsCompleted() {
		t.Fatal("pending future should not be IsCompleted immediately")
	}
	if pending.Value().IsSome() {
		t.Fatal("pending future's Value should be None")
	}
	if !pending.Ready(testTimeout) {
		t.Fatal("Ready should report completion within the timeout")
	}

	if _, _, ok := Never[int]().WaitTimeout(10 * time.Millisecond); ok {
		t.Fatal("WaitTimeout should time out on a Future that never completes")
	}
	if _, _, ok := Successful(1).WaitTimeout(testTimeout); !ok {
		t.Fatal("WaitTimeout should report completion for an already-completed Future")
	}
}

func TestCallbacks(t *testing.T) {
	done := make(chan int, 1)
	Successful(4).OnComplete(func(r result.Result[int]) {
		v, _ := r.Get()
		done <- v
	})
	select {
	case v := <-done:
		if v != 4 {
			t.Fatalf("OnComplete callback got %d, want 4", v)
		}
	case <-time.After(testTimeout):
		t.Fatal("OnComplete callback never ran")
	}

	fe := make(chan int, 1)
	Successful(6).Foreach(func(v int) { fe <- v })
	select {
	case v := <-fe:
		if v != 6 {
			t.Fatalf("Foreach got %d, want 6", v)
		}
	case <-time.After(testTimeout):
		t.Fatal("Foreach callback never ran")
	}

	sideEffect := make(chan struct{}, 1)
	out := Successful(1).AndThen(func(r result.Result[int]) { close(sideEffect) })
	if v, err := out.Wait(); v != 1 || err != nil {
		t.Fatal("AndThen should pass through the original result")
	}
	select {
	case <-sideEffect:
	case <-time.After(testTimeout):
		t.Fatal("AndThen side effect never ran")
	}
}

func TestTransformation(t *testing.T) {
	if v, _ := Successful(2).Map(func(v int) int { return v * 10 }).Wait(); v != 20 {
		t.Fatal("Map should transform the value")
	}
	if _, err := Failed[int](errBoom).Map(func(v int) int { return v * 10 }).Wait(); err != errBoom {
		t.Fatal("Map should propagate the error unchanged")
	}

	fm := Successful(2).FlatMap(func(v int) *Future[string] {
		return Successful("ok")
	})
	if v, _ := fm.Wait(); v != "ok" {
		t.Fatal("FlatMap should flatten the nested Future")
	}
	if _, err := Failed[int](errBoom).FlatMap(func(v int) *Future[string] {
		return Successful("unreached")
	}).Wait(); err != errBoom {
		t.Fatal("FlatMap should propagate the error unchanged")
	}

	tr := Successful(2).Transform(func(r result.Result[int]) result.Result[string] {
		v, ok := r.Get()
		if !ok {
			return result.Err[string](r.UnwrapErr())
		}
		if v == 2 {
			return result.Ok("two")
		}
		return result.Err[string](errBoom)
	})
	if v, err := tr.Wait(); err != nil || v != "two" {
		t.Fatal("Transform should apply fn to the Result")
	}

	tw := Successful(2).TransformWith(func(r result.Result[int]) *Future[int] {
		v, _ := r.Get()
		return Successful(v + 1)
	})
	if v, _ := tw.Wait(); v != 3 {
		t.Fatal("TransformWith should flatten fn's Future")
	}
}

func TestFilter(t *testing.T) {
	if v, err := Successful(4).Filter(func(v int) bool { return v == 4 }).Wait(); v != 4 || err != nil {
		t.Fatal("Filter should keep a matching value")
	}
	if _, err := Successful(4).Filter(func(v int) bool { return v == 5 }).Wait(); err != ErrFilterFalse {
		t.Fatal("Filter should fail with ErrFilterFalse for a non-matching value")
	}
	if _, err := Failed[int](errBoom).Filter(func(v int) bool { return true }).Wait(); err != errBoom {
		t.Fatal("Filter should leave an already-failed Future unchanged")
	}
}

func TestRecovery(t *testing.T) {
	if v, err := Failed[int](errBoom).Recover(func(err error) (int, bool) {
		return 42, true
	}).Wait(); v != 42 || err != nil {
		t.Fatal("Recover should replace the error with fn's value")
	}
	if _, err := Failed[int](errBoom).Recover(func(err error) (int, bool) {
		return 0, false
	}).Wait(); err != errBoom {
		t.Fatal("Recover should leave the Future failed if fn declines")
	}
	if v, err := Successful(1).Recover(func(err error) (int, bool) {
		return 42, true
	}).Wait(); v != 1 || err != nil {
		t.Fatal("Recover should not affect a successful Future")
	}

	if v, err := Failed[int](errBoom).RecoverWith(func(err error) (*Future[int], bool) {
		return Successful(9), true
	}).Wait(); v != 9 || err != nil {
		t.Fatal("RecoverWith should replace the Future with fn's Future")
	}
	if _, err := Failed[int](errBoom).RecoverWith(func(err error) (*Future[int], bool) {
		return nil, false
	}).Wait(); err != errBoom {
		t.Fatal("RecoverWith should leave the Future failed if fn declines")
	}

	if v, err := Failed[int](errBoom).FallbackTo(Successful(11)).Wait(); v != 11 || err != nil {
		t.Fatal("FallbackTo should use other's value when the Future fails")
	}
	if _, err := Failed[int](errBoom).FallbackTo(Failed[int](errors.New("other"))).Wait(); err != errBoom {
		t.Fatal("FallbackTo should keep the original error if both fail")
	}
	if v, err := Successful(1).FallbackTo(Successful(11)).Wait(); v != 1 || err != nil {
		t.Fatal("FallbackTo should not affect a successful Future")
	}
}

func TestZip(t *testing.T) {
	a := Successful(1)
	b := Successful("x")
	z := Zip(a, b)
	tup, err := z.Wait()
	if err != nil || tup.V1() != 1 || tup.V2() != "x" {
		t.Fatal("Zip should combine both values into a tuple")
	}

	if _, err := Zip(Failed[int](errBoom), Successful("x")).Wait(); err != errBoom {
		t.Fatal("Zip should fail if either Future fails")
	}

	zw := ZipWith(Successful(2), Successful(3), func(a, b int) int { return a + b })
	if v, _ := zw.Wait(); v != 5 {
		t.Fatal("ZipWith should combine both values with fn")
	}
}

func TestFreeFunctions(t *testing.T) {
	fs := []*Future[int]{Successful(1), Successful(2), Successful(3)}
	seq, err := Sequence(fs).Wait()
	if err != nil || len(seq) != 3 || seq[0] != 1 || seq[1] != 2 || seq[2] != 3 {
		t.Fatal("Sequence should collect every value in order")
	}

	failing := []*Future[int]{Successful(1), Failed[int](errBoom)}
	if _, err := Sequence(failing).Wait(); err != errBoom {
		t.Fatal("Sequence should fail with the first error")
	}

	tv, err := Traverse([]int{1, 2, 3}, func(v int) *Future[int] {
		return Successful(v * 2)
	}).Wait()
	if err != nil || len(tv) != 3 || tv[0] != 2 || tv[2] != 6 {
		t.Fatal("Traverse should map then combine")
	}

	fc, err := FirstCompletedOf([]*Future[int]{
		Go(func() (int, error) { time.Sleep(50 * time.Millisecond); return 1, nil }),
		Successful(2),
	}).Wait()
	if err != nil || fc != 2 {
		t.Fatal("FirstCompletedOf should complete with the fastest Future's result")
	}

	found, err := Find([]*Future[int]{Successful(1), Successful(2), Successful(3)}, func(v int) bool {
		return v == 2
	}).Wait()
	if err != nil {
		t.Fatal("Find should not fail")
	}
	if v, ok := found.Get(); !ok || v != 2 {
		t.Fatal("Find should return the matching value")
	}
	notFound, _ := Find([]*Future[int]{Successful(1), Successful(3)}, func(v int) bool {
		return v == 2
	}).Wait()
	if notFound.IsSome() {
		t.Fatal("Find should return None if no value matches")
	}

	sum, err := FoldLeft([]*Future[int]{Successful(1), Successful(2), Successful(3)}, 10,
		func(acc, v int) int { return acc + v }).Wait()
	if err != nil || sum != 16 {
		t.Fatal("FoldLeft should accumulate starting from zero")
	}

	product, err := ReduceLeft([]*Future[int]{Successful(2), Successful(3), Successful(4)},
		func(a, b int) int { return a * b }).Wait()
	if err != nil || product != 24 {
		t.Fatal("ReduceLeft should combine starting from the first value")
	}
	if _, err := ReduceLeft([]*Future[int]{}, func(a, b int) int { return a }).Wait(); err != ErrReduceLeftEmpty {
		t.Fatal("ReduceLeft should fail with ErrReduceLeftEmpty when empty")
	}

	inner := Successful(Successful(5))
	if v, err := Flatten(inner).Wait(); err != nil || v != 5 {
		t.Fatal("Flatten should unwrap the nested Future")
	}
}

func TestString(t *testing.T) {
	if Successful(1).String() != "Future(Ok(1))" {
		t.Fatalf("unexpected String() for a completed Future: %s", Successful(1).String())
	}
	if Never[int]().String() != "Future(<pending>)" {
		t.Fatal("String() should report <pending> for an incomplete Future")
	}
}
