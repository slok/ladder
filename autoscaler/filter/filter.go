package filter

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
	// creators is where we store all the registered filters creators in order to
	// retrieve them
	creators = make(map[string]Creator)
)

// Creator is an interface that will create Solvers, this is used by
// the register this way we can create filters by string.
// Same approach as SQL package, for the analogy this would be the driver interface
type Creator interface {
	Create(ctx context.Context, opts map[string]interface{}) (Filterer, error)
}

// Register registers a filter creator on the registry, it will panic if receives nil or the
// filter creator is already registered
func Register(name string, c Creator) {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	if c == nil {
		panic(fmt.Sprintf("%s filter creator can't register as a nil", name))
	}

	if _, dup := creators[name]; dup {
		panic(fmt.Sprintf("%s filter creator already registered", name))
	}

	log.Logger.Infof("Filter creator registered: %s", name)
	creators[name] = c
}

// UnregisterAllCreators flushes teh list of registered creators, used mainly
// on tests
func UnregisterAllCreators() {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	creators = make(map[string]Creator)
}

// Create creates a new filter based on the received name
func Create(ctx context.Context, name string, opts map[string]interface{}) (Filterer, error) {
	creatorMu.RLock()
	c, ok := creators[name]
	creatorMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s filter creator not registered", name)
	}

	f, err := c.Create(ctx, opts)

	if err != nil {
		return nil, err
	}
	log.Logger.Debugf("%s filter created", name)
	return f, nil
}

// Filterer is the interface needed to be implemented by all the filters
type Filterer interface {
	// Filter receives the new quantity that needs to scale and the current one,
	// returns a new one and error, an error stops the chain and returning break also
	Filter(ctx context.Context, currentQ, newQ types.Quantity) (q types.Quantity, br bool, err error)
}
