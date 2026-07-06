package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/ops/probe"
	"github.com/omcrgnt/ops/transport/http/oapi"
)

// Handler is the ops HTTP transport mounted by srv-http.
type Handler struct {
	prober   probe.Prober
	metricer metrics.Metricer
	handler  http.Handler
}

func (h *Handler) Deps() []any {
	return []any{
		(*probe.Prober)(nil),
		(*metrics.Metricer)(nil),
	}
}

func (h *Handler) Inject(args []any) {
	for _, arg := range args {
		switch v := arg.(type) {
		case probe.Prober:
			h.prober = v
		case metrics.Metricer:
			h.metricer = v
		}
	}
	strict := oapi.NewStrictHandler(&strictOps{prober: h.prober, metricer: h.metricer}, nil)
	mux := chi.NewRouter()
	h.handler = oapi.HandlerFromMux(strict, mux)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

type strictOps struct {
	prober   probe.Prober
	metricer metrics.Metricer
}

func (s *strictOps) GetLivez(ctx context.Context, _ oapi.GetLivezRequestObject) (oapi.GetLivezResponseObject, error) {
	if err := s.prober.ProbeLive(ctx); err != nil {
		return oapi.GetLivez503TextResponse(err.Error()), nil
	}
	return oapi.GetLivez200TextResponse("ok"), nil
}

func (s *strictOps) GetReadyz(ctx context.Context, _ oapi.GetReadyzRequestObject) (oapi.GetReadyzResponseObject, error) {
	if err := s.prober.ProbeReady(ctx); err != nil {
		return oapi.GetReadyz503TextResponse(err.Error()), nil
	}
	return oapi.GetReadyz200TextResponse("ok"), nil
}

func (s *strictOps) GetMetrics(ctx context.Context, _ oapi.GetMetricsRequestObject) (oapi.GetMetricsResponseObject, error) {
	body, err := s.metricer.MetricsMetrics(ctx)
	if err != nil {
		return oapi.GetMetrics503TextResponse(err.Error()), nil
	}
	return oapi.GetMetrics200TextResponse(body), nil
}

// HandlerConfig builds the ops HTTP handler for res registration.
type HandlerConfig struct{}

func (HandlerConfig) Build() (any, error) {
	return &Handler{}, nil
}
