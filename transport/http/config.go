package http

import (
	commonv1 "github.com/omcrgnt/proto/gen/go/common/v1"
)

// Config is the non-generic ops HTTP server configuration (proto leaf types).
// ecfg fills and protovalidates Label, Host, and Port before Build is called.
type Config struct {
	Label commonv1.Label
	Host  commonv1.Host
	Port  commonv1.Port
}

// DefaultConfig returns org defaults: ops label, all interfaces, port 8080.
func DefaultConfig() *Config {
	return &Config{
		Label: commonv1.Label{Value: "ops"},
		Host:  commonv1.Host{Value: "0.0.0.0"},
		Port:  commonv1.Port{Value: 8080},
	}
}

// Build returns a user ops HTTP server for res.Add and builder.Build.
// Does not bind; srv-http listen runs lazily on SDI Resolve (see systemServer).
func (c *Config) Build() (any, error) {
	return &systemServer{
		label: c.Label.GetValue(),
		host:  c.Host.GetValue(),
		port:  c.Port.GetValue(),
	}, nil
}
