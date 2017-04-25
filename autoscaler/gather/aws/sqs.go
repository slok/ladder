package aws

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	awsRegionOpt     = "aws_region"
	queueURLOpt      = "queue_url"
	queuePropertyOpt = "queue_property"

	// the name
	sqsRegName = "aws_sqs"

	// Number of calls for messages per gather
	sqsCallTimes = 3
)

// Generate sqs AWS API mocks running go generate
//go:generate mockgen -source ../../../vendor/github.com/aws/aws-sdk-go/service/sqs/sqsiface/interface.go -package sdk -destination ../../../mock/aws/sdk/sqsiface_mock.go

// SQS represents an object for gathering imputs from SQS
type SQS struct {
	session *session.Session
	client  sqsiface.SQSAPI

	// The queue Name
	QueueURL string

	// The queue property where we will get the information
	QueueProperty string

	log *log.Log // Custom logger
}

// sqsCreator creates the the sqs gatherer creator
type sqsCreator struct{}

func (a *sqsCreator) Create(ctx context.Context, opts map[string]interface{}) (gather.Gatherer, error) {
	return NewSQS(ctx, opts)
}

// Autoregister on gatherers creators
func init() {
	gather.Register(sqsRegName, &sqsCreator{})
}

// NewSQS creates an SQS gatherer
func NewSQS(ctx context.Context, opts map[string]interface{}) (s *SQS, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	s = &SQS{}

	// Prepare ops
	var ok bool

	// Set each option with the correct type
	if s.QueueURL, ok = opts[queueURLOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", queueURLOpt)
	}

	if s.QueueURL == "" {
		return nil, fmt.Errorf("%s configuration opt is required", queueURLOpt)
	}

	if s.QueueProperty, ok = opts[queuePropertyOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", queuePropertyOpt)
	}

	if s.QueueProperty == "" {
		return nil, fmt.Errorf("%s configuration opt is required", queuePropertyOpt)
	}

	// Check queue property correct
	var valid bool
	switch s.QueueProperty {
	case
		sqs.QueueAttributeNameApproximateNumberOfMessages,
		sqs.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
		sqs.QueueAttributeNameApproximateNumberOfMessagesDelayed:
		valid = true
	}

	if !valid {
		return nil, fmt.Errorf("%s configuration opt is wrong", queuePropertyOpt)
	}

	region, ok := opts[awsRegionOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", awsRegionOpt)
	}

	if region == "" {
		return nil, fmt.Errorf("%s configuration opt is required", awsRegionOpt)
	}

	// Create AWS session
	ss := session.New(&aws.Config{Region: aws.String(region)})
	if ss == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	// Create AWS SQS service client
	c := sqs.New(ss)
	s.session = ss
	s.client = c

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	s.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "gatherer",
		"name":       sqsRegName,
	})

	return
}

// sqsResult is an aux result for the goroutines channel
type sqsResult struct {
	result int
	err    error
}

// sqsGather calls to SQS and returns in a channel the result or error
func (s *SQS) sqsGather(resultChan chan<- sqsResult, wg *sync.WaitGroup) {
	defer wg.Done()
	s.log.Debugf("Gathering input from AWS API (running concurrently)")

	params := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.QueueURL),
		AttributeNames: []*string{
			aws.String(s.QueueProperty),
		},
	}

	// Get the sqs properties
	resp, err := s.client.GetQueueAttributes(params)
	if err != nil {
		resultChan <- sqsResult{0, err}
		return
	}

	// Get property
	prop, ok := resp.Attributes[s.QueueProperty]
	if !ok {
		resultChan <- sqsResult{0, fmt.Errorf("Error retrieving SQS queue properties")}
		return
	}
	// Convert string to integer
	propInt, err := strconv.Atoi(aws.StringValue(prop))
	if err != nil {
		resultChan <- sqsResult{0, fmt.Errorf("Error retrieving SQS queue properties")}
		return
	}
	resultChan <- sqsResult{propInt, nil}
	return
}

// Gather retrieves the SQS properties and returns the quantity of the desired property
func (s *SQS) Gather(_ context.Context) (types.Quantity, error) {
	q := types.Quantity{Q: 0}

	// Prepare sync and result channel for concurrent calls to the API
	var wg sync.WaitGroup
	wg.Add(sqsCallTimes)
	resChan := make(chan sqsResult)
	done := make(chan bool)
	results := []sqsResult{}

	// Make the multiple calls
	for i := 0; i < sqsCallTimes; i++ {
		go s.sqsGather(resChan, &wg)
	}

	// Run the result "aggregator"
	go func() {
		for r := range resChan {
			results = append(results, r)
		}
		// Finish getting all the results
		done <- true
	}()

	// wait for all the calls (Wait to process all the API calls)
	s.log.Debugf("Waiting SQS call results")
	wg.Wait()
	close(resChan)

	// Wait to drain the resultsChannel
	<-done

	// Get the maximum value of all (this discards 0s) or if errors then return
	var max float64
	for _, r := range results {
		// Return the first error
		if r.err != nil {
			return q, r.err
		}
		max = math.Max(max, float64(r.result))
	}
	q.Q = int64(max)

	s.log.Debugf("Retrieved sqs input: %s", q)

	return q, nil
}
