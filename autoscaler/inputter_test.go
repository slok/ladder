package autoscaler

import (
	"context"
	"testing"

	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

func TestNewInputter(t *testing.T) {
	// Register all dummies
	registerDummies()

	tests := []struct {
		cfg       *config.Inputter
		wantError bool
	}{
		{
			cfg: &config.Inputter{
				Name:    "test0_input",
				Arrange: config.Block{Kind: "test0"},
				Gather:  config.Block{Kind: "test0"}},
			wantError: false,
		},
		{
			cfg: &config.Inputter{
				Name:   "test0_input",
				Gather: config.Block{Kind: "test0"}},
			wantError: false,
		},
		{
			cfg: &config.Inputter{
				Name:    "test0_input",
				Arrange: config.Block{Kind: "test0"}},
			wantError: true,
		},
	}

	for _, test := range tests {
		ctx := context.TODO()
		ctx = context.WithValue(ctx, AutoscalerCtxKey, "test")
		_, err := newInputter(ctx, test.cfg)
		if !test.wantError {
			if err != nil {
				t.Errorf("\n- %+v\n  Inputter creation shouldn't give an error: %v", test, err)
			}
		} else {
			if err == nil {
				t.Errorf("\n- %+v\n  Inputter should give an error, it didn't", test)
			}
		}

	}
}

func TestCorrectInputterGatherAndArrange(t *testing.T) {
	// Testing arranger adds 1 if in is greater, removes 1 if is lesser than current
	// Testing gatherer adds 1 always
	tests := []struct {
		input       int64
		current     int64
		setArranger bool

		want int64
	}{
		{
			input: 15, current: 9, setArranger: true, want: 10,
		},
		{
			input: 5, current: 11, setArranger: true, want: 10,
		},
		{
			input: 9, current: 10, setArranger: true, want: 10,
		},
		{
			input: 15, current: 9, setArranger: false, want: 16,
		},
		{
			input: 5, current: 11, setArranger: false, want: 6,
		},
		{
			input: 9, current: 10, setArranger: false, want: 10,
		},
	}

	for _, test := range tests {
		var ar arrange.Arranger
		if test.setArranger {
			ar = &testArranger{}
		}
		i := &inputter{
			config:   &config.Inputter{},
			name:     "test",
			gatherer: &testGatherer{times: test.input},
			arranger: ar,
			log:      log.New(),
		}

		inQ, err := i.gatherAndArrange(context.TODO(), types.Quantity{Q: test.current})
		if err != nil {
			t.Errorf("\n- %+v\n  Gather shouldn't give error: %s", test, err)
		}

		if inQ.Q != test.want {
			t.Errorf("\n- %+v\n  result is not correct, got: %v, want: %v", test, inQ.Q, test.want)
		}
	}
}

func TestCorrectInputterGatherAndArrangeError(t *testing.T) {
	tests := []struct {
		input        int64
		current      int64
		gatherError  bool
		arrangeError bool

		want int
	}{
		{
			input: 15, current: 9, gatherError: true, arrangeError: false, want: 10,
		},
		{
			input: 5, current: 11, gatherError: true, arrangeError: false, want: 10,
		},
		{
			input: 9, current: 10, gatherError: true, arrangeError: true, want: 10,
		},
	}

	for _, test := range tests {

		i := &inputter{
			config:   &config.Inputter{},
			name:     "test",
			gatherer: &testGatherer{times: test.input, retError: test.gatherError},
			arranger: &testArranger{retError: test.arrangeError},
			log:      log.New(),
		}

		_, err := i.gatherAndArrange(context.TODO(), types.Quantity{Q: test.current})
		if err == nil {
			t.Errorf("\n- %+v\n  Gather should give error, it dind't", test)
		}
	}
}
