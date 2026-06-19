package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/omcrgnt/ops/probe"
	ophttp "github.com/omcrgnt/ops/transport/http"
	"github.com/omcrgnt/res"
	"github.com/omcrgnt/sdi"
)

type alwaysReady struct{}

func (alwaysReady) ProbeReady(context.Context) error { return nil }

func TestIntegration_ResolveAndServe(t *testing.T) {
	res.ResetDefault()
	t.Cleanup(res.ResetDefault)

	_ = res.Add(&probe.Actuator{})
	_ = res.Add(alwaysReady{})
	_ = res.Add(&ophttp.Handler{})

	if err := sdi.Resolve(res.Default); err != nil {
		t.Fatal(err)
	}

	hAny, err := res.GetOneByType(reflect.TypeOf((*ophttp.Handler)(nil)))
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
