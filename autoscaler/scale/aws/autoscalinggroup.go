package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/themotion/ladder/autoscaler/scale"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	awsRegionOpt                    = "aws_region"
	asgNameOpt                      = "auto_scaling_group_name"
	asgWaitUpDurationOpt            = "scale_up_wait_duration"
	asgWaitDownDurationOpt          = "scale_down_wait_duration"
	asgForceMinMaxOpt               = "force_min_max"
	asgRemainingClosestHourLimitDur = "remaining_closest_hour_limit_duration"
	asgMaxTimesRemainingNoDownscale = "max_no_downscale_rch_limit"

	// the name
	asgRegName = "aws_autoscaling_group"

	// internal constants
	asgDefaultWaiterInterval = 5 * time.Second
)

// Generate autoscaling EC2 API mocks running go generate
//go:generate mockgen -source ../../../vendor/github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface/interface.go -package sdk -destination ../../../mock/aws/sdk/autoscalingiface_mock.go
//go:generate mockgen -source ../../../vendor/github.com/aws/aws-sdk-go/service/ec2/ec2iface/interface.go -package sdk -destination ../../../mock/aws/sdk/ec2iface_mock.go

// ASG represents an object for scaling in a SecurityGroup
type ASG struct {
	session   *session.Session
	asClient  autoscalingiface.AutoScalingAPI
	ec2Client ec2iface.EC2API

	// Auto scaling group name
	asgName string

	// the waiting durations
	waitUpDuration   time.Duration
	waitDownDuration time.Duration

	// force setting the bounds (min and max)
	forceMinMax bool

	// The remaining time that an instances must bypass on last hour (billing hour)
	// to be valid to take down
	remainingClosestHourLimit time.Duration

	// Remaining closes hoyr limit counter
	remainingNoDownscaleC int

	// this sets the number of tries the remaining limit will be valid before not aplying it,
	// for example we have 10 instances and we want 5 instances, we apply th filter and none of
	// the instances met the requirement of running limit, it will set the deseired to 10. if
	// this happens N times then the limit will set to a conclusion that there is some kind of
	// problem, bad interval configuration or whatever, with this limit we dont keep running all instances
	// forever
	maxTimesRemainingNoDownscale int

	// The interval the waiter will check the quantity is met
	waiterInterval time.Duration

	// custom logger
	log *log.Log
}

// asgCreator creates the auto scaling group scaler creator
type asgCreator struct{}

func (a *asgCreator) Create(ctx context.Context, opts map[string]interface{}) (scale.Scaler, error) {
	return NewASG(ctx, opts)
}

// Autoregister on scaler creators
func init() {
	scale.Register(asgRegName, &asgCreator{})
}

// NewASG creates an ASG scaler
func NewASG(ctx context.Context, opts map[string]interface{}) (a *ASG, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}

	a = &ASG{
		log: log.WithFields(log.Fields{
			"autoscaler": asName,
			"kind":       "scaler",
			"name":       asgRegName,
		}),
		waiterInterval: asgDefaultWaiterInterval,
	}

	// Set each option with the correct type
	if a.asgName, ok = opts[asgNameOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", asgNameOpt)
	}

	if a.asgName == "" {
		return nil, fmt.Errorf("%s configuration opt is required", asgNameOpt)
	}

	region, ok := opts[awsRegionOpt].(string)
	if !ok {
		return nil, fmt.Errorf("%s configuration opt is required", awsRegionOpt)
	}

	if region == "" {
		return nil, fmt.Errorf("%s configuration opt is required", awsRegionOpt)
	}

	// durations
	ts, ok := opts[asgWaitUpDurationOpt].(string)
	if ok {
		if a.waitUpDuration, err = time.ParseDuration(ts); err != nil {
			return
		}
	}

	ts, ok = opts[asgWaitDownDurationOpt].(string)
	if ok {
		if a.waitDownDuration, err = time.ParseDuration(ts); err != nil {
			return
		}
	}

	a.forceMinMax, _ = opts[asgForceMinMaxOpt].(bool)
	if a.forceMinMax {
		a.log.Warning("Scaler will force max and min bounds on autoscaling group")
	}

	// Set each option with the correct type
	if a.maxTimesRemainingNoDownscale, ok = opts[asgMaxTimesRemainingNoDownscale].(int); !ok {
		a.maxTimesRemainingNoDownscale = 0
	}

	ts, ok = opts[asgRemainingClosestHourLimitDur].(string)
	if ok {
		var err2 error
		if a.remainingClosestHourLimit, err2 = time.ParseDuration(ts); err2 != nil {
			a.remainingClosestHourLimit = 0
		}
	} else {
		a.remainingClosestHourLimit = 0
	}

	if a.remainingClosestHourLimit > 1*time.Hour {
		return nil, fmt.Errorf("%s can't be higher than 60m", asgRemainingClosestHourLimitDur)
	}

	if a.maxTimesRemainingNoDownscale <= 0 && a.remainingClosestHourLimit > 0 {
		return nil, fmt.Errorf("%s is required when %s is used", asgMaxTimesRemainingNoDownscale, asgRemainingClosestHourLimitDur)
	}

	// Create AWS session
	s := session.New(&aws.Config{Region: aws.String(region)})
	if s == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	// Create AWS Auto scaling group service client
	a.session = s
	a.asClient = autoscaling.New(a.session)
	a.ec2Client = ec2.New(a.session)

	return a, err
}

