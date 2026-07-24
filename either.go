package valkyrie

import "github.com/kigichang/valkyrie/either"

type Either[L, R any] = either.Either[L, R]

// Left creates an Either holding a Left value.
func Left[L, R any](value L) Either[L, R] {
	return either.Left[L, R](value)
}

// Right creates an Either holding a Right value.
func Right[L, R any](value R) Either[L, R] {
	return either.Right[L, R](value)
}

// Cond creates a Right containing right if test is true, or a Left containing left otherwise.
func Cond[L, R any](test bool, left L, right R) Either[L, R] {
	return either.Cond(test, left, right)
}

// CondFunc creates a Right containing the result of calling right if test is true, or a Left
// containing the result of calling left otherwise.
func CondFunc[L, R any](test bool, left func() L, right func() R) Either[L, R] {
	return either.CondFunc(test, left, right)
}
