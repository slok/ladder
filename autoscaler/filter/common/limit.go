package common

import (
	"context"
	"fmt"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	limitMaxOpt = "max"
	limitMinOpt = "min"

	// id name
	limitRegName = "limit"
)

// Limit will check that a value soesn't go out of limits
type Limit struct {
	Max int64
	Min int64

	log *log.Log // custom logger
}

type limitCreator struct{}

func (l *limitCreator) Create(ctx context.Context, opts map[string]interface{}) (filter.Filterer, error) {
	return NewLimit(ctx, opts)
}

// Autoregister on filterers creator
func init() {
	filter.Register(limitRegName, &limitCreator{})
}

// NewLimit creates a limit filterer
func NewLimit(ctx context.Context, opts map[string]interface{}) (l *Limit, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	l = &Limit{}

	// Set each option with the correct type
	v, ok := opts[limitMaxOpt]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", limitMaxOpt)
	}
	l.Max = types.I2Int64(v)

	v, ok = opts[limitMinOpt]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", limitMinOpt)
	}
	l.Min = types.I2Int64(v)

	if l.Max < 0 || l.Min < 0 {
		return nil, fmt.Errorf("%s or %s should'b be less than 0", limitMinOpt, limitMaxOpt)
	}

	// max limit should be greated than min limit
	if l.Max <= l.Min {
		return nil, fmt.Errorf("%s should be greater than %s", limitMaxOpt, limitMinOpt)
	}

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	l.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "filterer",
		"name":       limitRegName,
	})

	return l, nil
}

// Filter will filter the input based on a maximum of a minimum
func (l *Limit) Filter(_ context.Context, currentQ, newQ types.Quantity) (types.Quantity, bool, error) {
	switch {
	case newQ.Q > l.Max:
		l.log.Infof("Quantity highest than max limit, filtering from %d to %d", newQ.Q, l.Max)
		newQ.Q = l.Max
	case newQ.Q < l.Min:
		l.log.Infof("Quantity lesser than min limit, filtering from %d to %d", newQ.Q, l.Min)
		newQ.Q = l.Min
	default:
		l.log.Debugf("No limit filtered applied")
	}

	return newQ, false, nil
}
