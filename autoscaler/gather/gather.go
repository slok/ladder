package gather

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
	// creators is where we store all the registered gatherers creataors in order to
	// retrieve them
	creators = make(map[string]Creator)
)

// Creator is an interface that will create Gatherers, this is used by
// the register this way we can create Gatherers by string.
// Same approach as SQL package, for the analogy this would be the driver interface
type Creator interface {
	Create(ctx context.Context, opts map[string]interface{}) (Gatherer, error)
}

// Register resgisters a Gatherer creator on the registry, it will panic if receives nil or the
// gatherer creator is already registered
func Register(name string, c Creator) {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	if c == nil {
		panic(fmt.Sprintf("%s gatherer creator can't register as a nil", name))
	}

	if _, dup := creators[name]; dup {
		panic(fmt.Sprintf("%s gatherer creator already registered", name))
	}

	log.Logger.Infof("Gatherer creator registered: %s", name)
	creators[name] = c
}

// UnregisterAllCreators flushes the list of registered creators, used mainly
// on tests
func UnregisterAllCreators() {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	creators = make(map[string]Creator)
}

// Create creates a new Gatherer based on the received name
func Create(ctx context.Context, name string, opts map[string]interface{}) (Gatherer, error) {
	creatorMu.RLock()
	c, ok := creators[name]
	creatorMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s gatherer creator not registered", name)
	}

	g, err := c.Create(ctx, opts)

	if err != nil {
		return nil, err
	}
	log.Logger.Debugf("%s gatherer created", name)
	return g, nil
}

// Creators returns a sorted list of the Gatherers (creators)
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

// Gatherer implements the Gather method to obtain quantities from any source
type Gatherer interface {
	// Gather gathers quantitys of a given input
	Gather(ctx context.Context) (types.Quantity, error)
}
