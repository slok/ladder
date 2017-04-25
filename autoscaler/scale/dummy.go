package scale

import (
	"context"
	"fmt"
	"time"

	"github.com/themotion/ladder/types"
)

const (
	dummyWaitOpt = "wait_duration"
)

// Dummy is a dummy object that satisfies Scaler interface
type Dummy struct {
	currentQ     types.Quantity
	waitDuration time.Duration
}

// NewDummy creates a dummy object
func NewDummy(opts map[string]interface{}) (d *Dummy, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	d = &Dummy{}
	ts, ok := opts[dummyWaitOpt].(string)
	if ok {
		if d.waitDuration, err = time.ParseDuration(ts); err != nil {
			return
		}
	}

	return
}

// Current returns the current quantity of the scalation target
func (d *Dummy) Current(_ context.Context) (types.Quantity, error) {
	return d.currentQ, nil
}

// Scale does nothing, is dummy, returns 0
func (d *Dummy) Scale(_ context.Context, newQ types.Quantity) (types.Quantity, types.ScalingMode, error) {
	var mode types.ScalingMode

	switch {
	case newQ.Q > d.currentQ.Q:
		mode = types.ScalingUp
	case newQ.Q < d.currentQ.Q:
		mode = types.ScalingDown
	default:
		mode = types.NotScaling
	}

	return newQ, mode, nil
}

// Wait will wait a given time duration
func (d *Dummy) Wait(_ context.Context, _ types.Quantity, _ types.ScalingMode) error {
	time.Sleep(d.waitDuration)
	return nil
}

// DummyCreator is a dummy creation object that satisfies Creator interface
type DummyCreator struct{}

// Create creates an Scaler
func (d *DummyCreator) Create(_ context.Context, opts map[string]interface{}) (Scaler, error) {
	return NewDummy(opts)
}
