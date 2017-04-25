package aws

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/themotion/ladder/log"
	awsMock "github.com/themotion/ladder/mock/aws"
	"github.com/themotion/ladder/mock/aws/sdk"
	"github.com/themotion/ladder/types"
)

func TestECSRunningTasksCorrectCreation(t *testing.T) {
	tests := []struct {
		region        string
		cluster       string
		maxNotRunning int64
		maxChecks     int64
		errorOnMax    bool
		when          string

		correct bool
	}{
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "always", maxNotRunning: 5, maxChecks: 6, errorOnMax: true},
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "always", maxNotRunning: 5, maxChecks: 6, errorOnMax: false},
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "always", maxNotRunning: 10, maxChecks: 0, errorOnMax: true},
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "always", maxNotRunning: 0, maxChecks: 0, errorOnMax: true},
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "scale_up", maxNotRunning: 0, maxChecks: 0, errorOnMax: true},
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "scale_down", maxNotRunning: 0, maxChecks: 0, errorOnMax: true},
		{correct: true, region: "us-west-2", cluster: "test_cluster", when: "always"},
		{correct: false, cluster: "test_cluster", when: "always", maxNotRunning: 5, maxChecks: 6, errorOnMax: true},
		{correct: false, region: "us-west-2", when: "always", maxNotRunning: 5, maxChecks: 6, errorOnMax: true},
		{correct: false, region: "us-west-2", cluster: "test_cluster", when: "wrong", maxNotRunning: 0, maxChecks: 0, errorOnMax: true},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			ecsAwsRegionOpt:           test.region,
			ecsClusterNameOpt:         test.cluster,
			maxPendingTasksAllowedOpt: test.maxNotRunning,
			maxChecksOpt:              test.maxChecks,
			errorOnMaxCheckOpt:        test.errorOnMax,
			whenOpt:                   test.when,
		}

		e, err := NewECSRunningTasks(context.TODO(), opts)

		if test.correct {
			// Validations for a wanted correct creation
			if err != nil {
				t.Errorf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
			}

			if e.clusterName != test.cluster {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.cluster, e.clusterName)
			}

			if e.maxNotRunningTasks != test.maxNotRunning {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %d; got %d", test, test.maxNotRunning, e.maxNotRunningTasks)
			}

			if e.maxChecks != test.maxChecks {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %d; got %d", test, test.maxChecks, e.maxChecks)
			}

			if e.maxChecks != test.maxChecks {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %d; got %d", test, test.maxChecks, e.maxChecks)
			}

			if e.errorOnMaxCheck != test.errorOnMax {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %t; got %t", test, test.errorOnMax, e.errorOnMaxCheck)
			}
		} else {
			// Validations for a wanted incorrect creation
			if err == nil {
				t.Errorf("\n- %+v\n  Creation should give error, it didn't", test)
			}
		}
	}
}

func TestECSRunningTaskFilter(t *testing.T) {
	tests := []struct {
		currentQ      int64
		newQ          int64
		maxNotRunning int64
		maxChecks     int64
		errorOnMax    bool
		currentChecks int
		when          when

		apiReturnError     bool
		apiReturnNClusters int
		apiReturnedPending int

		wantCurrentChecks int
		wantError         bool
		wantBreak         bool
		wantQ             int64
	}{
		//  Error from the API
		{
			when:           always,
			apiReturnError: true,
			wantError:      true,
		},
		//  Multiple clusters from the API
		{
			when:               always,
			apiReturnNClusters: 2,
			wantError:          true,
		},
		//  We are good, no pending tasks
		{

			currentQ: 5, newQ: 10, when: always,
			apiReturnNClusters: 1,
			wantQ:              10,
		},
		//  We are good, no pending tasks using exceed
		{
			currentQ: 5, newQ: 10, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 10,
			maxNotRunning: 10,
			wantQ:         10,
		},
		{
			currentQ: 5, newQ: 10, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 10,
			maxNotRunning: 11,
			wantQ:         10,
		},
		//  Bad, wanted all tasks running
		{
			currentQ: 5, newQ: 10, currentChecks: 1, maxChecks: 0, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 1,
			wantBreak: true, wantQ: 5, wantCurrentChecks: 2,
		},
		//  Bad, wanted all tasks running, but... we are scaling up and only applies on scale down
		{
			currentQ: 5, newQ: 10, currentChecks: 1, maxChecks: 0, when: scaleDown,
			apiReturnNClusters: 1, apiReturnedPending: 1,
			wantBreak: false, wantQ: 10, wantCurrentChecks: 0,
		},
		//  Bad, wanted all tasks running, but... we are scaling down and only applies on scale up
		{
			currentQ: 5, newQ: 2, currentChecks: 1, maxChecks: 0, when: scaleUp,
			apiReturnNClusters: 1, apiReturnedPending: 1,
			wantBreak: false, wantQ: 2, wantCurrentChecks: 0,
		},
		// Bad! we have pending tasks with maxChecks deactivated
		{
			currentQ: 5, newQ: 10, currentChecks: 1, maxChecks: 0, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 10,
			maxNotRunning: 9,
			wantBreak:     true, wantQ: 5, wantCurrentChecks: 2,
		},
		// Bad! we have pending tasks with maxChecks activated and not exceed this checks
		{
			currentQ: 5, newQ: 10, currentChecks: 3, maxChecks: 5, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 10,
			maxNotRunning: 9,
			wantBreak:     true, wantQ: 5, wantCurrentChecks: 4,
		},
		// Bad! we have pending tasks with maxChecks activated and exceed this checks (no error activated)
		{
			currentQ: 5, newQ: 10, currentChecks: 5, maxChecks: 5, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 10,
			maxNotRunning: 9,
			wantBreak:     false, wantQ: 10, wantCurrentChecks: 0,
		},
		// Bad! we have pending tasks with maxChecks activated and exceed this checks (error activated)
		{
			currentQ: 5, newQ: 10, currentChecks: 5, maxChecks: 5, errorOnMax: true, when: always,
			apiReturnNClusters: 1, apiReturnedPending: 10,
			maxNotRunning:     9,
			wantCurrentChecks: 0, wantError: true,
		},
	}

	for _, test := range tests {

		// Mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSDescribeClusters(t, mockECS, test.apiReturnNClusters, 1, int64(test.apiReturnedPending), test.apiReturnError)

		e := &ECSRunningTasks{
			maxNotRunningTasks: test.maxNotRunning,
			maxChecks:          test.maxChecks,
			errorOnMaxCheck:    test.errorOnMax,
			currentChecks:      test.currentChecks,
			client:             mockECS,
			when:               test.when,
			log:                log.New(),
		}

		q, b, err := e.Filter(context.TODO(), types.Quantity{Q: test.currentQ}, types.Quantity{Q: test.newQ})

		if test.wantError {
			if err == nil {
				t.Errorf("\n- %+v\n  Filtering should give error, it didn't", test)
			}
		} else {
			if err != nil {
				t.Errorf("\n- %+v\n  Filtering shouldnt give error, it did: %s", test, err)
			}
			if b != test.wantBreak {
				t.Errorf("\n- %+v\n  Wrong break result, want: %t", test, test.wantBreak)
			}
			if q.Q != test.wantQ {
				t.Errorf("\n- %+v\n  Wrong quantity result, want: %d; got: %d", test, test.wantQ, q.Q)
			}
		}
		if e.currentChecks != test.wantCurrentChecks {
			t.Errorf("\n- %+v\n  Wrong internal check counter state, want: %d; got: %d", test,
				test.wantCurrentChecks, e.currentChecks)
		}
	}
}
