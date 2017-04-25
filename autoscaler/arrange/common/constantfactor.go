package common

import (
	"context"
	"fmt"
	"math"

	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/types"
)

// ConstFactor arranges the desired scale based on a constant factor division
// F.e for a 50 factor if we have an input of 500 then the result is 10
// and rounding up or down based on the round type, it will return an error if
// the result is our o bounds.
type ConstFactor struct {
	factor    int64
	roundType string
}

const (
	// Opts
	cfFactorOpt    = "factor"
	cfRoundTypeOpt = "round_type"

	// Name
	constFactorRegName = "constant_factor"

	// round types
	cfRTCeil  = "ceil"
	cfRTFloor = "floor"
)

type constFactorCreator struct{}

func (c *constFactorCreator) Create(ctx context.Context, opts map[string]interface{}) (arrange.Arranger, error) {
	return NewConstFactor(ctx, opts)
}

// Autoregister on arranger creators
func init() {
	arrange.Register(constFactorRegName, &constFactorCreator{})
}

// NewConstFactor will create a Const factor arranger
func NewConstFactor(_ context.Context, opts map[string]interface{}) (c *ConstFactor, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	c = &ConstFactor{}

	v, ok := opts[cfFactorOpt]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cfFactorOpt)
	}
	c.factor = types.I2Int64(v)

	if c.roundType, ok = opts[cfRoundTypeOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cfRoundTypeOpt)
	}

	// Check round type is correct
	switch c.roundType {
	case cfRTCeil,
		cfRTFloor:
		// correct
	default:
		return nil, fmt.Errorf("Wrong type of rounding")
	}

	return
}

// Arrange will calculate a new quantity based on a constant factor division
func (c *ConstFactor) Arrange(_ context.Context, inputQ, currentQ types.Quantity) (newQ types.Quantity, err error) {
	newQ = types.Quantity{}

	// Calculate
	n := float64(inputQ.Q) / float64(c.factor)

	switch c.roundType {
	case cfRTCeil:
		n = math.Ceil(n)

	case cfRTFloor:
		n = math.Floor(n)

	default:
		err = fmt.Errorf("Wrong type of rounding")
		return
	}

	newQ.Q = int64(n)

	return
}
