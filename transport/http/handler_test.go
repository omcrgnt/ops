package http_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	ophttp "github.com/omcrgnt/ops/transport/http"
)

type mockProber struct {
	live, ready error
}

func (m mockProber) ProbeLive(context.Context) error  { return m.live }
func (m mockProber) ProbeReady(context.Context) error { return m.ready }

type mockMetricer struct {
	body []byte
	err  error
}

func (m mockMetricer) MetricsMetrics(context.Context) ([]byte, error) {
	return m.body, m.err
}

func TestHandler_ProbeRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		mock   mockProber
		status int
		body   string
	}{
		{
			name:   "livez ok",
			path:   "/livez",
			mock:   mockProber{},
			status: http.StatusOK,
			body:   "ok",
		},
		{
			name:   "livez fail",
			path:   "/livez",
			mock:   mockProber{live: errors.New("down")},
			status: http.StatusServiceUnavailable,
			body:   "down",
		},
		{
			name:   "readyz ok",
			path:   "/readyz",
			mock:   mockProber{},
			status: http.StatusOK,
			body:   "ok",
		},
		{
			name:   "readyz fail",
			path:   "/readyz",
			mock:   mockProber{ready: errors.New("not ready")},
			status: http.StatusServiceUnavailable,
			body:   "not ready",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := &ophttp.Handler{}
			h.Inject([]any{tc.mock, mockMetricer{}})

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d", rec.Code, tc.status)
			}
			body, _ := io.ReadAll(rec.Body)
			if string(body) != tc.body {
				t.Fatalf("body %q, want %q", body, tc.body)
			}
		})
	}
}
