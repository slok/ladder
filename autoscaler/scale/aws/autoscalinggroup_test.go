package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"

	awsMock "github.com/themotion/ladder/mock/aws"
	"github.com/themotion/ladder/mock/aws/sdk"
	"github.com/themotion/ladder/types"
)

func TestAutoScalingGroupCorrectCreation(t *testing.T) {
	tests := []struct {
		awsRegion            string
		asgName              string
		waitUp               string
		waitDown             string
		forceMinMax          bool
		remainingClosestHour string
		maxRemainingCH       int
	}{
		{
			awsRegion:            "us-west-2",
			asgName:              "test",
			waitUp:               "3m",
			waitDown:             "1m",
			forceMinMax:          false,
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "us-west-2",
			asgName:              "test",
			waitUp:               "3m",
			waitDown:             "1m",
			forceMinMax:          false,
			remainingClosestHour: "",
			maxRemainingCH:       10,
		},
		{
			awsRegion:            "eu-west-1",
			asgName:              "test2",
			waitUp:               "31s",
			waitDown:             "15s",
			forceMinMax:          true,
			remainingClosestHour: "60m",
			maxRemainingCH:       50,
		},
		{
			awsRegion:            "us-west-1",
			asgName:              "test3",
			waitUp:               "90m",
			waitDown:             "1h",
			forceMinMax:          true,
			remainingClosestHour: "300s",
			maxRemainingCH:       1092,
		},
		{
			awsRegion:            "us-west-1",
			asgName:              "test3",
			waitUp:               "90m",
			waitDown:             "1h",
			forceMinMax:          true,
			remainingClosestHour: "0s",
			maxRemainingCH:       0,
		},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			asgNameOpt:                      test.asgName,
			awsRegionOpt:                    test.awsRegion,
			asgWaitUpDurationOpt:            test.waitUp,
			asgWaitDownDurationOpt:          test.waitDown,
			asgForceMinMaxOpt:               test.forceMinMax,
			asgRemainingClosestHourLimitDur: test.remainingClosestHour,
			asgMaxTimesRemainingNoDownscale: test.maxRemainingCH,
		}

		a, err := NewASG(context.TODO(), ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if aws.StringValue(a.session.Config.Region) != test.awsRegion {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.awsRegion, aws.StringValue(a.session.Config.Region))
		}

		if a.asgName != test.asgName {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.asgName, a.asgName)
		}
		ts, _ := time.ParseDuration(test.waitUp)
		if a.waitUpDuration != ts {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, a.waitUpDuration, ts)
		}
		ts, _ = time.ParseDuration(test.waitDown)
		if a.waitDownDuration != ts {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, a.waitDownDuration, ts)
		}

		if a.forceMinMax != test.forceMinMax {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.forceMinMax, a.forceMinMax)
		}

		ts, _ = time.ParseDuration(test.remainingClosestHour)
		if a.remainingClosestHourLimit != ts {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, a.remainingClosestHourLimit, ts)
		}

		if a.maxTimesRemainingNoDownscale != test.maxRemainingCH {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.maxRemainingCH, a.maxTimesRemainingNoDownscale)
		}
	}
}

func TestAutoScalingGroupWrongParameterCreation(t *testing.T) {
	tests := []struct {
		awsRegion            string
		asgName              string
		waitUp               string
		waitDown             string
		remainingClosestHour string
		maxRemainingCH       int
	}{
		{
			awsRegion:            "",
			asgName:              "test",
			waitUp:               "1s",
			waitDown:             "5s",
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "eu-west-1",
			asgName:              "",
			waitUp:               "1s",
			waitDown:             "5s",
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			asgName:              "test3",
			waitUp:               "1s",
			waitDown:             "5s",
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "eu-west-1",
			waitUp:               "1s",
			waitDown:             "5s",
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "us-west-1",
			asgName:              "test3",
			waitUp:               "3g",
			waitDown:             "5s",
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "us-west-1",
			asgName:              "test3",
			waitUp:               "1s",
			waitDown:             "3g",
			remainingClosestHour: "0m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "us-west-1",
			asgName:              "test",
			waitUp:               "1s",
			waitDown:             "5s",
			remainingClosestHour: "61m",
			maxRemainingCH:       1,
		},
		{
			awsRegion:            "eu-west-1",
			asgName:              "test",
			waitUp:               "1s",
			waitDown:             "5s",
			remainingClosestHour: "10m",
			maxRemainingCH:       0,
		},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			asgNameOpt:                      test.asgName,
			awsRegionOpt:                    test.awsRegion,
			asgWaitUpDurationOpt:            test.waitUp,
			asgWaitDownDurationOpt:          test.waitDown,
			asgRemainingClosestHourLimitDur: test.remainingClosestHour,
			asgMaxTimesRemainingNoDownscale: test.maxRemainingCH,
		}

		_, err := NewASG(context.TODO(), ops)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}
	}
}

