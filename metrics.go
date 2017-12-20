package forward

import (
	"sync"

	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
)

// Variables declared for monitoring.
var (
	RequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "forward",
		Name:      "request_count_total",
		Help:      "Counter of requests made per protocol, family and upstream.",
	}, []string{"proto", "family", "to"})
	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "forward",
		Name:      "request_duration_seconds",
		Buckets:   plugin.TimeBuckets,
		Help:      "Histogram of the time each request took.",
	}, []string{"proto", "family", "to"})
	HealthcheckFailureCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "forward",
		Name:      "healthcheck_failure_count_total",
		Help:      "Counter of the number of failed healtchecks.",
	}, []string{"to"})
	SocketGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "forward",
		Name:      "socket_count_total",
		Help:      "Guage of open sockets per upstream.",
	}, []string{"to"})
)

// OnStartupMetrics sets up the metrics on startup.
func OnStartupMetrics() error {
	metricsOnce.Do(func() {
		prometheus.MustRegister(RequestCount)
		prometheus.MustRegister(RequestDuration)
		prometheus.MustRegister(HealthcheckFailureCount)
		prometheus.MustRegister(SocketGauge)
	})
	return nil
}

// familyToString returns the string form of either 1, or 2. Returns
// empty string is not a known family
func familyToString(f int) string {
	if f == 1 {
		return "1"
	}
	if f == 2 {
		return "2"
	}
	return ""
}

var metricsOnce sync.Once
