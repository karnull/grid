package collage

import "math"

// computes the smallest near-square (rows, cols) whose product is at least n
func SquareDimensions(n int) (rows, cols int) {
	cols = int(math.Ceil(math.Sqrt(float64(n))))
	rows = int(math.Ceil(float64(n) / float64(cols)))
	return
}