func TestAutoScalingGroupCurrent(t *testing.T) {
	tests := []struct {
		running int64
		max     int64
		min     int64
		desired int64
		asgName string
		region  string
	}{
		{10, 10, 9, 5, "group1", "eu-west-1"},
		{8, 11, 1, 8, "group2", "eu-west-1"},
		{6, 11, 1, 4, "group3", "eu-west-1"},
		{5, 11, 1, 8, "group4", "eu-west-1"},
		{3, 11, 1, 1, "group5", "eu-west-1"},
		{1, 11, 1, 10, "group6", "eu-west-1"},
	}

	for _, test := range tests {

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockASG := sdk.NewMockAutoScalingAPI(ctrl)

		// Set our mock desired result
		awsMock.MockDescribeAutoScalingGroupsSingle(t, mockASG, test.running,
			test.max, test.min, test.desired)

		// Create our scaler
		a, err := NewASG(context.TODO(), map[string]interface{}{asgNameOpt: test.asgName, awsRegionOpt: test.region})

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		// Set our mock as client
		a.asClient = mockASG

		q, err := a.Current(context.TODO())

		if err != nil {
			t.Errorf("\n- %+v\n  current calling shouldn't give error: %v", test, err)
		}

		if int64(q.Q) != test.running {
			t.Errorf("\n- %+v\n  current didn't return the correct running instances; got: %d, want: %d", test, q.Q, test.running)
		}

	}
}

func TestAutoScalingGroupCurrentError(t *testing.T) {
	tests := []struct {
		groups  int64
		running int64
		max     int64
		min     int64
		desired int64
		asgName string
		region  string
	}{
		{2, 10, 10, 9, 5, "group1", "eu-west-1"},
	}

	for _, test := range tests {

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockASG := sdk.NewMockAutoScalingAPI(ctrl)

		// Set our mock desired result
		awsMock.MockDescribeAutoScalingGroupsMultiple(t, mockASG, test.groups, test.running,
			test.max, test.min, test.desired)

		// Create our scaler
		a, err := NewASG(context.TODO(), map[string]interface{}{asgNameOpt: test.asgName, awsRegionOpt: test.region})

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		// Set our mock as client
		a.asClient = mockASG

		_, err = a.Current(context.TODO())

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}

	}
}

