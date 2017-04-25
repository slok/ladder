package arrange

import (
	"context"
	"fmt"
	"testing"
)

func TestArrangeCreatorRegister(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 50

	// create our arrangers creators
	as := make([]Creator, q)
	for i := 0; i < q; i++ {
		as[i] = &DummyCreator{}
	}

	// Register all the arrangers creators
	for i, a := range as {
		Register(fmt.Sprintf("dummy-%d", i), a)
	}

	// Check all registered ok
	if len(creators) != q {
		t.Errorf("\n- Number of creators registered is wrong, got: %d, want: %d", len(creators), q)
	}

	for i, a := range as {
		name := fmt.Sprintf("dummy-%d", i)
		if creators[name] != a {
			t.Errorf("\n- Registered creator is not the expected one, got: %v, want: %v", creators[name], as[i])
		}
	}
}

func TestArrangeCreatorRegisterNil(t *testing.T) {
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

func TestArrangeCreatorRegisterTwice(t *testing.T) {
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

func TestArrangeCreatorCreate(t *testing.T) {
	UnregisterAllCreators()
	defer UnregisterAllCreators()
	q := 10

	// Register all the Arrangers creators
	for i := 0; i < q; i++ {
		Register(fmt.Sprintf("dummy-%d", i), &DummyCreator{})
	}

	// Create with each creator
	for i := 0; i < q; i++ {
		name := fmt.Sprintf("dummy-%d", i)

		at, err := Create(context.TODO(), name, map[string]interface{}{quantityOpt: 0})

		if err != nil {
			t.Errorf("\n- Arrange creation shouldn't give an error: %s", err)
		}

		if at == nil {
			t.Errorf("\n- Arrange creation shouldn't return nil")
		}
	}
}
