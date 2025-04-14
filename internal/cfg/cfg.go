package cfg

import (
	"github.com/impossiblecloud/pd-cert-assistant/internal/metrics"
)

// Config is the main app config struct
type AppConfig struct {
	Metrics metrics.AppMetrics
}
