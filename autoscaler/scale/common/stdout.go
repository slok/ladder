package common

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/themotion/ladder/autoscaler/scale"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	msgPrefixOpt = "message_prefix"

	// Name
	stdoutRegName = "stdout"
)

// Stdout representes an object for scaling in stdout
type Stdout struct {
	MsgPrefix string

	dst      io.Writer
	currentQ int64
}

type stdoutCreator struct{}

func (s *stdoutCreator) Create(_ context.Context, opts map[string]interface{}) (scale.Scaler, error) {
	return NewStdout(opts)
}

// Autoregister on arranger creators
func init() {
	scale.Register(stdoutRegName, &stdoutCreator{})
}

// NewStdout creates an stdout scaler
func NewStdout(opts map[string]interface{}) (s *Stdout, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	s = &Stdout{
		// Use stdout
		dst: os.Stdout,
	}

	// Prepare ops
	var ok bool

	// Set each option with the correct type
	if s.MsgPrefix, ok = opts[msgPrefixOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", msgPrefixOpt)
	}

	return
}

// Current returns the current quantity of the scaler
func (s *Stdout) Current(_ context.Context) (types.Quantity, error) {
	return types.Quantity{Q: s.currentQ}, nil
}

// Wait will do nothing
func (s *Stdout) Wait(_ context.Context, _ types.Quantity, _ types.ScalingMode) error {
	return nil
}

// Scale scales to the new received quantity
func (s *Stdout) Scale(ctx context.Context, newQ types.Quantity) (types.Quantity, types.ScalingMode, error) {
	actionStr := ""
	c, _ := s.Current(ctx)
	mode := types.NotScaling
	switch {

	case newQ.Q > c.Q:
		actionStr = "Scaling up"
		mode = types.ScalingUp
	case newQ.Q < c.Q:
		actionStr = "Scaling down"
		mode = types.ScalingDown
	default:
		actionStr = "Don't scalingup/down"
	}
	s.currentQ = newQ.Q
	fmt.Fprintf(s.dst, "%s %s: %d\n", s.MsgPrefix, actionStr, newQ.Q)
	return newQ, mode, nil
}
