package http_test

import (
	"testing"

	commonv1 "github.com/omcrgnt/proto/gen/go/common/v1"
	ophttp "github.com/omcrgnt/ops/transport/http"
)

func TestDefaultConfig(t *testing.T) {
	cfg := ophttp.DefaultConfig()
	if cfg.Port.Value != 8080 {
		t.Fatalf("port = %d, want 8080", cfg.Port.Value)
	}
	if cfg.Label.Value != "ops" {
		t.Fatalf("label = %q, want ops", cfg.Label.Value)
	}
}

func TestConfig_Build(t *testing.T) {
	cfg := ophttp.DefaultConfig()
	cfg.Port = commonv1.Port{Value: 0}

	built, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	if built == nil {
		t.Fatal("expected server resource")
	}
}

func TestDefaultServer(t *testing.T) {
	if ophttp.DefaultServer() == nil {
		t.Fatal("expected default server resource")
	}
}
