package probe

import "context"

// Prober is the stable ops surface API for probe aggregation.
type Prober interface {
	ProbeLive(ctx context.Context) error
	ProbeReady(ctx context.Context) error
	ProbeHealth(ctx context.Context) error
}
