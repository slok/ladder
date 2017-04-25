package common

import (
	"context"
	"fmt"

	"github.com/themotion/ladder/autoscaler/solve"
	"github.com/themotion/ladder/types"
)

type boundKind int

const (
	boundMax = iota
	boundMin
	boundUnknown
)

func parseBoundKind(kind string) (boundKind, error) {
	switch kind {
	case "max":
		return boundMax, nil
	case "min":
		return boundMin, nil
	default:
		return boundUnknown, fmt.Errorf("Wrong bound kind: %s", kind)
	}

}

const (
	// Opts
	boundKindOpt = "kind"

	// Type
	boundType = ""

	// id name
	boundRegName = "bound"
)

// Bound represent the bound solver that will solve the multiplevalues with min or max bounds
type Bound struct {
	kind boundKind
}

type boundCreator struct{}

func (b *boundCreator) Create(ctx context.Context, opts map[string]interface{}) (solve.Solver, error) {
	return NewBound(ctx, opts)
}

// Autoregister on solvers creator
func init() {
	solve.Register(boundRegName, &boundCreator{})
}

// NewBound creates a bound solver
func NewBound(ctx context.Context, opts map[string]interface{}) (b *Bound, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	b = &Bound{}

	// Prepare ops
	var ok bool

	// Set each option with the correct type
	var kind string
	if kind, ok = opts[boundKindOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", boundKindOpt)
	}
	b.kind, err = parseBoundKind(kind)
	if err != nil {
		return
	}

	return
}

// Solve will get the bound of all the results, for example maximum or minimum
func (b *Bound) Solve(_ context.Context, qs []types.Quantity) (types.Quantity, error) {
	if qs == nil || len(qs) == 0 {
		return types.Quantity{}, fmt.Errorf("qs param can't be empty")
	}
	res := qs[0]

	for _, q := range qs[1:] {
		switch b.kind {
		case boundMax:
			if q.Q > res.Q {
				res = q
			}
		case boundMin:
			if q.Q < res.Q {
				res = q
			}
		}
	}
	return res, nil
}
