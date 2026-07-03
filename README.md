# ops

Org library for application **operations surface**: probe and metrics actuators, HTTP transport.

Supersedes the experimental [actuator](https://github.com/omcrgnt/actuator) repo.

Stack: `res v0.22`, `sdi v0.21`, `srv-http` (HTTPMetrics auto-registers on import).

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
    _ "github.com/omcrgnt/ops/transport/http/use"
    srvhttp "github.com/omcrgnt/srv-http" // Server types; HTTPMetrics via init
)

app.Run(&appResources, app.Pipeline{Registry: unique.Global(), ...})
```

One blank import registers probe actuator, metrics actuator + shared `*prometheus.Registry`, ops HTTP handler, and default ops server (`DefaultServer`, `:8080`).

Import `srv-http` for domain `Server[T]` — `HTTPMetrics` singleton registers in `unique.Global` automatically (no separate `use` package).

Metrics without ops REST: `_ "github.com/omcrgnt/ops/metrics/use"` only (registry + actuator, no `/metrics` route).

Optional override: `*ophttp.Server` in AppResources with `ecfg` tags (resource → `Config` spec → `*Server`); materialize dedup removes the replaceable `DefaultServer`.

## Codegen

```bash
task gen   # oapi-codegen from transport/http/openapi/
```

See [docs/architecture.md](docs/architecture.md).
