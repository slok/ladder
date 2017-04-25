package health

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/version"
)

var (
	// MainCheck is the checker that will have all the checkers of the app
	MainCheck *Check

	// startTime is the time the app started
	startTime time.Time
)

// Register will register a checker on the app health check
func Register(name, group string, checker Checker) {
	MainCheck.Register(name, group, checker)
}

// Status will run all the app checks and return a checkStatus as result
func Status() *CheckStatus {
	return MainCheck.Status()
}

// Uptime gets the uptime of the application
func Uptime() time.Duration {
	return time.Now().UTC().Sub(startTime)
}

// On first import this will create the app checker and set the uptime
// Important, this should be import on apps main
func init() {
	log.Logger.Infof("Initializing health check and uptime")
	// Set start time to now
	startTime = time.Now().UTC()

	MainCheck = NewCheck()
}

// HCStatus is the health check status
type HCStatus int

const (
	// HCOk health check ok status
	HCOk HCStatus = iota
	// HCError health check error status
	HCError
)

func (h HCStatus) String() string {
	switch h {
	case HCOk:
		return "Ok"
	case HCError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Checker is the interface so it can be check
type Checker interface {
	Check() (string, error)
}

// Check represents a healtcheck object
type Check struct {
	// Checkers are a grouped
	checks map[string]map[string]Checker
}

// NewCheck returns a new check
func NewCheck() *Check {
	return &Check{
		checks: make(map[string]map[string]Checker),
	}
}

// Status will return the status of the checker
func (c *Check) Status() *CheckStatus {
	log.Logger.Debugf("Checking all healthcheck checks")
	oks := map[string]map[string]string{}
	errors := map[string]map[string]string{}

	res := &CheckStatus{
		Uptime:       Uptime(),
		CheckTs:      time.Now().UTC(),
		Status:       HCOk,
		OkResults:    oks,
		ErrorResults: errors,
	}

	// Run all checks
	for cg, cs := range c.checks {
		oks[cg] = map[string]string{}
		errors[cg] = map[string]string{}
		// For each group
		for cn, cc := range cs {
			r, err := cc.Check()

			if err != nil { // If check errored then add to checkers error group
				res.Status = HCError
				errors[cg][cn] = err.Error()
				continue
			}
			// If no message it should put Ok
			if r == "" {
				r = "Ok"
			}
			oks[cg][cn] = r
		}
	}

	return res
}

// Register registers a checker in the health check
func (c *Check) Register(name, group string, checker Checker) {
	log.Logger.Infof("Registered checker %s in group %s", name, group)
	if _, ok := c.checks[group]; !ok {
		c.checks[group] = map[string]Checker{}
	}
	c.checks[group][name] = checker
}

// CheckStatus is the status returned by the health checker
type CheckStatus struct {
	Status  HCStatus
	Uptime  time.Duration
	CheckTs time.Time

	OkResults    map[string]map[string]string
	ErrorResults map[string]map[string]string
}

// MarshalJSON implements json marshaller interface
func (c *CheckStatus) MarshalJSON() ([]byte, error) {

	return json.Marshal(&struct {
		Status  string                       `json:"status"`
		Uptime  string                       `json:"uptime"`
		CheckTs string                       `json:"check_ts"`
		Version string                       `json:"version"`
		Oks     map[string]map[string]string `json:"healthy"`
		Errors  map[string]map[string]string `json:"unhealthy"`
	}{
		Status:  fmt.Sprintf("%s", c.Status),
		Uptime:  fmt.Sprintf("%v", c.Uptime),
		CheckTs: fmt.Sprintf("%v", c.CheckTs),
		Version: version.Get().String(),
		Oks:     c.OkResults,
		Errors:  c.ErrorResults,
	})
}
