package autoscaler

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

func TestCorrectIntervalAutoscalerCreation(t *testing.T) {
	// Register all dummies
	registerDummies()

	tests := []struct {
		config *config.Autoscaler
		dryRun bool
	}{
		{
			config: &config.Autoscaler{
				Name:        "autoscaler-test",
				Description: "autoscaler-test description",
				Interval:    1 * time.Minute,
				Scale:       config.Block{Kind: "test0"},
				Solve:       config.Block{Kind: "test0"},
				Inputters: []config.Inputter{
					config.Inputter{Name: "test0_input", Gather: config.Block{Kind: "test0"}, Arrange: config.Block{Kind: "test0"}},
				},
			},
			dryRun: false,
		},
		{
			config: &config.Autoscaler{
				Name:        "autoscaler-test2",
				Description: "autoscaler-test2 description",
				Interval:    1 * time.Minute,
				Scale:       config.Block{Kind: "test0"},
				Solve:       config.Block{Kind: "test0"},
				Inputters: []config.Inputter{
					config.Inputter{Name: "test0_input", Gather: config.Block{Kind: "test0"}, Arrange: config.Block{Kind: "test0"}},
				},
			},
			dryRun: true,
		},
	}

	for _, test := range tests {
		a, err := NewIntervalAutoscaler(test.config, test.dryRun)

		if err != nil {
			t.Fatalf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}
		if a.Config != test.config || a.Name != test.config.Name ||
			a.Description != test.config.Description || a.DryRun != test.dryRun ||
			a.Interval != a.Config.Interval || a.ctx.Value(AutoscalerCtxKey).(string) != a.Name {
			want := &IntervalAutoscaler{
				Config:      test.config,
				Name:        test.config.Name,
				Description: test.config.Description,
				Interval:    test.config.Interval,
				DryRun:      test.dryRun,
			}
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %+v; got %+v", test, *want, *a)
		}
	}
}

func TestWrongIntervalAutoscalerCreation(t *testing.T) {

	tests := []struct {
		config *config.Autoscaler
		dryRun bool
	}{
		{
			config: nil,
			dryRun: false,
		},
		{
			config: &config.Autoscaler{
				Description: "autoscaler-test2 description",
				Interval:    14 * time.Second,
			},
			dryRun: true,
		},
		{
			config: &config.Autoscaler{
				Description: "autoscaler-test3 description",
				Interval:    14 * time.Second,
				Name:        "test3",
			},
			dryRun: true,
		},
		{
			config: &config.Autoscaler{
				Description: "autoscaler-test4 description",
				Interval:    14 * time.Second,
				Name:        "test4",
				Scale:       config.Block{Kind: "test0"},
			},
			dryRun: true,
		},
		{
			config: &config.Autoscaler{
				Description: "autoscaler-test4 description",
				Interval:    14 * time.Second,
				Name:        "test4",
				Scale:       config.Block{Kind: "test0"},
				Solve:       config.Block{Kind: "test0"},
			},
			dryRun: true,
		},
	}

	for _, test := range tests {
		_, err := NewIntervalAutoscaler(test.config, test.dryRun)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give an error", test)
		}

	}
}

func TestCorrectSolve(t *testing.T) {
	tests := []struct {
		inputs []int64
		want   int64
	}{
		// The solver will add all the values of all the inputters
		{inputs: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, want: 100},
		{inputs: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10}, want: 90},
		{inputs: []int64{10, 10, 10, 10, 10, 10, 10, 10}, want: 80},
		{inputs: []int64{10, 10, 10, 10, 10, 10, 10}, want: 70},
		{inputs: []int64{10, 10, 10, 10, 10, 10}, want: 60},
		{inputs: []int64{10, 10, 10, 10, 10}, want: 50},
		{inputs: []int64{10, 10, 10, 10}, want: 40},
		{inputs: []int64{10, 10, 10}, want: 30},
		{inputs: []int64{10, 10}, want: 20},
		{inputs: []int64{10}, want: 10},
	}

	for _, test := range tests {

		a := IntervalAutoscaler{
			Name:      "test",
			Inputters: []inputter{inputter{}, inputter{}},
			Solver:    &testSolver{},
			log:       log.New(),
		}

		inputs := make([]types.Quantity, len(test.inputs))
		errors := []error{}

		for i, in := range test.inputs {
			inputs[i] = types.Quantity{Q: in}
		}

		inQ, err := a.solve(inputs, errors)
		if err != nil {
			t.Errorf("\n- %+v\n  Solve shouldn't give error: %s", test, err)
		}

		if inQ.Q != test.want {
			t.Errorf("\n- %+v\n  result is not correct, got: %v, want: %v", test, inQ.Q, test.want)
		}
	}
}

