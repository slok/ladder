package solve

import (
	"context"

	"github.com/themotion/ladder/types"
)

// Dummy is a dummy object that satisfies solver interface
type Dummy struct{}

// NewDummy creates a dummy object
func NewDummy(opts map[string]interface{}) (*Dummy, error) { return &Dummy{}, nil }

// Solve returns always returns the sum of all inputs
func (d *Dummy) Solve(ctx context.Context, qs []types.Quantity) (types.Quantity, error) {
	res := types.Quantity{}
	for _, q := range qs {
		res.Q = res.Q + q.Q
	}
	return res, nil
}

// DummyCreator is a dummy creation object that satisfies Creator interface
type DummyCreator struct{}

// Create creates an solver
func (d *DummyCreator) Create(_ context.Context, opts map[string]interface{}) (Solver, error) {
	return NewDummy(opts)
}
