package aws

import (
	"errors"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/mock/gomock"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/mock/aws/sdk"
)

// MockGetQueueAttributes mocks the API call of getting the queue attributes from SQS
func MockGetQueueAttributes(t *testing.T, mockMatcher *sdk.MockSQSAPI,
	messages, notVisible, delayed int) {

	log.Logger.Warningf("Mocking AWS iface: GetQueueAttributes")

	// SQS API result for properties are strings
	strmsgs := strconv.Itoa(messages)
	strnvsb := strconv.Itoa(notVisible)
	strdly := strconv.Itoa(delayed)

	result := &sqs.GetQueueAttributesOutput{}
	result.Attributes = map[string]*string{
		sqs.QueueAttributeNameApproximateNumberOfMessages:           aws.String(strmsgs),
		sqs.QueueAttributeNameApproximateNumberOfMessagesDelayed:    aws.String(strdly),
		sqs.QueueAttributeNameApproximateNumberOfMessagesNotVisible: aws.String(strnvsb),
	}

	// Mock as expected with our result
	mockMatcher.EXPECT().GetQueueAttributes(gomock.Any()).Do(func(input interface{}) {
		gotInput := input.(*sqs.GetQueueAttributesInput)
		// Check API received parameters are fine
		if len(gotInput.AttributeNames) < 1 {
			t.Fatalf("Expected 1 or more attrs, got %d", len(gotInput.AttributeNames))
		}

		if aws.StringValue(gotInput.QueueUrl) == "" {
			t.Fatalf("Expected a queue URL, got nothing")
		}

	}).AnyTimes().Return(result, nil)
}

// MockGetQueueAttributesError mocks the API call of getting an error from SQS
func MockGetQueueAttributesError(t *testing.T, mockMatcher *sdk.MockSQSAPI) {
	log.Logger.Warningf("Mocking AWS iface: GetQueueAttributes")
	mockMatcher.EXPECT().GetQueueAttributes(gomock.Any()).AnyTimes().Return(&sqs.GetQueueAttributesOutput{}, errors.New("Wrong!"))
}

// MockGetQueueVisibleMessagesTimes mocks multiple calls of the API returning differet results in each call,
// if random error it will send an error between the calls in a random moment
func MockGetQueueVisibleMessagesTimes(t *testing.T, mockMatcher *sdk.MockSQSAPI, randomError bool,
	visibleMessages ...int) {

	var errIdx int
	if randomError {
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		errIdx = r.Intn(len(visibleMessages))
	}

	log.Logger.Warningf("Mocking AWS iface: GetQueueAttributes")
	calls := make([]*gomock.Call, len(visibleMessages))
	for i, m := range visibleMessages {
		// SQS API result for properties are strings
		strmsgs := strconv.Itoa(m)

		result := &sqs.GetQueueAttributesOutput{}
		result.Attributes = map[string]*string{
			sqs.QueueAttributeNameApproximateNumberOfMessages: aws.String(strmsgs),
		}

		var err error
		if randomError && i == errIdx {
			err = errors.New("Random error")
		}

		// Mock as expected with our result
		calls[i] = mockMatcher.EXPECT().GetQueueAttributes(gomock.Any()).Return(result, err)
	}
	gomock.InOrder(calls...)
}
