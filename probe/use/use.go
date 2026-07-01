// Package use registers the probe actuator in unique.Global.
//
// Import for side effects:
//
//	import _ "github.com/omcrgnt/ops/probe/use"
package use

import (
	"github.com/omcrgnt/ops/probe"
	"github.com/omcrgnt/res/unique"
)

func init() {
	unique.MustAddReplaceable(&probe.Actuator{})
}
