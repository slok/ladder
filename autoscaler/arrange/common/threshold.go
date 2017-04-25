package common

import (
	"context"
	"fmt"

	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

// Threshold arranger will scale up when the received quantity is higher than `thUpThreshold`
// by `thUpPercent`% of current value, if doesn't meet
// the `thUpMin` minimum quantity it will be increase in that.
// It will scale down by `thDownPercent`% of current value, if doesn't meet
// the `thDownMin` minimum quantity it will be decrease in that.
type Threshold struct {
	upTh        int64 // Scale up threshold is the value that will set the limit for an upscale
	downTh      int64 // Scale down threshold is the value that will set the limit for an downscale
	upPercent   int64 // The percent to add when upscaling
	downPercent int64 // The percent to rest when downscaling
	upMax       int64 // The maximum number to set when upscaling
	downMax     int64 // The minimum number to set when upscaling
	upMin       int64 // The maximum number to set when downscaling
	downMin     int64 // The minimum number to set when downscaling
	// inverse will upscale and downscale in inverse mode, this means if value is lesser than upscale threshold will scale and if
	// the value is greater than the downscale threshold it will downscale
	inverse bool

	log *log.Log // custom logger
}

const (
	// Opts
	thUpThreshold   = "scaleup_threshold"
	thDownThreshold = "scaledown_threshold"
	thInverseMode   = "inverse"
	thUpPercent     = "scaleup_percent"
	thDownPercent   = "scaledown_percent"
	thUpMax         = "scaleup_max_quantity"
	thDownMax       = "scaledown_max_quantity"
	thUpMin         = "scaleup_min_quantity"
	thDownMin       = "scaledown_min_quantity"

	// Name
	thresholdRegName = "threshold"
)

// ThresholdCreator can create a threshold arranger
type ThresholdCreator struct{}

// Create will create a threshold arrangers
func (t *ThresholdCreator) Create(ctx context.Context, opts map[string]interface{}) (arrange.Arranger, error) {
	return NewThreshold(ctx, opts)
}

// Autoregister on arranger creators
func init() {
	arrange.Register(thresholdRegName, &ThresholdCreator{})
}

// NewThreshold will create a Threshold arranger
func NewThreshold(ctx context.Context, opts map[string]interface{}) (t *Threshold, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	// Prepare ops
	var ok bool

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}

	t = &Threshold{
		log: log.WithFields(log.Fields{
			"autoscaler": asName,
			"kind":       "arranger",
			"name":       thresholdRegName,
		})}

	// Set each option with the correct type

	// inverse  mode
	if t.inverse, ok = opts[thInverseMode].(bool); !ok {
		t.log.Infof("Inverse not defined, regular mode active")
	}

	// Upper & lower thresholds
	v, ok := opts[thUpThreshold]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thUpThreshold)
	}
	t.upTh = types.I2Int64(v)

	v, ok = opts[thDownThreshold]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thDownThreshold)
	}
	t.downTh = types.I2Int64(v)

	if !t.inverse && t.downTh >= t.upTh {
		return nil, fmt.Errorf("%s configuration opt can't be greater or equal as %s on regular mode", thDownThreshold, thUpThreshold)
	}

	if t.inverse && t.downTh <= t.upTh {
		return nil, fmt.Errorf("%s configuration opt can't be lesser or equal as %s on inverse mode", thDownThreshold, thUpThreshold)
	}

	// Thresholds percent inc/decr
	v, ok = opts[thUpPercent]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thUpPercent)
	}
	t.upPercent = types.I2Int64(v)

	// Don't check upper than 100 bound because sometimes we want to upscale in higher than 100%
	if t.upPercent < 0 {
		return nil, fmt.Errorf("%s configuration opt must be higher than 0", thUpPercent)
	}

	v, ok = opts[thDownPercent]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thDownPercent)
	}
	t.downPercent = types.I2Int64(v)

	if t.downPercent < 0 || t.downPercent > 100 {
		return nil, fmt.Errorf("%s configuration opt must be between 0 and 100", thDownPercent)
	}

	// up/down minimum quantities
	v, ok = opts[thUpMin]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thUpMin)
	}
	t.upMin = types.I2Int64(v)

	if t.upMin < 0 {
		return nil, fmt.Errorf("%s configuration opt can't be lesser than 0", thUpMin)
	}

	v, ok = opts[thDownMin]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thDownMin)
	}

	t.downMin = types.I2Int64(v)
	if t.downMin < 0 {
		return nil, fmt.Errorf("%s configuration opt can't be lesser than 0", thDownMin)
	}

	// up/down maximum quantities
	v, ok = opts[thUpMax]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thUpMax)
	}
	t.upMax = types.I2Int64(v)

	if t.upMax < 0 {
		return nil, fmt.Errorf("%s configuration opt can't be lesser than 0", thUpMax)
	}

	v, ok = opts[thDownMax]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", thDownMax)
	}
	t.downMax = types.I2Int64(v)

	if t.downMax < 0 {
		return nil, fmt.Errorf("%s configuration opt can't be lesser than 0", thDownMax)
	}

	return
}

// Arrange will calculate the new quantity based on the upper and lower thresholds
func (t *Threshold) Arrange(_ context.Context, inputQ, currentQ types.Quantity) (newQ types.Quantity, err error) {
	newQ = currentQ
	mode := types.NotScaling
	var percent, deltaMin, deltaMax int64
	switch {
	case (!t.inverse && inputQ.Q > t.upTh) || (t.inverse && inputQ.Q < t.upTh): // Scale up
		mode = types.ScalingUp
		percent = t.upPercent
		deltaMin = t.upMin
		deltaMax = t.upMax
	case (!t.inverse && inputQ.Q < t.downTh) || (t.inverse && inputQ.Q > t.downTh): // Scale down
		mode = types.ScalingDown
		percent = t.downPercent
		deltaMin = t.downMin
		deltaMax = t.downMax
	default:
		t.log.Debugf("Not scaling")
		return
	}

	// calculate the percent
	wantDelta := percent * currentQ.Q / 100

	// check if min satisfies
	switch {
	case wantDelta < deltaMin:
		t.log.Warningf("New quantity delta is lower than min : %d < %d, using %[2]d", wantDelta, deltaMin)
		wantDelta = deltaMin

	case wantDelta > deltaMax:
		t.log.Warningf("New quantity delta is greater than max : %d > %d, using %[2]d", wantDelta, deltaMax)
		wantDelta = deltaMax
	}

	// Create the new quantity
	var total int64
	switch mode {
	case types.ScalingUp:
		total = currentQ.Q + wantDelta
	case types.ScalingDown:
		total = currentQ.Q - wantDelta
	default:
		err = fmt.Errorf("Wrong scaling mode reached: %s", mode)
		return
	}

	newQ.Q = total
	t.log.Debugf("New arrangment quantity: %s", newQ)

	return
}
