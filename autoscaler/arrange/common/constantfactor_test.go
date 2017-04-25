package common

import (
	"context"
	"testing"

	"github.com/themotion/ladder/types"
)

func TestConstFactorCorrectCreation(t *testing.T) {
	tests := []struct {
		factor    int64
		roundType string
	}{
		{5, "ceil"},
		{5, "floor"},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			cfFactorOpt:    test.factor,
			cfRoundTypeOpt: test.roundType,
		}

		cf, err := NewConstFactor(context.TODO(), opts)
		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if cf.factor != test.factor || cf.roundType != test.roundType {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test,
				[]interface{}{test.factor, test.roundType},
				[]interface{}{cf.factor, cf.roundType},
			)
		}
	}
}

func TestConstFactorWrongParameterCreation(t *testing.T) {
	tests := []struct {
		factor    interface{}
		roundType interface{}
	}{
		{5, "ceiling"},
		{5, "flooooor"},
		{105, ""},
		{"5", "ceil"},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			cfFactorOpt:    test.factor,
			cfRoundTypeOpt: test.roundType,
		}

		_, err := NewConstFactor(context.TODO(), opts)
		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give an error", test)
		}
	}
}

func TestConstFactorArrange(t *testing.T) {
	tests := []struct {
		factor    int64
		roundType string
		inputQ    int64

		wantQ int64
	}{
		{5, "ceil", 50, 10},
		{5, "floor", 50, 10},
		{3, "ceil", 50, 17},
		{3, "floor", 50, 16},
		{6, "ceil", 50, 9},
		{6, "floor", 50, 8},
		{2, "floor", 0, 0}, // 0 input
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			cfFactorOpt:    test.factor,
			cfRoundTypeOpt: test.roundType,
		}

		cf, err := NewConstFactor(context.TODO(), opts)
		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		inQ := types.Quantity{Q: test.inputQ}
		cuQ := types.Quantity{}

		newQ, err := cf.Arrange(context.TODO(), inQ, cuQ)
		if err != nil {
			t.Errorf("\n- %+v\n  Arrange shouldn't give error: %v", test, err)
		}

		if newQ.Q != test.wantQ {
			t.Errorf("\n- %+v\n  Results don't match, want: %v; got: %v", test, test.wantQ, newQ.Q)
		}
	}
}

func TestConstFactorArrangeError(t *testing.T) {
	tests := []struct {
		factor    int64
		roundType string
		inputQ    int64
	}{
		{5, "ceiling", 50},
		{5, "flooring", 50},
		{5, "", 50},
	}

	for _, test := range tests {

		cf := &ConstFactor{
			factor:    test.factor,
			roundType: test.roundType,
		}

		inQ := types.Quantity{Q: test.inputQ}
		cuQ := types.Quantity{}

		if _, err := cf.Arrange(context.TODO(), inQ, cuQ); err == nil {
			t.Errorf("\n- %+v\n  Arrange should give error", test)
		}
	}
}
