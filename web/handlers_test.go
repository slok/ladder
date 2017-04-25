package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/themotion/ladder/health"
)

type TestChecker bool

func (t TestChecker) Check() (string, error) {
	if bool(t) {
		return "Error", errors.New("Error")
	}

	return "No error", nil
}

func TestCheckHandlerOK(t *testing.T) {
	// Create a false healthcheck
	c := TestChecker(false)
	hc := health.NewCheck()
	hc.Register("test", "test", c)

	// Make the request
	req := httptest.NewRequest("GET", "/check", nil)
	w := httptest.NewRecorder()
	handler := healthCheckHandler(hc)
	handler(w, req, nil)

	// Check status code
	wantSt := http.StatusOK
	if w.Result().StatusCode != wantSt {
		t.Errorf("Wrong status code received by the healtcheck, want: %d, got: %d", wantSt, w.Result().StatusCode)
	}

	// Simple string check of the body
	d := w.Body.String()
	wantBd := `"healthy":{"test":{"test":"No error"}}`
	if !strings.Contains(d, wantBd) {
		t.Errorf("REsponse didn't cotain the expected body\n want:%s\n got:%s", d, wantBd)
	}
}

func TestCheckHandlerError(t *testing.T) {
	// Create a false healthcheck
	c := TestChecker(true)
	hc := health.NewCheck()
	hc.Register("test", "test", c)

	// Make the request
	req := httptest.NewRequest("GET", "/check", nil)
	w := httptest.NewRecorder()
	handler := healthCheckHandler(hc)
	handler(w, req, nil)

	// Check status code
	want := http.StatusInternalServerError
	if w.Result().StatusCode != want {
		t.Errorf("Wrong status code received by the healtcheck, want: %d, got: %d", want, w.Result().StatusCode)
	}

	// Simple string check of the body
	d := w.Body.String()
	wantBd := `"unhealthy":{"test":{"test":"Error"}}`
	if !strings.Contains(d, wantBd) {
		t.Errorf("REsponse didn't cotain the expected body\n want:%s\n got:%s", d, wantBd)
	}
}