func TestCorrectSolveSingleInputter(t *testing.T) {
	tests := []struct {
		inputs []int64
		want   int64
	}{
		// The solver will add all the values of all the inputters
		{inputs: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, want: 10},
		{inputs: []int64{10, 10, 10, 10}, want: 10},
	}

	for _, test := range tests {

		a := IntervalAutoscaler{
			Config:    &config.Autoscaler{Scale: config.Block{Kind: "test"}},
			Name:      "test",
			Solver:    &testSolver{},
			Inputters: []inputter{inputter{}},
			log:       log.New(),
		}

		inputs := make([]types.Quantity, len(test.inputs))
		errors := []error{}

		for i, in := range test.inputs {
			inputs[i] = types.Quantity{Q: in}
		}

		inQ, err := a.solve(inputs, errors)
		if err != nil {
			t.Errorf("\n- %+v\n  Solve shouldn't give error: %s", test, err)
		}

		if inQ.Q != test.want {
			t.Errorf("\n- %+v\n  result is not correct, got: %v, want: %v", test, inQ.Q, test.want)
		}
	}

}

func TestCorrectSolveError(t *testing.T) {
	tests := []struct {
		errors []string
		want   string
	}{
		// The solver will add all the values of all the inputters
		{errors: []string{"1", "2", "3", "4", "5"}, want: "solver got 5 errors from inputs: 1; 2; 3; 4; 5;"},
		{errors: []string{"wrong!", "and", "wrong"}, want: "solver got 3 errors from inputs: wrong!; and; wrong;"},
		{errors: []string{"wrong!"}, want: "solver got 1 errors from inputs: wrong!;"},
	}

	for _, test := range tests {
		l := log.New()
		bf := &bytes.Buffer{}
		l.Logger.Out = bf

		a := IntervalAutoscaler{
			Name:      "test",
			Inputters: []inputter{inputter{}, inputter{}},
			Solver:    &testSolver{},
			log:       l,
		}

		inputs := []types.Quantity{}
		errs := make([]error, len(test.errors))

		for i, e := range test.errors {
			errs[i] = errors.New(e)
		}

		_, err := a.solve(inputs, errs)
		if err == nil {
			t.Errorf("\n- %+v\n  Solve should give error, it didn't", test)
		}

		if !strings.Contains(bf.String(), test.want) {
			t.Errorf("\n- %+v\n  Inputter errors didn't log correctly", test)
		}
	}

}

func TestGatherWinningInputSingleInputter(t *testing.T) {
	tests := []struct {
		input   int64
		current int64

		want int64
	}{
		{
			input: 15, current: 9, want: 10,
		},
		{
			input: 5, current: 11, want: 10,
		},
		{
			input: 9, current: 10, want: 10,
		},
	}

	for _, test := range tests {
		a := IntervalAutoscaler{

			Name:   "test",
			Scaler: &testScaler{cQ: types.Quantity{Q: test.current}},
			Config: &config.Autoscaler{
				Scale: config.Block{Kind: "test"},
			},
			log: log.New(),
		}
		i := inputter{
			config: &config.Inputter{
				Gather:  config.Block{Kind: "test"},
				Arrange: config.Block{Kind: "test"},
			},
			name:     "test",
			gatherer: &testGatherer{times: test.input},
			arranger: &testArranger{},
			log:      log.New(),
		}
		a.Inputters = []inputter{i}

		inQ, err := a.gatherWinningInput(types.Quantity{Q: test.current})
		if err != nil {
			t.Errorf("\n- %+v\n  GatherWinningInput shouldn't give error: %s", test, err)
		}

		if inQ.Q != test.want {
			t.Errorf("\n- %+v\n  result is not correct, got: %v, want: %v", test, inQ.Q, test.want)
		}
	}
}

