package main

import (

	//_ "net/http/pprof" // Activate pprof for profiling
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/themotion/ladder/autoscaler"
	"github.com/themotion/ladder/autoscaler/arrange"
	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/autoscaler/scale"
	"github.com/themotion/ladder/autoscaler/solve"
	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/version"
	"github.com/themotion/ladder/web"
	"github.com/themotion/ladder/web/api/v1"

	// Register app healthcheck
	"github.com/themotion/ladder/health"

	// Register metrics
	_ "github.com/themotion/ladder/metrics"

	// Register blocks
	_ "github.com/themotion/ladder/autoscaler/arrange/common"
	_ "github.com/themotion/ladder/autoscaler/filter/aws"
	_ "github.com/themotion/ladder/autoscaler/filter/common"
	_ "github.com/themotion/ladder/autoscaler/gather/aws"
	_ "github.com/themotion/ladder/autoscaler/gather/common"
	_ "github.com/themotion/ladder/autoscaler/gather/metrics"
	_ "github.com/themotion/ladder/autoscaler/scale/aws"
	_ "github.com/themotion/ladder/autoscaler/scale/common"
	_ "github.com/themotion/ladder/autoscaler/solve/common"
)

func clean() error {
	log.Logger.Infof("Cleaning before exiting")
	time.Sleep(1 * time.Second)
	return nil
}

func main() {
	// Run main program
	exCode := Main()

	// Clean
	err := clean()

	if err != nil {
		log.Logger.Errorf("Cleanig error: %v", err)
		exCode = 1
	}

	if exCode == 0 {
		log.Logger.Infof("Good bye!")
	} else {
		log.Logger.Infof("Good bye! (with errors)")
	}

	os.Exit(exCode)
}

// Main is the main process of Ladder
func Main() int {
	// Parse command line flags
	if err := parse(os.Args[1:]); err != nil {
		log.Logger.Error(err)
		return 1
	}

	verStr := version.Get().String()
	// Show version
	if cfg.showVersion {
		fmt.Println(verStr)
		return 0
	}

	// Set up logger
	lFields := log.Fields{
		"dryrun":  cfg.dryRun,
		"version": verStr,
	}
	log.Setup(os.Stderr, lFields, cfg.jsonLog, cfg.debug)

	// Load debug if required
	if cfg.debug {
		// Register dummy blocks
		gather.Register("dummy", &gather.DummyCreator{})
		arrange.Register("dummy", &arrange.DummyCreator{})
		scale.Register("dummy", &scale.DummyCreator{})
		solve.Register("dummy", &solve.DummyCreator{})
		filter.Register("dummy", &filter.DummyCreator{})
	}

	// Load config file configuration
	c, err := config.LoadConfig(cfg.configFile)
	if err != nil {
		log.Logger.Errorf("Error starting ladder: %s", err)
		return 1
	}

	// Set the flags config on the global config so it passes around the program
	c.Debug = cfg.debug
	c.DryRun = cfg.dryRun

	log.Logger.Infof("Ladder ready to rock!")
	log.Logger.Infof("Starting...")
	if c.DryRun {
		log.Logger.Warningf("Ladder will run in dry run mode!!")
	}

	// check if there are no autoscalers
	if len(c.Autoscalers) == 0 {
		log.Logger.Warningf("No autoscalers present")
		return 0
	}

	// Create an error channel
	errChan := make(chan error)

	// Create all the scalers
	autoscalers := map[string]autoscaler.Autoscaler{}
	for _, ac := range c.Autoscalers {
		if ac.Disabled {
			log.Logger.Warningf("%s autoscaler is disabled, not loading it", ac.Name)
			continue
		}
		// We need to pass the config pointer to copy, if we pass directly
		// then all the autoscalers will point to the same object
		act := ac
		as, err2 := autoscaler.NewIntervalAutoscaler(&act, c.DryRun)
		if err2 != nil {
			log.Logger.Errorf("Error creating autoscaler %s: %v", ac.Name, err2)
			return 1
		}
		autoscalers[as.Name] = as
	}
	log.Logger.Debugf("%d autoscalers created", len(autoscalers))

	// If all autoscalers ok then run them
	for k := range autoscalers {
		// Don't iterate over autoscalers getting them so we don't have to copy it to
		// pass to the goroutine. Note that this is made because loops reuse the same variable
		// in all iterations.
		// Istead we take the key and get from the map directly
		go func(asName string) {
			errChan <- autoscalers[asName].Run()
		}(k)
	}
	log.Logger.Debugf("%d autoscalers running", len(autoscalers))

	// Create our HTTP handler
	api, err := v1.NewAPIV1(c.Global.APIV1Path, autoscalers)
	if err != nil {
		log.Logger.Errorf("Error creating API handler: %v", err)
		return 1
	}
	h, err := web.NewHandler(c, health.MainCheck, api)
	if err != nil {
		log.Logger.Errorf("Error creating http handler: %v", err)
		return 1
	}
	// Serve
	go func() {
		errChan <- h.Serve(cfg.listenAddress)
	}()

	// Wait until signal (ctr+c, SIGTERM...)
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for goroutines
	for {
		select {
		// Wait for errors
		case err := <-errChan:
			if err != nil {
				log.Logger.Error(err)
				return 1
			}
			// Wait for signal
		case <-sigC:
			return 0
		}
	}
}
