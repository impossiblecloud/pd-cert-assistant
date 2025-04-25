package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AppMetrics contains pointers to prometheus metrics objects
type AppMetrics struct {
	Listen   string
	Registry *prometheus.Registry

	// Gauges
	Config *prometheus.GaugeVec

	// Counters
	CertUpdateErrors *prometheus.CounterVec
}

func InitMetrics(version string) AppMetrics {

	am := AppMetrics{}
	am.Registry = prometheus.NewRegistry()

	// Config info
	am.Config = promauto.With(am.Registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pd_assistant",
			Name:      "config",
			Help:      "App config info",
		},
		[]string{"version"},
	)

	am.CertUpdateErrors = promauto.With(am.Registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pd_assistant",
			Name:      "cert_update_errors_total",
			Help:      "Total number of certificate update errors",
		},
		[]string{},
	)

	am.Config.WithLabelValues(version).Set(1)
	am.CertUpdateErrors.WithLabelValues().Add(0)

	am.Registry.MustRegister()
	return am
}
