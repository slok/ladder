package common

import (
	"context"
	"fmt"

	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

// InList is an arranger that will make the decision of scalation based on the
// lists of downscale and upscale, will check the given quantity is in one of
// the two lists.
type InList struct {
	MatchDownscale     []int64
	MatchUpscale       []int64
	MatchUpMagnitude   int64
	MatchDownMagnitude int64

	log *log.Log // custom logger
}

const (
	// Opts
	matchDownscaleOpt     = "match_downscale"
	matchUpscaleOpt       = "match_upscale"
	matchUpMagnitudeOpt   = "match_up_magnitude"
	matchDownMagnitudeOpt = "match_down_magnitude"

	// Name
	inListRegName = "in_list"
)

type inListCreator struct{}

func (i *inListCreator) Create(ctx context.Context, opts map[string]interface{}) (arrange.Arranger, error) {
	return NewInList(ctx, opts)
}

// Autoregister on arranger creators
func init() {
	arrange.Register(inListRegName, &inListCreator{})
}

// NewInList creates an InList arranger (Upscale has priority)
func NewInList(ctx context.Context, opts map[string]interface{}) (i *InList, err error) {

	// Recover from panic type conversions and return like a regular error
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	// Prepare ops
	var ok bool

	// Set each option with the correct type
	if _, ok = opts[matchDownscaleOpt]; !ok {
		return nil, fmt.Errorf("%s configuration opt is required", matchDownscaleOpt)
	}
	if _, ok = opts[matchUpscaleOpt]; !ok {
		return nil, fmt.Errorf("%s configuration opt is required", matchUpscaleOpt)
	}
	if _, ok = opts[matchDownMagnitudeOpt]; !ok {
		return nil, fmt.Errorf("%s configuration opt is required", matchDownMagnitudeOpt)
	}
	if _, ok = opts[matchUpMagnitudeOpt]; !ok {
		return nil, fmt.Errorf("%s configuration opt is required", matchUpMagnitudeOpt)
	}

	// We need to separate type assertions from the ifs because we have interface slices
	md := opts[matchDownscaleOpt]
	mu := opts[matchUpscaleOpt]
	mum := opts[matchUpMagnitudeOpt]
	mdm := opts[matchDownMagnitudeOpt]

	if len(md.([]interface{})) == 0 || len(mu.([]interface{})) == 0 {
		return nil, fmt.Errorf("%s or %s should be of len 0", matchDownMagnitudeOpt, matchUpMagnitudeOpt)
	}

	// Start type assertions T_T
	mdTA := make([]int64, len(md.([]interface{})))
	muTA := make([]int64, len(mu.([]interface{})))

	for i, n := range md.([]interface{}) {
		mdTA[i] = types.I2Int64(n)
	}

	for i, n := range mu.([]interface{}) {
		muTA[i] = types.I2Int64(n)
	}

	i = &InList{
		MatchDownscale:     mdTA,
		MatchUpscale:       muTA,
		MatchUpMagnitude:   types.I2Int64(mum),
		MatchDownMagnitude: types.I2Int64(mdm),
	}

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}

	i.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "arranger",
		"name":       inListRegName,
	})

	return
}

// Arrange checks if needs to scale up or down checking if the quantity is on the
// given quantity lists (Upscale has priority in case of being in both lists)
func (i *InList) Arrange(_ context.Context, inputQ, currentQ types.Quantity) (types.Quantity, error) {
	// Something multiplied by 0 is 0...
	var cMagnitude = currentQ.Q
	if cMagnitude == 0 {
		cMagnitude = 1
	}

	q := types.Quantity{Q: currentQ.Q}

	// scale up?
	for _, n := range i.MatchUpscale {
		if inputQ.Q == n {
			m := cMagnitude * i.MatchUpMagnitude / 100
			q.Q = currentQ.Q + m
			i.log.Debugf("%d increasing in %d%%: %d", cMagnitude, i.MatchUpMagnitude, m)

			return q, nil
		}
	}

	// scale down?
	for _, n := range i.MatchDownscale {
		if inputQ.Q == n {
			m := cMagnitude * i.MatchDownMagnitude / 100
			q.Q = currentQ.Q - m
			i.log.Debugf("%d decreasing in %d%%: %d", cMagnitude, i.MatchDownMagnitude, m)

			return q, nil
		}
	}

	// Don't scale
	return q, nil
}
