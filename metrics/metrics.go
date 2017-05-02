package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

var (
	msDurationHistogramBuckets     = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}                                        // ms buckets doubling, from 5ms to 10s
	msLongDurationHistogramBuckets = []float64{100, 250, 500, 1000, 2500, 5000, 10000, 20000, 40000, 80000, 160000, 320000, 640000, 1280000} // ms buckets doubling, from 100ms to 21,3m
)

// Our global metrics,
var (
	// Gathering metrics
	gathererQ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ladder_gatherer_quantity",
		Help: "The quantity got from a ladder gatherer",
	}, []string{"autoscaler", "inputter", "kind"})

	gathererDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_gatherer_duration_histogram_ms",
		Help:    "The latency in ms of a gathering process",
		Buckets: msDurationHistogramBuckets,
	}, []string{"autoscaler", "inputter", "kind"})

	gathererErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_gatherer_errors_total",
		Help: "The total gathering errors",
	}, []string{"autoscaler", "inputter", "kind"})

	// Inputter  metrics
	inputterQ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ladder_inputter_quantity",
		Help: "The quantity got from a ladder inputter",
	}, []string{"autoscaler", "inputter"})

	inputterDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_inputter_duration_histogram_ms",
		Help:    "The latency in ms of an inputter process, this is gather + arrange)",
		Buckets: msDurationHistogramBuckets,
	}, []string{"autoscaler", "inputter"})

	inputterErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_inputter_errors_total",
		Help: "The total inputter errors",
	}, []string{"autoscaler", "inputter"})

	// Solver metrics
	solverQ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ladder_solver_quantity",
		Help: "The quantity got from a ladder solver",
	}, []string{"autoscaler", "kind"})

	solverDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_solver_duration_histogram_ms",
		Help:    "The latency in ms of a solver process",
		Buckets: msDurationHistogramBuckets,
	}, []string{"autoscaler", "kind"})

	solverErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_solver_errors_total",
		Help: "The total solver errors",
	}, []string{"autoscaler", "kind"})

	// Filterer metrics
	filtererDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_filterer_duration_histogram_ms",
		Help:    "The latency in ms of a scaling current filtering process",
		Buckets: msDurationHistogramBuckets,
	}, []string{"autoscaler"})

	filtererErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_filterer_errors_total",
		Help: "The total filterer errors",
	}, []string{"autoscaler"})

	// Scaler metrics
	currentQ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ladder_scaler_current_quantity",
		Help: "The quantity currently from the scaler target",
	}, []string{"autoscaler", "kind"})

	currentDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_scaler_current_duration_histogram_ms",
		Help:    "The latency in ms of a scaling current gathering process",
		Buckets: msDurationHistogramBuckets,
	}, []string{"autoscaler", "kind"})

	currentErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_scaler_current_errors_total",
		Help: "The total curret errors",
	}, []string{"autoscaler", "kind"})

	scalerQ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ladder_scaler_quantity",
		Help: "The quantity scaled from a ladder scaler",
	}, []string{"autoscaler", "kind"})

	scalerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_scaler_duration_histogram_ms",
		Help:    "The latency in ms of a scaling process",
		Buckets: msDurationHistogramBuckets,
	}, []string{"autoscaler", "kind"})

	scalerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_scaler_errors_total",
		Help: "The total scaler errors",
	}, []string{"autoscaler", "kind"})

	// Autoscaler metrics
	autoScalerIteration = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_autoscaler_iterations_total",
		Help: "The total autoscaler iterations total",
	}, []string{"autoscaler"})

	autoScalerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ladder_autoscaler_errors_total",
		Help: "The total autoscaler errors",
	}, []string{"autoscaler"})

	autoscalerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ladder_autoscaler_duration_histogram_ms",
		Help:    "The latency in ms of an autoscaling whole iteration process",
		Buckets: msLongDurationHistogramBuckets,
	}, []string{"autoscaler"})

	autoscalerRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ladder_autoscaler_running",
		Help: "The state of the autoscaler in running state boolean",
	}, []string{"autoscaler"})
)

// Register all the metrics
func init() {
	// Gatherer
	prometheus.MustRegister(gathererQ)
	prometheus.MustRegister(gathererDuration)
	prometheus.MustRegister(gathererErrors)
	// Inputter
	prometheus.MustRegister(inputterQ)
	prometheus.MustRegister(inputterDuration)
	prometheus.MustRegister(inputterErrors)
	// Solver
	prometheus.MustRegister(solverQ)
	prometheus.MustRegister(solverDuration)
	prometheus.MustRegister(solverErrors)
	// scaler
	prometheus.MustRegister(currentQ)
	prometheus.MustRegister(currentErrors)
	prometheus.MustRegister(currentDuration)
	prometheus.MustRegister(scalerQ)
	prometheus.MustRegister(scalerErrors)
	prometheus.MustRegister(scalerDuration)
	// Filterer
	prometheus.MustRegister(filtererDuration)
	prometheus.MustRegister(filtererErrors)
	//autoscaler
	prometheus.MustRegister(autoScalerIteration)
	prometheus.MustRegister(autoScalerErrors)
	prometheus.MustRegister(autoscalerDuration)
	prometheus.MustRegister(autoscalerRunning)

	log.Logger.Infof("Registered metrics on prometheus")
}

