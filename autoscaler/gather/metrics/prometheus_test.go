package metrics

import (
	"errors"
	"math"
	"testing"
	"time"

	// TODO: Deprecated in 1.7, clean when updating github.com/prometheus/client_golang
	"golang.org/x/net/context"

	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"

	"github.com/themotion/ladder/log"
)

func TestPrometheusMetricCorrectCreation(t *testing.T) {
	tests := []struct {
		addresses []interface{}
		query     string

		wantError bool
	}{
		{[]interface{}{"http://10.0.2.235:9090", "http://10.0.2.236:9090"}, `rate(container_cpu_user_seconds_total{image=~".*prometheus.*"}[1m])`, false},
		{[]interface{}{"http://10.0.2.235:9090"}, "", true},
		{[]interface{}{}, `rate(container_cpu_user_seconds_total{image=~".*prometheus.*"}[1m])`, true},
		{[]interface{}{}, "", true},
	}

	for _, test := range tests {
		opts := map[string]interface{}{
			pmAddresses: test.addresses,
			pmQuery:     test.query,
		}
		p, err := NewPrometheusMetric(context.TODO(), opts)

		if test.wantError && err == nil {
			t.Fatalf("\n- %+v\n  Creation should give error, it dind't", test)
		}

		if !test.wantError {
			if err != nil {
				t.Fatalf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
			}

			if len(p.addresses) != len(test.addresses) {
				t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, test.addresses, p.addresses)
			}

			// Check each address
			for i, a := range test.addresses {
				if p.addresses[i] != a {
					t.Errorf("\n- %+v\n  Wrong parameters loaded on object, want: %v; got %v", test, a, p.addresses[i])
				}
			}

			if len(p.apiCs) != len(test.addresses) {
				t.Errorf("\n- %+v\n  Api Client should have the same length as the addresses; want: %d; got %d", test, len(test.addresses), len(p.apiCs))
			}
		}
	}
}

type queryAPITestClient struct {
	value     model.Value
	wantError bool
}

func (q *queryAPITestClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	var err error
	if q.wantError {
		err = errors.New("Wrong!")
	}
	return q.value, err
}

func (q *queryAPITestClient) QueryRange(ctx context.Context, query string, r prometheus.Range) (model.Value, error) {
	return nil, nil
}

func TestPrometheusMetricGather(t *testing.T) {
	tests := []struct {
		v        model.Value
		apiError bool

		wantValue int64
		wantError bool
	}{
		// Vectors with one result are good
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     14.98,
				},
			},
			apiError:  false,
			wantValue: 15,
			wantError: false,
		},
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     14.51,
				},
			},
			apiError:  false,
			wantValue: 15,
			wantError: false,
		},
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     14.50,
				},
			},
			apiError:  false,
			wantValue: 15,
			wantError: false,
		},
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     14.49,
				},
			},
			apiError:  false,
			wantValue: 14,
			wantError: false,
		},
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     50,
				},
			},
			apiError:  false,
			wantValue: 50,
			wantError: false,
		},
		// Vectors with multiple results are bad
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     14.98,
				},
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     15.98,
				},
			},
			apiError:  false,
			wantValue: 0,
			wantError: true,
		},
		// Vectors with no values are bad
		{
			v:         model.Vector{},
			apiError:  false,
			wantValue: 0,
			wantError: true,
		},
		// Matrix type metrics are bad
		{
			v:         model.Matrix{},
			apiError:  false,
			wantValue: 0,
			wantError: true,
		},
		// api errors are bad
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     14.98,
				},
			},
			apiError:  true,
			wantValue: 0,
			wantError: true,
		},
		// Error when no metric is present
		{
			v: model.Vector{
				&model.Sample{
					Metric:    model.Metric{},
					Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
					Value:     model.SampleValue(math.NaN()),
				},
			},
			apiError:  false,
			wantValue: 0,
			wantError: true,
		},
	}

	for _, test := range tests {
		p := &PrometheusMetric{
			addresses: []string{"http://example.test.org"},
			qry:       `test_metric`,
			log:       log.New(),
		}

		p.apiCs = make([]prometheus.QueryAPI, 1)
		p.apiCs[0] = &queryAPITestClient{
			value:     test.v,
			wantError: test.apiError,
		}

		q, err := p.Gather(context.TODO())

		if test.wantError && err == nil {
			t.Fatalf("\n- %+v\n  Gather should give error, it dind't", test)
		}

		if !test.wantError {
			if err != nil {
				t.Fatalf("\n- %+v\n  Creation shouldn't give error: %v", test, err)
			}

			if q.Q != test.wantValue {
				t.Errorf("\n- %+v\n  Wrong gathering retrieved value, want: %v; got %v", test, test.wantValue, q.Q)
			}
		}
	}
}

func TestPrometheusMetricGatherRetries(t *testing.T) {
	tests := []struct {
		endpointShouldError []bool
		wantError           bool
	}{
		// One correct endpoint
		{endpointShouldError: []bool{false}, wantError: false},
		// One wrong endpoint
		{endpointShouldError: []bool{true}, wantError: true},

		// Two endpoints, one at least correct
		{endpointShouldError: []bool{true, false}, wantError: false},
		{endpointShouldError: []bool{false, true}, wantError: false},

		// multiple endpoints
		{endpointShouldError: []bool{false, false, false, false}, wantError: false},
		{endpointShouldError: []bool{false, true, false, false}, wantError: false},
		{endpointShouldError: []bool{false, true, false, true}, wantError: false},
		{endpointShouldError: []bool{true, true, false, true}, wantError: false},
		{endpointShouldError: []bool{true, true, true, true}, wantError: true},
	}

	for _, test := range tests {
		p := &PrometheusMetric{
			qry: `test_metric`,
			log: log.New(),
		}

		p.apiCs = make([]prometheus.QueryAPI, len(test.endpointShouldError))
		v := model.Vector{
			&model.Sample{
				Metric:    model.Metric{},
				Timestamp: model.Time(time.Now().UTC().Nanosecond() / 1000000),
				Value:     14.98,
			},
		}

		// Create all endpoints
		p.apiCs = make([]prometheus.QueryAPI, len(test.endpointShouldError))
		for i, e := range test.endpointShouldError {
			p.apiCs[i] = &queryAPITestClient{
				value:     v,
				wantError: e,
			}
		}

		_, err := p.Gather(context.TODO())

		if test.wantError && err == nil {
			t.Fatalf("\n- %+v\n  Gather should give error, it dind't", test)
		}
	}

}
