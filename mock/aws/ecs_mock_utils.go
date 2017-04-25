package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/mock/aws/sdk"
)

// MockECSDescribeServicesMultiple will return N running services
func MockECSDescribeServicesMultiple(t *testing.T, mockMatcher *sdk.MockECSAPI,
	serviceNames []string, runningServiceInstances int64) {
	log.Logger.Warningf("Mocking AWS iface: DescribeServices")

	services := make([]*ecs.Service, len(serviceNames))

	for i, s := range serviceNames {
		services[i] = &ecs.Service{
			ServiceName:  aws.String(s),
			RunningCount: aws.Int64(runningServiceInstances),
		}
	}

	result := &ecs.DescribeServicesOutput{
		Services: services,
	}

	// Mock as expected with our result
	mockMatcher.EXPECT().DescribeServices(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*ecs.DescribeServicesInput)
		// Check API received parameters are fine
		if len(gotInput.Services) != 1 {
			t.Fatalf("Expected 1 service name, got %d", len(gotInput.Services))
		}
	}).AnyTimes().Return(result, nil)
}

// MockECSUpdateServiceCount will mock a call to AWS API setting the desired capacity of an ECS service
func MockECSUpdateServiceCount(t *testing.T, mockMatcher *sdk.MockECSAPI, desiredCount int64, wantError bool) {
	log.Logger.Warningf("Mocking AWS iface: DescribeServices")

	var err error
	if wantError {
		err = errors.New("Wrong!")
	}

	result := &ecs.UpdateServiceOutput{}

	// Mock as expected with our result
	mockMatcher.EXPECT().UpdateService(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*ecs.UpdateServiceInput)
		// Check API received parameters are fine
		if aws.Int64Value(gotInput.DesiredCount) != desiredCount {
			t.Fatalf("Wrong desired count, got %d; want %d", aws.Int64Value(gotInput.DesiredCount), desiredCount)
		}
	}).AnyTimes().Return(result, err)
}

// MockECSDescribeClusters will mock a call to AWS API setting the running and pending tasks of a cluster or multiple clusters
func MockECSDescribeClusters(t *testing.T, mockMatcher *sdk.MockECSAPI, numberOfClusters int, runningTasks, pendingTasks int64, wantError bool) {
	log.Logger.Warningf("Mocking AWS iface: DescribeClusters")
	var err error
	if wantError {
		err = errors.New("Wrong!")
	}

	// Create the clusters
	clusters := []*ecs.Cluster{}
	for i := 0; i < numberOfClusters; i++ {
		c := &ecs.Cluster{
			RunningTasksCount: aws.Int64(runningTasks),
			PendingTasksCount: aws.Int64(pendingTasks),
		}
		clusters = append(clusters, c)
	}
	result := &ecs.DescribeClustersOutput{Clusters: clusters}

	mockMatcher.EXPECT().DescribeClusters(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*ecs.DescribeClustersInput)
		// Check API received parameters are fine
		if len(gotInput.Clusters) != 1 {
			t.Fatalf("Expected 1 cluster name, got %d", len(gotInput.Clusters))
		}
	}).AnyTimes().Return(result, err)
}
