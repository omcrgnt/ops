# ops architecture

## Layers

```text
domain resources     implement probe.ProbeReadiness / ProbeLiveness / ProbeHealth → res.Add
probe.Actuator       SDI many → aggregate → Prober API
transport/http       SDI one Prober → OpenAPI routes → http.Handler
srv-http (lazy bind) listen on SDI Resolve, Start/Close via runner
runner (in app)      process lifecycle
```

| Package | Role |
|---------|------|
| `probe` | Ops facade for readiness/liveness/health |
| `transport/http` | HTTP transport (handlers only; server via Config.Build) |
| `srv-http` | TCP server, otel, request metrics |
| `res` + `sdi` | registry and wiring |

## Naming

SDI matches Many/One deps by `Implements` — any type in `res` with matching method signatures is collected. To avoid accidental duck-typing with business APIs, probe ports use the `{Actuator}{Action}` pattern:

| Port (inbound) | Method | Surface (outbound) |
|----------------|--------|-------------------|
| `ProbeReadiness` | `ProbeReady(ctx)` | `Prober.ProbeReady(ctx)` |
| `ProbeLiveness` | `ProbeLive(ctx)` | `Prober.ProbeLive(ctx)` |
| `ProbeHealth` | `ProbeHealth(ctx)` | `Prober.ProbeHealth(ctx)` |

**Org rule:** do not use bare `Live` / `Ready` / `Health` method names on types registered in `res` for non-probe purposes.

Transport `Handler` depends on `(*probe.Prober)(nil)` (interface One), not concrete `*Actuator`. Footgun on surface is mitigated by prefixed method names, not by concrete stub.

## Probe aggregation (v1)

`Actuator` runs implementors **sequentially**, **fail-fast** on first error. Latency for a probe kind is roughly the **sum** of check durations.

| Probe kind | Typical implementors | Notes |
|------------|---------------------|-------|
| `ProbeLive` | few, cheap | process liveness; parallel rarely needed |
| `ProbeReady` | infra deps (DB, cache, …) | may become slow with many checks |
| `ProbeHealth` | deeper operational checks | same as readiness when heavy |

**Org layout:** one shared resource per dependency type in `res` (e.g. one `db` connection pool for the whole app). Repositories do **not** implement probe ports — only the shared infra resource does. So duplicate DB pings from multiple repos is not expected; parallel vs sequential mainly matters when **several different** `ProbeReadiness` / `ProbeHealth` implementors exist (DB + cache + queue, etc.). Revisit aggregation policy once the first real implementor lands in an app or org-lib.

## Implementor contract

Implementors of `ProbeReadiness`, `ProbeLiveness`, and `ProbeHealth` must:

- respect `context.Context` cancellation and deadlines (kube probe timeout, HTTP request ctx);
- keep `ProbeLive` cheap — no heavy I/O;
- document whether `ProbeReady` / `ProbeHealth` are safe for **concurrent** invocation on the same receiver.

v1 always calls implementors **one at a time**, so concurrency safety is not required today. If optional **parallel aggregation** is added later (see backlog), the actuator may invoke multiple implementors of the same kind concurrently. Implementors that touch shared mutable state without synchronization may race; the library cannot infer safety from method signatures alone — **callers must know** how their resources behave. Prefer idempotent checks and connection-pool pings over ad-hoc shared flags.

## Actuator file convention

One directory per actuator (`probe/`, future `metrics/`):

| File | Contents |
|------|----------|
| `interface.go` | Surface API consumed by transports (`Prober`) |
| `deps.go` | Port interfaces implemented by domain |
| `actuator.go` | `Deps`/`Inject`, aggregation logic — no `Config` |
| `use/use.go` | `res.AddToGlobalWithTags` (TagReplaceable) |

## Registration (`use` packages)

```text
probe/use              → ActuatorConfig (TagReplaceable)
transport/http/use     → HandlerConfig (TagFixed) + DefaultServer (TagReplaceable, legacy resource until opssrv refactor)
```

Pattern: library use init registers **configs** (`AddToGlobalWithTags`); AppResources + `builder.Seed` / `builder.Build` materialize Builder entries (inherits tags). `DefaultServer()` remains a pre-built resource (not Builder) until opssrv phase 2.

**metrics.Recorder** — shared app dependency (srv-http One dep), **not** registered in ops/use. App AppResources field or blank-import `logger/use` / app metrics config.

`transport/http/use` blank-imports `probe/use` first.

Apps import `_ "github.com/omcrgnt/ops/transport/http/use"` and run `app.Run(&appResources, app.Pipeline{Registry: res.Global(), ...})`. No mandatory ops server field in AppResources.

Override host/port/label: optional `transport/http.Config` in AppResources with `ecfg` tags; after `builder.Build`, sdi pool-wide dedup removes the replaceable `DefaultServer`.

Default port **8080** is an org convention in `DefaultConfig()`, not enforced by the library. TCP bind errors (e.g. port in use) surface at **SDI Resolve** or **runner.Start**, not at import time.

## Proto

- `transport/http.Config` uses `common.v1.Label`, `Host`, `Port` (same as srv-http).
- Future HTTP/gRPC DTOs: define messages in `proto/ops/v1`, map in OpenAPI via `x-go-type`, configure `oapi-codegen.yaml` `additional-imports` / `import-mapping`.

v1 probe endpoints: no input DTO; `text/plain` responses.

## OpenAPI

- Spec: `transport/http/openapi/openapi.yaml`
- Config: `transport/http/openapi/oapi-codegen.yaml`
- Generated: `transport/http/oapi/gen.go`
- Regenerate: `task gen`

## Backlog

Org-wide items: [github.com/omcrgnt/backlog](https://github.com/omcrgnt/backlog)

| Theme | Item |
|-------|------|
| ops probe follow-ups (demo, grpc, metrics, surface/) | [ops-probe-v1-followups](https://github.com/omcrgnt/backlog/blob/main/items/ops-probe-v1-followups.md) |
| SDI Many warn / lifecycle lint | [sdi-v21-followups](https://github.com/omcrgnt/backlog/blob/main/items/sdi-v21-followups.md) |
| srv-http defer listen | [srv-http-defer-listen](https://github.com/omcrgnt/backlog/blob/main/items/srv-http-defer-listen.md) |
| shared Taskfiles / templates | [org-devtools-taskfiles](https://github.com/omcrgnt/backlog/blob/main/items/org-devtools-taskfiles.md) |

## References

- [actuator/docs/architecture.md](https://github.com/omcrgnt/actuator/blob/main/docs/architecture.md) — prior design discussion
