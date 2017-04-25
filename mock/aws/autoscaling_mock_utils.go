package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/golang/mock/gomock"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/mock/aws/sdk"
)

// MockDescribeAutoScalingGroupsSingle mocks DescribeAutoScalingGroups API call with a single group
func MockDescribeAutoScalingGroupsSingle(t *testing.T, mockMatcher *sdk.MockAutoScalingAPI,
	runningInstances, max, min, desired int64) {

	MockDescribeAutoScalingGroupsMultiple(t, mockMatcher, 1, runningInstances, max, min, desired)
}

// MockDescribeAutoScalingGroupsMultiple mocks DescribeAutoScalingGroups API call with multiple groups
func MockDescribeAutoScalingGroupsMultiple(t *testing.T, mockMatcher *sdk.MockAutoScalingAPI,
	groups, runningInstances, max, min, desired int64) {

	log.Logger.Warningf("Mocking AWS iface: DescribeAutoScalingGroups")

	// Only mock the things we need
	groupName := aws.String("")
	maxSize := aws.Int64(max)
	minSize := aws.Int64(min)
	desiredSize := aws.Int64(desired)

	instances := make([]*autoscaling.Instance, runningInstances)

	for i := 0; i < int(runningInstances); i++ {
		instances[i] = &autoscaling.Instance{
			LifecycleState: aws.String(autoscaling.LifecycleStateInService),
		}
	}

	result := &autoscaling.DescribeAutoScalingGroupsOutput{}
	result.AutoScalingGroups = make([]*autoscaling.Group, groups)

	for i := 0; i < int(groups); i++ {
		result.AutoScalingGroups[i] = &autoscaling.Group{
			AutoScalingGroupName: groupName,
			MaxSize:              maxSize,
			MinSize:              minSize,
			DesiredCapacity:      desiredSize,
			Instances:            instances,
		}

	}

	// Mock as expected with our result
	mockMatcher.EXPECT().DescribeAutoScalingGroups(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*autoscaling.DescribeAutoScalingGroupsInput)
		// Check API received parameters are fine
		if len(gotInput.AutoScalingGroupNames) != 1 {
			t.Fatalf("Expected 1 group name, got %d", len(gotInput.AutoScalingGroupNames))
		}
	}).AnyTimes().Return(result, nil)
}

// MockUpdateAutoScalingGroup mocks the call of setting a desired capacity on an scalation group
func MockUpdateAutoScalingGroup(t *testing.T, mockMatcher *sdk.MockAutoScalingAPI,
	checkDesired, checkMin, checkMax int64, checkBounds, wantError bool) {

	log.Logger.Warningf("Mocking AWS iface: UpdateAutoScalingGroup")
	result := &autoscaling.UpdateAutoScalingGroupOutput{}

	var err error
	if wantError {
		err = errors.New("Wrong!")
	}

	mockMatcher.EXPECT().UpdateAutoScalingGroup(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*autoscaling.UpdateAutoScalingGroupInput)
		// Check API received parameters are fine
		if checkBounds && aws.Int64Value(gotInput.MaxSize) != checkMax {
			t.Fatalf("Expected %d as ASG max, got %d", checkMax, aws.Int64Value(gotInput.MaxSize))
		}
		if checkBounds && aws.Int64Value(gotInput.MinSize) != checkMin {
			t.Fatalf("Expected %d as ASG min, got %d", checkMin, aws.Int64Value(gotInput.MinSize))
		}
		if aws.Int64Value(gotInput.DesiredCapacity) != checkDesired {
			t.Fatalf("Expected %d as ASG desired, got %d", checkDesired, aws.Int64Value(gotInput.DesiredCapacity))
		}
	}).AnyTimes().Return(result, err)

}
