package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"

	awsMock "github.com/themotion/ladder/mock/aws"
	"github.com/themotion/ladder/mock/aws/sdk"
	"github.com/themotion/ladder/types"
)

func TestECSCorrectCreation(t *testing.T) {
	tests := []struct {
		awsRegion string
		cluster   string
		service   string
	}{
		{"us-west-2", "testCluster1", "service1"},
		{"us-west-2", "testCluster2", "service2"},
		{"us-west-2", "testCluster3", "service3"},
		{"us-west-2", "testCluster4", "service4"},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			ecsAwsRegionOpt:   test.awsRegion,
			ecsClusterNameOpt: test.cluster,
			ecsServiceNameOpt: test.service,
		}

		a, err := NewECSService(context.TODO(), ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if aws.StringValue(a.session.Config.Region) != test.awsRegion {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.awsRegion, aws.StringValue(a.session.Config.Region))
		}

		if a.clusterName != test.cluster {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.cluster, a.clusterName)
		}

		if a.serviceName != test.service {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.service, a.serviceName)
		}
	}
}

func TestECSWrongCreation(t *testing.T) {
	tests := []struct {
		awsRegion string
		cluster   string
		service   string
	}{
		{"", "testCluster1", "service1"},
		{"us-west-2", "", "service2"},
		{"us-west-2", "testCluster3", ""},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			ecsAwsRegionOpt:   test.awsRegion,
			ecsClusterNameOpt: test.cluster,
			ecsServiceNameOpt: test.service,
		}

		_, err := NewECSService(context.TODO(), ops)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}
	}
}

func TestECSCurrent(t *testing.T) {
	tests := []struct {
		awsRegion    string
		cluster      string
		service      string
		runningCount int64
	}{
		{"us-west-2", "testCluster1", "testService1", 5},
		{"us-west-2", "testCluster2", "testService2", 100},
		{"us-west-2", "testCluster3", "testService3", 0},
	}

	for _, test := range tests {
		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockECSDescribeServicesMultiple(t, mockECS, []string{test.service}, test.runningCount)

		// Create our scaler

		ops := map[string]interface{}{
			ecsAwsRegionOpt:   test.awsRegion,
			ecsClusterNameOpt: test.cluster,
			ecsServiceNameOpt: test.service,
		}

		e, err := NewECSService(context.TODO(), ops)
		e.client = mockECS

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		q, err := e.Current(context.TODO())

		if err != nil {
			t.Errorf("\n- %+v\n  current calling shouldn't give error: %v", test, err)
		}

		if int64(q.Q) != test.runningCount {
			t.Errorf("\n- %+v\n  current didn't return the correct running instances; got: %d, want: %d", test, q.Q, test.runningCount)
		}

	}
}

func TestECSCurrentError(t *testing.T) {
	tests := []struct {
		awsRegion    string
		cluster      string
		services     []string
		runningCount int64
	}{
		{"us-west-2", "testCluster1", []string{"testService1", "testService2"}, 5},
	}

	for _, test := range tests {
		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockECSDescribeServicesMultiple(t, mockECS, test.services, test.runningCount)

		// Create our scaler

		ops := map[string]interface{}{
			ecsAwsRegionOpt:   test.awsRegion,
			ecsClusterNameOpt: test.cluster,
			ecsServiceNameOpt: test.services[0],
		}

		e, err := NewECSService(context.TODO(), ops)
		e.client = mockECS

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		_, err = e.Current(context.TODO())

		if err == nil {
			t.Errorf("\n- %+v\n  current calling should give error, it didn't", test)
		}
	}
}

func TestECSScale(t *testing.T) {
	tests := []struct {
		current   int64
		desired   int64
		wantError bool

		wantMode types.ScalingMode
		wantQ    int64
	}{
		{10, 10, false, types.NotScaling, 0},
		{10, 15, false, types.ScalingUp, 15},
		{10, 5, false, types.ScalingDown, 5},
		{0, 1, true, types.NotScaling, 0},
	}

	for _, test := range tests {
		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockECSDescribeServicesMultiple(t, mockECS, []string{"test"}, test.current)
		awsMock.MockECSUpdateServiceCount(t, mockECS, test.desired, test.wantError)

		// Create our scaler

		ops := map[string]interface{}{
			ecsAwsRegionOpt:   "us-west-2",
			ecsClusterNameOpt: "testCluster",
			ecsServiceNameOpt: "test",
		}

		e, err := NewECSService(context.TODO(), ops)
		e.client = mockECS

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		scaled, mode, err := e.Scale(context.TODO(), types.Quantity{Q: test.desired})

		if test.wantError && err == nil {
			t.Errorf("\n- %+v\n  scale calling should give error, it didn't", test)
		}

		if !test.wantError && mode != test.wantMode {
			t.Errorf("\n- %+v\n  scale returned mode is wrong; got: %s, want: %s", test, mode, test.wantMode)
		}

		if !test.wantError && scaled.Q != test.wantQ {
			t.Errorf("\n- %+v\n  scale returned scalation Q is wrong; got: %d, want: %d", test, scaled.Q, test.wantQ)
		}
	}
}

func TestECSWait(t *testing.T) {

	tests := []struct {
		current     int64
		desired     int64
		wantTimeout bool
	}{
		{10, 5, true},
		{10, 10, false},
	}

	for _, test := range tests {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockECSDescribeServicesMultiple(t, mockECS, []string{"test"}, int64(test.current))

		ops := map[string]interface{}{
			ecsAwsRegionOpt:   "us-west-2",
			ecsClusterNameOpt: "testCluster",
			ecsServiceNameOpt: "test",
		}

		e, err := NewECSService(context.TODO(), ops)
		if err != nil {
			t.Errorf("\n- Creation shouldn't give error: %v", err)
		}
		e.client = mockECS
		e.waiterInterval = 2 * time.Millisecond

		res := make(chan error)
		go func() {
			err := e.Wait(context.TODO(), types.Quantity{Q: test.desired}, types.NotScaling)
			if err != nil {
				t.Fatalf("Wait returned error, it shoudln't: %s", err)
			}
			res <- nil
		}()
		var timeout bool
		select {
		case <-time.After(10 * time.Millisecond):
			timeout = true
		case <-res:
		}
		close(res)

		if test.wantTimeout && !timeout {
			t.Error("Wait should timeout, it dind't")
		}

		if !test.wantTimeout && timeout {
			t.Error("Wait shouldn't timeout, it did")
		}
	}
}
