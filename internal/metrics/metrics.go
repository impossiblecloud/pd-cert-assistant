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
	Config   *prometheus.GaugeVec
	AllIPs   *prometheus.GaugeVec
	LocalIPs *prometheus.GaugeVec

	// Counters
	CertUpdateErrors       *prometheus.CounterVec
	PDAssistantFetchErrors *prometheus.CounterVec
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

	am.AllIPs = promauto.With(am.Registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pd_assistant",
			Name:      "all_ips_count",
			Help:      "List of all IP addresses from all PD Assistants",
		},
		[]string{},
	)

	am.LocalIPs = promauto.With(am.Registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pd_assistant",
			Name:      "local_ips_count",
			Help:      "List of IP addresses from Cilium nodes",
		},
		[]string{},
	)

	am.CertUpdateErrors = promauto.With(am.Registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pd_assistant",
			Name:      "cert_update_errors_total",
			Help:      "Total number of certificate update errors",
		},
		[]string{},
	)

	am.PDAssistantFetchErrors = promauto.With(am.Registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pd_assistant",
			Name:      "fetch_errors_total",
			Help:      "Total number of errors fetching data from PD Assistants",
		},
		[]string{"pd_assistant"},
	)

	am.Config.WithLabelValues(version).Set(1)
	am.CertUpdateErrors.WithLabelValues().Add(0)

	am.Registry.MustRegister()
	return am
}