func (a *ASG) getAutoscalingGroup() (*autoscaling.Group, error) {
	a.log.Debugf("Retrieving autoscaling group: %s", a.asgName)

	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(a.asgName),
		},
		MaxRecords: aws.Int64(1),
	}
	resp, err := a.asClient.DescribeAutoScalingGroups(params)

	if err != nil {
		return nil, err
	}

	// Check retrieval is correct
	if len(resp.AutoScalingGroups) != 1 {
		err = fmt.Errorf("Wrong number of Autoscaling groups retrieved: %d", len(resp.AutoScalingGroups))
		return nil, err
	}

	return resp.AutoScalingGroups[0], nil

}

// Current returns the current quantity of instances on th auto scaling group
func (a *ASG) Current(ctx context.Context) (q types.Quantity, err error) {
	q = types.Quantity{Q: 0}

	group, err := a.getAutoscalingGroup()
	if err != nil {
		return q, err
	}

	q.Q = int64(len(group.Instances))
	a.log.Debugf("%s auto scaling group has %d instances (desired: %d)", a.asgName, q.Q, aws.Int64Value(group.DesiredCapacity))

	return q, nil
}

// Scale sets the desired quantity on security group
func (a *ASG) Scale(ctx context.Context, newQ types.Quantity) (types.Quantity, types.ScalingMode, error) {
	mode := types.NotScaling
	currentQ, err := a.Current(ctx)
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
		return types.Quantity{}, mode, nil
	}

	// Limit the new valid to the ones closest to the finish hour that have run at least
	// the desired minutes of the running hour
	if a.remainingClosestHourLimit > 0 && mode == types.ScalingDown {
		a.log.Debugf("Filtering running instances with: %s", a.remainingClosestHourLimit)
		q, err2 := a.filterClosestHourQ(newQ)
		if err2 != nil {
			a.log.Error(err2)
		} else {
			newQ = q
		}
	}

	desired := int64(newQ.Q)
	params := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(a.asgName),
		DesiredCapacity:      aws.Int64(desired),
	}

	// If we want to force then set to desired
	if a.forceMinMax {
		a.log.Debugf("Forcing max & min with the same value as desired: %d ", desired)
		params.MaxSize = params.DesiredCapacity
		params.MinSize = params.DesiredCapacity
	}

	// Closest hour limit could change the wanted and not scale, check
	if currentQ.Q == newQ.Q {
		a.log.Infof("Want to scale to %d but didn't due to closest hour limiter, Iterations by filter without scaling: %d (max: %d)", newQ.Q, a.remainingNoDownscaleC, a.maxTimesRemainingNoDownscale)
		return types.Quantity{}, types.NotScaling, nil
	}

	_, err = a.asClient.UpdateAutoScalingGroup(params)

	if err != nil {
		return types.Quantity{}, types.NotScaling, err
	}

	a.log.Infof("Scaled %s group from %d to %d desired instances", a.asgName, currentQ.Q, newQ.Q)
	return newQ, mode, nil
}

