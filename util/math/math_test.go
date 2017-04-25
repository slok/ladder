package math

import (
	"testing"
)

func TestRoundInt64(t *testing.T) {
	tests := []struct {
		num      float64
		expected int64
	}{
		{1.0, 1},
		{5.1, 5},
		{5.5, 6},
		{5.51, 6},
		{5.49, 5},
		{-5.9, -6},
		{-5.499999, -5},
		{0, 0},
	}

	for _, test := range tests {
		got := RoundInt64(test.num)

		if got != test.expected {
			t.Errorf("- %+v\n -Expected %d, got %d", test, test.expected, got)
		}
	}
}
