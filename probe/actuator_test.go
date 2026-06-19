package probe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/omcrgnt/ops/probe"
)

type stubReadiness struct{ err error }

func (s stubReadiness) ProbeReady(context.Context) error { return s.err }

type stubLiveness struct{ err error }

func (s stubLiveness) ProbeLive(context.Context) error { return s.err }

type stubHealth struct{ err error }

func (s stubHealth) ProbeHealth(context.Context) error { return s.err }

func TestActuator_InjectAndProbe(t *testing.T) {
	t.Parallel()

	a := &probe.Actuator{}
	a.Inject([]any{
		[]probe.ProbeReadiness{stubReadiness{}},
		[]probe.ProbeLiveness{stubLiveness{}},
		[]probe.ProbeHealth{stubHealth{}},
	})

	ctx := context.Background()
	if err := a.ProbeReady(ctx); err != nil {
		t.Fatalf("ProbeReady: %v", err)
	}
	if err := a.ProbeLive(ctx); err != nil {
		t.Fatalf("ProbeLive: %v", err)
	}
	if err := a.ProbeHealth(ctx); err != nil {
		t.Fatalf("ProbeHealth: %v", err)
	}
}

func TestActuator_EmptySlicesOK(t *testing.T) {
	t.Parallel()

	a := &probe.Actuator{}
	ctx := context.Background()
	for _, fn := range []struct {
		name string
		run  func() error
	}{
		{"ProbeLive", func() error { return a.ProbeLive(ctx) }},
		{"ProbeReady", func() error { return a.ProbeReady(ctx) }},
		{"ProbeHealth", func() error { return a.ProbeHealth(ctx) }},
	} {
		if err := fn.run(); err != nil {
			t.Fatalf("%s: %v", fn.name, err)
		}
	}
}

func TestActuator_FailFast(t *testing.T) {
	t.Parallel()

	want := errors.New("probe failed")
	a := &probe.Actuator{}
	a.Inject([]any{
		[]probe.ProbeReadiness{stubReadiness{err: want}},
	})

	if err := a.ProbeReady(context.Background()); !errors.Is(err, want) {
		t.Fatalf("ProbeReady err = %v, want %v", err, want)
	}
}
