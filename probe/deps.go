package probe

import "context"

// ProbeReadiness is implemented by shared infra resources checked for traffic readiness
// (e.g. one db pool per app — not per repository). Must respect ctx cancellation.
// v1 calls implementors sequentially (see Actuator.ProbeReady).
type ProbeReadiness interface {
	ProbeReady(ctx context.Context) error
}

// ProbeLiveness is implemented by resources checked for process liveness.
// Must be cheap and respect ctx cancellation. v1 calls implementors sequentially.
type ProbeLiveness interface {
	ProbeLive(ctx context.Context) error
}
