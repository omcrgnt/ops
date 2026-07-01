package metrics

import "context"

// Metricer is the stable ops surface API for Prometheus scrape.
type Metricer interface {
	MetricsMetrics(ctx context.Context) ([]byte, error)
}
