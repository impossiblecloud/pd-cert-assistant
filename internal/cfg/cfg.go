package cfg

import (
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
)

// AppConfig is the main configuration structure for the application.
type AppConfig struct {
	// Metrics contains the application's metrics.
	Metrics metrics.AppMetrics
	//PDAddress is the address of the PD server.
	PDAddress string

	// TLS Parameters
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
	TLSInsecure bool

	// HTTPRequestTimeout is the timeout for HTTP requests in seconds.
	HTTPRequestTimeout int
}
