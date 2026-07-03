package use_test

import (
	"reflect"
	"testing"

	"github.com/omcrgnt/ops/metrics"
	"github.com/omcrgnt/res/unique"
	"github.com/prometheus/client_golang/prometheus"

	_ "github.com/omcrgnt/ops/transport/http/use"
)

func TestInit_registersMetricsPlatform(t *testing.T) {
	reg := unique.Global()

	if _, err := reg.GetOneByType(reflect.TypeOf((*prometheus.Registry)(nil))); err != nil {
		t.Fatalf("registry: %v", err)
	}
	if _, err := reg.GetOneByType(reflect.TypeOf((*metrics.Actuator)(nil))); err != nil {
		t.Fatalf("actuator: %v", err)
	}
}