func TestAutoScalingGroupScale(t *testing.T) {
	tests := []struct {
		desired              int64
		max                  int64
		min                  int64
		APIError             bool
		forceMinMax          bool
		APIEC2Error          bool
		instancesRunningTime []time.Duration
		remaining            time.Duration

		wantMax     int64
		wantMin     int64
		wantMode    types.ScalingMode
		wantQ       int64
		wantScaledQ int64
		wantError   bool
	}{
		// Ok, scaling up
		{
			5, 10, 1, false, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			10, 1, types.ScalingUp, 5, 5, false,
		},

		// Ok, not scaling
		{
			5, 10, 1, false, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			10, 1, types.NotScaling, 5, 0, false,
		},
		// Ok, scaling up
		{
			10, 10, 1, false, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			10, 1, types.ScalingUp, 10, 10, false,
		},
		// Ok, scaling down
		{
			1, 10, 1, false, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			10, 1, types.ScalingDown, 1, 1, false,
		},

		// API error
		{
			5, 10, 1, true, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			10, 1, types.NotScaling, 5, 0, false,
		},

		// desired is greater than max but sets bounds (Scaling up)
		{
			11, 10, 1, false, true, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			11, 11, types.ScalingUp, 11, 11, false,
		},

		// desired is smaller than min but sets bounds (Scaling down)
		{
			1, 10, 2, false, true, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			30 * time.Minute,
			1, 1, types.ScalingDown, 1, 1, false,
		},

		// Ok, scaling down, running limit, set desired to 4 (no downscaling)
		{
			1, 10, 1, false, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			10 * time.Minute,
			10, 1, types.NotScaling, 4, 0, false,
		},

		// Ok, scaling down, running limit, set desired to 2 (downscaling limited)
		{
			1, 10, 1, false, false, false,
			[]time.Duration{51 * time.Minute, 112 * time.Minute, 100 * time.Minute, 40 * time.Minute},
			10 * time.Minute,
			10, 1, types.ScalingDown, 2, 2, false,
		},
		// Ok, scaling down, no running limit activated
		{
			1, 10, 1, false, false, false,
			[]time.Duration{51 * time.Minute, 112 * time.Minute, 100 * time.Minute, 40 * time.Minute},
			0 * time.Minute,
			10, 1, types.ScalingDown, 1, 1, false,
		},
		// Ok, scaling down, running limit, error from limiter, set desired to 4 (no downscaling)
		{
			1, 10, 1, false, false, true,
			[]time.Duration{51 * time.Minute, 112 * time.Minute, 100 * time.Minute, 40 * time.Minute},
			10 * time.Minute,
			10, 1, types.ScalingDown, 1, 1, false,
		},
		// Ok, scaling up, running limit doesn't activate
		{
			10, 10, 1, false, false, false,
			[]time.Duration{40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute},
			10 * time.Minute,
			10, 1, types.ScalingUp, 10, 10, false,
		},
	}

	for _, test := range tests {

		instances := []*ec2.Instance{}

		for _, i := range test.instancesRunningTime {
			inst := &ec2.Instance{
				LaunchTime: aws.Time(time.Now().UTC().Add(-i)),
				State:      &ec2.InstanceState{Name: aws.String(ec2.InstanceStateNameRunning)},
			}

			instances = append(instances, inst)
		}

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockASG := sdk.NewMockAutoScalingAPI(ctrl)
		mockEC2 := sdk.NewMockEC2API(ctrl)

		// Set our mock desired result
		running := int64(len(test.instancesRunningTime))
		awsMock.MockDescribeAutoScalingGroupsSingle(t, mockASG, running, test.max, test.min, running)
		awsMock.MockUpdateAutoScalingGroup(t, mockASG, test.wantQ, test.wantMin, test.wantMax, test.forceMinMax, test.APIError)
		awsMock.MockEC2DescribeInstances(t, mockEC2, instances, test.APIEC2Error)

		// Create our scaler
		a, err := NewASG(
			context.TODO(),
			map[string]interface{}{
				asgNameOpt:                      "test",
				awsRegionOpt:                    "us-west-1",
				asgForceMinMaxOpt:               test.forceMinMax,
				asgRemainingClosestHourLimitDur: test.remaining.String(),
				asgMaxTimesRemainingNoDownscale: 1,
			},
		)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		// Set our mock as client
		a.asClient = mockASG
		a.ec2Client = mockEC2
		scaled, mode, err := a.Scale(context.TODO(), types.Quantity{Q: test.desired})

		if !test.wantError {
			if err != nil {
				t.Errorf("\n- %+v\n  Scalation shouldn't give error: %v", test, err)
			}

			if mode != test.wantMode {
				t.Errorf("\n- %+v\n  Scalation mode result is wrong; got: %s, want: %s", test, mode, test.wantMode)
			}

			if scaled.Q != test.wantScaledQ {
				t.Errorf("\n- %+v\n  Scaled Q result is wrong; got: %d, want: %d", test, scaled.Q, test.wantScaledQ)
			}
		} else {
			if err == nil {
				t.Errorf("\n- %+v\n  Creation should give error, it didn't", test)
			}
		}
	}
}

