// Package use registers the probe actuator singleton in res.Global.
//
// Import for side effects:
//
//	import _ "github.com/omcrgnt/ops/probe/use"
package use

import (
	"github.com/omcrgnt/ops/probe"
	"github.com/omcrgnt/res"
)

func init() {
	res.MustAddToGlobalWithTags(probe.ActuatorWire{}, res.TagReplaceable)
}
