// +build integration

package autoscaler

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/config"
)

func TestIntegrationCorrectAutoScalerRun(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    50 * time.Millisecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}
	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Run
	go func() {
		err = as.Run()
	}()

	// Give time to start first the background goroutine
	time.Sleep(5 * time.Millisecond)

	// Check no error on the first Run
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Stop in 1 second
	time.Sleep(1 * time.Second)
	as.stopForever()

	// Check scalation history
	expected := 20
	if len(as.Scaler.(*testScaler).history) != expected {
		t.Errorf("\n- run result of scalation has wrong length; got: %d; expected: %d", len(as.Scaler.(*testScaler).history), expected)
	}

	for i, got := range as.Scaler.(*testScaler).history {
		if got.Q != int64(i+1) {
			t.Errorf("\n- run result of scalation is wrong; got: %d; expected: %d", got.Q, i+1)
		}
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationCorrectAutoScalerRunFilterers(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    50 * time.Millisecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}
	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	f1 := int64(10)
	f2 := int64(20)
	// Set our filters
	as.Filterers = []filter.Filterer{&testFilterer{resAdd: f1}, &testFilterer{resAdd: f2}}

	// Run
	go func() {
		err = as.Run()
	}()

	// Give time to start first the background goroutine
	time.Sleep(5 * time.Millisecond)

	// Check no error on the first Run
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Stop in 1 second
	time.Sleep(1 * time.Second)
	as.stopForever()

	// Check scalation history, one beacuse only scales first time, then the filters
	// sets the same quantity over and over again
	expected := 20
	if len(as.Scaler.(*testScaler).history) != expected {
		t.Errorf("\n- run result of scalation has wrong length; got: %d; expected: %d", len(as.Scaler.(*testScaler).history), expected)
	}

	wantResult := []int64{31, 60, 89, 118, 147, 176, 205, 234, 263, 292, 321, 350, 379, 408, 437, 466, 495, 524, 553, 582}
	for i, got := range as.Scaler.(*testScaler).history {
		// Always same result set by our filter
		if got.Q != wantResult[i] {
			t.Errorf("\n- run result of scalation is wrong; got: %d; expected: %d", got.Q, i+1)
		}
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationCorrectAutoScalerRunWarmUp(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    50 * time.Millisecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Warmup:      500 * time.Millisecond,
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}
	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Run
	go func() {
		err = as.Run()
	}()

	// Give time to start first the background goroutine
	time.Sleep(5 * time.Millisecond)

	// Check no error on the first Run
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Stop in 1 second
	time.Sleep(1 * time.Second)
	as.stopForever()

	// Check scalation history
	expected := 11 // Warm up will allow scaling after 500ms, this is on the 9th iteration
	if len(as.Scaler.(*testScaler).history) != expected {
		t.Errorf("\n- run result of scalation has wrong length; got: %d; expected: %d", len(as.Scaler.(*testScaler).history), expected)
	}

	for i, got := range as.Scaler.(*testScaler).history {
		if got.Q != int64(i+1) {
			t.Errorf("\n- run result of scalation is wrong; got: %d; expected: %d", got.Q, i+1)
		}
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationCorrectAutoScalerDryRun(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    50 * time.Millisecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}
	as, err := NewIntervalAutoscaler(c, true)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Run
	go func() {
		err = as.Run()
	}()

	// Give time to start first the background goroutine
	time.Sleep(1 * time.Millisecond)

	// Check no error on the first Run
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	// Stop in 1 second
	time.Sleep(5 * time.Second)
	as.stopForever()

	// Check scalation history
	expected := 0
	if len(as.Scaler.(*testScaler).history) != expected {
		t.Errorf("\n- run result of scalation has wrong length; got: %d; expected: %d", len(as.Scaler.(*testScaler).history), expected)
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationMultipleRuns(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    50 * time.Millisecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}
	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	go func() {
		err = as.Run()
	}()

	// Give time to start first the background goroutine
	time.Sleep(1 * time.Millisecond)

	// Check no error on the first Run
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	//Check for an error
	err = as.Run()

	// Check error on second run
	if err == nil {
		t.Errorf("\n- IntervalAutoscaler should give an error, it didn't")
	}
	as.stopForever()

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationStop(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    1 * time.Minute,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}

	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	var finished bool
	go func() {
		as.Run()
		finished = true
	}()

	// Give time to start first the background goroutine
	time.Sleep(20 * time.Millisecond)

	// Check is running
	if !as.running {
		t.Errorf("\n- IntervalAutoscaler should be running")
	}

	// stop for 10ms
	as.Stop(10 * time.Millisecond)

	// wait 5ms and check
	time.Sleep(5 * time.Millisecond)

	// Check finished
	if as.running && finished {
		t.Errorf("\n- IntervalAutoscaler shouldn't be running")
	}

	// wait 10ms and check is running
	time.Sleep(10 * time.Millisecond)

	// Check finished and started
	if !as.running && finished {
		t.Errorf("\n- IntervalAutoscaler should be running after stop duration")
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationCancelStop(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    1 * time.Minute,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}

	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	var finished bool
	go func() {
		as.Run()
		finished = true
	}()

	// Give time to start first the background goroutine
	time.Sleep(20 * time.Millisecond)

	// Check is running
	if !as.running {
		t.Errorf("\n- IntervalAutoscaler should be running")
	}

	// stop for 50ms
	as.Stop(50 * time.Millisecond)

	// wait 5ms and check
	time.Sleep(5 * time.Millisecond)

	// Check finished
	if as.running && finished {
		t.Errorf("\n- IntervalAutoscaler shouldn't be running")
	}

	// Cancel stop
	if err := as.CancelStop(); err != nil {
		t.Errorf("\n- IntervalAutoscaler stop cancelation shouldn't give error")
	}

	// Wait a bit to start and check finished and started
	time.Sleep(2 * time.Millisecond)
	if !as.running && finished {
		t.Errorf("\n- IntervalAutoscaler should be running after stop cancelation")
	}

	// Stop & leave time to clean
	as.stopForever()
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationStopNotRunning(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    1 * time.Minute,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}

	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	err = as.Stop(10 * time.Millisecond)

	// Check finished
	if err == nil {
		t.Errorf("\n- Stop should error when not running, it didn't")
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationCancelStopNotStopped(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    1 * time.Minute,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0"},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}

	as, err := NewIntervalAutoscaler(c, false)
	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}
	go func() {
		as.Run()
	}()
	time.Sleep(5 * time.Millisecond)

	// Cancel stop
	if err := as.CancelStop(); err == nil {
		t.Errorf("\n- IntervalAutoscaler stop cancelation should give error when running")
	}

	as.stopForever()

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationWithError(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    10 * time.Nanosecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0", Config: map[string]interface{}{"return_error": true}},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}

	as, err := NewIntervalAutoscaler(c, false)
	// Set a custom logger
	lOut := &bytes.Buffer{}
	as.log.Logger.Out = lOut

	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	var finished bool
	go func() {
		as.Run()
		finished = true
	}()

	// Give time to start first the background goroutine
	time.Sleep(20 * time.Millisecond)

	// Check is running
	if !as.running {
		t.Errorf("\n- IntervalAutoscaler should be running")
	}

	as.stopForever()

	// Check finished
	if as.running && finished {
		t.Errorf("\n- IntervalAutoscaler shouldn't be running")
	}

	// Check error
	errMsgs := []string{
		"solver got 1 errors from inputs: error gathering input: Error!;",
		"Error on iteration: solver didn't receive any input values from the inputters",
	}
	for _, subStr := range errMsgs {
		if !strings.Contains(lOut.String(), subStr) {
			t.Errorf("\n- Error string not found on logger, it should appear")
		}
	}
	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}

func TestIntegrationWithMultipleInputtersSomeError(t *testing.T) {
	registerDummies()

	c := &config.Autoscaler{
		Name:        "autoscaler-test",
		Description: "autoscaler-test description",
		Interval:    10 * time.Nanosecond,
		Scale:       config.Block{Kind: "test0"},
		Solve:       config.Block{Kind: "test0"},
		Inputters: []config.Inputter{
			config.Inputter{
				Name:    "test0_input",
				Gather:  config.Block{Kind: "test0", Config: map[string]interface{}{"return_error": true}},
				Arrange: config.Block{Kind: "test0"},
			},
			config.Inputter{
				Name:    "test1_input",
				Gather:  config.Block{Kind: "test0", Config: map[string]interface{}{"return_error": false}},
				Arrange: config.Block{Kind: "test0"},
			},
			config.Inputter{
				Name:   "test2_input",
				Gather: config.Block{Kind: "test0", Config: map[string]interface{}{"return_error": false}},
			},
			config.Inputter{
				Name:    "test3_input",
				Gather:  config.Block{Kind: "test0", Config: map[string]interface{}{"return_error": true}},
				Arrange: config.Block{Kind: "test0"},
			},
			config.Inputter{
				Name:    "test4_input",
				Gather:  config.Block{Kind: "test0", Config: map[string]interface{}{"return_error": true}},
				Arrange: config.Block{Kind: "test0"},
			},
		},
	}

	as, err := NewIntervalAutoscaler(c, false)

	// Set a custom logger
	lOut := &bytes.Buffer{}
	as.log.Logger.Out = lOut

	if err != nil {
		t.Fatalf("\n- IntervalAutoscaler shouldn't give an error: %v", err)
	}

	var finished bool
	go func() {
		as.Run()
		finished = true
	}()

	// Give time to start first the background goroutine
	time.Sleep(20 * time.Millisecond)

	// Check is running
	if !as.running {
		t.Errorf("\n- IntervalAutoscaler should be running")
	}

	as.stopForever()

	// Check finished
	if as.running && finished {
		t.Errorf("\n- IntervalAutoscaler shouldn't be running")
	}
	// Check inputter error messages &  not the final solver error
	subStr := "solver got 3 errors from inputs: error gathering input: Error!; error gathering input: Error!; error gathering input: Error!;"
	if !strings.Contains(lOut.String(), subStr) {
		t.Errorf("\n- Inputters error string warning not found on logger, it should appear")
	}
	subStr = "Error on iteration: solver didn't receive any input values from the inputters"
	if strings.Contains(lOut.String(), subStr) {
		t.Errorf("\n- Error string found on logger, it shouldn't appear")
	}

	// Leave time to clean
	time.Sleep(10 * time.Millisecond)
}
