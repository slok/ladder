package common

import (
	"context"
	"testing"

	"github.com/themotion/ladder/types"
)

func TestBoundCreation(t *testing.T) {
	tests := []struct {
		kind string

		valid    bool
		wantKind boundKind
	}{
		{kind: "max", valid: true, wantKind: boundMax},
		{kind: "min", valid: true, wantKind: boundMin},
		{kind: "Min", valid: false},
		{kind: "Max", valid: false},
		{kind: "something", valid: false},
		{kind: "maax", valid: false},
		{kind: "miin", valid: false},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			boundKindOpt: test.kind,
		}

		b, err := NewBound(context.TODO(), opts)

		if test.valid {
			if err != nil {
				t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
			}

			if b.kind != test.wantKind {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.wantKind, b.kind)
			}
		}

		if !test.valid && err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}

	}
}

func TestBoundSolve(t *testing.T) {
	tests := []struct {
		kind   boundKind
		inputs []types.Quantity

		wantOutput types.Quantity
	}{
		{
			kind: boundMax,
			inputs: []types.Quantity{
				types.Quantity{Q: 1},
				types.Quantity{Q: 2},
				types.Quantity{Q: 3},
				types.Quantity{Q: 4},
				types.Quantity{Q: 5},
			},
			wantOutput: types.Quantity{Q: 5},
		},
		{
			kind: boundMax,
			inputs: []types.Quantity{
				types.Quantity{Q: 1},
				types.Quantity{Q: 2},
				types.Quantity{Q: 1},
				types.Quantity{Q: 0},
				types.Quantity{Q: -5},
			},
			wantOutput: types.Quantity{Q: 2},
		},
		{
			kind: boundMin,
			inputs: []types.Quantity{
				types.Quantity{Q: 1},
				types.Quantity{Q: 2},
				types.Quantity{Q: 3},
				types.Quantity{Q: 4},
				types.Quantity{Q: 5},
			},
			wantOutput: types.Quantity{Q: 1},
		},
		{
			kind: boundMax,
			inputs: []types.Quantity{
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
			},
			wantOutput: types.Quantity{Q: 76},
		},
		{
			kind: boundMin,
			inputs: []types.Quantity{
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
				types.Quantity{Q: 76},
			},
			wantOutput: types.Quantity{Q: 76},
		},
		{
			kind: boundMin,
			inputs: []types.Quantity{
				types.Quantity{Q: 76},
			},
			wantOutput: types.Quantity{Q: 76},
		},
		{
			kind: boundMax,
			inputs: []types.Quantity{
				types.Quantity{Q: 76},
			},
			wantOutput: types.Quantity{Q: 76},
		},
	}

	for _, test := range tests {
		b := &Bound{kind: test.kind}

		q, err := b.Solve(context.TODO(), test.inputs)
		if err != nil {
			t.Errorf("\n- %+v\n  Solve shouldn't give error: %v", test, err)
		}

		if q != test.wantOutput {
			t.Errorf("\n- %+v\n  Wrong result, want: %v; got %v", test, test.wantOutput, q)
		}

	}
}

func TestBoundSolveError(t *testing.T) {
	tests := []struct {
		inputs []types.Quantity
	}{
		{inputs: nil},
		{inputs: []types.Quantity{}},
	}

	for _, test := range tests {
		b := &Bound{}

		_, err := b.Solve(context.TODO(), test.inputs)
		if err == nil {
			t.Errorf("\n- %+v\n  Solve should give error", test)
		}

	}
}
