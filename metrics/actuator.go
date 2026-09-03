package metrics

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// Actuator aggregates metrics contributors and exposes scrape via Metricer.
type Actuator struct {
	reg *prometheus.Registry
}

var _ Metricer = (*Actuator)(nil)

func (a *Actuator) Deps() []any {
	return []any{
		(*prometheus.Registry)(nil),
		([]MetricsContributor)(nil),
	}
}

func (a *Actuator) Inject(args []any) {
	// Two passes, not one: sdi builds args in Deps()'s declared order
	// (*prometheus.Registry, then []MetricsContributor) and preserves it
	// into this call, but nothing enforces that order stays that way if
	// Deps() ever gets reordered — collecting both values first, then
	// acting, means Inject no longer depends on which one args puts first.
	var contributors []MetricsContributor
	for _, arg := range args {
		switch v := arg.(type) {
		case *prometheus.Registry:
			a.reg = v
		case []MetricsContributor:
			contributors = v
		}
	}

	if a.reg == nil {
		panic("metrics: Actuator.Inject: no *prometheus.Registry resolved — Deps() declares it required")
	}
	for _, c := range contributors {
		if err := c.RegisterMetrics(a.reg); err != nil {
			panic(fmt.Sprintf("metrics: RegisterMetrics: %v", err))
		}
	}
}

func (a *Actuator) MetricsMetrics(_ context.Context) ([]byte, error) {
	if a.reg == nil {
		return nil, errors.New("metrics: registry not configured")
	}
	mfs, err := a.reg.Gather()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
