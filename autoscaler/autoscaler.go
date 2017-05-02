package autoscaler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/autoscaler/scale"
	"github.com/themotion/ladder/autoscaler/solve"
	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/health"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/metrics"
	"github.com/themotion/ladder/types"
)

const (
	hCGroup = "ladder_autoscaler"
)

const (
	// AutoscalerCtxKey the key that will represent the autoscaler name on the context
	AutoscalerCtxKey = "autoscaler"
)

// State is the state the autoscaler can be
type State int

const (
	// StateRunning describes running state
	StateRunning State = iota
	// StateStopped describes stopped state
	StateStopped
	// StateDisabled describes disabled state
	StateDisabled
)

// String implements the stringer interface
func (s State) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// Status is the status of the autoscaler
type Status struct {
	// The current state of the autoscaler
	State State
	// StopDeadline is the deadline of stop state
	StopDeadline time.Time
}

func (s Status) String() string {
	switch s.State {
	case StateStopped:
		return fmt.Sprintf("stopped for %s more", s.StopDeadline.Sub(time.Now().UTC()))
	default:
		return s.State.String()
	}
}

// Autoscaler is an interface with the methods needed to implment by an autoscaler
type Autoscaler interface {
	// Run will start the loop where the autoscaler will execute its logic
	Run() error
	// Stop will stop the loop where the autoscaler will execute its logic for the desired time
	Stop(duration time.Duration) error
	// CancelStop will cancel the current stop state
	CancelStop() error
	// Running will check if the autoscaler is running
	Running() bool
	// Status will return the status of the autoscaler
	Status() (Status, error)
}

// IntervalAutoscaler is the one that has the logic of detecting the downscale/upscale
// and triggering the scaler in regular intervals
type IntervalAutoscaler struct {
	Config             *config.Autoscaler
	Name               string
	Description        string
	DryRun             bool
	Interval           time.Duration
	Warmup             time.Duration
	ScalingWaitTimeout time.Duration

	// Autoscaler blocks
	Solver    solve.Solver      // The solver that will solve all the inputs, solver will deide who goes to the scaler
	Scaler    scale.Scaler      // The scaler that will scale the winning valu of all the inputs
	Inputters []inputter        // Multiple inputters for the autoscalers
	Filterers []filter.Filterer // Multiple filterers to filter before scaling

	// internal control fields
	intervalT    *time.Ticker // Ticker that sets the autoscaler pace
	startTime    time.Time    // Autoscaler start time
	running      bool         // Checker for the state of the autoscaler
	stateMu      *sync.Mutex
	ctx          context.Context    // Autoscaler iteration context
	cancel       context.CancelFunc // Cancel function
	done         chan struct{}      // channel to finish the infinite loop
	stopDone     chan struct{}      // channel to finish the waiting of stoped autoscaler
	stopDeadline time.Time          // the timestamp when the autoscaler will start running again

	log *log.Log // custom logger based on the global one
}

// NewIntervalAutoscaler returns a new autoscaler based on the configuration
func NewIntervalAutoscaler(c *config.Autoscaler, dryRun bool) (*IntervalAutoscaler, error) {
	if c == nil {
		return nil, fmt.Errorf("Configuration can't be nil")
	}

	if c.Name == "" {
		return nil, fmt.Errorf("An autoscaler name must be provided")
	}

	a := &IntervalAutoscaler{
		Config:             c,
		Name:               c.Name,
		Description:        c.Description,
		DryRun:             dryRun,
		Interval:           c.Interval,
		Warmup:             c.Warmup,
		ScalingWaitTimeout: c.ScalingWaitTimeout,

		running: false,
		stateMu: &sync.Mutex{},

		stopDone: make(chan struct{}),

		log: log.WithFields(log.Fields{
			"autoscaler": c.Name,
			"kind":       "autoscaler",
			"name":       c.Name,
		}),
	}

	// Renew the context
	a.renewContext()

	if err := a.setScaler(); err != nil {
		return nil, err
	}

	// Set the filterers
	fs := make([]filter.Filterer, len(c.Filters))
	for i, fc := range c.Filters {
		f, err := newFilter(a.ctx, &fc)
		if err != nil {
			return nil, fmt.Errorf("error creating the filters: %s", err)
		}
		fs[i] = *f
	}
	a.Filterers = fs
	a.log.Debugf("Created %d filters", len(fs))

	// If 1 inputter then no solvers
	if len(c.Inputters) < 2 {
		a.log.Warningf("No solver loaded, you only have one inputter")
	} else {
		if err := a.setSolver(); err != nil {
			return nil, err
		}
	}

	// Set the inputters
	if len(c.Inputters) == 0 {
		return nil, fmt.Errorf("no inputters on autoscaler")
	}
	inputters := make([]inputter, len(c.Inputters))
	for i, inC := range c.Inputters {
		in, err := newInputter(a.ctx, &inC)
		if err != nil {
			return nil, fmt.Errorf("error creating the inputters: %s", err)
		}
		inputters[i] = *in
	}

	a.Inputters = inputters

	a.log.Infof("Autoscaler '%s' created", a.Name)
	if a.Interval == 0 {
		a.log.Warningf("Interval for '%s' is 0", a.Name)
	}

	// Register autoscaler on healthcheck
	health.Register(c.Name, hCGroup, a)

	return a, nil
}

