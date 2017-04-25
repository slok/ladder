package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"

	"github.com/themotion/ladder/autoscaler/scale"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	ecsAwsRegionOpt   = "aws_region"
	ecsClusterNameOpt = "cluster_name"
	ecsServiceNameOpt = "service_name"

	// the name
	ecsRegName = "aws_ecs_service"

	// internal constants
	ecsDefaultWaiterInterval = 5 * time.Second
)

// Generate autoscaling EC2 API mocks running go generate
//go:generate mockgen -source ../../../vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -package sdk -destination ../../../mock/aws/sdk/ecsiface_mock.go

// ECSService represents an object for scaling an ECS service
type ECSService struct {
	session *session.Session
	client  ecsiface.ECSAPI

	clusterName    string        // ECS cluster name
	serviceName    string        // Service name
	serviceResp    *ecs.Service  // current ECS service response from the api
	waiterInterval time.Duration // Waiter check interval
	log            *log.Log      // custom logger
}

// ecsCreator creates the auto scaling group scaler creator
type ecsCreator struct{}

func (e *ecsCreator) Create(ctx context.Context, opts map[string]interface{}) (scale.Scaler, error) {
	return NewECSService(ctx, opts)
}

// Autoregister on scaler creators
func init() {
	scale.Register(ecsRegName, &ecsCreator{})
}

// NewECSService creates an ECSService scaler
func NewECSService(ctx context.Context, opts map[string]interface{}) (e *ECSService, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	e = &ECSService{
		waiterInterval: ecsDefaultWaiterInterval,
	}

	// Prepare ops
	var ok bool

	// Set each option with the correct type
	if e.clusterName, ok = opts[ecsClusterNameOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", ecsClusterNameOpt)
	}

	if e.clusterName == "" {
		return nil, fmt.Errorf("%s configuration opt is required", ecsClusterNameOpt)
	}

	if e.serviceName, ok = opts[ecsServiceNameOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", ecsServiceNameOpt)
	}

	if e.serviceName == "" {
		return nil, fmt.Errorf("%s configuration opt is required", ecsServiceNameOpt)
	}

	region, ok := opts[ecsAwsRegionOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", ecsAwsRegionOpt)
	}

	if region == "" {
		return nil, fmt.Errorf("%s configuration opt is required", ecsAwsRegionOpt)
	}

	// Create AWS session
	s := session.New(&aws.Config{Region: aws.String(region)})
	if s == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	// Create the ECS service client
	e.session = s
	e.client = ecs.New(e.session)

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	e.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "scaler",
		"name":       ecsRegName,
	})

	return
}

// Current returns the number of current ECS service instances
func (e *ECSService) Current(_ context.Context) (q types.Quantity, err error) {
	q = types.Quantity{Q: 0}

	e.log.Debugf("Retrieving current instances of %s service", e.serviceName)

	params := &ecs.DescribeServicesInput{
		Services: []*string{aws.String(e.serviceName)},
		Cluster:  aws.String(e.clusterName),
	}
	resp, err := e.client.DescribeServices(params)
	if err != nil {
		return
	}

	// Check description is ok
	if len(resp.Services) != 1 {
		err = fmt.Errorf("wrong number of services retrieved: %d", len(resp.Services))
		return
	}

	e.serviceResp = resp.Services[0]
	q.Q = aws.Int64Value(e.serviceResp.RunningCount)
	e.log.Debugf("%s service has %d running instances (desired %d)", e.serviceName, q.Q, aws.Int64Value(e.serviceResp.DesiredCount))

	return
}

// Scale sets the desired count of instances on the cluster service, it limits
// to the number of ECS machines of the cluster
func (e *ECSService) Scale(ctx context.Context, newQ types.Quantity) (types.Quantity, types.ScalingMode, error) {
	mode := types.NotScaling
	currentQ, err := e.Current(ctx)
	if err != nil {
		return types.Quantity{}, mode, err
	}

	// No change
	switch {
	case newQ.Q > currentQ.Q:
		mode = types.ScalingUp
	case newQ.Q < currentQ.Q:
		mode = types.ScalingDown
	default:
		return types.Quantity{}, mode, err
	}

	params := &ecs.UpdateServiceInput{
		Service:      aws.String(e.serviceName),
		Cluster:      aws.String(e.clusterName),
		DesiredCount: aws.Int64(newQ.Q),
	}
	_, err = e.client.UpdateService(params)
	if err != nil {
		return types.Quantity{}, mode, err
	}

	e.log.Infof("Scaled %s service from %d to %d desired instances", e.serviceName, currentQ.Q, newQ.Q)
	return newQ, mode, nil
}

// Wait will wait a given time
func (e *ECSService) Wait(ctx context.Context, scaledQ types.Quantity, mode types.ScalingMode) error {
	t := time.NewTicker(e.waiterInterval)

	e.log.Debugf("Start waiting for ECS services meet the scaler desired quantity...")
	for range t.C {
		q, err := e.Current(ctx)
		if err != nil {
			return err
		}
		// If met the desired ones then exit
		if q.Q == scaledQ.Q {
			return nil
		}
	}

	// Timeout will handle this infinite loop if something happens, so we are not reaching here never
	return nil
}
