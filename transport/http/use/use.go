// Package use registers probe actuator and ops HTTP transport defaults in unique.Global.
//
// Import for side effects at the app composition root:
//
//	import _ "github.com/omcrgnt/ops/transport/http/use"
//
// Also import _ "github.com/omcrgnt/ops/metrics/use" and _ "github.com/omcrgnt/srv-http/use" for metrics.
package use

import (
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res/unique"

	_ "github.com/omcrgnt/ops/probe/use"
)

func init() {
	unique.MustAddFixed(&ophttp.Handler{})
	unique.MustAddReplaceable(ophttp.DefaultServer())
}