func TestAutoScalingWaitTime(t *testing.T) {
	tests := []struct {
		waitUp   time.Duration
		waitDown time.Duration
	}{
		{50 * time.Millisecond, 100 * time.Millisecond},
		{25 * time.Millisecond, 10 * time.Millisecond},
		{100 * time.Millisecond, 400 * time.Millisecond},
		{250 * time.Millisecond, 100 * time.Millisecond},
	}

	for _, test := range tests {
		a, err := NewASG(context.TODO(), map[string]interface{}{
			asgNameOpt:             "test",
			awsRegionOpt:           "us-west-1",
			asgWaitDownDurationOpt: test.waitDown.String(),
			asgWaitUpDurationOpt:   test.waitUp.String(),
		})

		// Up mode
		s := time.Now().UTC()
		err = a.Wait(context.TODO(), types.Quantity{}, types.ScalingUp)
		if err != nil {
			t.Errorf("\n- %+v\n  Scaling up wait shouldn't give error: %v", test, err)
		}
		if time.Now().UTC().Sub(s) < test.waitUp {
			t.Errorf("\n- %+v\n  Scaling up wait should wait an amount of time, it didn't correctly", test)
		}

		// Down mode
		s = time.Now().UTC()
		err = a.Wait(context.TODO(), types.Quantity{}, types.ScalingDown)
		if err != nil {
			t.Errorf("\n- %+v\n  Scaling down wait shouldn't give error: %v", test, err)
		}
		if time.Now().UTC().Sub(s) < test.waitDown {
			t.Errorf("\n- %+v\n  Scaling down wait should wait an amount of time, it didn't correctly", test)
		}
	}
}

func TestAutoScalingWaitDesired(t *testing.T) {
	tests := []struct {
		current int64
		desired int64

		wantTimeout bool
	}{
		{10, 5, true},
		{10, 10, false},
	}

	for _, test := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockASG := sdk.NewMockAutoScalingAPI(ctrl)

		// Set our mock desired result
		awsMock.MockDescribeAutoScalingGroupsSingle(t, mockASG, test.current, test.desired, test.desired, test.desired)

		a, err := NewASG(context.TODO(), map[string]interface{}{asgNameOpt: "test", awsRegionOpt: "us-west-1"})
		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}
		a.asClient = mockASG
		a.waiterInterval = 2 * time.Millisecond

		waitRes := make(chan struct{})
		go func() {
			err := a.Wait(context.TODO(), types.Quantity{Q: test.desired}, types.ScalingDown)
			if err != nil {
				t.Fatalf("Wait returned error, it shoudln't: %s", err)
			}
			waitRes <- struct{}{}
		}()

		timeout := false
		select {
		case <-waitRes:
		case <-time.After(10 * time.Millisecond):
			timeout = true
		}

		if test.wantTimeout && !timeout {
			t.Error("Wait should timeout, it didn't")
		}

		if !test.wantTimeout && timeout {
			t.Error("Wait shouldn't timeout, it did")
		}
	}
}

