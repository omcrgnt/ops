package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omcrgnt/ops/probe"
	"github.com/omcrgnt/ops/transport/http/oapi"
)

// Handler is the ops HTTP transport mounted by srv-http.
type Handler struct {
	prober  probe.Prober
	handler http.Handler
}

func (h *Handler) Deps() []any {
	return []any{
		(*probe.Prober)(nil),
	}
}

func (h *Handler) Inject(args []any) {
	for _, arg := range args {
		if p, ok := arg.(probe.Prober); ok {
			h.prober = p
		}
	}
	strict := oapi.NewStrictHandler(&strictProbe{prober: h.prober}, nil)
	mux := chi.NewRouter()
	h.handler = oapi.HandlerFromMux(strict, mux)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

type strictProbe struct {
	prober probe.Prober
}

func (s *strictProbe) GetLivez(ctx context.Context, _ oapi.GetLivezRequestObject) (oapi.GetLivezResponseObject, error) {
	if err := s.prober.ProbeLive(ctx); err != nil {
		return oapi.GetLivez503TextResponse(err.Error()), nil
	}
	return oapi.GetLivez200TextResponse("ok"), nil
}

func (s *strictProbe) GetReadyz(ctx context.Context, _ oapi.GetReadyzRequestObject) (oapi.GetReadyzResponseObject, error) {
	if err := s.prober.ProbeReady(ctx); err != nil {
		return oapi.GetReadyz503TextResponse(err.Error()), nil
	}
	return oapi.GetReadyz200TextResponse("ok"), nil
}

func (s *strictProbe) GetHealthz(ctx context.Context, _ oapi.GetHealthzRequestObject) (oapi.GetHealthzResponseObject, error) {
	if err := s.prober.ProbeHealth(ctx); err != nil {
		return oapi.GetHealthz503TextResponse(err.Error()), nil
	}
	return oapi.GetHealthz200TextResponse("ok"), nil
}

// HandlerConfig builds the ops HTTP handler for res registration.
type HandlerConfig struct{}

func (HandlerConfig) Build() (any, error) {
	return &Handler{}, nil
}
