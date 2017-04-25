package aws

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/golang/mock/gomock"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/mock/aws/sdk"
)

// MockGetMetricStatistics will mock GetMetricStatistics cloudwatch API call
func MockGetMetricStatistics(t *testing.T, mockMatcher *sdk.MockCloudWatchAPI, metric float64, statistic string, numberOfMetrics int) {
	log.Logger.Warningf("Mocking AWS iface: GetMetricStatistics")
	if numberOfMetrics == 0 {
		numberOfMetrics = 1
	}

	result := &cloudwatch.GetMetricStatisticsOutput{}
	result.Label = aws.String("fake")
	result.Datapoints = make([]*cloudwatch.Datapoint, numberOfMetrics)
	for i := 0; i < numberOfMetrics; i++ {
		d := &cloudwatch.Datapoint{}

		switch statistic {
		case cloudwatch.StatisticMaximum:
			d.Maximum = aws.Float64(metric)
		case cloudwatch.StatisticMinimum:
			d.Minimum = aws.Float64(metric)
		case cloudwatch.StatisticSum:
			d.Sum = aws.Float64(metric)
		case cloudwatch.StatisticSampleCount:
			d.SampleCount = aws.Float64(metric)
		case cloudwatch.StatisticAverage:
			d.Average = aws.Float64(metric)
		default:
			t.Fatalf("Wrong metric statistic: %s", statistic)
		}
		d.Timestamp = aws.Time(time.Now().UTC())
		d.Unit = aws.String(cloudwatch.StandardUnitPercent)
		result.Datapoints[i] = d
	}

	// Mock as expected with our result
	mockMatcher.EXPECT().GetMetricStatistics(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*cloudwatch.GetMetricStatisticsInput)
		// Check API received parameters are fine
		if aws.StringValue(gotInput.Namespace) == "" {
			t.Fatalf("Expected namespace, got nothing")
		}

		if aws.StringValue(gotInput.MetricName) == "" {
			t.Fatalf("Expected metric name, got nothing")
		}

		if aws.StringValue(gotInput.Unit) == "" {
			t.Fatalf("Expected unit, got nothing")
		}

		if len(gotInput.Statistics) != 1 {
			t.Fatalf("Wrong statistics name")
		}

	}).AnyTimes().Return(result, nil)

}

// MockGetMetricStatisticsError mocks the API call of getting an error from Cloudwatch
func MockGetMetricStatisticsError(t *testing.T, mockMatcher *sdk.MockCloudWatchAPI) {
	log.Logger.Warningf("Mocking AWS iface: GetQueueAttributes")
	mockMatcher.EXPECT().GetMetricStatistics(gomock.Any()).AnyTimes().Return(&cloudwatch.GetMetricStatisticsOutput{}, errors.New("Wrong!"))
}
