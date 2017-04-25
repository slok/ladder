package common

import (
	"context"
	"testing"

	"github.com/themotion/ladder/types"
)

func TestInListCorrectCreation(t *testing.T) {
	tests := []struct {
		matchDown          []interface{}
		matchUp            []interface{}
		matchUpMagnitude   interface{}
		matchDownMagnitude interface{}
	}{
		{
			matchDown:          []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)},
			matchUp:            []interface{}{int64(6), int64(7), int64(8), int64(9), int64(10)},
			matchUpMagnitude:   int64(50),
			matchDownMagnitude: int64(10),
		},
		{
			matchDown:          []interface{}{int64(2), int64(4), int64(5), int64(6)},
			matchUp:            []interface{}{int64(6), int64(7), int64(8), int64(9), int64(10)},
			matchUpMagnitude:   int64(50),
			matchDownMagnitude: int64(10),
		},
	}
	for _, test := range tests {
		ops := map[string]interface{}{
			matchDownscaleOpt:     test.matchDown,
			matchUpscaleOpt:       test.matchUp,
			matchDownMagnitudeOpt: test.matchDownMagnitude,
			matchUpMagnitudeOpt:   test.matchUpMagnitude,
		}

		i, err := NewInList(context.TODO(), ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if i.MatchDownscale == nil || i.MatchUpscale == nil || i.MatchDownMagnitude != test.matchDownMagnitude ||
			i.MatchUpMagnitude != test.matchUpMagnitude {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test,
				[]interface{}{i.MatchDownscale, i.MatchUpscale, i.MatchDownMagnitude, i.MatchUpMagnitude},
				[]interface{}{test.matchDown, test.matchUp, test.matchDownMagnitude, test.matchUpMagnitude})
		}
	}
}

func TestInListWrongParameterCreation(t *testing.T) {
	tests := []struct {
		matchDown          []interface{}
		matchUp            []interface{}
		matchUpMagnitude   interface{}
		matchDownMagnitude interface{}
	}{
		{
			matchUp:            []interface{}{6, 7, 8, 9, 10},
			matchUpMagnitude:   50,
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{1, 2, 3, 4, 5},
			matchUpMagnitude:   50,
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{1, 2, 3, 4, 5},
			matchUp:            []interface{}{6, 7, 8, 9, 10},
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{1, 2, 3, 4, 5},
			matchUp:            []interface{}{6, 7, 8, 9, 10},
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{"a", "b"},
			matchUp:            []interface{}{6, 7, 8, 9, 10},
			matchUpMagnitude:   50,
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{1, 2, 3, 4, 5},
			matchUp:            []interface{}{"a", "b"},
			matchUpMagnitude:   50,
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{1, 2, 3, 4, 5},
			matchUp:            []interface{}{6, 7, 8, 9, 10},
			matchUpMagnitude:   "a",
			matchDownMagnitude: 10,
		},
		{
			matchDown:          []interface{}{1, 2, 3, 4, 5},
			matchUp:            []interface{}{6, 7, 8, 9, 10},
			matchUpMagnitude:   50,
			matchDownMagnitude: "b",
		},
	}
	for _, test := range tests {
		ops := map[string]interface{}{
			matchDownscaleOpt:     test.matchDown,
			matchUpscaleOpt:       test.matchUp,
			matchDownMagnitudeOpt: test.matchDownMagnitude,
			matchUpMagnitudeOpt:   test.matchUpMagnitude,
		}

		_, err := NewInList(context.TODO(), ops)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}
	}
}