// Wait will wait a given time
// TODO: Quantity checks as instance number same as desired, etc...
func (a *ASG) Wait(ctx context.Context, scaledQ types.Quantity, mode types.ScalingMode) error {
	// First check the wait time is set and if its then wait
	switch mode {
	case types.ScalingUp:
		if a.waitUpDuration > 0 {
			a.log.Debugf("Waiting %s after scalation", a.waitUpDuration)
			time.Sleep(a.waitUpDuration)
			return nil
		}
	case types.ScalingDown:
		if a.waitDownDuration > 0 {
			a.log.Debugf("Waiting %s after scalation", a.waitDownDuration)
			time.Sleep(a.waitDownDuration)
			return nil
		}
	}

	// if no time specified then we need to wait until the scaled quantity matches the running instance machines
	t := time.NewTicker(a.waiterInterval)

	a.log.Debugf("Waiting for ASG desired instances meet the scaler desired quantity...")
	for range t.C {
		q, err := a.Current(ctx)
		if err != nil {
			return err
		}
		// If met the desired ones then exit
		if q.Q == scaledQ.Q {
			return nil
		}
	}
	return nil
}

// filterClosesHourQ will limit the desired number of machines to only the ones that can keep running
// this filtering is made based on the premise that all running machines need to keep running until
// remainingClosestHourLimit instance attr duration remains to finish the closing hour. For example:
// We have remainingClosestHourLimit attr set to 10m. there are 10 machines running and want to downscale
// to 2 machines, this should take down 8 machines, but 5 of those machines have been running for 30m and
// the remaining 3 for 52m. This would filter to 7 machines because 60-10 = 50m, at least each machine
// needs to be running 50m. and only 3 met the requirement (10-3 = 7)
func (a *ASG) filterClosestHourQ(newQ types.Quantity) (types.Quantity, error) {
	// Before anything check if we reached to the maximum of tries without downscaling due to this filter
	if a.remainingNoDownscaleC >= a.maxTimesRemainingNoDownscale {
		a.remainingNoDownscaleC = 0
		a.log.Warningf("Closest hour limit didn't downscale in %d times, reached maximum limit of no downscaling limit", a.maxTimesRemainingNoDownscale)
		return newQ, nil
	}

	group, err := a.getAutoscalingGroup()
	if err != nil {
		return newQ, err
	}

	// Get current instances
	instIDs := make([]*string, len(group.Instances))
	for i, inst := range group.Instances {
		instIDs[i] = inst.InstanceId
	}

	var validDownInsts, currentInst int64
	params := &ec2.DescribeInstancesInput{
		InstanceIds: instIDs,
	}
	for {
		resp, err := a.ec2Client.DescribeInstances(params)
		if err != nil {
			return newQ, err
		}

		for _, r := range resp.Reservations {
			for _, i := range r.Instances {
				// Only valid instance if are in running or pending state
				// (pending state ones will be running in few minutes)
				ist := aws.StringValue(i.State.Name)
				if ist == ec2.InstanceStateNamePending || ist == ec2.InstanceStateNameRunning {
					currentInst++
					// Get the run minutes from the last billing hour
					tRunning := time.Now().UTC().Sub(aws.TimeValue(i.LaunchTime))
					rlbh := tRunning % (1 * time.Hour) // run last billing hour
					// if the instance has run at least N time in the last billing hour then
					// allow to destroy
					if rlbh >= ((1 * time.Hour) - a.remainingClosestHourLimit) {
						validDownInsts++
					}
				}
			}
		}
		if resp.NextToken == nil || aws.StringValue(resp.NextToken) == "" {
			break
		}
		params.NextToken = resp.NextToken
	}

	// if there are too much instances to take down then set the current number to
	//only take down the number of instances that can remove
	wantDown := currentInst - newQ.Q
	if wantDown > validDownInsts {
		newQLim := types.Quantity{Q: currentInst - validDownInsts}
		a.log.Infof("Filtered instances by closest hour limit running time(%s) from %s to %s", a.remainingClosestHourLimit, newQ, newQLim)

		// If the filter decided to not downscale, then increase the counter
		if newQLim.Q == currentInst {
			a.remainingNoDownscaleC++
		}
		return newQLim, nil
	}

	// Reset counter
	a.remainingNoDownscaleC = 0
	return newQ, nil
}
