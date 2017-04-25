package filter

import (
	"context"

	"github.com/themotion/ladder/types"
)

// Dummy is a dummy object that satisfies filterer interface
type Dummy struct {
	returnQ types.Quantity
}

// NewDummy creates a dummy object
func NewDummy(opts map[string]interface{}) (*Dummy, error) { return &Dummy{}, nil }

// Filter returns always the specified number
func (d *Dummy) Filter(_ context.Context, currentQ, newQ types.Quantity) (types.Quantity, bool, error) {
	return d.returnQ, false, nil
}

// DummyCreator is a dummy creation object that satisfies Creator interface
type DummyCreator struct{}

// Create creates an filterer
func (d *DummyCreator) Create(_ context.Context, opts map[string]interface{}) (Filterer, error) {
	return NewDummy(opts)
}