func TestGatherWinningInputMultipleInputterWithSolver(t *testing.T) {

	tests := []struct {
		inputs  []int64
		current int64
		want    int64
	}{
		// gather will add 1 to inputs and pass to arranger
		// Arrangers will compare with current, if higher than current then add 1 to current, if lesser subtract 1, nothing if equal
		// The solver will add all the values of all the inputters
		{
			inputs:  []int64{9, 9, 9, 9, 9},
			current: 10,
			want:    50,
		},
		{
			inputs:  []int64{350, 200, 500, 600, 399, 800, 1032, 100, 900, 864, 923},
			current: 400,
			want:    4404,
		},
	}

	for _, test := range tests {
		a := IntervalAutoscaler{
			Name:   "test",
			Scaler: &testScaler{cQ: types.Quantity{Q: test.current}},
			Solver: &testSolver{},
			Config: &config.Autoscaler{
				Scale: config.Block{Kind: "test"},
			},
			log: log.New(),
		}
		inputters := make([]inputter, len(test.inputs))
		for i, in := range test.inputs {
			nIn := inputter{
				config: &config.Inputter{
					Gather:  config.Block{Kind: "test"},
					Arrange: config.Block{Kind: "test"},
				},
				name:     "test",
				gatherer: &testGatherer{times: in},
				arranger: &testArranger{},
				log:      log.New(),
			}

			inputters[i] = nIn
		}

		a.Inputters = inputters

		inQ, err := a.gatherWinningInput(types.Quantity{Q: test.current})
		if err != nil {
			t.Errorf("\n- %+v\n  GatherWinningInput shouldn't give error: %s", test, err)
		}

		if inQ.Q != test.want {
			t.Errorf("\n- %+v\n  result is not correct, got: %v, want: %v", test, inQ.Q, test.want)
		}
	}
}

func TestGatherWinningInputSingleInputterError(t *testing.T) {
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
		a := IntervalAutoscaler{
			Name:   "test",
			Scaler: &testScaler{cQ: types.Quantity{Q: test.current}},
			Config: &config.Autoscaler{
				Scale: config.Block{Kind: "test"},
			},
			log: log.New(),
		}
		i := inputter{
			config:   &config.Inputter{},
			name:     "test",
			gatherer: &testGatherer{times: test.input, retError: test.gatherError},
			arranger: &testArranger{retError: test.arrangeError},
			log:      log.New(),
		}
		a.Inputters = []inputter{i}

		_, err := a.gatherWinningInput(types.Quantity{Q: test.current})
		if err == nil {
			t.Errorf("\n- %+v\n  GatherWinningInput should give error, it didn't", test)
		}
	}
}

func TestCorrectFilter(t *testing.T) {
	tests := []struct {
		currentQ  types.Quantity
		newQ      types.Quantity
		filterers []int64

		want int64
	}{
		{
			currentQ:  types.Quantity{Q: 1000},
			newQ:      types.Quantity{Q: 1000},
			filterers: []int64{1, 2, 3, 4, 5, 6},
			want:      1021,
		},
		{
			currentQ:  types.Quantity{Q: 1000},
			newQ:      types.Quantity{Q: 5000},
			filterers: []int64{1, 9, 4, 1, 0},
			want:      5015,
		},
		{
			currentQ:  types.Quantity{Q: 5000},
			newQ:      types.Quantity{Q: 1000},
			filterers: []int64{},
			want:      1000,
		},
	}

	for _, test := range tests {
		a := IntervalAutoscaler{
			Name: "test",
			Config: &config.Autoscaler{
				Scale: config.Block{Kind: "test"},
			},
			log: log.New(),
		}

		// create filterers
		fs := make([]filter.Filterer, len(test.filterers))
		for i, f := range test.filterers {
			fs[i] = &testFilterer{resAdd: f}
		}
		a.Filterers = fs

		res, err := a.filter(test.currentQ, test.newQ)
		if err != nil {
			t.Errorf("\n- %+v\n  filter shouldn't give error: %s", test, err)
		}

		if res.Q != test.want {
			t.Errorf("\n- %+v\n  filter wrong result; want:%d, got: %d", test, test.want, res.Q)
		}
	}
}

