package scale

import (
	"context"
	"fmt"
	"testing"
)

func TestScaleCreatorRegister(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 50

	// create our scaler creators
	gs := make([]Creator, q)
	for i := 0; i < q; i++ {
		gs[i] = &DummyCreator{}
	}

	// Register all the scaler creators
	for i, g := range gs {
		Register(fmt.Sprintf("dummy-%d", i), g)
	}

	// Check all registered ok
	if len(creators) != q {
		t.Errorf("\n- Number of creators registered is wrong, got: %d, want: %d", len(creators), q)
	}

	for i, g := range gs {
		name := fmt.Sprintf("dummy-%d", i)
		if creators[name] != g {
			t.Errorf("\n- Registered creator is not the expected one, got: %v, want: %v", creators[name], gs[i])
		}
	}
}

func TestScaleCreatorRegisterNil(t *testing.T) {
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

func TestScaleCreatorRegisterTwice(t *testing.T) {
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

func TestScaleCreatorCreate(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 10

	// Register all the scaler creators
	for i := 0; i < q; i++ {
		Register(fmt.Sprintf("dummy-%d", i), &DummyCreator{})
	}

	// Create with each creator
	for i := 0; i < q; i++ {
		name := fmt.Sprintf("dummy-%d", i)

		gt, err := Create(context.TODO(), name, map[string]interface{}{})

		if err != nil {
			t.Errorf("\n- Scale creation shouldn't give an error: %s", err)
		}

		if gt == nil {
			t.Errorf("\n- Scale creation shouldn't return nil")
		}
	}
}
