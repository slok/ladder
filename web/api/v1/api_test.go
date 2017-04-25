package v1

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/themotion/ladder/autoscaler"
)

type mockAutoscaler struct {
	runCntr      int
	cancelCntr   int
	stopCntr     int
	wantErrRun   bool
	wantErrStop  bool
	wantErrCheck bool
	running      bool
	duration     time.Duration
}

func (a *mockAutoscaler) Run() error {
	a.runCntr++
	if a.wantErrStop {
		return errors.New("Want stop error!")
	}
	a.running = true
	return nil
}
func (a *mockAutoscaler) Stop(duration time.Duration) error {
	a.stopCntr++
	if a.wantErrStop {
		return errors.New("Want stop error!")
	}
	a.running = false
	a.duration = duration
	return nil
}
func (a *mockAutoscaler) Running() bool { return a.running }
func (a *mockAutoscaler) CancelStop() error {
	a.cancelCntr++
	a.Run()
	return nil
}
func (a *mockAutoscaler) Status() (autoscaler.Status, error) {
	if a.wantErrCheck {
		return autoscaler.Status{}, errors.New("Want check error!")
	}

	if a.running {
		return autoscaler.Status{State: autoscaler.StateRunning}, nil
	}

	return autoscaler.Status{
		State:        autoscaler.StateStopped,
		StopDeadline: time.Now().UTC().Add(a.duration)}, nil
}

func makeMockAutoscalers() map[string]autoscaler.Autoscaler {
	return map[string]autoscaler.Autoscaler{
		"running_run_ok_stop_ok_check_ok":  &mockAutoscaler{running: true, wantErrRun: false, wantErrStop: false, wantErrCheck: false},
		"running_run_err_stop_ok_check_ok": &mockAutoscaler{running: true, wantErrRun: true, wantErrStop: false, wantErrCheck: false},
		"running_run_ok_stop_err_check_ok": &mockAutoscaler{running: true, wantErrRun: false, wantErrStop: true, wantErrCheck: false},
		"running_run_ok_stop_ok_check_err": &mockAutoscaler{running: true, wantErrRun: false, wantErrStop: false, wantErrCheck: true},
		"stopped_run_ok_stop_ok_check_ok":  &mockAutoscaler{running: false, wantErrRun: false, wantErrStop: false, wantErrCheck: false},
		"stopped_run_err_stop_ok_check_ok": &mockAutoscaler{running: false, wantErrRun: true, wantErrStop: false, wantErrCheck: false},
		"stopped_run_ok_stop_err_check_ok": &mockAutoscaler{running: false, wantErrRun: false, wantErrStop: true, wantErrCheck: false},
		"stopped_run_ok_stop_ok_check_err": &mockAutoscaler{running: false, wantErrRun: false, wantErrStop: false, wantErrCheck: true},
	}
}

func TestFixPrefix(t *testing.T) {
	tests := []struct {
		prefix     string
		wantPrefix string
		shouldErr  bool
	}{
		{"/api/v1", "/api/v1", false},
		{"/api/v1/", "/api/v1", false},
		{"api/v1/", "/api/v1", false},
		{"api", "/api", false},
		{"/api", "/api", false},
		{"/api/", "/api", false},
		{`/api!"·$%&/()"/`, "/api", true},
	}

	for _, test := range tests {
		got, err := fixPrefix(test.prefix)

		if !test.shouldErr {
			if err != nil {
				t.Errorf("%+v\n - Fixing the prefix shouldn't error: %s", test, err)
			}

			if got != test.wantPrefix {
				t.Errorf("%+v\n - fixed prefix isn't the expected, got: %s; want: %s", test, got, test.wantPrefix)
			}
		} else {
			if err == nil {
				t.Errorf("%+v\n - Fixing the prefix should errored, it didn't", test)
			}
		}
	}
}

