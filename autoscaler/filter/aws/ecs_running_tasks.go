package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	// Opts
	ecsAwsRegionOpt           = "aws_region"
	ecsClusterNameOpt         = "cluster_name"
	maxPendingTasksAllowedOpt = "max_pending_tasks_allowed"
	maxChecksOpt              = "max_checks"
	errorOnMaxCheckOpt        = "error_on_max_checks"
	whenOpt                   = "when"

	// id name
	ecsRunningTasksRegName = "ecs_running_tasks"
)

type when int

const (
	always when = iota
	scaleUp
	scaleDown
	unknown
)

// getWhen returns a constant when based on a string
func getWhen(v string) when {
	switch v {
	case "always":
		return always
	case "scale_up":
		return scaleUp
	case "scale_down":
		return scaleDown
	default:
		return unknown
	}
}

// ECSRunningTasks will check the running tasks on an ECS cluster and if the not running
// tasks exceeds the limit then it will break  the filter chain with the received value,
// usually this should go as one of the firsts filters
type ECSRunningTasks struct {
	session *session.Session
	client  ecsiface.ECSAPI

	clusterName        string // the name of the cluster to check
	maxNotRunningTasks int64  // max of not running tasks, if greater than this then trigger a break
	maxChecks          int64  // max simultaneous checks that breaked the chain (if 0 then no max checks)
	// if true will error when max checks is reached, if false then will not break the chain and
	// will  let the autoscaler do its job as a regular scaling iteration
	errorOnMaxCheck bool
	when            when     // can be: always, scale_up, scale_down, based on the setting it will apply the filter only when it's needed
	currentChecks   int      // The number of continued checks that broke the filter chain
	log             *log.Log // custom logger
}

// Autoregister on filterers creator
func init() {
	filter.Register(ecsRunningTasksRegName, &ecsRTCreator{})
}

type ecsRTCreator struct{}

func (e *ecsRTCreator) Create(ctx context.Context, opts map[string]interface{}) (filter.Filterer, error) {
	return NewECSRunningTasks(ctx, opts)
}

// NewECSRunningTasks creates a ECSRunningTasks object
func NewECSRunningTasks(ctx context.Context, opts map[string]interface{}) (e *ECSRunningTasks, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var ok bool

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}

	e = &ECSRunningTasks{
		log: log.WithFields(log.Fields{
			"autoscaler": asName,
			"kind":       "filterer",
			"name":       ecsRunningTasksRegName,
		}),
	}

	// Set each option with the correct type
	if e.clusterName, ok = opts[ecsClusterNameOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", ecsClusterNameOpt)
	}
	if e.clusterName == "" {
		return nil, fmt.Errorf("%s configuration opt is required", ecsClusterNameOpt)
	}
	var when string
	if when, ok = opts[whenOpt].(string); !ok {
		return nil, fmt.Errorf("%s configuration opt is required", whenOpt)
	}
	if when == "" {
		return nil, fmt.Errorf("%s configuration opt is required", whenOpt)
	}

	if e.when = getWhen(when); e.when == unknown {
		return nil, fmt.Errorf("%s configuration opt is wrong, should be one of: always, scale_up or scale_down", whenOpt)
	}

	v, ok := opts[maxPendingTasksAllowedOpt]
	if !ok {
		v = 0
		e.log.Warning("Maximum not running tasks set to 0 on ECS running tasks filter, a.k.a always all tasks running")
	}
	e.maxNotRunningTasks = types.I2Int64(v)

	// No error, if not set or 0 then disabled
	v, ok = opts[maxChecksOpt]
	if !ok {
		v = 0
		e.log.Warning("Maximum checks disabled on ECS running tasks filter")
	}
	e.maxChecks = types.I2Int64(v)

	e.errorOnMaxCheck, ok = opts[errorOnMaxCheckOpt].(bool)
	if !e.errorOnMaxCheck {
		e.log.Warning("Error on max check disabled on ECS running task filter")
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

	return
}

// Filter will check if the running tasks stisfies with the required ones
func (e *ECSRunningTasks) Filter(_ context.Context, currentQ, newQ types.Quantity) (types.Quantity, bool, error) {

	// Check if the filter is need to apply (if when is always then passing this check)
	switch e.when {
	case scaleUp:
		// If only applies on scale up and we aren't scaling up then don't execute the filter
		if newQ.Q <= currentQ.Q {
			e.log.Debugf("Filter only applies on scale up, we aren't scaling up: filter ignored")
			e.currentChecks = 0
			return newQ, false, nil
		}
	case scaleDown:
		// If only applies on scale down and we aren't scaling down then don't execute the filter
		if newQ.Q >= currentQ.Q {
			e.log.Debugf("Filter only applies on scale down, we aren't scaling down: filter ignored")
			e.currentChecks = 0
			return newQ, false, nil
		}
	}

	params := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(e.clusterName)},
	}

	resp, err := e.client.DescribeClusters(params)
	if err != nil {
		return currentQ, true, err
	}

	// Check if we have our cluster
	if len(resp.Clusters) != 1 {
		err = fmt.Errorf("Wrong number of clusters retrieved, should be one, got: %d", len(resp.Clusters))
		return currentQ, true, err
	}

	// Check the pending tasks (a.k.a not running tasks)
	got := aws.Int64Value(resp.Clusters[0].PendingTasksCount)
	if got > e.maxNotRunningTasks {
		e.log.Infof("Maximum of not running task permited exceed, max: %d, got: %d", e.maxNotRunningTasks, got)
		// Increment the checks
		e.currentChecks++

		// no max continued checks exceed (or max checks disabled), break without error
		if e.maxChecks == 0 || int64(e.currentChecks) <= e.maxChecks {
			return currentQ, true, nil
		}

		// We reached the limits of continued checks
		// Do we need to error?
		if e.errorOnMaxCheck {
			e.currentChecks = 0
			err = fmt.Errorf("Max checks of not running tasks on cluster reached")
			return currentQ, true, err
		}
		// the max continued checks are active and we dont need to error so we don't break
		// the chain and continue as always
		e.log.Infof("Although max pending tasks exceed, the max continued check also exceed, ignoring filter and continue scaling")
	}
	// All ok, you shall continue, but first we need to reset the continued checks counter
	e.log.Debugf("No %s filtered applied (pending: %d, max: %d)", ecsRunningTasksRegName, got, e.maxNotRunningTasks)
	e.currentChecks = 0
	return newQ, false, nil
}
