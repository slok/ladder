package solve

import (
	"context"
	"fmt"
	"testing"
)

func TestSolveCreatorRegister(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 50

	// create our solver creators
	ss := make([]Creator, q)
	for i := 0; i < q; i++ {
		ss[i] = &DummyCreator{}
	}

	// Register all the solver creators
	for i, s := range ss {
		Register(fmt.Sprintf("dummy-%d", i), s)
	}

	// Check all registered ok
	if len(creators) != q {
		t.Errorf("\n- Number of creators registered is wrong, got: %d, want: %d", len(creators), q)
	}

	for i, s := range ss {
		name := fmt.Sprintf("dummy-%d", i)
		if creators[name] != s {
			t.Errorf("\n- Registered creator is not the expected one, got: %v, want: %v", creators[name], ss[i])
		}
	}
}

func TestSolveCreatorRegisterNil(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("\n- Registering a nil should panic, it didn't")
		}
	}()

	Register("test", nil)

	t.Errorf("\n- Registering a nil should panic, it didn't")
}

func TesSolveCreatorRegisterTwice(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("\n- Registering a nil should panic, it didn't")
		}
	}()

	Register("test", &DummyCreator{})
	Register("test", &DummyCreator{})

	t.Errorf("\n- Registering a nil should panic, it didn't")
}

func TesSolveCreatorCreate(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 10

	// Register all the solver creators
	for i := 0; i < q; i++ {
		Register(fmt.Sprintf("dummy-%d", i), &DummyCreator{})
	}

	// Create with each creator
	for i := 0; i < q; i++ {
		name := fmt.Sprintf("dummy-%d", i)

		gt, err := Create(context.TODO(), name, map[string]interface{}{})

		if err != nil {
			t.Errorf("\n-solve creation shouldn't give an error: %s", err)
		}

		if gt == nil {
			t.Errorf("\n-solve creation shouldn't return nil")
		}
	}
}