func TestCancelStopAutoscaler(t *testing.T) {
	tests := []struct {
		autoscaler string
		method     string
		wantCode   int
		wantBody   string
		wantCall   bool
	}{
		{"stopped_run_ok_stop_ok_check_ok", "PUT", 202, `{"autoscaler":"stopped_run_ok_stop_ok_check_ok","msg":"Autoscaler stop cancel request sent"}`, true},
		{"running_run_ok_stop_ok_check_ok", "PUT", 400, `{"data":{"autoscaler":"running_run_ok_stop_ok_check_ok","msg":"Autoscaler is not stopped"},"error":"Autoscaler is not stopped"}`, false},
		{"nope", "PUT", 400, `{"data":{"autoscaler":"nope","msg":"nope is not a valid autoscaler"},"error":"nope is not a valid autoscaler"}`, false},
	}

	for _, test := range tests {
		autoscalers := makeMockAutoscalers()
		api := APIV1{autoscalers: autoscalers}
		router := httprouter.New()
		api.Register(router)

		// Make the request
		url := fmt.Sprintf("/autoscalers/%s/cancel-stop", test.autoscaler)
		req := httptest.NewRequest(test.method, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		time.Sleep(1 * time.Millisecond) // Make time to let the goroutine execute

		// Check the request
		if w.Code != test.wantCode {
			t.Errorf("%+v\n -Received code from the API is wrong, got: %d, want: %d", test, w.Code, test.wantCode)
		}

		gotB := strings.TrimSpace(w.Body.String())
		if gotB != test.wantBody {
			t.Errorf("%+v\n -Received body from the API is wrong, \ngot: %s, \nwant: %s", test, gotB, test.wantBody)
		}

		if test.wantCall {
			aa, ok := autoscalers[test.autoscaler].(*mockAutoscaler)
			if !ok {
				t.Errorf("%+v\n -Autoscaler '%s' should exist, id din't", test, test.autoscaler)
			}

			if aa.cancelCntr <= 0 {
				t.Errorf("%+v\n -Autoscaler '%s' stop should be cancelled at least once: %d", test, test.autoscaler, aa.runCntr)
			}

			if aa.runCntr <= 0 {
				t.Errorf("%+v\n -Autoscaler '%s' should be run at least once: %d", test, test.autoscaler, aa.runCntr)
			}
		}
	}

}

func TestStopAutoscaler(t *testing.T) {
	tests := []struct {
		autoscaler string
		duration   string
		method     string
		wantCode   int
		wantBody   string
		wantCall   bool
	}{
		{
			"stopped_run_ok_stop_ok_check_ok", "10ms",
			"PUT",
			409,
			fmt.Sprintf(`{"data":{"autoscaler":"stopped_run_ok_stop_ok_check_ok","deadline":%d,"msg":"Autoscaler already stopped","required-action":"Need to cancel current stop state first"},"error":"Autoscaler already stopped"}`, time.Now().UTC().Add(10*time.Millisecond).Unix()),
			false,
		},
		{
			"nope",
			"10ms",
			"PUT",
			400,
			`{"data":{"autoscaler":"nope","msg":"nope is not a valid autoscaler"},"error":"nope is not a valid autoscaler"}`,
			false,
		},
		{
			"running_run_ok_stop_ok_check_ok",
			"10ms",
			"PUT",
			202,
			`{"autoscaler":"running_run_ok_stop_ok_check_ok","msg":"Autoscaler stop request sent"}`,
			true,
		},
		{
			"running_run_ok_stop_ok_check_ok",
			"1h15m45s",
			"PUT",
			202,
			`{"autoscaler":"running_run_ok_stop_ok_check_ok","msg":"Autoscaler stop request sent"}`,
			true,
		},
		{
			"running_run_ok_stop_ok_check_ok",
			"24h",
			"PUT",
			202,
			`{"autoscaler":"running_run_ok_stop_ok_check_ok","msg":"Autoscaler stop request sent"}`,
			true,
		},
		{
			"running_run_ok_stop_ok_check_ok",
			"",
			"PUT",
			404,
			`404 page not found`,
			false,
		},
		{
			"running_run_ok_stop_ok_check_ok",
			"something",
			"PUT",
			400,
			`{"data":{"autoscaler":"running_run_ok_stop_ok_check_ok","msg":"duration parameter is not valid"},"error":"duration parameter is not valid"}`,
			false,
		},
	}

	for _, test := range tests {
		autoscalers := makeMockAutoscalers()
		api := APIV1{autoscalers: autoscalers}
		router := httprouter.New()
		api.Register(router)

		// Make the request
		url := fmt.Sprintf("/autoscalers/%s/stop/%s", test.autoscaler, test.duration)
		req := httptest.NewRequest(test.method, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		time.Sleep(1 * time.Millisecond) // Make time to let the goroutine execute

		// Check the request
		if w.Code != test.wantCode {
			t.Errorf("%+v\n -Received code from the API is wrong, got: %d, want: %d", test, w.Code, test.wantCode)
		}

		gotB := strings.TrimSpace(w.Body.String())
		if gotB != test.wantBody {
			t.Errorf("%+v\n -Received body from the API is wrong, \ngot: %s, \nwant: %s", test, gotB, test.wantBody)
		}

		if test.wantCall {
			aa, ok := autoscalers[test.autoscaler].(*mockAutoscaler)
			if !ok {
				t.Errorf("%+v\n -Autoscaler '%s' should exist, id din't", test, test.autoscaler)
			}

			if aa.stopCntr <= 0 {
				t.Errorf("%+v\n -Autoscaler '%s' should be stopped at least once: %d", test, test.autoscaler, aa.runCntr)
			}
			td, _ := time.ParseDuration(test.duration)
			if aa.duration != td {
				t.Errorf("%+v\n -Autoscaler '%s' duration expected: %s; got: %s", test, test.autoscaler, td, aa.duration)
			}
		}
	}

}

func TestListAutoscaler(t *testing.T) {
	type test struct {
		name   string
		status autoscaler.State
	}

	tests := []struct {
		autoscalers []test
		wantBody    string
		wantCode    int
	}{
		{
			autoscalers: []test{
				test{"asg1", autoscaler.StateStopped},
			},
			wantBody: `{"autoscalers":{"asg1":{"status":"stopped"}}}`,
			wantCode: 200,
		},
		{
			autoscalers: []test{
				test{"asg1", autoscaler.StateRunning},
				test{"asg2", autoscaler.StateRunning},
			},
			wantBody: `{"autoscalers":{"asg1":{"status":"running"},"asg2":{"status":"running"}}}`,
			wantCode: 200,
		},
		{
			autoscalers: []test{
				test{"asg1", autoscaler.StateRunning},
				test{"asg2", autoscaler.StateStopped},
				test{"asg3", autoscaler.StateRunning},
				test{"asg4", autoscaler.StateStopped},
			},
			wantBody: `{"autoscalers":{"asg1":{"status":"running"},"asg2":{"status":"stopped"},"asg3":{"status":"running"},"asg4":{"status":"stopped"}}}`,
			wantCode: 200,
		},
		{
			autoscalers: []test{},
			wantBody:    `{"autoscalers":{}}`,
			wantCode:    200,
		},
	}

	for _, test := range tests {
		autoscalers := map[string]autoscaler.Autoscaler{}
		for _, as := range test.autoscalers {
			running := false
			if as.status == autoscaler.StateRunning {
				running = true
			}
			autoscalers[as.name] = &mockAutoscaler{running: running}
		}

		api := APIV1{autoscalers: autoscalers}
		router := httprouter.New()
		api.Register(router)

		// Make the request
		req := httptest.NewRequest("GET", "/autoscalers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the request
		if w.Code != test.wantCode {
			t.Errorf("%+v\n -Received code from the API is wrong, got: %d, want: %d", test, w.Code, test.wantCode)
		}

		gotB := strings.TrimSpace(w.Body.String())
		if gotB != test.wantBody {
			t.Errorf("%+v\n -Received body from the API is wrong, \ngot: %s, \nwant: %s", test, gotB, test.wantBody)
		}

	}
}
