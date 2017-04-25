package common

import (
	"context"
	"testing"
	"time"

	"github.com/themotion/ladder/types"
)

func TestScalingKindIntervalCreation(t *testing.T) {
	tests := []struct {
		upDur   string
		downDur string

		wantUpDur   time.Duration
		wantDownDur time.Duration
		correct     bool
	}{
		{
			upDur:       "30s",
			downDur:     "3m",
			wantUpDur:   30 * time.Second,
			wantDownDur: 3 * time.Minute,
			correct:     true,
		},
		{
			upDur:       "30s",
			downDur:     "3m15s",
			wantUpDur:   30 * time.Second,
			wantDownDur: 3*time.Minute + 15*time.Second,
			correct:     true,
		},
		{
			upDur:       "30g",
			downDur:     "3m",
			wantUpDur:   0 * time.Second,
			wantDownDur: 0 * time.Second,
			correct:     false,
		},
		{
			upDur:       "30s",
			downDur:     "wrong",
			wantUpDur:   0 * time.Second,
			wantDownDur: 0 * time.Second,
			correct:     false,
		},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			skiUpDurationOpt:   test.upDur,
			skiDownDurationOpt: test.downDur,
		}

		ski, err := NewScalingKindInterval(context.TODO(), opts)
		if test.correct {
			if err != nil {
				t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
				return
			}

			if ski.upDuration != test.wantUpDur ||
				ski.downDuration != test.wantDownDur {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object", test)
			}
		}

		if !test.correct && err == nil {
			t.Errorf("\n- %+v\n  Creation should give error, it didn't", test)
		}

	}
}

func TestScalingKindIntervalFilter(t *testing.T) {
	tests := []struct {
		upDur       string
		downDur     string
		mode        types.ScalingMode
		modeStarted time.Time
		currentQ    int64
		newQ        int64

		wantQ            int64
		wantMode         types.ScalingMode
		wantModeTChanged bool
	}{
		// From notscaling to upscaling mode, butnot trigering scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.NotScaling,
			modeStarted:      time.Now().UTC().Add(-10 * time.Second), // Doesn't matter should change
			currentQ:         10,
			newQ:             15,
			wantQ:            10,
			wantMode:         types.ScalingUp,
			wantModeTChanged: true,
		},
		// From notscaling to downscaling mode, but not triggering scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.NotScaling,
			modeStarted:      time.Now().UTC().Add(-10 * time.Second), // Doesn't matter should change
			currentQ:         20,
			newQ:             15,
			wantQ:            20,
			wantMode:         types.ScalingDown,
			wantModeTChanged: true,
		},
		// being in upscaling mode, required time passed, trigger scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.ScalingUp,
			modeStarted:      time.Now().UTC().Add(-35 * time.Second),
			currentQ:         10,
			newQ:             15,
			wantQ:            15,
			wantMode:         types.ScalingUp,
			wantModeTChanged: false,
		},
		// being in upscaling mode, not required time passed, don't trigger scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.ScalingUp,
			modeStarted:      time.Now().UTC().Add(-25 * time.Second),
			currentQ:         10,
			newQ:             15,
			wantQ:            10,
			wantMode:         types.ScalingUp,
			wantModeTChanged: false,
		},
		// being in downscaling mode, required time passed, trigger scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.ScalingDown,
			modeStarted:      time.Now().UTC().Add(-6 * time.Minute),
			currentQ:         20,
			newQ:             15,
			wantQ:            15,
			wantMode:         types.ScalingDown,
			wantModeTChanged: false,
		},
		// being in downscaling mode, not required time passed, don't trigger scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.ScalingDown,
			modeStarted:      time.Now().UTC().Add(-4 * time.Minute),
			currentQ:         20,
			newQ:             15,
			wantQ:            20,
			wantMode:         types.ScalingDown,
			wantModeTChanged: false,
		},
		// being in downscaling mode pass to upscaling mode, not required time passed, don't trigger scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.ScalingDown,
			modeStarted:      time.Now().UTC().Add(-10 * time.Second), // Doesn't matter should change
			currentQ:         10,
			newQ:             15,
			wantQ:            10,
			wantMode:         types.ScalingUp,
			wantModeTChanged: true,
		},
		// being in upscaling mode pass to downscaling mode, not required time passed, don't trigger scalation
		{
			upDur:            "30s",
			downDur:          "5m",
			mode:             types.ScalingUp,
			modeStarted:      time.Now().UTC().Add(-10 * time.Second), // Doesn't matter should change
			currentQ:         15,
			newQ:             10,
			wantQ:            15,
			wantMode:         types.ScalingDown,
			wantModeTChanged: true,
		},
		// being in downscaling mode pass to upscaling mode, required time passed, trigger scalation
		{
			upDur:            "0s",
			downDur:          "5m",
			mode:             types.ScalingDown,
			modeStarted:      time.Now().UTC().Add(-10 * time.Second), // Doesn't matter should change
			currentQ:         10,
			newQ:             15,
			wantQ:            15,
			wantMode:         types.ScalingUp,
			wantModeTChanged: true,
		},
		// being in downscaling mode pass to upscaling mode, required time passed, trigger scalation
		{
			upDur:            "20s",
			downDur:          "0s",
			mode:             types.ScalingUp,
			modeStarted:      time.Now().UTC().Add(-10 * time.Second), // Doesn't matter should change
			currentQ:         15,
			newQ:             10,
			wantQ:            10,
			wantMode:         types.ScalingDown,
			wantModeTChanged: true,
		},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			skiUpDurationOpt:   test.upDur,
			skiDownDurationOpt: test.downDur,
		}

		ski, err := NewScalingKindInterval(context.TODO(), opts)
		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
			return
		}

		// Set correct state
		ski.mode = test.mode
		ski.modeStarted = test.modeStarted

		gotQ, _, err := ski.Filter(context.TODO(), types.Quantity{Q: test.currentQ}, types.Quantity{Q: test.newQ})
		if err != nil {
			t.Errorf("\n- %+v\n  filtering shouldn't give error: %v", test, err)
			return
		}

		if gotQ.Q != test.wantQ {
			t.Errorf("\n- %+v\n  filtering result is wrong; want: %d, got: %d", test, test.wantQ, gotQ.Q)
		}

		if test.wantModeTChanged && test.modeStarted == ski.modeStarted {
			t.Errorf("\n- %+v\n  filtering should change started mode timestamp, it didn't", test)
		}

		if !test.wantModeTChanged && test.modeStarted != ski.modeStarted {
			t.Errorf("\n- %+v\n  filtering shouldn't change started mode timestamp, it did", test)
		}

	}

}
