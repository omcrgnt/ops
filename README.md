# ops

Org library for application **operations surface**: probe and metrics actuators, HTTP transport.

Supersedes the experimental [actuator](https://github.com/omcrgnt/actuator) repo.

Stack: `res v0.22`, `sdi v0.21`, `srv-http` (local v0.21+ with HTTPMetrics).

## Layout

```text
probe/                 # readiness / liveness / health actuator
metrics/               # Prometheus registry actuator + Metricer scrape surface
transport/http/        # OpenAPI HTTP transport + srv-http server config
```

Each actuator:

```text
interface.go   # surface API (Prober / Metricer)
deps.go        # domain port interfaces
actuator.go    # SDI Compatible implementation
use/use.go     # unique.MustAddFixed / MustAddReplaceable
```

## Zero-config

```go
import (
    _ "github.com/omcrgnt/srv-http/use"
    _ "github.com/omcrgnt/ops/metrics/use"
    _ "github.com/omcrgnt/ops/transport/http/use"
)

app.Run(&appResources, app.Pipeline{Registry: unique.Global(), ...})
```

Registers probe actuator, metrics actuator + shared `*prometheus.Registry`, ops HTTP handler, and default ops server (`DefaultServer`, `:8080`). HTTP metrics recorder: `srv-http.HTTPMetrics` via `srv-http/use` (not ops).

Optional override: `transport/http.Config` in AppResources with `ecfg` tags for ops listen port/host; materialize dedup removes the replaceable `DefaultServer`.

## Codegen

```bash
task gen   # oapi-codegen from transport/http/openapi/
```

See [docs/architecture.md](docs/architecture.md).
