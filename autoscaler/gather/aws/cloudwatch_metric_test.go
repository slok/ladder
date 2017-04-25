package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"

	awsMock "github.com/themotion/ladder/mock/aws"
	"github.com/themotion/ladder/mock/aws/sdk"
)

func TestCWMetricCorrectCreation(t *testing.T) {
	tests := []struct {
		region     string
		metricName string
		namespace  string
		statistic  string
		unit       string
		offset     string
		dimensions []interface{}

		wantOffset time.Duration
	}{
		{
			"us-west-2", "CPUReservation", "AWS/ECS", "Maximum", "Percent", "0s",
			[]interface{}{
				map[interface{}]interface{}{cwDimensionsNameOpt: "ClusterName", cwDimensionsValueOpt: "slok-ECSCluster1-15OBYPKBNXIO6"},
			},
			0 * time.Second,
		},
		{
			"eu-west-1", "DiskReadBytes", "AWS/EC2", "Minimum", "Bytes", "-1m",
			[]interface{}{
				map[interface{}]interface{}{cwDimensionsNameOpt: "InstanceId", cwDimensionsValueOpt: "i-003efbc77add1544a"},
				map[interface{}]interface{}{cwDimensionsNameOpt: "InstanceType", cwDimensionsValueOpt: "t2.micro"},
			},
			-1 * time.Minute,
		},
		{
			"eu-west-1", "DiskReadBytes", "AWS/EC2", "Minimum", "Bytes", "-30s",
			[]interface{}{},
			-30 * time.Second,
		},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			cwAwsRegionOpt:  test.region,
			cwMetricNameOpt: test.metricName,
			cwNamespaceOpt:  test.namespace,
			cwStatisticOpt:  test.statistic,
			cwUnitOpt:       test.unit,
			cwDimensionsOpt: test.dimensions,
			cwTimeOffsetOpt: test.offset,
		}

		c, err := NewCWMetric(context.TODO(), opts)
		if err != nil {
			t.Fatalf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if aws.StringValue(c.session.Config.Region) != test.region {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.region, aws.StringValue(c.session.Config.Region))
		}

		if c.metricName != test.metricName {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.metricName, c.metricName)
		}

		if c.namespace != test.namespace {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.namespace, c.namespace)
		}

		if c.statistic != test.statistic {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.statistic, c.statistic)
		}

		if c.unit != test.unit {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.unit, c.unit)
		}

		if c.offset != test.wantOffset {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.wantOffset, c.offset)
		}

		if len(c.dimensions) != len(test.dimensions) {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object length, want: %v; got %v", test, len(test.dimensions), len(c.dimensions))
		}

		for i, d := range test.dimensions {
			n := d.(map[interface{}]interface{})[cwDimensionsNameOpt]
			v := d.(map[interface{}]interface{})[cwDimensionsValueOpt]

			if c.dimensions[i].name != n {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, n, c.dimensions[i].name)
			}
			if c.dimensions[i].value != v {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, v, c.dimensions[i].value)
			}
		}
	}
}

func TestCWMetricWrongParameterCreation(t *testing.T) {
	tests := []struct {
		region     string
		metricName string
		namespace  string
		statistic  string
		unit       string
		offset     string
		dimensions []interface{}
	}{
		// Missing params
		{metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", namespace: "AWS/ECS", statistic: "Maximum", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", statistic: "Maximum", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "Percent", dimensions: []interface{}{}},

		// wrong params
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximun", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Max", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "wrong", unit: "Percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "percent", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "Percents", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "wrong", offset: "-30s", dimensions: []interface{}{}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "Percent", offset: "-30g", dimensions: []interface{}{}},

		// wrong types
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "wrong", offset: "-30s", dimensions: []interface{}{1, 2, 3, 4}},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "wrong", offset: "-30s", dimensions: nil},
		{region: "us-west-2", metricName: "CPUReservation", namespace: "AWS/ECS", statistic: "Maximum", unit: "Percent", offset: "30s", dimensions: []interface{}{}},
		{
			region:     "us-west-2",
			metricName: "CPUReservation",
			namespace:  "AWS/ECS",
			statistic:  "Maximum",
			unit:       "wrong",
			dimensions: []interface{}{
				map[interface{}]interface{}{
					1: 2,
					2: 3,
				},
			},
		},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			cwAwsRegionOpt:  test.region,
			cwMetricNameOpt: test.metricName,
			cwNamespaceOpt:  test.namespace,
			cwStatisticOpt:  test.statistic,
			cwUnitOpt:       test.unit,
			cwDimensionsOpt: test.dimensions,
			cwTimeOffsetOpt: test.offset,
		}

		_, err := NewCWMetric(context.TODO(), opts)
		if err == nil {
			t.Fatalf("\n- %+v\n  Creation should give error, it didn't", test)
		}

	}
}

