package probe

import "context"

// Actuator aggregates probe implementors from res via SDI many-deps.
type Actuator struct {
	readiness []ProbeReadiness
	liveness  []ProbeLiveness
}

var _ Prober = (*Actuator)(nil)

func (a *Actuator) Deps() []any {
	return []any{
		([]ProbeReadiness)(nil),
		([]ProbeLiveness)(nil),
	}
}

func (a *Actuator) Inject(args []any) {
	for _, arg := range args {
		switch v := arg.(type) {
		case []ProbeReadiness:
			a.readiness = v
		case []ProbeLiveness:
			a.liveness = v
		}
	}
}

func (a *Actuator) ProbeLive(ctx context.Context) error {
	for _, l := range a.liveness {
		if err := l.ProbeLive(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *Actuator) ProbeReady(ctx context.Context) error {
	for _, r := range a.readiness {
		if err := r.ProbeReady(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ActuatorConfig builds the probe actuator for res registration.
type ActuatorConfig struct{}

func (ActuatorConfig) Build() (any, error) {
	return &Actuator{}, nil
}
