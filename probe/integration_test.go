package probe_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/omcrgnt/ops/probe"
	"github.com/omcrgnt/res"
	"github.com/omcrgnt/sdi"
)

type fakeReady struct{}

func (fakeReady) ProbeReady(context.Context) error { return nil }

func TestActuator_SDIResolve(t *testing.T) {
	res.ResetDefault()
	t.Cleanup(res.ResetDefault)

	_ = res.Add(&probe.Actuator{})
	_ = res.Add(fakeReady{})

	if err := sdi.Resolve(res.Default); err != nil {
		t.Fatal(err)
	}

	actAny, err := res.GetOneByType(reflect.TypeOf((*probe.Actuator)(nil)))
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
	res.ResetDefault()
	t.Cleanup(res.ResetDefault)

	_ = res.Add(&probe.Actuator{})
	_ = res.Add(fakeReady{})
	_ = res.Add(failReady{})

	if err := sdi.Resolve(res.Default); err != nil {
		t.Fatal(err)
	}

	actAny, err := res.GetOneByType(reflect.TypeOf((*probe.Actuator)(nil)))
	if err != nil {
		t.Fatal(err)
	}
	act := actAny.(*probe.Actuator)
	if err := act.ProbeReady(context.Background()); err == nil {
		t.Fatal("expected readiness failure")
	}
}