func newFilter(ctx context.Context, c *config.Block) (*filter.Filterer, error) {
	if c.Kind == "" {
		err := fmt.Errorf("filter kind can't be empty")
		return nil, err
	}

	// Create new filterer using the registry and the creators
	f, err := filter.Create(ctx, c.Kind, c.Config)

	if err != nil {
		return nil, err
	}

	return &f, nil
}

// setScaler returns the correct Scaler
func (a *IntervalAutoscaler) setScaler() error {
	if a.Scaler == nil {

		if a.Config.Scale.Kind == "" {
			err := fmt.Errorf("scaler type can't be empty")
			return err
		}

		// Create new scaler using the registry and the creators
		s, err := scale.Create(a.ctx, a.Config.Scale.Kind, a.Config.Scale.Config)

		if err != nil {
			return err
		}
		a.Scaler = s
		a.log.Debugf("Scaler created")
		return nil
	}
	a.log.Debugf("Scaler already created, ignoring")
	return nil
}

// setsolver returns the correct solver
func (a *IntervalAutoscaler) setSolver() error {
	if a.Solver == nil {

		if a.Config.Solve.Kind == "" {
			err := fmt.Errorf("solver type can't be empty")
			return err
		}

		// Create new solver using the registry and the creators
		s, err := solve.Create(a.ctx, a.Config.Solve.Kind, a.Config.Solve.Config)

		if err != nil {
			return err
		}
		a.Solver = s
		a.log.Debugf("Solver created")
		return nil
	}
	a.log.Debugf("Solver already created, ignoring")
	return nil
}

// gatherWinningInput will start the gathering the inputs
func (a *IntervalAutoscaler) gatherWinningInput(currentQ types.Quantity) (types.Quantity, error) {
	a.log.Debugf("Get the winning input for the scaler, start running %d inputters", len(a.Inputters))

	// The inputs will send their results through this channel
	inputChan := make(chan types.Quantity)

	// The inputs will send their errors through this channel if they error obviously
	errChan := make(chan error)

	// Sync all the results
	wg := sync.WaitGroup{}
	wg.Add(len(a.Inputters) + 1) // Add one for the results gatherer

	// the results
	inputs := []types.Quantity{}
	inputErrors := []error{}

	// Gather all inputs
	for _, in := range a.Inputters {
		go func(in inputter) {
			defer wg.Done()
			startS := time.Now().UTC()
			inQ, errS := in.gatherAndArrange(a.ctx, currentQ) // use different name for err to avoid annoying message of shadowing variable of go vet
			metrics.ObserveInputterDuration(time.Now().UTC().Sub(startS), a.Name, in.name)

			if errS != nil {
				metrics.AddInputterErrors(1, a.Name, in.name)
				errChan <- errS
				return
			}
			metrics.SetInputterQ(inQ, a.Name, in.name)
			inputChan <- inQ
		}(in)
	}
	// grab all the results
	go func() {
		// We know the number of results
		for i := 0; i < len(a.Inputters); i++ {
			select {
			case errS := <-errChan: // use different name for err to avoid annoying message of shadowing variable of go vet
				inputErrors = append(inputErrors, errS)
			case input := <-inputChan:
				inputs = append(inputs, input)
			}
		}
		wg.Done()
	}()

	// Wait for all the inputs
	wg.Wait()
	close(inputChan)
	close(errChan)

	// Solve all the inputs
	start := time.Now().UTC()
	newQ, err := a.solve(inputs, inputErrors)
	metrics.ObserveSolverDuration(time.Now().UTC().Sub(start), a.Name, a.Config.Solve.Kind)

	if err != nil {
		metrics.AddSolverErrors(1, a.Name, a.Config.Solve.Kind)
		a.log.Errorf("error gathering inputs: %s", err)
	} else {
		metrics.SetSolverQ(newQ, a.Name, a.Config.Solve.Kind)
		a.log.Infof("Winner inputter of %s set new input: %s", a.Name, newQ)
	}

	return newQ, err
}

// Renews the context of the autoscaler
func (a *IntervalAutoscaler) renewContext() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, AutoscalerCtxKey, a.Name)

	a.ctx, a.cancel = context.WithCancel(ctx)
}