func TestInListArrangeScaleUp(t *testing.T) {
	tests := []struct {
		opts     map[string]interface{}
		inputQ   types.Quantity
		currentQ types.Quantity
		resultQ  types.Quantity
	}{
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1, 5, 6, 7}, matchDownscaleOpt: []interface{}{0, 2, 100, 29}, matchUpMagnitudeOpt: 100, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 5},
			currentQ: types.Quantity{Q: 100},
			resultQ:  types.Quantity{Q: 200},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 100, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 100},
			resultQ:  types.Quantity{Q: 200},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 50, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 150},
			resultQ:  types.Quantity{Q: 225},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 15, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 5000},
			resultQ:  types.Quantity{Q: 5750},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 92, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 254},
			resultQ:  types.Quantity{Q: 487},
		},
		// Upscale has priority ofer downscale when quantity in both sides
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{1}, matchUpMagnitudeOpt: 92, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 254},
			resultQ:  types.Quantity{Q: 487},
		},
	}
	for _, test := range tests {
		i, err := NewInList(context.TODO(), test.opts)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}
		res, err := i.Arrange(context.TODO(), test.inputQ, test.currentQ)

		if err != nil {
			t.Errorf("\n- %+v\n  Arrange shouldn't give error: %v", test, err)
		}

		if res != test.resultQ {
			t.Errorf("\n- %+v\n  Results don't match, want: %v; got: %v", test, test.resultQ, res)
		}

	}
}

func TestInListArrangeScaleDown(t *testing.T) {
	tests := []struct {
		opts     map[string]interface{}
		inputQ   types.Quantity
		currentQ types.Quantity
		resultQ  types.Quantity
	}{
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 0, matchDownMagnitudeOpt: 25},
			inputQ:   types.Quantity{Q: 0},
			currentQ: types.Quantity{Q: 100},
			resultQ:  types.Quantity{Q: 75},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 0, matchDownMagnitudeOpt: 200},
			inputQ:   types.Quantity{Q: 0},
			currentQ: types.Quantity{Q: 150},
			resultQ:  types.Quantity{Q: -150},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 0, matchDownMagnitudeOpt: 92},
			inputQ:   types.Quantity{Q: 0},
			currentQ: types.Quantity{Q: 5000},
			resultQ:  types.Quantity{Q: 400},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 0, matchDownMagnitudeOpt: 3},
			inputQ:   types.Quantity{Q: 0},
			currentQ: types.Quantity{Q: 254},
			resultQ:  types.Quantity{Q: 247},
		},
	}
	for _, test := range tests {
		i, err := NewInList(context.TODO(), test.opts)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}
		res, err := i.Arrange(context.TODO(), test.inputQ, test.currentQ)

		if err != nil {
			t.Errorf("\n- %+v\n  Arrange shouldn't give error: %v", test, err)
		}

		if res != test.resultQ {
			t.Errorf("\n- %+v\n  Results don't match, want: %v; got: %v", test, test.resultQ, res)
		}

	}

}

func TestInListArrangeNotScale(t *testing.T) {
	tests := []struct {
		opts     map[string]interface{}
		inputQ   types.Quantity
		currentQ types.Quantity
		resultQ  types.Quantity
	}{
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 25, matchDownMagnitudeOpt: 25},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 1},
			resultQ:  types.Quantity{Q: 1},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 25, matchDownMagnitudeOpt: 25},
			inputQ:   types.Quantity{Q: 0},
			currentQ: types.Quantity{Q: 1},
			resultQ:  types.Quantity{Q: 1},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 0, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 1},
			currentQ: types.Quantity{Q: 150},
			resultQ:  types.Quantity{Q: 150},
		},
		{
			opts:     map[string]interface{}{matchUpscaleOpt: []interface{}{1}, matchDownscaleOpt: []interface{}{0}, matchUpMagnitudeOpt: 0, matchDownMagnitudeOpt: 0},
			inputQ:   types.Quantity{Q: 0},
			currentQ: types.Quantity{Q: 150},
			resultQ:  types.Quantity{Q: 150},
		},
	}
	for _, test := range tests {
		i, err := NewInList(context.TODO(), test.opts)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}
		res, err := i.Arrange(context.TODO(), test.inputQ, test.currentQ)

		if err != nil {
			t.Errorf("\n- %+v\n  Arrange shouldn't give error: %v", test, err)
		}

		if res != test.resultQ {
			t.Errorf("\n- %+v\n  Results don't match, want: %v; got: %v", test, test.resultQ, res)
		}

	}
}
