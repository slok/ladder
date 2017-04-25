package filter

import (
	"context"
	"fmt"
	"testing"
)

func TestFilterCreatorRegister(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 50

	// create our filterer creators
	fs := make([]Creator, q)
	for i := 0; i < q; i++ {
		fs[i] = &DummyCreator{}
	}

	// Register all the filterer creators
	for i, f := range fs {
		Register(fmt.Sprintf("dummy-%d", i), f)
	}

	// Check all registered ok
	if len(creators) != q {
		t.Errorf("\n- Number of creators registered is wrong, got: %d, want: %d", len(creators), q)
	}

	for i, f := range fs {
		name := fmt.Sprintf("dummy-%d", i)
		if creators[name] != f {
			t.Errorf("\n- Registered creator is not the expected one, got: %v, want: %v", creators[name], fs[i])
		}
	}
}

func TestFilterCreatorRegisterNil(t *testing.T) {
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

func TesFilterCreatorRegisterTwice(t *testing.T) {
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

func TesFilterCreatorCreate(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 10

	// Register all the filterer creators
	for i := 0; i < q; i++ {
		Register(fmt.Sprintf("dummy-%d", i), &DummyCreator{})
	}

	// Create with each creator
	for i := 0; i < q; i++ {
		name := fmt.Sprintf("dummy-%d", i)

		gt, err := Create(context.TODO(), name, map[string]interface{}{})

		if err != nil {
			t.Errorf("\n-filterer creation shouldn't give an error: %s", err)
		}

		if gt == nil {
			t.Errorf("\n-filterer creation shouldn't return nil")
		}
	}
}