func TestFilterBreak(t *testing.T) {
	tests := []struct {
		currentQ  types.Quantity
		newQ      types.Quantity
		filterers []int64
		breakOn   int

		want int64
	}{
		{
			currentQ:  types.Quantity{Q: 1000},
			newQ:      types.Quantity{Q: 1000},
			filterers: []int64{1, 2, 3, 4, 5, 6},
			breakOn:   3,
			want:      1010,
		},
		{
			currentQ:  types.Quantity{Q: 1000},
			newQ:      types.Quantity{Q: 5000},
			filterers: []int64{1, 9, 4, 1, 0},
			breakOn:   1,
			want:      5010,
		},
		{
			currentQ:  types.Quantity{Q: 5000},
			newQ:      types.Quantity{Q: 1000},
			filterers: []int64{},
			want:      1000,
		},
	}

	for _, test := range tests {
		a := IntervalAutoscaler{
			Name: "test",
			Config: &config.Autoscaler{
				Scale: config.Block{Kind: "test"},
			},
			log: log.New(),
		}

		// create filterers
		fs := make([]filter.Filterer, len(test.filterers))
		for i, f := range test.filterers {
			var br bool
			// Break
			if test.breakOn == i {
				br = true
			}
			fs[i] = &testFilterer{resAdd: f, br: br}
		}
		a.Filterers = fs

		res, err := a.filter(test.currentQ, test.newQ)
		if err != nil {
			t.Errorf("\n- %+v\n  filter shouldn't give error: %s", test, err)
		}

		if res.Q != test.want {
			t.Errorf("\n- %+v\n  filter wrong result; want:%d, got: %d", test, test.want, res.Q)
		}
	}
}

func TestFilterError(t *testing.T) {
	tests := []struct {
		currentQ  types.Quantity
		newQ      types.Quantity
		filterers []int64
		errorOn   int

		want int64
	}{
		{
			currentQ:  types.Quantity{Q: 1000},
			newQ:      types.Quantity{Q: 1000},
			filterers: []int64{1, 2, 3, 4, 5, 6},
			errorOn:   3,
			want:      1010,
		},
		{
			currentQ:  types.Quantity{Q: 1000},
			newQ:      types.Quantity{Q: 5000},
			filterers: []int64{1, 9, 4, 1, 0},
			errorOn:   1,
			want:      5010,
		},
		{
			currentQ:  types.Quantity{Q: 5000},
			newQ:      types.Quantity{Q: 1000},
			filterers: []int64{1, 9, 4, 1, 0},
			errorOn:   1,
			want:      1010,
		},
	}

	for _, test := range tests {
		a := IntervalAutoscaler{
			Name: "test",
			Config: &config.Autoscaler{
				Scale: config.Block{Kind: "test"},
			},
			log: log.New(),
		}

		// create filterers
		fs := make([]filter.Filterer, len(test.filterers))
		for i, f := range test.filterers {
			var erN bool
			// Break
			if test.errorOn == i {
				erN = true
			}
			fs[i] = &testFilterer{resAdd: f, retError: erN}
		}
		a.Filterers = fs

		res, err := a.filter(test.currentQ, test.newQ)
		if err == nil {
			t.Errorf("\n- %+v\n  filter should give error, it didn't", test)
		}

		// Check that the error was returned when breaking the chain
		if res.Q != test.want {
			t.Errorf("\n- %+v\n  filter wrong result; want:%d, got: %d", test, test.want, res.Q)
		}
	}

}

func TestCorrectScaler(t *testing.T) {
	tests := []struct {
		currentQ types.Quantity
		newQ     types.Quantity
		wantMode types.ScalingMode
	}{
		{
			currentQ: types.Quantity{Q: 1000},
			newQ:     types.Quantity{Q: 1000},
			wantMode: types.NotScaling,
		},
		{
			currentQ: types.Quantity{Q: 999},
			newQ:     types.Quantity{Q: 1000},
			wantMode: types.ScalingUp,
		},
		{
			currentQ: types.Quantity{Q: 1000},
			newQ:     types.Quantity{Q: 999},
			wantMode: types.ScalingDown,
		},
	}

	for _, test := range tests {
		a := &IntervalAutoscaler{
			Name:   "test",
			Config: &config.Autoscaler{},
			Scaler: &testScaler{cQ: test.currentQ},
			log:    log.New(),
		}

		mode, err := a.scale(test.newQ)

		if err != nil {
			t.Errorf("\n- %+v\n  Scale shouldn' give an error: %v", test, err)
		}

		if mode != test.wantMode {
			t.Errorf("\n- %+v\n  result is not correct, got: %+v, want: %+v", test, mode, test.wantMode)
		}
	}
}

