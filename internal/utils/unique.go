package utils

import "iter"

func FirstDuplicate[T any, K comparable](seq iter.Seq[T], key func(T) K, of T) (T, bool) {
	want := key(of)

	for other := range seq {
		if key(other) == want {
			return other, true
		}
	}

	var none T

	return none, false
}
