// Package use registers probe actuator and ops HTTP transport defaults in res.Default.
//
// Import for side effects at the app composition root:
//
//	import _ "github.com/omcrgnt/ops/transport/http/use"
package use

import (
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res"

	_ "github.com/omcrgnt/ops/probe/use"
)

func init() {
	_ = res.AddWithTags(&ophttp.Handler{}, res.TagFixed)
	_ = res.AddWithTags(ophttp.DefaultServer(), res.TagReplaceable)
}
