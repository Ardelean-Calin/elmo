package common

// General-purpose iterator interface
// Workaround until GOEXPERIMENT=rangefunc is merged
type Iterator[T any] interface {
	Next() *T
	HasNext() bool
}

func Clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
