package types

import "fmt"

// Quantity is an abstract quantity type that will be passed between gatherers, arrangers and scalers
// type wrapped around a struct in case we want to add metadata in the future
type Quantity struct {
	// The quantity itself
	Q int64
}

func (q Quantity) String() string {
	return fmt.Sprintf("%dQ", q.Q)
}

// ScalingMode is the type of scale action,up, down or not scaling
type ScalingMode int

const (
	// ScalingUp describes scaling up mode
	ScalingUp ScalingMode = iota
	// ScalingDown describes scaling down mode
	ScalingDown
	// NotScaling describes scaling rest mode (not scaling)
	NotScaling
)

func (s ScalingMode) String() string {
	switch s {
	case ScalingDown:
		return "scaling down"
	case ScalingUp:
		return "scaling up"
	case NotScaling:
		return "not scaling"
	default:
		return "unknown"
	}
}
