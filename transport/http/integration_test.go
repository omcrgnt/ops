package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/ops/probe"
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res/unique"
	"github.com/omcrgnt/sdi"
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
	cfg.Port.Value = 9090
	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(built); err != nil {
		t.Fatal(err)
	}

	if err := sdi.Resolve(reg); err == nil {
		t.Fatal("expected wire error without handler and metrics deps")
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