func TestFilterClosestHourQ(t *testing.T) {
	tests := []struct {
		newQ                    int64
		instancesRunningTime    []time.Duration
		instancesNotRunningTime []time.Duration
		remaining               time.Duration
		wantQ                   int64
		wantError               bool
	}{
		// all machines running 40m, need 50m at least -> no one
		{
			newQ: 7,
			instancesRunningTime: []time.Duration{
				40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute,
				40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute,
				40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute,
			},
			remaining: 10 * time.Minute,
			wantQ:     12,
			wantError: false,
		},
		// all machines running 40m, need 30m at least -> all
		{
			newQ: 7,
			instancesRunningTime: []time.Duration{
				40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute,
				40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute,
				40 * time.Minute, 40 * time.Minute, 40 * time.Minute, 40 * time.Minute,
			},
			remaining: 30 * time.Minute,
			wantQ:     7,
			wantError: false,
		},
		// all machines running 40m (in the last hour), need 30m at least -> all
		{
			newQ: 7,
			instancesRunningTime: []time.Duration{
				40 * time.Minute, 100 * time.Minute, 160 * time.Minute, 220 * time.Minute,
				280 * time.Minute, 340 * time.Minute, 400 * time.Minute, 460 * time.Minute,
				520 * time.Minute, 580 * time.Minute, 640 * time.Minute, 700 * time.Minute,
			},
			remaining: 30 * time.Minute,
			wantQ:     7,
			wantError: false,
		},
		// 3 machines running more than 30m, 9 running 10m, need 30m at least -> 3
		{
			newQ: 5,
			instancesRunningTime: []time.Duration{
				35 * time.Minute, 162 * time.Minute, 98 * time.Minute, 70 * time.Minute,
				10 * time.Minute, 10 * time.Minute, 10 * time.Minute, 10 * time.Minute,
				70 * time.Minute, 70 * time.Minute, 70 * time.Minute, 70 * time.Minute,
			},
			remaining: 30 * time.Minute,
			wantQ:     9,
			wantError: false,
		},
		// 1 machine running more than 51m, 11 running 50m, need 51m at least -> 1
		{
			newQ: 5,
			instancesRunningTime: []time.Duration{
				51 * time.Minute, 50 * time.Minute, 50 * time.Minute, 50 * time.Minute,
				50 * time.Minute, 50 * time.Minute, 50 * time.Minute, 50 * time.Minute,
				50 * time.Minute, 50 * time.Minute, 50 * time.Minute, 50 * time.Minute,
			},
			remaining: 9 * time.Minute,
			wantQ:     11,
			wantError: false,
		},
		// 12 machine running more than 1m, need 50m at least but we want more than we have -> 0
		{
			newQ: 20,
			instancesRunningTime: []time.Duration{
				1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute,
				1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute,
				1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute,
			},
			remaining: 10 * time.Minute,
			wantQ:     20,
			wantError: false,
		},
		// 3 machines running 40m, need 30m at least -> 9 (except not running ones)
		{
			newQ: 1,
			instancesRunningTime: []time.Duration{
				10 * time.Minute, 10 * time.Minute, 10 * time.Minute, 40 * time.Minute,
				10 * time.Minute, 10 * time.Minute, 10 * time.Minute, 40 * time.Minute,
				10 * time.Minute, 10 * time.Minute, 10 * time.Minute, 40 * time.Minute,
			},
			instancesNotRunningTime: []time.Duration{
				10 * time.Minute, 10 * time.Minute, 40 * time.Minute, 40 * time.Minute,
			},
			remaining: 30 * time.Minute,
			wantQ:     9,
			wantError: false,
		},
	}

	for _, test := range tests {

		instances := []*ec2.Instance{}

		for _, i := range test.instancesRunningTime {
			inst := &ec2.Instance{
				LaunchTime: aws.Time(time.Now().UTC().Add(-i)),
				State:      &ec2.InstanceState{Name: aws.String(ec2.InstanceStateNameRunning)},
			}

			instances = append(instances, inst)
		}

		for _, i := range test.instancesNotRunningTime {
			inst := &ec2.Instance{
				LaunchTime: aws.Time(time.Now().UTC().Add(-i)),
				State:      &ec2.InstanceState{Name: aws.String(ec2.InstanceStateNameShuttingDown)},
			}

			instances = append(instances, inst)
		}

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockEC2 := sdk.NewMockEC2API(ctrl)
		mockASG := sdk.NewMockAutoScalingAPI(ctrl)
		linst := int64(len(instances))
		awsMock.MockDescribeAutoScalingGroupsSingle(t, mockASG, linst, linst, linst, linst)
		awsMock.MockEC2DescribeInstances(t, mockEC2, instances, test.wantError)

		a, err := NewASG(
			context.TODO(),
			map[string]interface{}{
				asgNameOpt:                      "test",
				awsRegionOpt:                    "us-west-1",
				asgRemainingClosestHourLimitDur: test.remaining.String(),
				asgMaxTimesRemainingNoDownscale: 1,
			},
		)

		a.ec2Client = mockEC2
		a.asClient = mockASG

		gotQ, err := a.filterClosestHourQ(types.Quantity{Q: test.newQ})

		if !test.wantError {
			if err != nil {
				t.Errorf("\n- %+v\n Shouldn't error, it did: %v", test, err)
			}
			if gotQ.Q != test.wantQ {
				t.Errorf("\n- %+v\n filtering didn't return expected result, got: %d, want: %d", test, gotQ.Q, test.wantQ)
			}
		} else {
			if err == nil {
				t.Errorf("\n- %+v\n Should error, it didn't", test)
			}
		}
	}
}