func TestCWMetricGather(t *testing.T) {
	tests := []struct {
		metric     float64
		wantMetric int64
	}{
		{1000, 1000},
		{1000000, 1000000},
		{9.45, 9},
		{48.9, 48},
	}

	for _, test := range tests {

		opts := map[string]interface{}{
			cwAwsRegionOpt:  "us-west-2",
			cwMetricNameOpt: "CPUReservation",
			cwNamespaceOpt:  "AWS/ECS",
			cwStatisticOpt:  "Maximum",
			cwUnitOpt:       "Bytes",
			cwDimensionsOpt: []interface{}{},
			cwTimeOffsetOpt: "0s",
		}

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockCWMetric := sdk.NewMockCloudWatchAPI(ctrl)
		c, err := NewCWMetric(context.TODO(), opts)
		c.client = mockCWMetric

		// Set our mock desired result
		awsMock.MockGetMetricStatistics(t, mockCWMetric, test.metric, c.statistic, 1)

		if err != nil {
			t.Fatalf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		res, err := c.Gather(context.TODO())
		if err != nil {
			t.Errorf("\n- %+v\n  Gathering shouldn't give error: %v", test, err)
		}

		if res.Q != test.wantMetric {
			t.Errorf("\n- %+v\n  Gathered quantity doesn't look good, got: %d want: %d", test, res.Q, test.wantMetric)

		}
	}
}

func TestCWMetricGatherWrong(t *testing.T) {
	times := 10

	for i := 0; i < times; i++ {

		opts := map[string]interface{}{
			cwAwsRegionOpt:  "us-west-2",
			cwMetricNameOpt: "CPUReservation",
			cwNamespaceOpt:  "AWS/ECS",
			cwStatisticOpt:  "Maximum",
			cwUnitOpt:       "Bytes",
			cwDimensionsOpt: []interface{}{},
			cwTimeOffsetOpt: "0s",
		}

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockCWMetric := sdk.NewMockCloudWatchAPI(ctrl)
		c, err := NewCWMetric(context.TODO(), opts)
		c.client = mockCWMetric

		// Set our mock desired result
		awsMock.MockGetMetricStatisticsError(t, mockCWMetric)

		if err != nil {
			t.Fatalf("\n-  Creation shouldn't give error: %v", err)
		}

		_, err = c.Gather(context.TODO())
		if err == nil {
			t.Errorf("\n-  Gathering should give error, it didn't")
		}
	}
}

func TestCWMetricGatherMultipleMessages(t *testing.T) {
	times := 10

	for i := 2; i < times+2; i++ {

		opts := map[string]interface{}{
			cwAwsRegionOpt:  "us-west-2",
			cwMetricNameOpt: "CPUReservation",
			cwNamespaceOpt:  "AWS/ECS",
			cwStatisticOpt:  "Maximum",
			cwUnitOpt:       "Bytes",
			cwDimensionsOpt: []interface{}{},
			cwTimeOffsetOpt: "0s",
		}

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockCWMetric := sdk.NewMockCloudWatchAPI(ctrl)
		c, err := NewCWMetric(context.TODO(), opts)
		c.client = mockCWMetric

		// Set our mock desired result
		awsMock.MockGetMetricStatistics(t, mockCWMetric, 0, "Sum", i)

		if err != nil {
			t.Fatalf("\n-  Creation shouldn't give error: %v", err)
		}

		_, err = c.Gather(context.TODO())
		if err == nil {
			t.Errorf("\n-  Gathering should give error, it didn't")
		}
	}
}
