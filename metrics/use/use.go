// Package use registers the shared Prometheus registry and metrics actuator in unique.Global.
//
// Import for side effects:
//
//	import _ "github.com/omcrgnt/ops/metrics/use"
package use

import (
	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/res/unique"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	unique.MustAddFixed(prometheus.NewRegistry())
	unique.MustAddReplaceable(&metrics.Actuator{})
}
