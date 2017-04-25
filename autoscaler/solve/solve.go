package solve

import (
	"context"
	"fmt"
	"sync"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

var (
	// rwmutex for the creator registry
	creatorMu sync.RWMutex
	// creators is where we store all the registered solvers creators in order to
	// retrieve them
	creators = make(map[string]Creator)
)

// Creator is an interface that will create Solvers, this is used by
// the register this way we can create solvers by string.
// Same approach as SQL package, for the analogy this would be the driver interface
type Creator interface {
	Create(ctx context.Context, opts map[string]interface{}) (Solver, error)
}

// Register registers a solver creator on the registry, it will panic if receives nil or the
// solver creator is already registered
func Register(name string, c Creator) {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	if c == nil {
		panic(fmt.Sprintf("%s solver creator can't register as a nil", name))
	}

	if _, dup := creators[name]; dup {
		panic(fmt.Sprintf("%s solver creator already registered", name))
	}

	log.Logger.Infof("Solver creator registered: %s", name)
	creators[name] = c
}

// UnregisterAllCreators flushes teh list of registered creators, used mainly
// on tests
func UnregisterAllCreators() {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	creators = make(map[string]Creator)
}

// Create creates a new solver based on the received name
func Create(ctx context.Context, name string, opts map[string]interface{}) (Solver, error) {
	creatorMu.RLock()
	c, ok := creators[name]
	creatorMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s solver creator not registered", name)
	}

	s, err := c.Create(ctx, opts)

	if err != nil {
		return nil, err
	}
	log.Logger.Debugf("%s solver created", name)
	return s, nil
}

// Solver is the interface needed to be implemented by all the solvers
type Solver interface {
	// Solve receives multiple quantities and retunrs only on based on the others
	Solve(ctx context.Context, qs []types.Quantity) (types.Quantity, error)
}
