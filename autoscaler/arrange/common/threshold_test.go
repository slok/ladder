package common

import (
	"context"
	"testing"

	"github.com/themotion/ladder/types"
)

func TestThresholdCreation(t *testing.T) {
	tests := []struct {
		upTh    int64
		downTh  int64
		upPer   int64
		downPer int64
		upMax   int64
		downMax int64
		upMin   int64
		downMin int64
		inverse bool

		correct bool
	}{
		{80, 60, 20, 10, 20, 10, 2, 1, false, true},
		{60, 80, 20, 10, 20, 10, 2, 1, true, true},
		{60, 80, 20, 10, 20, 10, 2, 1, false, false},
		{80, 60, 20, 10, 20, 10, 2, 1, true, false},
		{60, 60, 20, 10, 20, 10, 2, 1, false, false},
		{80, 60, 20, 10, 20, 10, -2, 1, false, false},
		{80, 60, 20, 10, -20, 10, 2, 1, false, false},
		{80, 60, 20, 10, 20, -10, 2, 1, false, false},

		// Check percent thresholds
		{80, 60, 100, 10, 20, 10, 2, 1, false, true},
		{80, 60, 0, 10, 20, 10, 2, 1, false, true},
		{80, 60, -1, 10, 20, 10, 2, 1, false, false},
		{80, 60, 20, 0, 20, 10, 2, 1, false, true},
		{80, 60, 20, 100, 20, 10, 2, 1, false, true},
		{80, 60, 20, -1, 20, 10, 2, 1, false, false},
		{80, 60, 20, 101, 20, 10, 2, 1, false, false},
		{80, 60, 0, 0, 20, 10, 2, 1, false, true},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			thUpThreshold:   test.upTh,
			thDownThreshold: test.downTh,
			thUpPercent:     test.upPer,
			thDownPercent:   test.downPer,
			thUpMax:         test.upMax,
			thDownMax:       test.downMax,
			thUpMin:         test.upMin,
			thDownMin:       test.downMin,
			thInverseMode:   test.inverse,
		}

		th, err := NewThreshold(context.TODO(), opts)
		if test.correct {
			if err != nil {
				t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
				return
			}

			if th.upTh != test.upTh ||
				th.downTh != test.downTh ||
				th.upPercent != test.upPer ||
				th.downPercent != test.downPer ||
				th.upMax != test.upMax ||
				th.downMax != test.downMax ||
				th.upMin != test.upMin ||
				th.downMin != test.downMin ||
				th.inverse != test.inverse {

				t.Errorf("\n- %+v\n  Wrong parameters loaded on object", test)
			}
		}

		if !test.correct && err == nil {
			t.Errorf("\n- %+v\n  Creation should give error, it didn't", test)
		}

	}
}

func TestThresholdArrange(t *testing.T) {
	tests := []struct {
		inputQ   int64
		currentQ int64
		upTh     int64
		downTh   int64
		upPer    int64
		downPer  int64
		upMin    int64
		downMin  int64
		upMax    int64
		downMax  int64
		inverse  bool

		wantQ int64
	}{
		// Don't scale, the input is between the values
		{
			inputQ: 300, currentQ: 10,
			upTh: 400, downTh: 250,
			wantQ: 10,
		},
		// scaleup, the input bypass up threshold, Scales by 200% of 10 = 20 , so 20 + 10 = 30
		{
			inputQ: 500, currentQ: 10,
			upTh: 400, downTh: 250,
			upPer: 200,
			upMax: 50,
			wantQ: 30,
		},
		// scaleup, the input bypass up threshold, Scales by 5% of 10 = 0,5, take minimum: 2 , so 2 + 10 = 12
		{
			inputQ: 500, currentQ: 10,
			upTh: 400, downTh: 250,
			upPer: 10,
			upMin: 2, upMax: 50,
			wantQ: 12,
		},
		// scaleup, the input bypass up threshold, Scales by 100% of 10 = 20, take max: 5 , so 5 + 10 = 15
		{
			inputQ: 500, currentQ: 10,
			upTh: 400, downTh: 250,
			upPer: 100,
			upMax: 5,
			wantQ: 15,
		},
		// scaledown,the input bypass down threshold, Scales by 20% of 10 = 2 , so 10 - 2 = 8
		{
			inputQ: 200, currentQ: 10,
			upTh: 400, downTh: 250,
			downPer: 20,
			downMax: 100,
			wantQ:   8,
		},
		// scaledown, the input bypass down threshold, Scales by 5% of 10 = 0,5, take minimum: 3 , so 10 - 3 = 7
		{
			inputQ: 200, currentQ: 10,
			upTh: 400, downTh: 250,
			downPer: 5,
			downMin: 3, downMax: 100,
			wantQ: 7,
		},
		// scaledown, the input bypass down threshold, Scales by 100% of 10 = 10, take max: 6 , so 10 - 6 = 4
		{
			inputQ: 200, currentQ: 10,
			upTh: 400, downTh: 250,
			downPer: 100,
			downMax: 6,
			wantQ:   4,
		},
		// Inverse mode: Don't scale, the input is between the values
		{
			inputQ: 300, currentQ: 10,
			upTh: 250, downTh: 400,
			wantQ:   10,
			inverse: true,
		},
		// Inverse mode: scaleup, the input bypass up threshold, Scales by 200% of 10 = 20 , so 20 + 10 = 30
		{
			inputQ: 200, currentQ: 10,
			upTh: 250, downTh: 400,
			upPer:   200,
			upMax:   50,
			wantQ:   30,
			inverse: true,
		},
		// Inverse mode: scaledown,the input bypass down threshold, Scales by 20% of 10 = 2 , so 10 - 2 = 8
		{
			inputQ: 500, currentQ: 10,
			upTh: 250, downTh: 400,
			downPer: 20,
			downMax: 100,
			wantQ:   8,
			inverse: true,
		},
		// Fixed value on scaling up: the input bypass up threshold, Scales by 50 , so 50 + 10 = 60
		{
			inputQ: 500, currentQ: 10,
			upTh: 400, downTh: 250,
			upPer: 0,
			upMax: 50,
			upMin: 50,
			wantQ: 60,
		},
		// Fixed value on scaling down: the input bypass down threshold, Scales by 6, so 10 - 6 = 4
		{
			inputQ: 200, currentQ: 10,
			upTh: 400, downTh: 250,
			downPer: 0,
			downMax: 6,
			downMin: 6,
			wantQ:   4,
		},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			thUpThreshold:   test.upTh,
			thDownThreshold: test.downTh,
			thUpPercent:     test.upPer,
			thDownPercent:   test.downPer,
			thUpMax:         test.upMax,
			thDownMax:       test.downMax,
			thUpMin:         test.upMin,
			thDownMin:       test.downMin,
			thInverseMode:   test.inverse,
		}

		th, err := NewThreshold(context.TODO(), opts)
		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
			return
		}

		newQ, err := th.Arrange(context.TODO(), types.Quantity{Q: test.inputQ}, types.Quantity{Q: test.currentQ})
		if err != nil {
			t.Errorf("\n- %+v\n  Arrange shouldn't give error: %v", test, err)
		}

		if newQ.Q != test.wantQ {
			t.Errorf("\n- %+v\n  Arrange returned wrong arrangement output; want: %d; got: %d", test, test.wantQ, newQ.Q)
		}

	}
}
