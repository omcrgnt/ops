// Package use registers probe actuator and ops HTTP transport defaults in res.Global.
//
// Import for side effects at the app composition root:
//
//	import _ "github.com/omcrgnt/ops/transport/http/use"
package use

import (
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res/unique"

	_ "github.com/omcrgnt/ops/probe/use"
)

func init() {
	unique.MustAddFixed(ophttp.HandlerWire{})
	unique.MustAddReplaceable(ophttp.DefaultServerWire{})
}
