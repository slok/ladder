package config

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// MainCfg Temporal config to load the yaml configuration
type MainCfg struct {
	Global          Global          `yaml:"global"`
	AutoscalerFiles AutoscalerFiles `yaml:"autoscaler_files"`

	// GlobalDefaults are the global defaults that will be applied if no data is set
	GlobalDefaults *Global
}

// Load will load a Global configuration object and return
func (m *MainCfg) Load(content []byte) (*Global, *AutoscalerFiles, error) {
	if m.GlobalDefaults == nil {
		m.GlobalDefaults = &Global{}
	}
	// Copy defaults
	d := *m.GlobalDefaults
	m.Global = d
	m.AutoscalerFiles = AutoscalerFiles{}

	if err := yaml.Unmarshal(content, m); err != nil {
		return nil, nil, err
	}

	// Check Global config requirements
	if err := m.check(); err != nil {
		return nil, nil, err
	}

	return &m.Global, &m.AutoscalerFiles, nil
}

// Global is the main global configuration of the service
type Global struct {
	MetricsPath        string        `yaml:"metrics_path,omitempty"`
	ConfigPath         string        `yaml:"config_path,omitempty"`
	HealthCheckPath    string        `yaml:"health_check_path,omitempty"`
	APIV1Path          string        `yaml:"api_v1_path,omitempty"`
	Interval           time.Duration `yaml:"interval,omitempty"`
	Warmup             time.Duration `yaml:"warmup,omitempty"`
	ScalingWaitTimeout time.Duration `yaml:"scaling_wait_timeout,omitempty"`
}

// AutoscalerFiles are the file path matchs of the autoscalers
type AutoscalerFiles []string

// Check will check for requirements met after loading the config file
func (m *MainCfg) check() error {
	if len(m.AutoscalerFiles) == 0 {
		return fmt.Errorf("Autoscaler files can't be empty")
	}
	return nil
}