func TestFilterClosestHourQMultipleIterations(t *testing.T) {
	tests := []struct {
		newQIters                 []int64
		instancesRunningTimeIters [][]time.Duration
		remaining                 time.Duration
		maxIterations             int
		wantQIters                []int64
		wantError                 bool
		wantCounterIters          []int
	}{
		// Filter reaches the limit of no downscaling
		{
			newQIters: []int64{2, 2, 2, 2, 2},
			instancesRunningTimeIters: [][]time.Duration{
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
			},
			remaining:        10 * time.Minute,
			maxIterations:    3,
			wantQIters:       []int64{4, 4, 4, 2, 4},
			wantError:        false,
			wantCounterIters: []int{1, 2, 3, 0, 1},
		},
		// Filter doesn't reach the limit of no downscaling
		{
			newQIters: []int64{2, 2, 2, 2, 2},
			instancesRunningTimeIters: [][]time.Duration{
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
			},
			remaining:        10 * time.Minute,
			maxIterations:    3,
			wantQIters:       []int64{2, 2, 2, 2, 2},
			wantError:        false,
			wantCounterIters: []int{0, 0, 0, 0, 0},
		},
		// Filter doesn't reach the limit of no downscaling but the first 2 doesn't downscale
		{
			newQIters: []int64{2, 2, 2, 2, 2},
			instancesRunningTimeIters: [][]time.Duration{
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
				[]time.Duration{1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
				[]time.Duration{53 * time.Minute, 53 * time.Minute, 53 * time.Minute, 53 * time.Minute},
			},
			remaining:        10 * time.Minute,
			maxIterations:    3,
			wantQIters:       []int64{4, 4, 2, 2, 2},
			wantError:        false,
			wantCounterIters: []int{1, 2, 0, 0, 0},
		},
	}

	for _, test := range tests {

		a, _ := NewASG(
			context.TODO(),
			map[string]interface{}{
				asgNameOpt:                      "test",
				awsRegionOpt:                    "us-west-1",
				asgRemainingClosestHourLimitDur: test.remaining.String(),
				asgMaxTimesRemainingNoDownscale: test.maxIterations,
			},
		)

		// For each iteration
		for i, irti := range test.instancesRunningTimeIters {
			instances := []*ec2.Instance{}

			for _, i := range irti {
				inst := &ec2.Instance{
					LaunchTime: aws.Time(time.Now().UTC().Add(-i)),
					State:      &ec2.InstanceState{Name: aws.String(ec2.InstanceStateNameRunning)},
				}

				instances = append(instances, inst)
			}

			// Create mock for AWS API
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockEC2 := sdk.NewMockEC2API(ctrl)
			mockASG := sdk.NewMockAutoScalingAPI(ctrl)
			linst := int64(len(instances))
			awsMock.MockDescribeAutoScalingGroupsSingle(t, mockASG, linst, linst, linst, linst)
			awsMock.MockEC2DescribeInstances(t, mockEC2, instances, test.wantError)

			a.ec2Client = mockEC2
			a.asClient = mockASG

			gotQ, err := a.filterClosestHourQ(types.Quantity{Q: test.newQIters[i]})

			if !test.wantError {
				if err != nil {
					t.Errorf("\n- %+v\n Shouldn't error, it did: %v", test, err)
				}
				if gotQ.Q != test.wantQIters[i] {
					t.Errorf("\n- %+v\n filtering didn't return expected result, got: %d, want: %d", test, gotQ.Q, test.wantQIters[i])
				}
				if a.remainingNoDownscaleC != test.wantCounterIters[i] {
					t.Errorf("\n- %+v\n State of filter counter doesn't match , got: %d, want: %d", test, a.remainingNoDownscaleC, test.wantCounterIters[i])
				}
			} else {
				if err == nil {
					t.Errorf("\n- %+v\n Should error, it didn't", test)
				}
			}
		}
	}
}
