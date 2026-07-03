package http

import (
	"context"
	"fmt"
	"sync"

	commonv1 "github.com/omcrgnt/proto/gen/go/common/v1"
	srvhttp "github.com/omcrgnt/srv-http"
)

type systemServer struct {
	label string
	host  string
	port  uint32

	mu       sync.Mutex
	inner    *srvhttp.Server[*Handler]
	buildErr error
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
	server, ok := built.(*srvhttp.Server[*Handler])
	if !ok {
		s.buildErr = fmt.Errorf("ops/http: Config.Build: got %T, want *srvhttp.Server[*Handler]", built)
		return
	}
	s.inner = server
}

func (s *systemServer) Deps() []any {
	s.ensureBuilt()
	if s.buildErr != nil {
		return nil
	}
	if s.inner == nil {
		return nil
	}
	return s.inner.Deps()
}

func (s *systemServer) Inject(args []any) {
	s.ensureBuilt()
	if s.buildErr != nil {
		return
	}
	if s.inner == nil {
		return
	}
	s.inner.Inject(args)
}

func (s *systemServer) Start(ctx context.Context) error {
	s.ensureBuilt()
	if s.buildErr != nil {
		return s.buildErr
	}
	if s.inner == nil {
		return nil
	}
	return s.inner.Start(ctx)
}

func (s *systemServer) Close(ctx context.Context) error {
	s.mu.Lock()
	inner := s.inner
	s.mu.Unlock()

	if inner == nil {
		return nil
	}
	return inner.Close(ctx)
}