// solve will solve the system with multiple inputs
func (a *IntervalAutoscaler) solve(inputs []types.Quantity, inputErrors []error) (newQ types.Quantity, err error) {
	// Notify about the errors
	if len(inputErrors) > 0 {
		// Create a massive error
		errStr := fmt.Sprintf("solver got %d errors from inputs:", len(inputErrors))

		for _, err := range inputErrors {
			errStr = fmt.Sprintf("%s %s;", errStr, err)
		}
		a.log.Warningf(errStr)
	}
	if len(inputs) == 0 {
		return newQ, fmt.Errorf("solver didn't receive any input values from the inputters")
	}

	// resolve the conflict of multiple inputs if necessary
	if len(a.Inputters) > 1 {
		newQ, err = a.Solver.Solve(a.ctx, inputs)
		if err != nil {
			err = fmt.Errorf("solver error: %s", err)
		}
	} else {
		newQ = inputs[0]
	}
	return
}

// filter will pass the newQ through all the filters to create a new filtered quantity or not
func (a *IntervalAutoscaler) filter(currentQ, newQ types.Quantity) (filteredQ types.Quantity, err error) {
	a.log.Debugf("Start %d filterers chain", len(a.Filterers))
	filteredQ = newQ
	var br bool
	// Apply all the filterers
	for _, f := range a.Filterers {
		filteredQ, br, err = f.Filter(a.ctx, currentQ, filteredQ)
		if err != nil {
			// Breaking with an error will stop this iteration of the autoscaler
			a.log.Warnf("filterer breaked the chain with an error: %v", err)
			return
		}
		if br {
			a.log.Infof("filterer breaked the chain")
			return
		}
	}

	return
}

// scale will scale the system with arranges new value (or not)
func (a *IntervalAutoscaler) scale(newQ types.Quantity) (types.ScalingMode, error) {
	// Scale with the new value if not on dry run
	if !a.DryRun {
		start := time.Now().UTC()
		scaledQ, mode, err := a.Scaler.Scale(a.ctx, newQ)
		metrics.ObserveScalerDuration(time.Now().UTC().Sub(start), a.Name, a.Config.Scale.Kind)

		if err != nil {
			return mode, fmt.Errorf("error scalating: %s", err)
		}

		if mode != types.NotScaling {
			a.log.Infof("Scaler scalated to %s", newQ)

			// If 0 then don't timeout
			if a.ScalingWaitTimeout == 0 {
				if err := a.Scaler.Wait(a.ctx, scaledQ, mode); err != nil {
					return mode, fmt.Errorf("error waiting after scalation: %s", err)
				}
			} else {
				// Wait after a successful scalation, allow timeouts
				waitChan := make(chan error)
				go func() {
					a.log.Infof("Scaler will wait until scalation confirmed...")
					waitChan <- a.Scaler.Wait(a.ctx, scaledQ, mode)
				}()

				select {
				// Timeout
				case <-time.After(a.ScalingWaitTimeout):
					return mode, fmt.Errorf("Timeout of waiting for scalation wait process")
					// Wait result
				case err := <-waitChan:
					if err != nil {
						return mode, fmt.Errorf("error waiting after scalation: %s", err)
					}
				}
			}
		} else {
			a.log.Infof("Scaler didn't scalated")
		}

		return mode, nil
	}

	a.log.Infof("Dry run mode: Not scaling '%s' new quantity", newQ)
	return types.NotScaling, nil
}

// Run will run the whole autoscaler logic
func (a *IntervalAutoscaler) Run() error {

	a.stateMu.Lock()

	// Check is running
	if a.running {
		a.stateMu.Unlock()
		return fmt.Errorf("Autoscaler '%s' is already running", a.Name)
	}

	// Set to running
	a.running = true
	a.stateMu.Unlock()
	// When finished set to not running
	defer func() {
		a.log.Infof("Cleaning '%s' autoscaler", a.Name)
		a.stateMu.Lock()
		a.running = false
		a.stateMu.Unlock()
		a.log.Infof("Autoscaler '%s' finished running", a.Name)
	}()

	a.log.Infof("Start running '%s' autoscaler", a.Name)
	a.startTime = time.Now().UTC()
	a.intervalT = time.NewTicker(a.Interval)

	// Set status metric
	metrics.SetAutoscalerRunning(true, a.Name)

	// Create a finish channel to stop main infinite loop, don't use context, context are used
	// only per loop iteration
	a.done = make(chan struct{})

	// Avoid creating over and over again the same variables in an infinite loop, relaxing the GC a little bit
	var start time.Time
	var err error

	// Start our main loop every tick of the ticker until done
	for {
		select {
		case <-a.intervalT.C:
			// Renew the context for current iteration
			a.renewContext()

			metrics.AddAutoscalerIteration(1, a.Name)

			start = time.Now().UTC()
			err = a.autoscalerIterationLogic()
			metrics.ObserveAutoscalerDuration(time.Now().UTC().Sub(start), a.Name)
			if err != nil {
				metrics.AddAutoscalerErrors(1, a.Name)
				a.log.Errorf("Error on iteration: %v", err)
			}
		case <-a.done:
			return nil
		}
	}
}

