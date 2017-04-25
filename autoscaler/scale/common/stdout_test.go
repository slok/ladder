package common

import (
	"bytes"
	"context"
	"testing"

	"github.com/themotion/ladder/types"
)

func TestStdoutCorrectCreation(t *testing.T) {
	tests := []struct {
		msgPrefix string
	}{
		{msgPrefix: "test"},
		{msgPrefix: "_test1"},
		{msgPrefix: "[test]"},
		{msgPrefix: ""},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			msgPrefixOpt: test.msgPrefix,
		}

		s, err := NewStdout(ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if s.MsgPrefix != test.msgPrefix {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.msgPrefix, s.MsgPrefix)
		}
	}
}

func TestStdoutWrongParameterCreation(t *testing.T) {
	ops := map[string]interface{}{}

	_, err := NewStdout(ops)

	if err == nil {
		t.Errorf("\n-  Creation should give error")
	}
}

func TestStodutScaler(t *testing.T) {
	tests := []struct {
		msgPrefix string
		currentQ  int64
		newQ      types.Quantity
		want      string
		mode      types.ScalingMode
	}{
		{
			msgPrefix: "*",
			currentQ:  10,
			newQ:      types.Quantity{Q: 11},
			want:      "* Scaling up: 11\n",
			mode:      types.ScalingUp,
		},
		{
			msgPrefix: "[test]",
			currentQ:  10,
			newQ:      types.Quantity{Q: 9},
			want:      "[test] Scaling down: 9\n",
			mode:      types.ScalingDown,
		},
		{
			msgPrefix: "",
			currentQ:  10,
			newQ:      types.Quantity{Q: 10},
			want:      " Don't scalingup/down: 10\n",
			mode:      types.NotScaling,
		},
	}

	for _, test := range tests {
		var b bytes.Buffer

		s := &Stdout{
			MsgPrefix: test.msgPrefix,
			currentQ:  test.currentQ,
			dst:       &b,
		}
		scaled, mode, err := s.Scale(context.TODO(), test.newQ)
		if err != nil {
			t.Errorf("\n- %+v\n  Scalation shouldn't give an error: %v", test, err)
		}

		if mode != test.mode {
			t.Errorf("\n- %+v\n  Scalation mode doesn't match, want:%v go: %v", test, test.mode, mode)
		}

		if scaled.Q != test.newQ.Q {
			t.Errorf("\n- %+v\n  Scalation scaled Q doesn't match, want:%d go: %d", test, test.newQ.Q, scaled.Q)
		}

		if string(b.Bytes()) != test.want {
			t.Errorf("\n- %+v\n  Scalation message doesn't match, want: '%v'; got: '%s'", test, test.want, string(b.Bytes()))
		}
	}
}