func TestCorrectScalerWait(t *testing.T) {
	tests := []struct {
		currentQ types.Quantity
		newQ     types.Quantity
		wantMode types.ScalingMode
	}{
		{
			currentQ: types.Quantity{Q: 999},
			newQ:     types.Quantity{Q: 1000},
			wantMode: types.ScalingUp,
		},
		{
			currentQ: types.Quantity{Q: 1000},
			newQ:     types.Quantity{Q: 1001},
			wantMode: types.ScalingUp,
		},
		{
			currentQ: types.Quantity{Q: 1000},
			newQ:     types.Quantity{Q: 999},
			wantMode: types.ScalingDown,
		},
	}

	for _, test := range tests {
		scaler := &testScaler{cQ: test.currentQ}
		a := &IntervalAutoscaler{
			Name:   "test",
			Config: &config.Autoscaler{},
			Scaler: scaler,
			log:    log.New(),
		}

		mode, err := a.scale(test.newQ)

		if err != nil {
			t.Errorf("\n- %+v\n  Scale shouldn' give an error: %v", test, err)
		}

		if mode != test.wantMode {
			t.Errorf("\n- %+v\n  result is not correct, got: %+v, want: %+v", test, mode, test.wantMode)
		}

		if !scaler.calledWait {
			t.Errorf("\n- %+v\n  Should be called scaler wait, it didn't", test)
		}

		if test.newQ.Q != scaler.calledWaitQ.Q {
			t.Errorf("\n- %+v\n  Scaler wait called with wrong Q param, want: %d; got: %d", test, test.newQ.Q, scaler.calledWaitQ.Q)
		}
	}
}

func TestCorrectScalerWaitTimeout(t *testing.T) {
	tests := []struct {
		currentQ      types.Quantity
		newQ          types.Quantity
		wait          time.Duration
		timeout       time.Duration
		shouldTimeout bool
	}{
		{
			currentQ:      types.Quantity{Q: 1000},
			newQ:          types.Quantity{Q: 1000},
			wait:          10 * time.Millisecond,
			timeout:       50 * time.Millisecond,
			shouldTimeout: false,
		},
		{
			currentQ:      types.Quantity{Q: 1000},
			newQ:          types.Quantity{Q: 1000},
			wait:          60 * time.Millisecond,
			timeout:       50 * time.Millisecond,
			shouldTimeout: false,
		},
		{
			currentQ:      types.Quantity{Q: 999},
			newQ:          types.Quantity{Q: 1000},
			wait:          10 * time.Millisecond,
			timeout:       50 * time.Millisecond,
			shouldTimeout: false,
		},
		{
			currentQ:      types.Quantity{Q: 999},
			newQ:          types.Quantity{Q: 1000},
			wait:          60 * time.Millisecond,
			timeout:       50 * time.Millisecond,
			shouldTimeout: true,
		},
		{
			currentQ:      types.Quantity{Q: 1000},
			newQ:          types.Quantity{Q: 1001},
			wait:          10 * time.Millisecond,
			timeout:       50 * time.Millisecond,
			shouldTimeout: false,
		},
		{
			currentQ:      types.Quantity{Q: 1000},
			newQ:          types.Quantity{Q: 1001},
			wait:          60 * time.Millisecond,
			timeout:       50 * time.Millisecond,
			shouldTimeout: true,
		},
	}

	for _, test := range tests {
		scaler := &testScaler{cQ: test.currentQ, waitDuration: test.wait}
		a := &IntervalAutoscaler{
			Name:               "test",
			Config:             &config.Autoscaler{},
			Scaler:             scaler,
			ScalingWaitTimeout: test.timeout,
			log:                log.New(),
		}

		_, err := a.scale(test.newQ)

		if test.shouldTimeout && err == nil {
			t.Errorf("\n- %+v\n  Scale should error by timeout, it didn't", test)
		}
		if !test.shouldTimeout && err != nil {
			t.Errorf("\n- %+v\n  Scale shouldn't error, it did: %v", test, err)
		}

	}
}
