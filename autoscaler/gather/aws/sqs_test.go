package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"

	awsMock "github.com/themotion/ladder/mock/aws"
	"github.com/themotion/ladder/mock/aws/sdk"
)

func TestSQSCorrectCreation(t *testing.T) {
	tests := []struct {
		awsRegion     string
		queueURL      string
		queueProperty string
	}{
		{
			awsRegion:     "us-west-2",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue1",
			queueProperty: "ApproximateNumberOfMessages",
		},
		{
			awsRegion:     "eu-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue2",
			queueProperty: "ApproximateNumberOfMessagesNotVisible",
		},
		{
			awsRegion:     "us-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue3",
			queueProperty: "ApproximateNumberOfMessagesDelayed",
		},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			queueURLOpt:      test.queueURL,
			queuePropertyOpt: test.queueProperty,
			awsRegionOpt:     test.awsRegion,
		}

		s, err := NewSQS(context.TODO(), ops)

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		if aws.StringValue(s.session.Config.Region) != test.awsRegion {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.awsRegion, aws.StringValue(s.session.Config.Region))
		}

		if s.QueueURL != test.queueURL {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.queueURL, s.QueueURL)
		}

		if s.QueueProperty != test.queueProperty {
			t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.queueProperty, s.QueueProperty)
		}
	}
}

func TestSQSWrongParameterCreation(t *testing.T) {
	tests := []struct {
		awsRegion     string
		queueURL      string
		queueProperty string
	}{
		{
			awsRegion:     "us-west-2",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue1",
			queueProperty: "wrongParameter",
		},
		{
			awsRegion:     "",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue2",
			queueProperty: "ApproximateNumberOfMessagesNotVisible",
		},
		{
			awsRegion:     "us-west-1",
			queueURL:      "",
			queueProperty: "ApproximateNumberOfMessagesDelayed",
		},
		{
			awsRegion:     "us-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue3",
			queueProperty: "",
		},
		{
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue4",
			queueProperty: "ApproximateNumberOfMessagesDelayed",
		},
		{
			awsRegion:     "us-west-1",
			queueProperty: "ApproximateNumberOfMessagesDelayed",
		},
		{
			awsRegion: "us-west-1",
			queueURL:  "https://sqs.us-west-2.amazonaws.com/00000000000/queue4",
		},
	}

	for _, test := range tests {
		ops := map[string]interface{}{
			queueURLOpt:      test.queueURL,
			queuePropertyOpt: test.queueProperty,
			awsRegionOpt:     test.awsRegion,
		}

		_, err := NewSQS(context.TODO(), ops)

		if err == nil {
			t.Errorf("\n- %+v\n  Creation should give error", test)
		}
	}
}

func TestSQSGather(t *testing.T) {
	tests := []struct {
		awsRegion     string
		queueURL      string
		queueProperty string
		messages      int
		notVisible    int
		delayed       int

		wantQ int64
	}{
		{
			awsRegion:     "us-west-2",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue1",
			queueProperty: "ApproximateNumberOfMessages",
			messages:      100,
			notVisible:    200,
			delayed:       300,
			wantQ:         100,
		},
		{
			awsRegion:     "eu-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue2",
			queueProperty: "ApproximateNumberOfMessagesNotVisible",
			messages:      100,
			notVisible:    200,
			delayed:       300,
			wantQ:         200,
		},
		{
			awsRegion:     "us-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue3",
			queueProperty: "ApproximateNumberOfMessagesDelayed",
			messages:      100,
			notVisible:    200,
			delayed:       300,
			wantQ:         300,
		},
	}

	for _, test := range tests {

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSQS := sdk.NewMockSQSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockGetQueueAttributes(t, mockSQS, test.messages, test.notVisible, test.delayed)

		ops := map[string]interface{}{
			queueURLOpt:      test.queueURL,
			queuePropertyOpt: test.queueProperty,
			awsRegionOpt:     test.awsRegion,
		}

		s, err := NewSQS(context.TODO(), ops)
		s.client = mockSQS

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		// Get property
		res, err := s.Gather(context.TODO())

		if err != nil {
			t.Errorf("\n- %+v\n  Gathering shouldn't give error: %v", test, err)
		}
		if res.Q != test.wantQ {
			t.Errorf("\n- %+v\n  Gathered quantity doesn't look good, got: %d want: %d", test, res.Q, test.wantQ)
		}
	}
}

func TestSQSGatherError(t *testing.T) {
	tests := []struct {
		awsRegion     string
		queueURL      string
		queueProperty string
	}{
		{
			awsRegion:     "us-west-2",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue1",
			queueProperty: "ApproximateNumberOfMessages",
		},
		{
			awsRegion:     "eu-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue2",
			queueProperty: "ApproximateNumberOfMessagesNotVisible",
		},
		{
			awsRegion:     "us-west-1",
			queueURL:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue3",
			queueProperty: "ApproximateNumberOfMessagesDelayed",
		},
	}

	for _, test := range tests {

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSQS := sdk.NewMockSQSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockGetQueueAttributesError(t, mockSQS)

		ops := map[string]interface{}{
			queueURLOpt:      test.queueURL,
			queuePropertyOpt: test.queueProperty,
			awsRegionOpt:     test.awsRegion,
		}

		s, err := NewSQS(context.TODO(), ops)
		s.client = mockSQS

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		// Get property
		_, err = s.Gather(context.TODO())

		if err == nil {
			t.Errorf("\n- %+v\n  Gathering should give error", test)
		}
	}
}

func TestSQSGatherMaxFromMultipleConcurrentCalls(t *testing.T) {
	tests := []struct {
		messages  []int
		wantQ     int64
		wantError bool
	}{
		{[]int{0, 0, 0}, 0, false},
		{[]int{0, 1, 2}, 2, false},
		{[]int{0, 1091, 1092}, 1092, false},
		{[]int{0, 1092, 1091}, 1092, false},
		{[]int{1092, 1091, 1092}, 1092, false},
		{[]int{1092, 1090, 1091}, 1092, false},
		{[]int{1092, 1093, 1091}, 1093, false},
		{[]int{1092, 0, 0}, 1092, false},
		{[]int{1092, 1091, 1092}, 0, true},
		{[]int{0, 1, 2}, 0, true},
		{[]int{0, 1091, 1092}, 0, true},
		{[]int{0, 0, 0}, 0, true},
	}

	for _, test := range tests {

		// Create mock for AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSQS := sdk.NewMockSQSAPI(ctrl)

		// Set our mock desired result
		awsMock.MockGetQueueVisibleMessagesTimes(t, mockSQS, test.wantError, test.messages...)

		ops := map[string]interface{}{
			queueURLOpt:      "https://sqs.us-west-2.amazonaws.com/00000000000/queue",
			queuePropertyOpt: "ApproximateNumberOfMessages",
			awsRegionOpt:     "us-west-1",
		}

		s, err := NewSQS(context.TODO(), ops)
		s.client = mockSQS

		if err != nil {
			t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
		}

		// Get property
		res, err := s.Gather(context.TODO())

		if !test.wantError {
			if err != nil {
				t.Errorf("\n- %+v\n  Gathering shouldn't give error: %v", test, err)
			}
			if res.Q != test.wantQ {
				t.Errorf("\n- %+v\n  Gathered quantity doesn't look good, got: %d want: %d", test, res.Q, test.wantQ)
			}
		} else {
			if err == nil {
				t.Errorf("\n- %+v\n  Gathering should give error, it didn't: %+v", test, res)
			}
		}
	}
}
