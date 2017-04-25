package common

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	maxLimitOpt = "max_limit"
	minLimitOpt = "min_limit"

	// id name
	randomRegName = "random"
)

// Random will create a new random number each time
type Random struct {
	MaxLimit int64
	MinLimit int64
}

type randomCreator struct{}

func (r *randomCreator) Create(_ context.Context, opts map[string]interface{}) (gather.Gatherer, error) {
	return NewRandom(opts)
}

// Autoregister on gatherers creator
func init() {
	gather.Register(randomRegName, &randomCreator{})
}

// NewRandom creates a random Gatherer
func NewRandom(opts map[string]interface{}) (ra *Random, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	r := &Random{}

	v, ok := opts[maxLimitOpt]
	// Set each option with the correct type
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", maxLimitOpt)
	}
	r.MaxLimit = types.I2Int64(v)

	v, ok = opts[minLimitOpt]
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", minLimitOpt)
	}
	r.MinLimit = types.I2Int64(v)

	if r.MaxLimit < 0 || r.MinLimit < 0 {
		return nil, fmt.Errorf("%s or %s should'b be less than 0", minLimitOpt, maxLimitOpt)
	}

	// max limit should be greated than min limit
	if r.MaxLimit <= r.MinLimit {
		return nil, fmt.Errorf("%s should be greated than %s", minLimitOpt, maxLimitOpt)
	}

	return r, nil
}

// Gather returns a random quantity
func (r *Random) Gather(_ context.Context) (types.Quantity, error) {
	s := rand.NewSource(time.Now().UnixNano())
	rr := rand.New(s)
	rNumber := rr.Int63n(r.MaxLimit) + r.MinLimit
	res := types.Quantity{Q: rNumber}
	return res, nil
}
