package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	cwAwsRegionOpt       = "aws_region"
	cwMetricNameOpt      = "metric_name"
	cwNamespaceOpt       = "namespace"
	cwStatisticOpt       = "statistic"
	cwUnitOpt            = "unit"
	cwDimensionsOpt      = "dimensions"
	cwDimensionsNameOpt  = "name"
	cwDimensionsValueOpt = "value"
	cwTimeOffsetOpt      = "offset"

	// the name
	cwRegName = "aws_cloudwatch_metric"
)

// Generate sqs AWS API mocks running go generate
//go:generate mockgen -source ../../../vendor/github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface/interface.go -package sdk -destination ../../../mock/aws/sdk/cloudwatchiface_mock.go

type dimension struct {
	name  string
	value string
}

// checkCwStatistic checks if is a valid cloudwatch statistic value
func checkCwStatistic(statistic string) error {
	switch statistic {
	case cloudwatch.StatisticMaximum,
		cloudwatch.StatisticMinimum,
		cloudwatch.StatisticSum,
		cloudwatch.StatisticSampleCount,
		cloudwatch.StatisticAverage:
		return nil
	default:
		return fmt.Errorf("%s is not a valid cloudwatch statistic", statistic)
	}
}

// checkCwUnit checks if is a valid cloudwatch unit value
func checkCwUnit(unit string) error {
	switch unit {
	case cloudwatch.StandardUnitBits,
		cloudwatch.StandardUnitBitsSecond,
		cloudwatch.StandardUnitBytes,
		cloudwatch.StandardUnitBytesSecond,
		cloudwatch.StandardUnitCount,
		cloudwatch.StandardUnitCountSecond,
		cloudwatch.StandardUnitGigabits,
		cloudwatch.StandardUnitGigabitsSecond,
		cloudwatch.StandardUnitGigabytes,
		cloudwatch.StandardUnitGigabytesSecond,
		cloudwatch.StandardUnitKilobits,
		cloudwatch.StandardUnitKilobitsSecond,
		cloudwatch.StandardUnitKilobytes,
		cloudwatch.StandardUnitKilobytesSecond,
		cloudwatch.StandardUnitMegabits,
		cloudwatch.StandardUnitMegabitsSecond,
		cloudwatch.StandardUnitMegabytes,
		cloudwatch.StandardUnitMegabytesSecond,
		cloudwatch.StandardUnitMicroseconds,
		cloudwatch.StandardUnitMilliseconds,
		cloudwatch.StandardUnitNone,
		cloudwatch.StandardUnitPercent,
		cloudwatch.StandardUnitSeconds,
		cloudwatch.StandardUnitTerabits,
		cloudwatch.StandardUnitTerabitsSecond,
		cloudwatch.StandardUnitTerabytes,
		cloudwatch.StandardUnitTerabytesSecond:
		return nil
	default:
		return fmt.Errorf("%s is not a valid cloudwatch unit", unit)
	}
}

// CWMetric represents an object for gathering imputs from cloudwatch
type CWMetric struct {
	session *session.Session
	client  cloudwatchiface.CloudWatchAPI

	dimensions []*dimension  // dimensions
	metricName string        // metric name
	namespace  string        // namespace
	statistic  string        // statistic (Sum, Maximum, Minimum, SampleCount, Average)
	unit       string        // unit # Seconds, Microseconds, Milliseconds, Bytes, Kilobytes, Megabytes, Gigabytes, Terabytes, Bits, Kilobits, Megabits, Gigabits, Terabits, Percent, Count, Bytes/Second, Kilobytes/Second, Megabytes/Second, Gigabytes/Second, Terabytes/Second, Bits/Second, Kilobits/Second, Megabits/Second, Gigabits/ Second, Terabits/Second, Count/Second, None
	offset     time.Duration // The offset to apply to the query, this is usually because AWS doesn't have the values ready for the time now
	log        *log.Log      // custom logger
}

// cwMetricCreator creates the cloudwatch metric gatherer creator
type cwMetricCreator struct{}

func (a *cwMetricCreator) Create(ctx context.Context, opts map[string]interface{}) (gather.Gatherer, error) {
	return NewCWMetric(ctx, opts)
}

// Autoregister on gatherers creators
func init() {
	gather.Register(cwRegName, &cwMetricCreator{})
}

