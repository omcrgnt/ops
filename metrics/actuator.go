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
	for _, arg := range args {
		switch v := arg.(type) {
		case *prometheus.Registry:
			a.reg = v
		case []MetricsContributor:
			if a.reg == nil {
				continue
			}
			for _, c := range v {
				if err := c.RegisterMetrics(a.reg); err != nil {
					panic(fmt.Sprintf("metrics: RegisterMetrics: %v", err))
				}
			}
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
