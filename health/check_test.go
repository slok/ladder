package health

import (
	"errors"
	"fmt"
	"testing"
)

type testChecker struct {
	msg       string
	wantError bool
}

func (t testChecker) Check() (msg string, err error) {
	msg = t.msg
	if t.wantError {
		err = errors.New(t.msg)
	}
	return
}

func TestHealthCheckNoCheckers(t *testing.T) {
	c := &Check{}
	status := c.Status()

	if status.Status != HCOk {
		t.Errorf("A health check with no checkers status should be healthy, it wasn't")
	}
}

func TestHealthCheckCheckersHealthy(t *testing.T) {
	c := &Check{
		checks: map[string]map[string]Checker{},
	}

	groupsN := 5
	checkersN := 3
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		for n := 0; n < checkersN; n++ {
			name := fmt.Sprintf("checker_%d_%d", g, n)
			cc := &testChecker{
				msg: fmt.Sprintf("%s is ok", name),
			}
			c.Register(name, group, cc)
		}
	}

	status := c.Status()

	if status.Status != HCOk {
		t.Fatalf("A health check with no checkers status should be healthy, it wasn't")
	}

	// Check groups length
	if len(status.OkResults) != groupsN {
		t.Errorf("The number of groups in healthy ones is wrong, want: %d; got: %d", groupsN, len(status.OkResults))
	}

	if len(status.ErrorResults) != groupsN {
		t.Errorf("The number of groups in unhealthy ones is wrong, want: %d; got: %d", groupsN, len(status.ErrorResults))
	}

	// Check healthy check results
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		if len(status.OkResults[group]) != checkersN {
			t.Errorf("The number of healthy checks in group %s should be %d, got: %d", group, checkersN, len(status.OkResults[group]))
		}
		for n := 0; n < checkersN; n++ {
			name := fmt.Sprintf("checker_%d_%d", g, n)

			r, ok := status.OkResults[group][name]

			if !ok {
				t.Errorf("Check %s in group %s not present in healthy checks, it should", name, group)
			}

			msg := fmt.Sprintf("%s is ok", name)
			if r != msg {
				t.Errorf("Check %s in healthy group %s doesn't have the expected message: want: '%s', got: '%s'", name, group, r, msg)
			}
		}
	}

	// Check unhealthy check results
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		if len(status.ErrorResults[group]) != 0 {
			t.Errorf("The number of unhealthy checks in group %s should be 0, got: %d", group, len(status.ErrorResults[group]))
		}
	}
}

func TestHealthCheckCheckersUnhealthy(t *testing.T) {
	c := &Check{
		checks: map[string]map[string]Checker{},
	}

	groupsN := 5
	checkersN := 3
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		for n := 0; n < checkersN; n++ {
			name := fmt.Sprintf("checker_%d_%d", g, n)
			cc := &testChecker{
				msg:       fmt.Sprintf("%s errored", name),
				wantError: true,
			}
			c.Register(name, group, cc)
		}
	}

	status := c.Status()

	if status.Status != HCError {
		t.Fatalf("A health check with no checkers status should be unhealthy, it wasn't")
	}

	// Check groups length
	if len(status.OkResults) != groupsN {
		t.Errorf("The number of groups in healthy ones is wrong, want: %d; got: %d", groupsN, len(status.OkResults))
	}

	if len(status.ErrorResults) != groupsN {
		t.Errorf("The number of groups in unhealthy ones is wrong, want: %d; got: %d", groupsN, len(status.ErrorResults))
	}

	// Check healthy check results
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		if len(status.OkResults[group]) != 0 {
			t.Errorf("The number of healthy checks in group %s should be 0, got: %d", group, len(status.OkResults[group]))
		}
	}

	// Check unhealthy check results
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		if len(status.ErrorResults[group]) != checkersN {
			t.Errorf("The number of unhealthy checks in group %s should be %d, got: %d", group, checkersN, len(status.ErrorResults[group]))
		}
		for n := 0; n < checkersN; n++ {
			name := fmt.Sprintf("checker_%d_%d", g, n)

			r, ok := status.ErrorResults[group][name]

			if !ok {
				t.Errorf("Check %s in group %s not present in unhealthy checks, it should", name, group)
			}

			msg := fmt.Sprintf("%s errored", name)
			if r != msg {
				t.Errorf("Check %s in unhealthy group %s doesn't have the expected message: want: '%s', got: '%s'", name, group, r, msg)
			}
		}
	}
}

func TestHealthCheckCheckersHealthyOneUnhealthy(t *testing.T) {
	c := &Check{
		checks: map[string]map[string]Checker{},
	}

	groupsN := 5
	checkersN := 3
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		for n := 0; n < checkersN; n++ {
			name := fmt.Sprintf("checker_%d_%d", g, n)
			cc := &testChecker{
				msg: fmt.Sprintf("%s is ok", name),
			}
			c.Register(name, group, cc)
		}
	}

	cc := &testChecker{
		msg:       "errored",
		wantError: true,
	}
	c.Register("error_check", "error_group", cc)

	status := c.Status()

	if status.Status != HCError {
		t.Fatalf("A health check with no checkers status should be healthy, it wasn't")
	}

	// Check groups length
	if len(status.OkResults) != groupsN+1 {
		t.Errorf("The number of groups in healthy ones is wrong, want: %d; got: %d", groupsN, len(status.OkResults))
	}

	if len(status.ErrorResults) != groupsN+1 {
		t.Errorf("The number of groups in unhealthy ones is wrong, want: %d; got: %d", groupsN, len(status.ErrorResults))
	}

	// Check healthy check results
	for g := 0; g < groupsN; g++ {
		group := fmt.Sprintf("group_%d", g)
		if len(status.OkResults[group]) != checkersN {
			t.Errorf("The number of healthy checks in group %s should be %d, got: %d", group, checkersN, len(status.OkResults[group]))
		}

		for n := 0; n < checkersN; n++ {
			name := fmt.Sprintf("checker_%d_%d", g, n)
			r, ok := status.OkResults[group][name]

			if !ok {
				t.Errorf("Check %s in group %s not present in healthy checks, it should", name, group)
			}

			msg := fmt.Sprintf("%s is ok", name)
			if r != msg {
				t.Errorf("Check %s in healthy group %s doesn't have the expected message: want: '%s', got: '%s'", name, group, r, msg)
			}
		}
	}

	// Check unhealthy check results
	r, ok := status.ErrorResults["error_group"]["error_check"]
	if !ok {
		t.Errorf("errored check is not present, it should")
	}

	if r != "errored" {
		t.Errorf("errored check result is wrong, want: %s; got: %s", "errored", r)
	}

}
