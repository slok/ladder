package arrange

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
	// creators is where we store all the registered arrangers creataors in order to
	// retrieve them
	creators = make(map[string]Creator)
)

// Creator is an interface that will create Arrengers, this is used by
// the register this way we can create Arrangers by string.
// Same approach as SQL package, for the analogy this would be the driver interface
type Creator interface {
	Create(ctx context.Context, opts map[string]interface{}) (Arranger, error)
}

// Register registers an arranger creator on the registry, it will panic if receives nil or the
// arranger creator is already registered
func Register(name string, c Creator) {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	if c == nil {
		panic(fmt.Sprintf("%s arranger creator can't register as a nil", name))
	}

	if _, dup := creators[name]; dup {
		panic(fmt.Sprintf("%s arranger creator already registered", name))
	}

	log.Logger.Infof("Arranger creator registered: %s", name)
	creators[name] = c
}

// UnregisterAllCreators flushes the list of registered creators, used mainly
// on tests
func UnregisterAllCreators() {
	creatorMu.Lock()
	defer creatorMu.Unlock()

	creators = make(map[string]Creator)
}

// Create creates a new arranger based on the received name
func Create(ctx context.Context, name string, opts map[string]interface{}) (Arranger, error) {
	creatorMu.RLock()
	c, ok := creators[name]
	creatorMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s arranger creator not registered", name)
	}

	a, err := c.Create(ctx, opts)

	if err != nil {
		return nil, err
	}
	log.Logger.Debugf("%s arranger created", name)
	return a, nil
}

// Creators returns a sorted list of the arrangers (creators)
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

// Arranger implements the interface who makes the final decisions of scaling
// out/in or nothing
type Arranger interface {
	// Arrange receives 2 quantities, an input quantity and a current quantity,
	// based on this parameters it will decide a new quantity. Usually the input quantity
	// will be get from the gatherer and the current quantity from the Scaler
	Arrange(ctx context.Context, inputQ, currentQ types.Quantity) (newQ types.Quantity, err error)
}