// autoscalerIterationLogic has the logc of gathering the input and scaling that will be in a constant loop
func (a *IntervalAutoscaler) autoscalerIterationLogic() error {
	// Get current value
	start := time.Now().UTC()
	currentQ, err := a.Scaler.Current(a.ctx)
	metrics.ObserveCurrentDuration(time.Now().UTC().Sub(start), a.Name, a.Config.Scale.Kind)
	if err != nil {
		metrics.AddCurrentErrors(1, a.Name, a.Config.Scale.Kind)
		err = fmt.Errorf("error gathering current scaler value: %s", err)
		return err
	}
	metrics.SetCurrentQ(currentQ, a.Name, a.Config.Scale.Kind)

	// Get the input for the scaler
	newQ, err := a.gatherWinningInput(currentQ)

	if err != nil {

		return err
	}

	// Apply all the filters
	start = time.Now().UTC()
	newFQ, err := a.filter(currentQ, newQ)
	metrics.ObserveFiltererDuration(time.Now().UTC().Sub(start), a.Name)
	if err != nil {
		metrics.AddFiltererErrors(1, a.Name)
		return fmt.Errorf("error filtering new  value: %s", err)
	}
	if newFQ != newQ {
		a.log.Infof("Filterers set new scaling output from %v to %v", newQ, newFQ)
		newQ = newFQ
	}

	// check for warmup, if warming up don't call scale
	if time.Now().UTC().Sub((a.startTime)) < a.Warmup {
		a.log.Infof("Warming up, not calling scale with %v", newQ)
		return nil
	}

	// Scale! (or not)
	metrics.SetScalerQ(newQ, a.Name, a.Config.Scale.Kind)
	_, err = a.scale(newQ)
	if err != nil {
		metrics.AddScalerErrors(1, a.Name, a.Config.Scale.Kind)
		return err
	}
	return nil
}

// stopForever stops the execution of the autoscaler
func (a *IntervalAutoscaler) stopForever() error {
	// Check is running
	if !a.running {
		return fmt.Errorf("Autoscaler '%s' is not running", a.Name)
	}

	// Stop main infinite loop
	a.intervalT.Stop()
	close(a.done)

	// Stop current iteration
	a.cancel()

	// Set status metric
	metrics.SetAutoscalerRunning(false, a.Name)
	a.log.Debugf("Autoscaler stopped")

	return nil
}

// Stop stops the execution of the autoscaler for the time desired
func (a *IntervalAutoscaler) Stop(duration time.Duration) error {

	// Stop the autoscaler
	if err := a.stopForever(); err != nil {
		return err
	}

	// Start the autoscaler after duration
	go func() {
		// Set when are we going to start again
		a.stateMu.Lock()
		a.stopDeadline = time.Now().UTC().Add(duration)
		a.stateMu.Unlock()
		a.log.Infof("Stopping autoscaler for %s (until %s)", duration, a.stopDeadline)
	StopLoop:
		for {
			select {
			case <-time.After(duration):
				a.log.Infof("Autoscaler stop deadline reached")
				break StopLoop
			case <-a.stopDone:
				a.log.Infof("Autoscaler stop candeled")
				break StopLoop
			}
		}
		a.Run()
	}()

	return nil
}

// CancelStop will cancel the current stop
func (a *IntervalAutoscaler) CancelStop() error {
	// Check is running
	if a.running {
		return fmt.Errorf("Autoscaler '%s' is in running", a.Name)
	}

	a.log.Infof("Cancelling '%s' autoscaler stop", a.Name)
	a.stopDone <- struct{}{}

	return nil
}

// Running returns true if the autoscaler is running, false if not
func (a *IntervalAutoscaler) Running() bool {
	return a.running
}

// Status returns the status of the autoscaler
func (a *IntervalAutoscaler) Status() (Status, error) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()

	var st Status
	if a.running {
		st = Status{
			State: StateRunning,
		}
	} else {
		st = Status{
			State:        StateStopped,
			StopDeadline: a.stopDeadline,
		}
	}
	return st, nil
}

// Check implements the Checker interface (for the healthchecks)
func (a *IntervalAutoscaler) Check() (string, error) {
	st, err := a.Status()
	return st.String(), err
}
