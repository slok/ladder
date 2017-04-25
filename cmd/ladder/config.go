package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	// defaults
	defaultDebug         = false
	defaultDryRun        = false
	defaultConfigFile    = "ladder.yml"
	defaultListenAddress = ":9094"
	defaultJSONLog       = false
	defaultShowVersion   = false
)

var cfg = struct {
	fs *flag.FlagSet

	debug         bool
	dryRun        bool
	configFile    string
	listenAddress string
	jsonLog       bool
	showVersion   bool
}{}

// init will load all the flags
func init() {
	cfg.fs = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg.fs.StringVar(
		&cfg.configFile, "config.file", defaultConfigFile,
		"Ladder configuration file name.",
	)

	cfg.fs.BoolVar(
		&cfg.debug, "debug", defaultDebug,
		"Run Ladder in debug mode",
	)

	cfg.fs.BoolVar(
		&cfg.dryRun, "dry.run", defaultDryRun,
		"Run Ladder in dry run mode",
	)

	cfg.fs.BoolVar(
		&cfg.jsonLog, "json.log", defaultJSONLog,
		"Log messages in json format",
	)

	cfg.fs.BoolVar(
		&cfg.showVersion, "version", defaultShowVersion,
		"Show version",
	)

	cfg.fs.StringVar(
		&cfg.listenAddress, "listen.address", defaultListenAddress,
		"Address to listen on for the web interface",
	)
}

func parse(args []string) error {
	err := cfg.fs.Parse(args)
	if err != nil {
		return err
	}

	if len(cfg.fs.Args()) != 0 {
		err = fmt.Errorf("Invalid command line arguments. Help: %s -h", os.Args[0])
	}

	return err
}
