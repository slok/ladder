package config

import (
	"reflect"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func TestAutoscalersConfigLoad(t *testing.T) {
	content := []byte(`
autoscalers:
- name: test
  interval: 50s
  warmup: 50m
  scaling_wait_timeout: 10m
  description: "test"
  scale:
    kind: test
    config:
      something: "test"

  inputters:
  - name: test
    description: "test"
    gather:
      kind: test
      config:
        something: "test"
    arrange:
      kind: test
      config:
        something: "test"

- name: test2
  interval: 120s
  disabled: true
  warmup: 30m
  scaling_wait_timeout: 4m
  description: "test2"
  scale:
    kind: test2
    config:
      something: "test2"
  solve:
    kind: test2
    config:
      something: "test2"

  inputters:
  - name: test2
    description: "test2"
    gather:
      kind: test2
      config:
        something: "test2"
    arrange:
      kind: test2
      config:
        something: "test2"
  - name: test2a
    description: "test2a"
    gather:
      kind: test2a
      config:
        something: "test2a"
    arrange:
      kind: test2a
      config:
        something: "test2a"

`)
	defaultConfig := Autoscaler{Interval: 5 * time.Minute}
	expectedConfig := []Autoscaler{
		Autoscaler{
			Name:               "test",
			Description:        "test",
			Disabled:           false,
			Warmup:             50 * time.Minute,
			ScalingWaitTimeout: 10 * time.Minute,
			Interval:           50 * time.Second,
			Scale:              Block{Kind: "test", Config: map[string]interface{}{"something": "test"}},
			Inputters: []Inputter{
				Inputter{
					Name: "test", Description: "test",
					Gather:  Block{Kind: "test", Config: map[string]interface{}{"something": "test"}},
					Arrange: Block{Kind: "test", Config: map[string]interface{}{"something": "test"}},
				},
			},
		},
		Autoscaler{
			Name:               "test2",
			Description:        "test2",
			Disabled:           true,
			Warmup:             30 * time.Minute,
			ScalingWaitTimeout: 4 * time.Minute,
			Interval:           120 * time.Second,
			Scale:              Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
			Solve:              Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
			Inputters: []Inputter{
				Inputter{
					Name: "test2", Description: "test2",
					Gather:  Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
					Arrange: Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
				},
				Inputter{
					Name: "test2a", Description: "test2a",
					Gather:  Block{Kind: "test2a", Config: map[string]interface{}{"something": "test2a"}},
					Arrange: Block{Kind: "test2a", Config: map[string]interface{}{"something": "test2a"}},
				},
			},
		},
	}

	aCfg := &AutoscalersCfg{Defaults: &defaultConfig}
	a, err := aCfg.Load(content)
	if err != nil {
		t.Errorf("Config load shouldn't give an error, it did: %v", err)
	}

	bgot, err := yaml.Marshal(a)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err := yaml.Marshal(expectedConfig)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected config result: \n\n%s\n expected\n\n%s", content, bgot, bwant)
	}
}

