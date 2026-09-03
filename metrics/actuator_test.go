package metrics_test

import (
	"context"
	"errors"
	"fmt"
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

type failingContributor struct{ err error }

func (f failingContributor) RegisterMetrics(*prometheus.Registry) error { return f.err }

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

// TestActuator_Inject_panicsOnRegisterMetricsError was previously the one
// truly dangerous path in this file with zero coverage: a contributor's
// RegisterMetrics failing is the only failure Inject reports at all (fail
// fast via panic, by design — sdi has no other way to abort Resolve from
// inside Inject), and nothing exercised it.
func TestActuator_Inject_panicsOnRegisterMetricsError(t *testing.T) {
	wantErr := errors.New("boom")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic, got none")
		}
		if !strings.Contains(fmt.Sprint(r), wantErr.Error()) {
			t.Fatalf("panic value = %v, want it to contain %v", r, wantErr)
		}
	}()

	a := &metrics.Actuator{}
	a.Inject([]any{prometheus.NewRegistry(), []metrics.MetricsContributor{failingContributor{err: wantErr}}})
}

// TestActuator_Inject_panicsWithoutRegistry: Deps declares
// *prometheus.Registry as a required single dependency, so sdi.Resolve
// itself should never let Inject run without one — but Inject is a public
// method reachable directly (as this test does), and a silent skip there
// used to hide contributors never getting registered at all. Panicking
// matches the RegisterMetrics-error case's fail-fast style in the same
// method.
func TestActuator_Inject_panicsWithoutRegistry(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic when Inject runs without a *prometheus.Registry")
		}
	}()

	a := &metrics.Actuator{}
	a.Inject([]any{[]metrics.MetricsContributor{stubContributor{}}})
}

// TestActuator_Inject_orderIndependent pins the actual fix: Inject must not
// depend on args being in Deps()'s declared order (*prometheus.Registry
// before []MetricsContributor) — sdi happens to preserve that order today,
// but nothing enforces it stays that way if Deps() is ever reordered.
func TestActuator_Inject_orderIndependent(t *testing.T) {
	reg := prometheus.NewRegistry()

	a := &metrics.Actuator{}
	// Deliberately reversed vs Deps()'s own order.
	a.Inject([]any{[]metrics.MetricsContributor{stubContributor{}}, reg})

	body, err := a.MetricsMetrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "test_counter_total") {
		t.Fatalf("body %q missing test_counter_total — contributor was not registered with reversed arg order", body)
	}
}
