package autoscaler

import (
	"context"
	"fmt"
	"time"

	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/autoscaler/scale"
	"github.com/themotion/ladder/autoscaler/solve"
	"github.com/themotion/ladder/types"
)

// custom test Gather
type testGatherer struct {
	times    int64
	retError bool
}

func (tg *testGatherer) Gather(_ context.Context) (types.Quantity, error) {
	q := types.Quantity{Q: tg.times}
	if tg.retError {
		return q, fmt.Errorf("Error!")
	}
	tg.times++
	return types.Quantity{Q: tg.times}, nil
}

type testGathererCreator struct{}

func (t *testGathererCreator) Create(_ context.Context, opts map[string]interface{}) (gather.Gatherer, error) {
	e, _ := opts["return_error"].(bool)
	res := &testGatherer{retError: e}
	return res, nil
}

// custom test arrange that will incrementif the input is greater than the current,
// decrement if the input is lower than the current and if equal, nothing
type testArranger struct {
	retError bool
}

func (ta *testArranger) Arrange(_ context.Context, inputQ, currentQ types.Quantity) (types.Quantity, error) {
	res := types.Quantity{Q: currentQ.Q}
	if ta.retError {
		return res, fmt.Errorf("Error!")
	}
	if inputQ.Q > currentQ.Q {
		res.Q++
	} else if inputQ.Q < currentQ.Q {
		res.Q--
	}
	return res, nil
}

type testArrangerCreator struct{}

func (t *testArrangerCreator) Create(_ context.Context, opts map[string]interface{}) (arrange.Arranger, error) {
	e, _ := opts["return_error"].(bool)
	res := &testArranger{retError: e}
	return res, nil
}

type testScaler struct {
	cQ             types.Quantity
	waitDuration   time.Duration
	calledWait     bool
	calledWaitQ    types.Quantity
	calledWaitMode types.ScalingMode

	history []types.Quantity

	currentErr bool
	scaleErr   bool
	waitErr    bool
}

func (ts *testScaler) Current(_ context.Context) (types.Quantity, error) {
	if ts.currentErr {
		return ts.cQ, fmt.Errorf("Error!")
	}
	return ts.cQ, nil
}

func (ts *testScaler) Scale(_ context.Context, newQ types.Quantity) (scaledQ types.Quantity, mode types.ScalingMode, err error) {
	ts.calledWait = false
	mode = types.NotScaling
	if ts.scaleErr {
		return types.Quantity{}, mode, fmt.Errorf("Error!")
	}
	switch {
	case newQ.Q > ts.cQ.Q:
		mode = types.ScalingUp
	case newQ.Q < ts.cQ.Q:
		mode = types.ScalingDown
	default:
		return types.Quantity{}, mode, nil
	}
	ts.cQ = newQ
	ts.history = append(ts.history, newQ)
	return newQ, mode, nil
}
func (ts *testScaler) Wait(_ context.Context, q types.Quantity, m types.ScalingMode) error {
	if ts.waitErr {
		return fmt.Errorf("Error!")
	}
	ts.calledWait = true
	ts.calledWaitQ = q
	ts.calledWaitMode = m
	time.Sleep(ts.waitDuration)

	return nil
}

type testScalerCreator struct{}

func (t *testScalerCreator) Create(_ context.Context, opts map[string]interface{}) (scale.Scaler, error) {
	eC, _ := opts["return_error_current"].(bool)
	eS, _ := opts["return_error_scale"].(bool)
	eW, _ := opts["return_error_wait"].(bool)
	res := &testScaler{
		cQ:         types.Quantity{Q: 0},
		history:    []types.Quantity{},
		currentErr: eC,
		scaleErr:   eS,
		waitErr:    eW,
	}
	return res, nil
}

type testSolver struct {
	solveErr bool
}

func (t *testSolver) Solve(_ context.Context, qs []types.Quantity) (types.Quantity, error) {
	res := types.Quantity{}
	if t.solveErr {
		return res, fmt.Errorf("Error!")
	}

	for _, q := range qs {
		res.Q = res.Q + q.Q
	}
	return res, nil
}

type testSolverCreator struct{}

func (t *testSolverCreator) Create(_ context.Context, opts map[string]interface{}) (solve.Solver, error) {
	e, _ := opts["return_error"].(bool)
	res := &testSolver{solveErr: e}
	return res, nil
}

type testFilterer struct {
	resAdd   int64
	retError bool
	br       bool
}

func (t *testFilterer) Filter(_ context.Context, currentQ, newQ types.Quantity) (types.Quantity, bool, error) {
	res := types.Quantity{Q: newQ.Q + t.resAdd}
	if t.retError {
		return res, false, fmt.Errorf("Error!")
	}
	if t.br {
		return res, true, nil
	}
	return res, false, nil
}

type testFiltererCreator struct{}

func (t *testFiltererCreator) Create(_ context.Context, opts map[string]interface{}) (filter.Filterer, error) {
	e, _ := opts["return_error"].(bool)
	res := &testFilterer{retError: e}
	return res, nil
}

// Register dummies for each block
func registerDummies() {
	gather.UnregisterAllCreators()
	arrange.UnregisterAllCreators()
	scale.UnregisterAllCreators()
	solve.UnregisterAllCreators()

	// Register dummies
	n := 2
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("test%d", i)
		gather.Register(name, &testGathererCreator{})
		arrange.Register(name, &testArrangerCreator{})
		scale.Register(name, &testScalerCreator{})
		solve.Register(name, &testSolverCreator{})
	}
}
