package autoscaler

import (
	"context"
	"fmt"
	"time"

	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/metrics"
	"github.com/themotion/ladder/types"
)

type inputter struct {
	config *config.Inputter // Inputter configuration
	name   string           // The name of the inputter (also in the configuration)

	// Our gathering and arrangement objects
	gatherer gather.Gatherer
	arranger arrange.Arranger

	asName string   // The name of the autoscaler
	log    *log.Log // custom logger
}

// NewInputter creates a inputter
func newInputter(ctx context.Context, cfg *config.Inputter) (*inputter, error) {
	asName := ctx.Value(AutoscalerCtxKey).(string)
	i := &inputter{
		config: cfg,
		name:   cfg.Name,
		asName: asName,
		log: log.WithFields(log.Fields{
			"autoscaler": asName,
			"kind":       "inputter",
			"name":       cfg.Name,
		}),
	}

	// Set the main pieces
	if err := i.setGatherer(ctx); err != nil {
		return nil, err
	}
	if err := i.setArranger(ctx); err != nil {
		return nil, err
	}
	return i, nil
}

// setGatherer sets the correct Gatherer
func (i *inputter) setGatherer(ctx context.Context) error {
	if i.gatherer == nil {
		if i.config.Gather.Kind == "" {
			err := fmt.Errorf("gatherer type can't be empty")
			i.log.Errorf("error creating gatherer: %v", err)
			return err
		}

		// Create new gatherer using the registry and the creators
		g, err := gather.Create(ctx, i.config.Gather.Kind, i.config.Gather.Config)

		if err != nil {
			i.log.Errorf("error creating gatherer: %v", err)
			return err
		}
		i.gatherer = g
		i.log.Debugf("Gatherer created")
		return nil
	}
	i.log.Debugf("Gatherer already created, ignoring")
	return nil
}

// setArranger returns the correct Arranger
func (i *inputter) setArranger(ctx context.Context) error {
	if i.arranger == nil {
		// Allow empty arranger
		if i.config.Arrange.Kind == "" {
			i.log.Warningf("Arranger not specified, gatherer value will be passed transparently to the solver")
			return nil
		}

		// Create new arranger using the registry and the creators
		ar, err := arrange.Create(ctx, i.config.Arrange.Kind, i.config.Arrange.Config)

		if err != nil {
			i.log.Errorf("error creating arranger: %v", err)
			return err
		}
		i.arranger = ar
		i.log.Debugf("Arranger created")
		return nil
	}
	i.log.Debugf("Arranger already created, ignoring")
	return nil
}

// Gathers and arranges the input and returns them so the solver can make a decision
func (i *inputter) gatherAndArrange(ctx context.Context, currentQ types.Quantity) (types.Quantity, error) {
	newQ := types.Quantity{}

	// Gather the input
	start := time.Now().UTC()
	inQ, err := i.gatherer.Gather(ctx)
	metrics.ObserveGathererDuration(time.Now().UTC().Sub(start), i.asName, i.name, i.config.Gather.Kind)
	if err != nil {
		metrics.AddGathererErrors(1, i.asName, i.name, i.config.Gather.Kind)
		err = fmt.Errorf("error gathering input: %s", err)
		return newQ, err
	}
	metrics.SetGathererQ(inQ, i.asName, i.name, i.config.Gather.Kind)
	i.log.Debugf("Gatherer %s:%s gathered: %s", i.name, i.config.Gather.Kind, inQ)

	// Make a decision to up/down scale or not (if not arranger then pass it transparently)
	if i.arranger != nil {
		newQ, err = i.arranger.Arrange(ctx, inQ, currentQ)
		if err != nil {
			err = fmt.Errorf("error making a decision: %s", err)
			return newQ, err
		}
	} else {
		newQ = inQ
	}
	i.log.Infof("Arranger %s:%s arranged: %s", i.name, i.config.Arrange.Kind, newQ)
	return newQ, nil
}