// NewCWMetric creates an Cloudwatch gatherer
func NewCWMetric(ctx context.Context, opts map[string]interface{}) (c *CWMetric, err error) {
	c = &CWMetric{}

	// Recover from panic type conversions and return like a regular error
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	// type assertions of the dimesions conf
	d := opts[cwDimensionsOpt]
	lenD := len(d.([]interface{}))
	td := make([]map[interface{}]interface{}, lenD)

	for i, dd := range d.([]interface{}) {
		td[i] = dd.(map[interface{}]interface{})
	}

	c.dimensions = make([]*dimension, lenD)
	for i, k := range td {
		n, ok := k[cwDimensionsNameOpt]
		if !ok {
			return nil, fmt.Errorf("%s configuration opt is invalid", cwDimensionsOpt)
		}
		v, ok := k[cwDimensionsValueOpt]
		if !ok {
			return nil, fmt.Errorf("%s configuration opt is invalid", cwDimensionsOpt)
		}

		c.dimensions[i] = &dimension{n.(string), v.(string)}
	}

	var ok bool

	// Check metric name
	if c.metricName, ok = opts[cwMetricNameOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cwMetricNameOpt)
	}

	if c.metricName == "" {
		return nil, fmt.Errorf("%s configuration opt is required", cwMetricNameOpt)
	}

	// Check namespace
	if c.namespace, ok = opts[cwNamespaceOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cwNamespaceOpt)
	}

	if c.namespace == "" {
		return nil, fmt.Errorf("%s configuration opt is required", cwNamespaceOpt)
	}

	// Check statistic
	if c.statistic, ok = opts[cwStatisticOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cwStatisticOpt)
	}

	if err = checkCwStatistic(c.statistic); err != nil {
		return nil, fmt.Errorf("%s configuration opt is invalid", cwStatisticOpt)
	}

	// Check unit
	if c.unit, ok = opts[cwUnitOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cwUnitOpt)
	}

	if err = checkCwUnit(c.unit); err != nil {
		return nil, fmt.Errorf("%s configuration opt is invalid", cwUnitOpt)
	}

	// Check region
	region, ok := opts[cwAwsRegionOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", cwAwsRegionOpt)
	}

	if region == "" {
		return nil, fmt.Errorf("%s configuration opt is required", cwAwsRegionOpt)
	}

	// durations
	ts, ok := opts[cwTimeOffsetOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is wrong", cwTimeOffsetOpt)
	}
	if c.offset, err = time.ParseDuration(ts); err != nil {
		return nil, err
	}
	if c.offset > 0 {
		return nil, fmt.Errorf("%s should be 0 or negative", cwTimeOffsetOpt)
	}

	// Create AWS session
	ss := session.New(&aws.Config{Region: aws.String(region)})
	if ss == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	// Create AWS cloudwatch service client
	cc := cloudwatch.New(ss)
	c.session = ss
	c.client = cc

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	c.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "gatherer",
		"name":       cwRegName,
	})

	return c, nil
}

// Gather gets the metrics from cloudwatch
func (c *CWMetric) Gather(_ context.Context) (types.Quantity, error) {
	q := types.Quantity{}

	ds := make([]*cloudwatch.Dimension, len(c.dimensions))
	for i, d := range c.dimensions {
		ds[i] = &cloudwatch.Dimension{
			Name:  aws.String(d.name),
			Value: aws.String(d.value),
		}
	}

	// Get the latest aggregated metric (one minite)
	c.log.Debugf("Retrieving cloudwatch metric")
	end := time.Now().UTC().Add(c.offset) // apply offset
	start := end.Add(-1 * time.Minute)
	params := &cloudwatch.GetMetricStatisticsInput{
		EndTime:    aws.Time(end),
		MetricName: aws.String(c.metricName),
		Namespace:  aws.String(c.namespace),
		Period:     aws.Int64(60),
		StartTime:  aws.Time(start),
		Dimensions: ds,
		Unit:       aws.String(c.unit),
		Statistics: []*string{aws.String(c.statistic)},
	}
	resp, err := c.client.GetMetricStatistics(params)
	if err != nil {
		return q, err
	}

	// Error if more than  1 metric
	if len(resp.Datapoints) != 1 {
		return q, fmt.Errorf("Wrong value of metrics retrieved: %d", len(resp.Datapoints))
	}
	// Only take the first datapoint
	d := resp.Datapoints[0]
	var res *float64
	switch c.statistic {
	case cloudwatch.StatisticMaximum:
		res = d.Maximum
	case cloudwatch.StatisticMinimum:
		res = d.Minimum
	case cloudwatch.StatisticSum:
		res = d.Sum
	case cloudwatch.StatisticSampleCount:
		res = d.SampleCount
	case cloudwatch.StatisticAverage:
		res = d.Average
	}

	q.Q = int64(aws.Float64Value(res))

	c.log.Debugf("Retrieved cloudwatch metric input: %s", q)

	return q, nil
}
