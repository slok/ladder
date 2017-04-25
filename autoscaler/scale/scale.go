package scale

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

var (
	// rwmutex for the creator registry
	creatorMu sync.RWMutex
	// creators is where we store all the registered scaler creataors in order to
	// retrieve them
	creators = make(map[string]Creator)
)

// Creator is an interface that will create Scalers, this is used by
// the register this way we can create Scalers by string.
// Same approach as SQL package, for the analogy this would be the driver interface
type Creator interface {
	Create(ctx context.Context, opts map[string]interface{}) (Scaler, error)
}

// Register registers a Scaler creator on the registry, it will panic if receives nil or the
// scaler creator is already registered
func Register(name string, c Creator) {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	if c == nil {
		panic(fmt.Sprintf("%s scaler creator can't register as a nil", name))
	}

	if _, dup := creators[name]; dup {
		panic(fmt.Sprintf("%s scaler creator already registered", name))
	}

	log.Logger.Infof("Scaler creator registered: %s", name)
	creators[name] = c
}

// UnregisterAllCreators flushes teh list of registered creators, used mainly
// on tests
func UnregisterAllCreators() {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	creators = make(map[string]Creator)
}

// Create creates a new Scaler based on the received name
func Create(ctx context.Context, name string, opts map[string]interface{}) (Scaler, error) {
	creatorMu.RLock()
	c, ok := creators[name]
	creatorMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s scaler creator not registered", name)
	}

	g, err := c.Create(ctx, opts)

	if err != nil {
		return nil, err
	}
	log.Logger.Debugf("%s scaler created", name)
	return g, nil
}

// Creators returns a sorted list of the Scalers (creators)
func Creators() []string {
	creatorMu.RLock()
	defer creatorMu.RUnlock()

	var c []string

	for name := range creators {
		c = append(c, name)
	}

	sort.Strings(c)

	return c
}

// Scaler is the interface needed to be implemented by all the scalers
type Scaler interface {
	// Current gets the quantity of the current scaling target
	Current(ctx context.Context) (types.Quantity, error)

	// Scale will scale based on the new quantity, it will return finally if it
	// the scalation was run in what mode and quantity or not
	Scale(ctx context.Context, newQ types.Quantity) (scaledQ types.Quantity, mode types.ScalingMode, err error)

	// Wait will wait until the scalation has been made
	Wait(ctx context.Context, scaledQ types.Quantity, mode types.ScalingMode) error
}
