package config

import (
	"reflect"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func TestConfigCorrect(t *testing.T) {
	testConfFile := "testdata/good.conf.yml"

	expectedConfig := Config{
		Global: Global{
			MetricsPath:        "/metrics",
			ConfigPath:         "/config",
			HealthCheckPath:    "/check",
			APIV1Path:          "/api/v1",
			Interval:           30 * time.Second,
			Warmup:             3 * time.Minute,
			ScalingWaitTimeout: 1 * time.Minute,
		},
		Autoscalers: []Autoscaler{
			Autoscaler{
				Name:               "random_even_odd_print_autoscaler",
				Description:        "A simple autoscaler that will print when scaling up or down. It scales up when random number is odd and scales down when even between numbers 0 and 9 ",
				Interval:           5 * time.Second,
				Warmup:             1 * time.Minute,
				ScalingWaitTimeout: 5 * time.Minute,
				Disabled:           true,
				Scale: Block{
					Kind: "stdout",
					Config: map[string]interface{}{
						"message_prefix": "[RANDOM_SCALER]",
					},
				},
				Solve: Block{
					Kind:   "max",
					Config: map[string]interface{}{},
				},

				Filters: []Block{
					Block{
						Kind: "scaling_kind_interval",
						Config: map[string]interface{}{
							"scale_up_duration":   "30s",
							"scale_down_duration": "1m",
						},
					},
					Block{
						Kind: "limit",
						Config: map[string]interface{}{
							"max": 25,
							"min": 5,
						},
					},
				},

				Inputters: []Inputter{
					Inputter{
						Name:        "random_even_odd",
						Description: "up when even, down when odd",
						Gather: Block{
							Kind: "random",
							Config: map[string]interface{}{
								"max_limit": 10,
								"min_limit": 0,
							},
						},
						Arrange: Block{
							Kind: "in_list",
							Config: map[string]interface{}{
								"match_downscale":      []int{0, 2, 4, 6, 8},
								"match_upscale":        []int{1, 2, 5, 7, 9},
								"match_up_magnitude":   200,
								"match_down_magnitude": 50,
							},
						},
					},
					Inputter{
						Name:        "dummy",
						Description: "Dummy inputter",
						Gather: Block{
							Kind: "dummy",
							Config: map[string]interface{}{
								"quantity": 10,
							},
						},
						Arrange: Block{
							Kind: "dummy",
							Config: map[string]interface{}{
								"quantity": 15,
							},
						},
					},
				},
			},
			Autoscaler{
				Name:        "render_amis_autoscaler",
				Description: "This auto scaler will scale based on the SQS visible messages having in mind that we want to process all the queue in one hour, knowing that 1 machine can process 10 messgaes per hour, it will scale the scalation group based on this information",
				// These 3 values are defaults
				Interval:           30 * time.Second,
				Warmup:             3 * time.Minute,
				ScalingWaitTimeout: 1 * time.Minute,
				Scale: Block{
					Kind: "aws_autoscaling_group",
					Config: map[string]interface{}{
						"aws_region":              "us-west-2",
						"auto_scaling_group_name": "slok-ECSAutoScalingGroup-1PNI4RX8BD5XU",
					},
				},
				Inputters: []Inputter{
					Inputter{
						Name:        "aws_sqs_constant_factor",
						Description: "Will get a number based on the queue messages and a constant factor division",
						Gather: Block{
							Kind: "aws_sqs",
							Config: map[string]interface{}{
								"queue_url":      "https://sqs.us-west-2.amazonaws.com/016386521566/slok-render-jobs",
								"queue_property": "ApproximateNumberOfMessages",
								"aws_region":     "us-west-2",
							},
						},
						Arrange: Block{
							Kind: "constant_factor",
							Config: map[string]interface{}{
								"factor":     10,
								"round_type": "ceil",
							},
						},
					},
				},
			},
		},
	}
	c, err := LoadConfig(testConfFile)
	if err != nil {
		t.Fatalf("Error loading %s configuration: %v", testConfFile, err)
	}
	// We don't need to test the original files string
	c.Originals = map[string]string{}

	bgot, err := yaml.Marshal(c)
	if err != nil {
		t.Fatalf("%s", err)
	}

	bwant, err := yaml.Marshal(expectedConfig)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !reflect.DeepEqual(bgot, bwant) {
		t.Fatalf("%s: unexpected config result: \n\n%s\n expected\n\n%s", testConfFile, bgot, bwant)
	}

}

func TestConfigWithErrors(t *testing.T) {
	tests := []string{
		"testdata/bad.conf.1.yml", // No autoscalers_files
		"testdata/bad.conf.2.yml", // No autoscalers on autoscalers_file
		"testdata/bad.conf.3.yml", // Multiple inputters, no solver
		"testdata/bad.conf.4.yml", // No scaler
		"testdata/bad.conf.5.yml", // No name on autoscaler
		"testdata/bad.conf.6.yml", // Multiple autoscalers, same names
	}

	for _, test := range tests {
		if _, err := LoadConfig(test); err == nil {
			t.Fatalf("Loading %s configuration didn't give error, it should", test)
		}
	}
}
