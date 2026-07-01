package metrics_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/res/unique"
	"github.com/omcrgnt/sdi"
	"github.com/prometheus/client_golang/prometheus"
)

type stubContributor struct{}

func (stubContributor) RegisterMetrics(reg *prometheus.Registry) error {
	return reg.Register(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter_total",
		Help: "test",
	}))
}

func TestActuator_SDIResolve(t *testing.T) {
	reg := unique.New()
	reg.MustAddFixed(prometheus.NewRegistry())
	reg.MustAddReplaceable(&metrics.Actuator{})
	if err := reg.Add(stubContributor{}); err != nil {
		t.Fatal(err)
	}

	if err := sdi.Resolve(reg); err != nil {
		t.Fatal(err)
	}

	actAny, err := reg.GetOneByType(reflect.TypeOf((*metrics.Actuator)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	act := actAny.(*metrics.Actuator)

	body, err := act.MetricsMetrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "test_counter_total") {
		t.Fatalf("body %q missing test_counter_total", body)
	}
}

func TestActuator_MetricsMetrics_noRegistry(t *testing.T) {
	var act metrics.Actuator
	if _, err := act.MetricsMetrics(context.Background()); err == nil {
		t.Fatal("expected error without registry")
	}
}
