package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/health"
	"github.com/themotion/ladder/log"
)

// Handlers
func prometheusHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	prometheus.Handler().ServeHTTP(w, r)
}

func configHandler(cfg *config.Config) httprouter.Handle {
	res := ""
	for k, v := range cfg.Originals {
		res = fmt.Sprintf("%s[%s]\n%s\n\n", res, k, v)
	}

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprint(w, res)
	}
}

func healthCheckHandler(hc *health.Check) httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		hc := hc.Status()

		b, err := json.Marshal(hc)

		if err != nil {
			log.Logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the json
		if hc.Status != health.HCOk {
			http.Error(w, string(b), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}
