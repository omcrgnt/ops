package http_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/ops/probe"
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res/unique"
	"github.com/omcrgnt/sdi"
	srvhttp "github.com/omcrgnt/srv-http"
	"github.com/prometheus/client_golang/prometheus"
)

type alwaysReady struct{}

func (alwaysReady) ProbeReady(context.Context) error { return nil }

func TestIntegration_ResolveAndServe(t *testing.T) {
	reg := unique.New()

	reg.MustAddReplaceable(&probe.Actuator{})
	if err := reg.Add(alwaysReady{}); err != nil {
		t.Fatal(err)
	}
	reg.MustAddFixed(prometheus.NewRegistry())
	reg.MustAddReplaceable(&metrics.Actuator{})
	reg.MustAddFixed(&ophttp.Handler{})

	if err := sdi.Resolve(reg); err != nil {
		t.Fatal(err)
	}

	hAny, err := reg.GetOneByType(reflect.TypeOf((*ophttp.Handler)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	h := hAny.(*ophttp.Handler)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "ok" {
		t.Fatalf("body %q", body)
	}
}

func TestIntegration_DefaultServerOverrideDedup(t *testing.T) {
	reg := unique.New()

	reg.MustAddReplaceable(ophttp.DefaultServer())

	cfg := ophttp.DefaultConfig()
	cfg.Port.Value = 0 // ephemeral port; avoid clash with running apps on :9090
	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(built); err != nil {
		t.Fatal(err)
	}

	typ := reflect.TypeOf(built)
	if len(reg.GetByType(typ)) != 1 {
		t.Fatalf("expected 1 server after dedup, got %d", len(reg.GetByType(typ)))
	}
	got, err := reg.GetOneByType(typ)
	if err != nil {
		t.Fatal(err)
	}
	if got != built {
		t.Fatal("expected explicit Config.Build server to remain after dedup")
	}
}

func TestIntegration_OwnEndpointNeverGated(t *testing.T) {
	// A not-ready gate would 503 a normal srv-http server (see
	// srvhttp.TestConfig_Build_gate) — this proves ops's own /readyz stays
	// reachable regardless, because ensureBuilt calls DisableGate on its
	// inner server.
	reg := unique.New()

	reg.MustAddReplaceable(&probe.Actuator{})
	if err := reg.Add(alwaysReady{}); err != nil {
		t.Fatal(err)
	}
	reg.MustAddFixed(prometheus.NewRegistry())
	reg.MustAddReplaceable(&metrics.Actuator{})
	reg.MustAddFixed(&srvhttp.HTTPMetrics{})
	reg.MustAddFixed(&ophttp.Handler{})
	reg.MustAddFixed(&fakeGate{ready: false})

	const port = 18476
	cfg := ophttp.DefaultConfig()
	cfg.Port.Value = port
	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	reg.MustAddReplaceable(built)

	if err := sdi.Resolve(reg); err != nil {
		t.Fatal(err)
	}

	srv := built.(*ophttp.Server)
	stop, err := srv.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stop(context.Background()) })

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/readyz", port))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d (gate must not block ops's own endpoint)", resp.StatusCode, http.StatusOK)
	}
}

func TestIntegration_MetricsRoute(t *testing.T) {
	reg := unique.New()
	reg.MustAddReplaceable(&probe.Actuator{})
	reg.MustAddFixed(prometheus.NewRegistry())
	reg.MustAddReplaceable(&metrics.Actuator{})
	reg.MustAddFixed(&ophttp.Handler{})

	if err := sdi.Resolve(reg); err != nil {
		t.Fatal(err)
	}

	hAny, err := reg.GetOneByType(reflect.TypeOf((*ophttp.Handler)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	h := hAny.(*ophttp.Handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}
