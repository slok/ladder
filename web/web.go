package web

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/health"
	"github.com/themotion/ladder/log"
	apiv1 "github.com/themotion/ladder/web/api/v1"
)

// Handler will serve HTTP endpoints (including prometheus metrics)
type Handler struct {
	router *httprouter.Router

	cfg   *config.Config
	hc    *health.Check
	apiV1 *apiv1.APIV1
}

// NewHandler returns a new handler with the registered routes
func NewHandler(cfg *config.Config, hc *health.Check, apiV1 *apiv1.APIV1) (*Handler, error) {
	h := &Handler{
		cfg:   cfg,   // configuration
		hc:    hc,    // healthcheck
		apiV1: apiV1, // API v1
	}

	router := httprouter.New()
	if router == nil {
		return nil, fmt.Errorf("Error creation httprouter")
	}
	h.router = router

	if err := h.registerRoutes(); err != nil {
		return nil, err
	}

	log.Logger.Debugf("Handler created")
	return h, nil
}

func (h *Handler) registerRoutes() error {

	// Metrics
	h.router.GET(h.cfg.Global.MetricsPath, prometheusHandler)

	// Configuration
	h.router.GET(h.cfg.Global.ConfigPath, configHandler(h.cfg))

	// Health check
	h.router.GET(h.cfg.Global.HealthCheckPath, healthCheckHandler(h.hc))

	// Register API v1
	h.apiV1.Register(h.router)

	return nil
}

// Serve will serve the http endpoint
func (h *Handler) Serve(addr string) error {
	log.Logger.Infof("Listening server on: %s", addr)
	return http.ListenAndServe(addr, h.router)
}
