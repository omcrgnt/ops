package http_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/ops/probe"
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res/unique"
	"github.com/omcrgnt/sdi"
	srvhttp "github.com/omcrgnt/srv-http"
	"github.com/prometheus/client_golang/prometheus"
)

func TestServer_ProbeReady_delegates(t *testing.T) {
	cfg := ophttp.DefaultConfig()
	cfg.Port.Value = 0
	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	srv := built.(*ophttp.Server)
	if _, ok := any(srv).(probe.ProbeReadiness); !ok {
		t.Fatal("Server must implement probe.ProbeReadiness")
	}

	_ = srv.Deps()

	if err := srv.ProbeReady(context.Background()); err != nil {
		t.Fatalf("ProbeReady before Start: %v", err)
	}
}

// TestServer_buildErr_propagatesToAllMethods forces ensureBuilt's error
// path (an out-of-range port fails srv-http's underlying net.Listen) and
// checks all four methods that branch on buildErr: Deps/Inject must no-op
// quietly (nothing to wire once building failed), Start/ProbeReady must
// surface the error instead of behaving as if nothing were configured.
func TestServer_buildErr_propagatesToAllMethods(t *testing.T) {
	cfg := ophttp.DefaultConfig()
	cfg.Port.Value = 99999 // out of the valid TCP port range
	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	srv := built.(*ophttp.Server)

	if deps := srv.Deps(); deps != nil {
		t.Errorf("Deps() = %v, want nil after a Build failure", deps)
	}

	// Must not panic: Inject is a silent no-op once buildErr is set.
	srv.Inject([]any{prometheus.NewRegistry()})

	if cleanup, err := srv.Start(context.Background()); err == nil {
		t.Error("Start: expected the Build error to propagate")
	} else if cleanup != nil {
		t.Error("Start: expected a nil cleanup alongside the Build error")
	}

	if err := srv.ProbeReady(context.Background()); err == nil {
		t.Error("ProbeReady: expected the Build error to propagate")
	}
}

func TestServer_LastStart_isMarkerOnly(t *testing.T) {
	// LastStart has no body — its only role is the runner.LastStarter duck
	// type Runner.Run checks for. This just pins that it exists and is
	// callable without panicking, in case a future edit gives it real work.
	(&ophttp.Server{}).LastStart()
}

// fakeGate satisfies srvhttp.Server's gate dependency (Ready() bool) —
// every registry using srvhttp.Server (ops's inner server included) must
// supply one; this org never uses srv-http without runner.Gate in a real
// deployment, only in isolated tests like this one.
type fakeGate struct{ ready bool }

func (g *fakeGate) Ready() bool { return g.ready }

func TestServer_SDIResolve_withActuatorCycle(t *testing.T) {
	reg := unique.New()

	reg.MustAddReplaceable(&probe.Actuator{})
	reg.MustAddFixed(prometheus.NewRegistry())
	reg.MustAddReplaceable(&metrics.Actuator{})
	reg.MustAddFixed(&srvhttp.HTTPMetrics{})
	reg.MustAddFixed(&ophttp.Handler{})
	reg.MustAddFixed(&fakeGate{})

	cfg := ophttp.DefaultConfig()
	cfg.Port.Value = 0
	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	reg.MustAddReplaceable(built)

	if err := sdi.Resolve(reg); err != nil {
		t.Fatalf("Resolve with Actuator cycle: %v", err)
	}

	actAny, err := reg.GetOneByType(reflect.TypeOf((*probe.Actuator)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	act := actAny.(*probe.Actuator)
	if err := act.ProbeReady(context.Background()); err != nil {
		t.Fatalf("ProbeReady after resolve: %v", err)
	}
}
