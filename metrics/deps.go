package metrics

import "github.com/prometheus/client_golang/prometheus"

// MetricsContributor registers Prometheus collectors into the shared registry.
// Domain and transport packages (e.g. srv-http HTTPMetrics) implement this port.
type MetricsContributor interface {
	RegisterMetrics(reg *prometheus.Registry) error
}
