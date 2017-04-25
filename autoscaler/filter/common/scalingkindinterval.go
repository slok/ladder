package common

import (
	"context"
	"fmt"
	"time"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	skiUpDurationOpt   = "scale_up_duration"
	skiDownDurationOpt = "scale_down_duration"

	// id name
	skiRegName = "scaling_kind_interval"
)

// ScalingKindInterval will check if scaling kind has been for a required interval
type ScalingKindInterval struct {
	upDuration   time.Duration
	downDuration time.Duration

	// The current scaling mode and when started
	mode        types.ScalingMode
	modeStarted time.Time
	log         *log.Log // custom logger
}

type scalingKindIntervalCreator struct{}

func (s *scalingKindIntervalCreator) Create(ctx context.Context, opts map[string]interface{}) (filter.Filterer, error) {
	return NewScalingKindInterval(ctx, opts)
}

// Autoregister on filterers creator
func init() {
	filter.Register(skiRegName, &scalingKindIntervalCreator{})
}

// NewScalingKindInterval creates an scalingKindInterval filterer
func NewScalingKindInterval(ctx context.Context, opts map[string]interface{}) (s *ScalingKindInterval, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	s = &ScalingKindInterval{}

	// durations
	ts, ok := opts[skiUpDurationOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is wrong", skiUpDurationOpt)
	}
	if s.upDuration, err = time.ParseDuration(ts); err != nil {
		return
	}

	ts, ok = opts[skiDownDurationOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is wrong", skiDownDurationOpt)
	}
	if s.downDuration, err = time.ParseDuration(ts); err != nil {
		return
	}

	// set initial
	s.mode = types.NotScaling
	s.modeStarted = time.Now().UTC()

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	s.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "filterer",
		"name":       skiRegName,
	})

	return
}

// Filter will return newQ only if the scaling type has been for a required period of time
func (s *ScalingKindInterval) Filter(_ context.Context, currentQ, newQ types.Quantity) (types.Quantity, bool, error) {
	var newMode types.ScalingMode
	var err error

	switch {
	case newQ.Q > currentQ.Q:
		newMode = types.ScalingUp
	case newQ.Q < currentQ.Q:
		newMode = types.ScalingDown
	default:
		// We don't need to scale anything
		s.mode = types.NotScaling
		s.log.Debugf("Not scaling mode set")
		return newQ, false, err
	}

	// Check if the mode changed, if changed then assing new start mode timestamp
	if s.mode != newMode {
		s.log.Debugf("Changing scaling mode from %s to %s", s.mode, newMode)
		s.mode = newMode
		s.modeStarted = time.Now().UTC()
	}

	// At this moment we should check if the duration of the mode is greater or equal,
	// in that case then we should scale
	tPassed := time.Now().UTC().Sub(s.modeStarted)
	var shouldScale bool

	switch s.mode {
	case types.ScalingDown:
		// Do we being in scaledown for the required time?
		if tPassed >= s.downDuration {
			shouldScale = true
		}
	case types.ScalingUp:
		// Do we being in scaleup for the required time?
		if tPassed >= s.upDuration {
			shouldScale = true
		}
	}

	// If shouldn't scale then don't do anything
	if !shouldScale {
		s.log.Infof("Not scaling triggered, continue in %s mode, don't use new quantity (%d), filtering to current (%d)", s.mode, newQ.Q, currentQ.Q)
		return currentQ, false, nil
	}

	s.log.Infof("Autoscaler has been in %s mode for %s, scaling allowed from %v to %v", s.mode, tPassed, currentQ, newQ)

	return newQ, false, err
}