func TestAutoscalersConfigLoadWithDefaults(t *testing.T) {
	content := []byte(`
autoscalers:
- name: test
  warmup: 50m
  scaling_wait_timeout: 10m
  description: "test"
  scale:
    kind: test
    config:
      something: "test"

  inputters:
  - name: test
    description: "test"
    gather:
      kind: test
      config:
        something: "test"
    arrange:
      kind: test
      config:
        something: "test"

- name: test2
  disabled: true
  description: "test2"
  scale:
    kind: test2
    config:
      something: "test2"
  solve:
    kind: test2
    config:
      something: "test2"
  inputters:
  - name: test2
    description: "test2"
    gather:
      kind: test2
      config:
        something: "test2"
    arrange:
      kind: test2
      config:
        something: "test2"
  - name: test2a
    description: "test2a"
    gather:
      kind: test2a
      config:
        something: "test2a"
    arrange:
      kind: test2a
      config:
        something: "test2a"

`)
	defaultConfig := Autoscaler{
		Interval:           99 * time.Minute,
		Warmup:             88 * time.Second,
		ScalingWaitTimeout: 77 * time.Minute,
	}
	expectedConfig := []Autoscaler{
		Autoscaler{
			Name:               "test",
			Description:        "test",
			Disabled:           false,
			Warmup:             50 * time.Minute,
			ScalingWaitTimeout: 10 * time.Minute,
			Interval:           99 * time.Minute,
			Scale:              Block{Kind: "test", Config: map[string]interface{}{"something": "test"}},
			Inputters: []Inputter{
				Inputter{
					Name: "test", Description: "test",
					Gather:  Block{Kind: "test", Config: map[string]interface{}{"something": "test"}},
					Arrange: Block{Kind: "test", Config: map[string]interface{}{"something": "test"}},
				},
			},
		},
		Autoscaler{
			Name:               "test2",
			Description:        "test2",
			Disabled:           true,
			Interval:           99 * time.Minute,
			Warmup:             88 * time.Second,
			ScalingWaitTimeout: 77 * time.Minute,
			Solve:              Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
			Scale:              Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
			Inputters: []Inputter{
				Inputter{
					Name: "test2", Description: "test2",
					Gather:  Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
					Arrange: Block{Kind: "test2", Config: map[string]interface{}{"something": "test2"}},
				},
				Inputter{
					Name: "test2a", Description: "test2a",
					Gather:  Block{Kind: "test2a", Config: map[string]interface{}{"something": "test2a"}},
					Arrange: Block{Kind: "test2a", Config: map[string]interface{}{"something": "test2a"}},
				},
			},
		},
	}

	aCfg := &AutoscalersCfg{Defaults: &defaultConfig}
	a, err := aCfg.Load(content)
	if err != nil {
		t.Errorf("Config load shouldn't give an error, it did: %v", err)
	}

	bgot, err := yaml.Marshal(a)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err := yaml.Marshal(expectedConfig)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected config result: \n\n%s\n expected\n\n%s", content, bgot, bwant)
	}
}

func TestAutoscalersConfigLoadWithErrors(t *testing.T) {

	tests := []struct {
		content []byte
	}{
		// >1 inputter, no solvers
		{
			content: []byte(`
autoscalers:
- name: test
  warmup: 50m
  scaling_wait_timeout: 10m
  description: "test"
  scale:
    kind: test
    config:
      something: "test"

  inputters:
  - name: test
    description: "test"
    gather:
      kind: test
      config:
        something: "test"
    arrange:
      kind: test
      config:
        something: "test"

- name: test2
  disabled: true
  description: "test2"
  scale:
    kind: test2
    config:
      something: "test2"

  inputters:
  - name: test2
    description: "test2"
    gather:
      kind: test2
      config:
        something: "test2"
    arrange:
      kind: test2
      config:
        something: "test2"
  - name: test2a
    description: "test2a"
    gather:
      kind: test2a
      config:
        something: "test2a"
    arrange:
      kind: test2a
      config:
        something: "test2a"`),
		},

		// No scaler
		{
			content: []byte(`
autoscalers:
- name: test
  warmup: 50m
  scaling_wait_timeout: 10m
  description: "test"

  inputters:
  - name: test
    description: "test"
    gather:
      kind: test
      config:
        something: "test"
    arrange:
      kind: test
      config:
        something: "test"
        `),
		},
		// No name on autoscaler
		{
			content: []byte(`
autoscalers:
- warmup: 50m
  scaling_wait_timeout: 10m
  description: "test"

  inputters:
  - name: test
	description: "test"
	gather:
	  kind: test
	  config:
		something: "test"
	arrange:
	  kind: test
	  config:
		something: "test"
		`),
		},
	}

	for _, test := range tests {

		aCfg := &AutoscalersCfg{}
		if _, err := aCfg.Load(test.content); err == nil {
			t.Errorf("- %v\n  Config load should give an error, it didn't", string(test.content))
		}
	}

}
