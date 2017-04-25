package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/themotion/ladder/log"
)

var (
	// When loading global settings will replace this defaults, because global
	// are the new defaults for every service

	// DefaultMetricsPath is the default metrics path
	DefaultMetricsPath = "/metrics"
	// DefaultConfigPath is the default config path
	DefaultConfigPath = "/config"
	// DefaultHealthCheckPath is the default health check path
	DefaultHealthCheckPath = "/check"
	// DefaultAPIV1Path is the default api version 1 path
	DefaultAPIV1Path = "/api/v1"
	// DefaultInterval is the default autoscaler interval
	DefaultInterval = 30 * time.Second
	// DefaultWarmup is the default autoscaler warmap
	DefaultWarmup = 30 * time.Second
	// DefaultScalingWaitTimeout is the default autoscaler scaling wait timeout
	DefaultScalingWaitTimeout = 2 * time.Minute
)

// LoadConfig parses the given YAML file into a Config.
func LoadConfig(filename string) (*Config, error) {
	log.Logger.Infof("Loading main config file: %s", filename)

	// If the entire config body is empty the UnmarshalYAML method is
	// never called. We thus have to set the DefaultConfig at the entry
	// point as well.
	cfg := &Config{
		Global:      Global{},
		Autoscalers: []Autoscaler{},
		Originals:   map[string]string{},
	}

	// Read the global configuration
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Create default values for global configuration
	gDefault := Global{
		MetricsPath:        DefaultMetricsPath,
		ConfigPath:         DefaultConfigPath,
		HealthCheckPath:    DefaultHealthCheckPath,
		APIV1Path:          DefaultAPIV1Path,
		Interval:           DefaultInterval,
		Warmup:             DefaultWarmup,
		ScalingWaitTimeout: DefaultScalingWaitTimeout,
	}

	// Load global configuration
	mainCfgTemp := &MainCfg{GlobalDefaults: &gDefault}
	g, afs, err := mainCfgTemp.Load(content)
	if err != nil {
		return nil, err
	}
	cfg.Global = *g
	cfg.Originals[filename] = string(content)

	// We have the global configuration, we need to load all the Autoscalers
	// Get autoscalers cfg files
	aSFiles := []string{}
	for _, f := range *afs {
		ms, err2 := filepath.Glob(f)
		if err2 != nil {
			return nil, err
		}
		aSFiles = append(aSFiles, ms...)
	}

	// Set the default configuration from the global config (confign inheritance)
	aDefault := Autoscaler{
		Interval:           cfg.Global.Interval,
		Warmup:             cfg.Global.Warmup,
		ScalingWaitTimeout: cfg.Global.ScalingWaitTimeout,
	}
	asConfigTemp := &AutoscalersCfg{Defaults: &aDefault}

	// Load autoscalers
	as := []Autoscaler{}
	for _, f := range aSFiles {
		log.Logger.Infof("Loading Autoscaler config file: %s", f)

		// Read the autoscaler configuration
		content, err = ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		if len(content) > 0 {
			asC, err2 := asConfigTemp.Load(content)
			if err2 != nil {
				return nil, err2
			}

			for _, a := range asC {
				as = append(as, *a)
			}
			cfg.Originals[f] = string(content)
		}

	}
	cfg.Autoscalers = as
	log.Logger.Infof("Loaded %d autoscalers", len(cfg.Autoscalers))

	// Check All the config as a whole
	if err = cfg.Check(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Config is the root level configuration
type Config struct {
	Global      Global
	Autoscalers []Autoscaler

	// DryRun will set if needs to be run the logic or no
	DryRun bool

	// Debug on/off
	Debug bool

	// Originals are the original config files
	Originals map[string]string
}

// Check will check that the configuraiton is ok
func (c *Config) Check() error {
	if len(c.Autoscalers) == 0 {
		return fmt.Errorf("No autoscalers loaded")
	}

	// Check different autoscaler names
	names := map[string]struct{}{}
	for _, a := range c.Autoscalers {
		if _, ok := names[a.Name]; ok {
			return fmt.Errorf("Autoscaler with name '%s' declared multiple times", a.Name)
		}
		names[a.Name] = struct{}{}
	}

	return nil
}
