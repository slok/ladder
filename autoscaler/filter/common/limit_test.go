package common

import (
	"context"
	"testing"

	"github.com/themotion/ladder/types"
)

func TestLimitCorrectCreation(t *testing.T) {

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
			limitMaxOpt: test.max,
			limitMinOpt: test.min,
		}

		l, err := NewLimit(context.TODO(), ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if l.Max != test.max || l.Min != test.min {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, []int64{test.min, test.max}, []int64{l.Min, l.Max})
		}
	}
}

func TestLimitWrongParameterCreation(t *testing.T) {
	tests := []struct {
		opts map[string]interface{}
	}{
		{opts: map[string]interface{}{limitMaxOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: 11, limitMaxOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: 10, limitMaxOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: 0.1, limitMaxOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: -4, limitMaxOpt: -1}},
		{opts: map[string]interface{}{limitMinOpt: -4, limitMaxOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: "wrong", limitMaxOpt: 10}},
		{opts: map[string]interface{}{limitMinOpt: 10, limitMaxOpt: true}},
		{opts: map[string]interface{}{limitMinOpt: 10, limitMaxOpt: []int{1}}},
	}

	for _, test := range tests {

		_, err := NewLimit(context.TODO(), test.opts)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}
	}
}

func TestLimitFilter(t *testing.T) {
	tests := []struct {
		max  int64
		min  int64
		new  int64
		want int64
	}{
		{10, 2, 5, 5},
		{10, 2, 15, 10},
		{10, 2, 1, 2},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			limitMaxOpt: test.max,
			limitMinOpt: test.min,
		}
		l, err := NewLimit(context.TODO(), opts)
		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		q, _, err := l.Filter(context.TODO(), types.Quantity{}, types.Quantity{Q: test.new})
		if err != nil {
			t.Errorf("\n- %+v\n  Filter shouldn't give error: %v", test, err)
		}

		if q.Q != test.want {
			t.Errorf("\n- %+v\n  Wrong limit result want: %d; got %d", test, test.want, q.Q)
		}
	}
}
