module github.com/omcrgnt/ops

go 1.26.2

require (
	github.com/go-chi/chi/v5 v5.3.0
	github.com/omcrgnt/proto/gen/go v0.3.0
	github.com/omcrgnt/res v0.8.1
	github.com/omcrgnt/sdi v1.3.0
	github.com/omcrgnt/srv-http v0.4.3
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20260415201107-50325440f8f2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/slok/go-http-metrics v0.13.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/omcrgnt/res => /opt/github/res
	github.com/omcrgnt/sdi => /opt/github/sdi
)
