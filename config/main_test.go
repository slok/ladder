package config

import (
	"reflect"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func TestMainConfigDefaultLoad(t *testing.T) {
	content := []byte(`
global:
  metrics_path: /metrics-custom
  config_path: /config-custom
  health_check_path: /check-custom
  api_v1_path: /api/custom/v1
  interval: 30s
  warmup: 3m
  scaling_wait_timeout: 1m

autoscaler_files:
  - testdata/as1/*.yml
  - testdata/as2/as2a/*.yml
`)
	defaultGlobalCfg := Global{MetricsPath: "/metrics"}

	expectedGlobalCfg := Global{
		MetricsPath:        "/metrics-custom",
		ConfigPath:         "/config-custom",
		HealthCheckPath:    "/check-custom",
		APIV1Path:          "/api/custom/v1",
		Interval:           30 * time.Second,
		Warmup:             3 * time.Minute,
		ScalingWaitTimeout: 1 * time.Minute,
	}
	expectedAFsCfg := AutoscalerFiles{
		"testdata/as1/*.yml",
		"testdata/as2/as2a/*.yml",
	}

	mCfg := &MainCfg{GlobalDefaults: &defaultGlobalCfg}
	g, afs, err := mCfg.Load(content)
	if err != nil {
		t.Fatalf("Config load shouldn't give an error, it did: %v", err)
	}

	// Check global config
	bgot, err := yaml.Marshal(g)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err := yaml.Marshal(expectedGlobalCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected global config result: \n\n%s\n expected\n\n%s", content, bgot, bwant)
	}

	// Check autoscaler files config
	bgot, err = yaml.Marshal(afs)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err = yaml.Marshal(expectedAFsCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected autoscaler files config result: \n\n%s\n expected\n\n%s", content, bgot, bwant)
	}
}

func TestMainConfigLoadWithDefaults(t *testing.T) {
	content := []byte(`
autoscaler_files:
  - testdata/as1/*.yml
  - testdata/as2/as2a/*.yml
`)

	expectedGlobalCfg := Global{
		MetricsPath:        "/metrics",
		ConfigPath:         "/config",
		HealthCheckPath:    "/check",
		Interval:           100 * time.Second,
		Warmup:             300 * time.Minute,
		ScalingWaitTimeout: 9 * time.Minute,
	}
	expectedAFsCfg := AutoscalerFiles{
		"testdata/as1/*.yml",
		"testdata/as2/as2a/*.yml",
	}

	mCfg := &MainCfg{GlobalDefaults: &expectedGlobalCfg}
	g, afs, err := mCfg.Load(content)
	if err != nil {
		t.Fatalf("Config load shouldn't give an error, it did: %v", err)
	}

	// Check global config
	bgot, err := yaml.Marshal(g)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err := yaml.Marshal(expectedGlobalCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected global config result: \n\n%s\n expected\n\n%s", content, bgot, bwant)
	}

	// Check autoscaler files config
	bgot, err = yaml.Marshal(afs)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err = yaml.Marshal(expectedAFsCfg)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected autoscaler files config result: \n\n%s\n expected\n\n%s", content, bgot, bwant)
	}

}
func TestGlobalConfigWithErrors(t *testing.T) {

	tests := []struct {
		content []byte
	}{
		// No autoscalers
		{
			content: []byte(`
global:
  metrics_path: /metrics
  interval: 30s
  warmup: 3m
  scaling_wait_timeout: 1m
autoscaler_files:
    `),
		},
	}

	for _, test := range tests {

		mCfg := &MainCfg{}
		if _, _, err := mCfg.Load(test.content); err == nil {
			t.Errorf("- %v\n  Config load should give an error, it didn't", string(test.content))
		}
	}
}
