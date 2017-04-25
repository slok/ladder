package config

import (
	"errors"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// AutoscalersCfg Temporal config to load the yaml configuration
type AutoscalersCfg struct {
	Cfg []Autoscaler `yaml:"autoscalers"`

	Defaults *Autoscaler
}

// Load will load autoscalers configuration object and return
func (a *AutoscalersCfg) Load(content []byte) ([]*Autoscaler, error) {
	// Load the autoscaler
	a.Cfg = []Autoscaler{}
	if err := yaml.Unmarshal(content, a); err != nil {
		return nil, err
	}

	// Set default values
	result := make([]*Autoscaler, len(a.Cfg))
	for i, as := range a.Cfg {
		if a.Defaults != nil {
			if as.Interval == 0 {
				as.Interval = a.Defaults.Interval
			}

			if as.Warmup == 0 {
				as.Warmup = a.Defaults.Warmup
			}

			if as.ScalingWaitTimeout == 0 {
				as.ScalingWaitTimeout = a.Defaults.ScalingWaitTimeout
			}
		}
		asCopy := as
		result[i] = &asCopy
	}

	// Check autoscalers config requirements
	if err := a.check(); err != nil {
		return nil, err
	}

	return result, nil
}

// Inputter represents a gatherer + arranger
type Inputter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Gather      Block  `yaml:"gather"`
	Arrange     Block  `yaml:"arrange"`
}

// Block will represent a block configuration, could be a solver, a gatherer, an arranger or an scaler
type Block struct {
	Kind   string                 `yaml:"kind"`
	Config map[string]interface{} `yaml:"config"`
}

// Autoscaler is an interface for the different configurations
type Autoscaler struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description,omitempty"`
	Scale       Block      `yaml:"scale"`
	Solve       Block      `yaml:"solve,omitempty"` // Should be error if not solve and more than one Inputter
	Filters     []Block    `yaml:"filters,omitempty"`
	Inputters   []Inputter `yaml:"inputters"`
	Disabled    bool       `yaml:"disabled"`

	// Note: if adding new global defaults settings here you will need to set
	// them also on (a *AutoscalerConfig) UnmarshalYAML
	Interval time.Duration `yaml:"interval,omitempty"`

	// Warmup is the time that will be waited before accepting an input for the scale process
	Warmup time.Duration `yaml:"warmup,omitempty"`

	// ScalingWaitTimeout is the time that will wait to wait after scaling before timing out
	ScalingWaitTimeout time.Duration `yaml:"scaling_wait_timeout,omitempty"`

	// The defaults for the autoscaler
	defaults *Autoscaler
}

// UnmarshalYAML implements yml lib interface in order to load defaults from global config
func (a *Autoscaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Autoscaler
	ac := &Autoscaler{}
	err := unmarshal((*plain)(ac))
	if err != nil {
		return err
	}

	// Cant set correctly here the defaults without setting global vars (we don't like that...)
	*a = *ac
	return nil
}

func (a *AutoscalersCfg) check() error {
	for _, a := range a.Cfg {
		// Check solvers are ok
		if a.Solve.Kind == "" && len(a.Inputters) > 1 {
			return errors.New("When using multiple inputters you need a solver")
		}

		// Check scale is present
		if a.Scale.Kind == "" {
			return errors.New("Scaler missing")
		}

		// Check name is present
		if a.Name == "" {
			return errors.New("Autoscaler should have a name")
		}
	}
	return nil
}
