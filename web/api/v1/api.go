package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/themotion/ladder/autoscaler"
	"github.com/themotion/ladder/log"
)

// Routes
const (
	cancelStopAutoscalerRT = "/autoscalers/:autoscaler/cancel-stop"
	stopAutoscalerRT       = "/autoscalers/:autoscaler/stop/:duration"
	autoscalersRT          = "/autoscalers"
)

// apiError is a wrapper around an api error
type apiResult struct {
	code int
	data interface{}
}
type apiError struct {
	apiResult
	err error
}

type apiFunc func(r *http.Request, ps httprouter.Params) (*apiResult, *apiError)

func createAPIHandler(f apiFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Call api function
		result, apiErr := f(r, ps)
		if apiErr != nil {
			data := map[string]interface{}{
				"error": apiErr.err.Error(),
				"data":  apiErr.data,
			}
			b, err := json.Marshal(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, string(b), apiErr.code)
			return
		}

		// Marshall the result
		b, err := json.Marshal(result.data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(result.code)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

func fixPrefix(prefix string) (string, error) {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimRight(prefix, "/")
	}

	// Check is a correct url
	if _, err := url.Parse(prefix); err != nil {
		return "", err
	}

	return prefix, nil
}

// APIV1 implements the v1 version of Ladder API
type APIV1 struct {
	prefix      string
	autoscalers map[string]autoscaler.Autoscaler
}

// NewAPIV1 returns a new object of APIV1
func NewAPIV1(prefix string, autoscalers map[string]autoscaler.Autoscaler) (*APIV1, error) {
	prefix, err := fixPrefix(prefix)
	if err != nil {
		return nil, err
	}
	return &APIV1{
		prefix:      prefix,
		autoscalers: autoscalers,
	}, nil
}

// Register will register all the handlers on all the routes
func (a *APIV1) Register(r *httprouter.Router) {
	r.GET(a.prefix+autoscalersRT, createAPIHandler(a.autoscalersList))
	r.PUT(a.prefix+cancelStopAutoscalerRT, createAPIHandler(a.cancelStopAutoscaler))
	r.PUT(a.prefix+stopAutoscalerRT, createAPIHandler(a.stopAutoscaler))
	log.Logger.Infof("Registered API v1 endopoints with prefix: %s", a.prefix)
}

// cancelStopAutoscaler will cancel the stopped autoscaler status
func (a *APIV1) cancelStopAutoscaler(r *http.Request, ps httprouter.Params) (*apiResult, *apiError) {
	aName := ps.ByName("autoscaler")
	log.Logger.Debugf("Called cancel autoscaler stop API v1 endpoint on autoscaler: %s", aName)

	as, ok := a.autoscalers[aName]
	if !ok {
		err := fmt.Errorf("%s is not a valid autoscaler", aName)
		return nil, &apiError{
			apiResult: apiResult{
				code: http.StatusBadRequest,
				data: map[string]string{
					"msg":        err.Error(),
					"autoscaler": aName,
				},
			},
			err: err,
		}
	}

	// Check if running
	if as.Running() {
		err := fmt.Errorf("Autoscaler is not stopped")
		return nil, &apiError{
			apiResult: apiResult{
				code: http.StatusBadRequest,
				data: map[string]string{
					"msg":        err.Error(),
					"autoscaler": aName,
				},
			},
			err: err,
		}
	}

	// Cancel
	go as.CancelStop()
	return &apiResult{
		code: http.StatusAccepted,
		data: map[string]string{
			"msg":        "Autoscaler stop cancel request sent",
			"autoscaler": aName,
		},
	}, nil
}

// stopAutoscaler will stop an autoscaler
func (a *APIV1) stopAutoscaler(r *http.Request, ps httprouter.Params) (*apiResult, *apiError) {
	aName := ps.ByName("autoscaler")
	log.Logger.Debugf("Called stop atusocaler API v1 endpoint on autoscaler: %s", aName)

	as, ok := a.autoscalers[aName]
	if !ok {
		err := fmt.Errorf("%s is not a valid autoscaler", aName)
		return nil, &apiError{
			apiResult: apiResult{
				code: http.StatusBadRequest,
				data: map[string]string{
					"msg":        err.Error(),
					"autoscaler": aName,
				},
			},
			err: err,
		}
	}

	// Get duration
	duration, err := time.ParseDuration(ps.ByName("duration"))
	if err != nil {
		err := fmt.Errorf("duration parameter is not valid")
		return nil, &apiError{
			apiResult: apiResult{
				code: http.StatusBadRequest,
				data: map[string]string{
					"msg":        err.Error(),
					"autoscaler": aName,
				},
			},
			err: err,
		}
	}

	// Check if not running
	if !as.Running() {
		st, err := as.Status()
		if err != nil {
			err = fmt.Errorf("Error getting the autoscaler status")
			return nil, &apiError{
				apiResult: apiResult{
					code: http.StatusInternalServerError,
					data: map[string]string{
						"msg":        err.Error(),
						"autoscaler": aName,
					},
				},
				err: err,
			}
		}
		err = fmt.Errorf("Autoscaler already stopped")
		return nil, &apiError{
			apiResult: apiResult{
				code: http.StatusConflict,
				data: map[string]interface{}{
					"msg":             err.Error(),
					"required-action": "Need to cancel current stop state first",
					"autoscaler":      aName,
					"deadline":        st.StopDeadline.Unix(),
				},
			},
			err: err,
		}
	}

	// Stop
	go as.Stop(duration)
	return &apiResult{
		code: http.StatusAccepted,
		data: map[string]string{
			"msg":        "Autoscaler stop request sent",
			"autoscaler": aName,
		},
	}, nil
}

// stopAutoscaler will stop an autoscaler
func (a *APIV1) autoscalersList(r *http.Request, ps httprouter.Params) (*apiResult, *apiError) {
	autoscalers := map[string]interface{}{}
	for aName, as := range a.autoscalers {
		st, err := as.Status()
		if err != nil {
			err = fmt.Errorf("Error getting the autoscaler status")
			return nil, &apiError{
				apiResult: apiResult{
					code: http.StatusInternalServerError,
					data: map[string]string{
						"msg":        err.Error(),
						"autoscaler": aName,
					},
				},
				err: err,
			}
		}
		autoscalers[aName] = map[string]string{
			"status": st.State.String(),
		}
	}

	return &apiResult{
		code: http.StatusOK,
		data: map[string]interface{}{"autoscalers": autoscalers},
	}, nil
}