// SetGathererQ sets the current value of the gatherer input as a gauge
func SetGathererQ(q types.Quantity, autoscalerName, inputterName, gathererKind string) {
	gathererQ.WithLabelValues(autoscalerName, inputterName, gathererKind).Set(float64(q.Q))
}

// ObserveGathererDuration observes the duration of a gathering process
func ObserveGathererDuration(duration time.Duration, autoscalerName, inputterName, gathererKind string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	gathererDuration.WithLabelValues(autoscalerName, inputterName, gathererKind).Observe(float64(ms))
}

// AddGathererErrors adds a number of errors to the counter
func AddGathererErrors(numberErrors int, autoscalerName, inputterName, gathererKind string) {
	gathererErrors.WithLabelValues(autoscalerName, inputterName, gathererKind).Add(float64(numberErrors))
}

// SetInputterQ sets the current value of the inputter input as a gauge
func SetInputterQ(q types.Quantity, autoscalerName, inputterName string) {
	inputterQ.WithLabelValues(autoscalerName, inputterName).Set(float64(q.Q))
}

// ObserveInputterDuration observes the duration of a gathering process
func ObserveInputterDuration(duration time.Duration, autoscalerName, inputterName string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	inputterDuration.WithLabelValues(autoscalerName, inputterName).Observe(float64(ms))
}

// AddInputterErrors adds a number of errors to the counter
func AddInputterErrors(numberErrors int, autoscalerName, inputterName string) {
	inputterErrors.WithLabelValues(autoscalerName, inputterName).Add(float64(numberErrors))
}

// SetSolverQ sets the current value of the solver output as a gauge
func SetSolverQ(q types.Quantity, autoscalerName, solverKind string) {
	solverQ.WithLabelValues(autoscalerName, solverKind).Set(float64(q.Q))
}

// ObserveSolverDuration observes the duration of a solver process
func ObserveSolverDuration(duration time.Duration, autoscalerName, solverKind string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	solverDuration.WithLabelValues(autoscalerName, solverKind).Observe(float64(ms))
}

// AddSolverErrors adds a number of errors to the counter
func AddSolverErrors(numberErrors int, autoscalerName, solverKind string) {
	solverErrors.WithLabelValues(autoscalerName, solverKind).Add(float64(numberErrors))
}

// ObserveFiltererDuration observes the duration of a filterer process
func ObserveFiltererDuration(duration time.Duration, autoscalerName string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	filtererDuration.WithLabelValues(autoscalerName).Observe(float64(ms))
}

// AddFiltererErrors adds a number of errors to the counter
func AddFiltererErrors(numberErrors int, autoscalerName string) {
	filtererErrors.WithLabelValues(autoscalerName).Add(float64(numberErrors))
}

// SetCurrentQ sets the current value of the current output as a gauge
func SetCurrentQ(q types.Quantity, autoscalerName, scalerKind string) {
	currentQ.WithLabelValues(autoscalerName, scalerKind).Set(float64(q.Q))
}

// ObserveCurrentDuration observes the duration of a current process
func ObserveCurrentDuration(duration time.Duration, autoscalerName, scalerKind string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	currentDuration.WithLabelValues(autoscalerName, scalerKind).Observe(float64(ms))
}

// AddCurrentErrors adds a number of errors to the counter
func AddCurrentErrors(numberErrors int, autoscalerName, scalerKind string) {
	currentErrors.WithLabelValues(autoscalerName, scalerKind).Add(float64(numberErrors))
}

// SetScalerQ sets the current value of the scaler output as a gauge
func SetScalerQ(q types.Quantity, autoscalerName, scalerKind string) {
	scalerQ.WithLabelValues(autoscalerName, scalerKind).Set(float64(q.Q))
}

// ObserveScalerDuration observes the duration of a scaling process
func ObserveScalerDuration(duration time.Duration, autoscalerName, scalerKind string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	scalerDuration.WithLabelValues(autoscalerName, scalerKind).Observe(float64(ms))
}

// AddScalerErrors adds a number of errors to the counter
func AddScalerErrors(numberErrors int, autoscalerName, scalerKind string) {
	scalerErrors.WithLabelValues(autoscalerName, scalerKind).Add(float64(numberErrors))
}

// AddAutoscalerIteration adds a number of iterations to the counter
func AddAutoscalerIteration(numberIterations int, autoscalerName string) {
	autoScalerIteration.WithLabelValues(autoscalerName).Add(float64(numberIterations))
}

// ObserveAutoscalerDuration observes the duration of a autoscaling iteration process
func ObserveAutoscalerDuration(duration time.Duration, autoscalerName string) {
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	autoscalerDuration.WithLabelValues(autoscalerName).Observe(float64(ms))
}

// AddAutoscalerErrors adds a number of errors to the counter
func AddAutoscalerErrors(numberErrors int, autoscalerName string) {
	autoScalerErrors.WithLabelValues(autoscalerName).Add(float64(numberErrors))
}

// SetAutoscalerRunning sets the state of the autoscaler running
func SetAutoscalerRunning(running bool, autoscalerName string) {
	var state float64
	if running {
		state = 1
	}
	autoscalerRunning.WithLabelValues(autoscalerName).Set(state)
}
