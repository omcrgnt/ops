# ops

Org library for application **operations surface**: probe actuators and HTTP transport.

Supersedes the experimental [actuator](https://github.com/omcrgnt/actuator) repo.

Stack: `res v0.20.2`, `sdi v0.20.4`, `srv-http v0.20.1`.

## Layout

```text
probe/                 # readiness / liveness / health actuator
transport/http/        # OpenAPI HTTP transport + srv-http server config
```

Each actuator:

```text
interface.go   # surface API (Prober: ProbeLive / ProbeReady / ProbeHealth)
deps.go        # domain port interfaces (ProbeReadiness / ProbeLiveness / ProbeHealth)
actuator.go    # SDI Compatible implementation
use/use.go     # res.AddToGlobalWithTags (TagReplaceable for probe)
```

Domain types implement `ProbeReady`, `ProbeLive`, and/or `ProbeHealth` — not bare `Ready`/`Live`/`Health`. Infra resources (e.g. one shared `db` per app) implement probe ports; repositories typically do not.

v1 aggregation is sequential fail-fast. Implementors must respect context cancellation; document concurrency safety if parallel aggregation is enabled later — see [docs/architecture.md](docs/architecture.md).

## Zero-config

```go
import _ "github.com/omcrgnt/ops/transport/http/use"

app.Run(&appResources, app.Pipeline{Registry: res.Global(), ...})
```

Registers default probe actuator (replaceable), HTTP handler, and default ops server (`DefaultServer`, `:8080`). Shared `metrics.Recorder` — app (not ops/use).

Optional override: `transport/http.Config` in AppResources with `ecfg` tags for ops listen port/host; after `builder.Build`, sdi dedup removes the replaceable default.

## Codegen

```bash
task gen   # oapi-codegen from transport/http/openapi/
```

## Proto-first DTO

Structured request/response types for ops HTTP live in `github.com/omcrgnt/proto`; OpenAPI uses `x-go-type` and `oapi-codegen.yaml` import-mapping. v1 probe routes are `GET` + `text/plain` only.

See [docs/architecture.md](docs/architecture.md).
