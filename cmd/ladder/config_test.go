package main

import (
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		flags []string
		valid bool
	}{
		{flags: []string{"--dry.run"}, valid: true},
		{flags: []string{"-dry.run"}, valid: true},
		{flags: []string{"--dry.run", "--debug", "-json.log"}, valid: true},
		{flags: []string{"--dry.run", "--version", "-json.log"}, valid: true},
		{flags: []string{"--dry.run", "-debug"}, valid: true},
		{flags: []string{"--dry.run", "-version"}, valid: true},
		{flags: []string{"--listen.address", ":8888"}, valid: true},
		{flags: []string{"--json.log"}, valid: true},
		{flags: []string{"--listen.address"}, valid: false},
		{flags: []string{"--dry.run", "-debug", "-config.file", "something.yml"}, valid: true},
		{flags: []string{"-debug", "--config.file", "something.yml"}, valid: true},
		{flags: []string{"-wrong", "--config.file", "something.yml"}, valid: false},
		{flags: []string{"--config.file"}, valid: false},
		{flags: []string{"-dryrun"}, valid: false},
		{flags: []string{"--jsonlog"}, valid: false},
	}

	for _, test := range tests {
		err := parse(test.flags)
		if err != nil && test.valid {
			t.Errorf("\n- %+v\n  Parse shouldn't raise an error, it did: %v", test, err)
		} else if err == nil && !test.valid {
			t.Errorf("\n- %+v\n  Parse should raise an error, it didn't", test)
		}
	}
}
