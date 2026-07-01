package probe_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/omcrgnt/ops/probe"
	"github.com/omcrgnt/res/unique"
	"github.com/omcrgnt/sdi"
)

type fakeReady struct{}

func (fakeReady) ProbeReady(context.Context) error { return nil }

func TestActuator_SDIResolve(t *testing.T) {
	reg := unique.New()

	reg.MustAddReplaceable(&probe.Actuator{})
	if err := reg.Add(fakeReady{}); err != nil {
		t.Fatal(err)
	}

	if err := sdi.Resolve(reg); err != nil {
		t.Fatal(err)
	}

	actAny, err := reg.GetOneByType(reflect.TypeOf((*probe.Actuator)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	act := actAny.(*probe.Actuator)
	if err := act.ProbeReady(context.Background()); err != nil {
		t.Fatalf("ProbeReady after resolve: %v", err)
	}
}

type failReady struct{}

func (failReady) ProbeReady(context.Context) error { return errors.New("not ready") }

func TestActuator_SDIResolveMany(t *testing.T) {
	reg := unique.New()

	reg.MustAddReplaceable(&probe.Actuator{})
	if err := reg.Add(fakeReady{}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(failReady{}); err != nil {
		t.Fatal(err)
	}

	if err := sdi.Resolve(reg); err != nil {
		t.Fatal(err)
	}

	actAny, err := reg.GetOneByType(reflect.TypeOf((*probe.Actuator)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	act := actAny.(*probe.Actuator)
	if err := act.ProbeReady(context.Background()); err == nil {
		t.Fatal("expected readiness failure")
	}
}
