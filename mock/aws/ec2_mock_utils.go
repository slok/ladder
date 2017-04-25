package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/mock/aws/sdk"
)

// MockEC2DescribeInstances will mock ec2 DescribeInstances API call
func MockEC2DescribeInstances(t *testing.T, mockMatcher *sdk.MockEC2API, instances []*ec2.Instance, wantError bool) {
	log.Logger.Warningf("Mocking AWS EC2 iface: DescribeInstances")

	var err error
	if wantError {
		err = errors.New("Error wanted!")
	}

	result := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			&ec2.Reservation{
				Instances: instances,
			},
		},
	}

	// Mock as expected with our result
	mockMatcher.EXPECT().DescribeInstances(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*ec2.DescribeInstancesInput)
		// Check API received parameters are fine
		if len(gotInput.InstanceIds) == 0 {
			t.Fatalf("Expected at least 1 instance ID, got %d", len(gotInput.InstanceIds))
		}
	}).AnyTimes().Return(result, err)
}
