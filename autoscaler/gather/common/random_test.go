package common

import (
	"context"
	"testing"
)

func TestRandomCorrectCreation(t *testing.T) {
	tests := []struct {
		max int64
		min int64
	}{
		{min: 0, max: 1},
		{min: 5, max: 10},
		{min: 98, max: 99},
		{min: 1234, max: 12345},
		{min: 0, max: 999999999999999999},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			maxLimitOpt: test.max,
			minLimitOpt: test.min,
		}

		r, err := NewRandom(ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if r.MaxLimit != test.max || r.MinLimit != test.min {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, []int64{test.min, test.max}, []int64{r.MinLimit, r.MaxLimit})
		}
	}

}

func TestRandomWrongParameterCreation(t *testing.T) {
	tests := []struct {
		opts map[string]interface{}
	}{
		{opts: map[string]interface{}{maxLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: 11, maxLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: 10, maxLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: 0.1, maxLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: -4, maxLimitOpt: -1}},
		{opts: map[string]interface{}{minLimitOpt: -4, maxLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: "wrong", maxLimitOpt: 10}},
		{opts: map[string]interface{}{minLimitOpt: 10, maxLimitOpt: true}},
		{opts: map[string]interface{}{minLimitOpt: 10, maxLimitOpt: []int64{1}}},
	}

	for _, test := range tests {

		_, err := NewRandom(test.opts)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}
	}
}

func TestRandomGather(t *testing.T) {
	ops := map[string]interface{}{
		minLimitOpt: 0,
		maxLimitOpt: 99999999999999999,
	}

	r, err := NewRandom(ops)
	if err != nil {
		t.Fatalf("\n- Creation shouldn't give error: %v", err)
	}

	n1, err := r.Gather(context.TODO())
	if err != nil {
		t.Errorf("\n- Gather shouldn't give error: %v", err)
	}

	n2, err := r.Gather(context.TODO())
	if err != nil {
		t.Errorf("\n- Gather shouldn't give error: %v", err)
	}

	if n2.Q == n1.Q {
		t.Errorf("\n- %+v\n- 2 random sequential numbers should be different", []int64{n1.Q, n2.Q})
	}
}
