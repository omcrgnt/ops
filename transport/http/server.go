package http

import (
	"context"
	"sync"

	commonv1 "github.com/omcrgnt/proto/gen/go/common/v1"
	srvhttp "github.com/omcrgnt/srv-http"
)

type systemServer struct {
	label string
	host  string
	port  uint32

	mu       sync.Mutex
	inner    any
	buildErr error
}

type depsInjector interface {
	Deps() []any
	Inject(args []any)
}

type starter interface {
	Start(ctx context.Context) error
}

type closer interface {
	Close(ctx context.Context) error
}

// DefaultServer returns the system ops HTTP server for transport/http/use registration.
// Bind happens lazily on first SDI Deps/Inject (before runner.Start), not in init.
func DefaultServer() any {
	cfg := DefaultConfig()
	return &systemServer{
		label: cfg.Label.GetValue(),
		host:  cfg.Host.GetValue(),
		port:  cfg.Port.GetValue(),
	}
}

func (s *systemServer) ensureBuilt() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inner != nil || s.buildErr != nil {
		return
	}

	built, err := (&srvhttp.Config[*Handler]{
		Label: commonv1.Label{Value: s.label},
		Host:  commonv1.Host{Value: s.host},
		Port:  commonv1.Port{Value: s.port},
	}).Build()
	if err != nil {
		s.buildErr = err
		return
	}
	s.inner = built
}

func (s *systemServer) Deps() []any {
	s.ensureBuilt()
	if s.buildErr != nil {
		return nil
	}
	d, ok := s.inner.(depsInjector)
	if !ok {
		return nil
	}
	return d.Deps()
}

func (s *systemServer) Inject(args []any) {
	s.ensureBuilt()
	if s.buildErr != nil {
		return
	}
	d, ok := s.inner.(depsInjector)
	if !ok {
		return
	}
	d.Inject(args)
}

func (s *systemServer) Start(ctx context.Context) error {
	s.ensureBuilt()
	if s.buildErr != nil {
		return s.buildErr
	}
	st, ok := s.inner.(starter)
	if !ok {
		return nil
	}
	return st.Start(ctx)
}

func (s *systemServer) Close(ctx context.Context) error {
	s.mu.Lock()
	inner := s.inner
	s.mu.Unlock()

	if inner == nil {
		return nil
	}
	cl, ok := inner.(closer)
	if !ok {
		return nil
	}
	return cl.Close(ctx)
}
