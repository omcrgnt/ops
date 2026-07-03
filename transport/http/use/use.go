// Package use registers probe actuator, metrics platform, and ops HTTP transport defaults in unique.Global.
//
// Import for side effects at the app composition root:
//
//	import _ "github.com/omcrgnt/ops/transport/http/use"
//
// Pulls in probe actuator and shared Prometheus registry + metrics actuator transitively.
// HTTP request metrics: import github.com/omcrgnt/srv-http for Server types (HTTPMetrics registers via init).
package use

import (
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res/unique"

	_ "github.com/omcrgnt/ops/metrics/use"
	_ "github.com/omcrgnt/ops/probe/use"
)

func init() {
	unique.MustAddFixed(&ophttp.Handler{})
	unique.MustAddReplaceable(ophttp.DefaultServer())
}
