package math

// RoundInt64 rounds floats into integer numbers.
func RoundInt64(v float64) int64 {
	if v < 0 {
		return int64(v - 0.5)
	}
	return int64(v + 0.5)

}
