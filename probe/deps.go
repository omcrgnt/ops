package probe

import "context"

// ProbeReadiness is implemented by shared infra resources checked for traffic readiness
// (e.g. one db pool per app — not per repository). Must respect ctx cancellation.
// If parallel aggregation is enabled in the future, ProbeReady may run concurrently
// with other implementors; document whether the receiver is concurrency-safe.
type ProbeReadiness interface {
	ProbeReady(ctx context.Context) error
}

// ProbeLiveness is implemented by resources checked for process liveness.
// Must be cheap and respect ctx cancellation. v1 calls implementors sequentially.
type ProbeLiveness interface {
	ProbeLive(ctx context.Context) error
}

// ProbeHealth is implemented by shared infra resources for operational health checks.
// Same concurrency and ctx rules as ProbeReadiness apply.
type ProbeHealth interface {
	ProbeHealth(ctx context.Context) error
}
