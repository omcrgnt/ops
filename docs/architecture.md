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
| `use/use.go` | `res.AddWithTags(&Actuator{}, TagReplaceable)` |

## Registration (`use` packages)

```text
probe/use              → Actuator (TagReplaceable — app may override with explicit res.Add)
transport/http/use     → Handler (TagFixed) + DefaultServer (TagReplaceable)
```

Pattern: `DefaultServer()` in init (no panic, lazy bind on Resolve); `Config.Build()` for port/host override via `builder.Build`. Dedup of replaceable default vs explicit `*systemServer` requires **pool-wide Replaceable dedup in sdi** (not deps-triggered only) — see [sdi](https://github.com/omcrgnt/sdi) backlog; no ServerAnchor.

**metrics.Recorder** — shared app dependency (srv-http One dep), **not** registered in ops/use. App `MetricsConfig` or future `res/core/use` meta pack (see backlog).

`transport/http/use` blank-imports `probe/use` first.

Apps import `_ "github.com/omcrgnt/ops/transport/http/use"`. No mandatory `AppConfig` ops server field.

Override host/port/label: optional `transport/http.Config` in `AppConfig` with `ecfg` tags (`builder.Build` → sdi dedup removes replaceable default once pool-wide dedup lands in sdi).

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

- Demo integration
- `transport/grpc`
- `/metrics` on ops HTTP
- Meta `ops/use` imported from `res/core/use` (shared `metrics.Recorder`, logger, telemetry defaults)
- `proto/ops/v1` messages when structured probes appear
- **`ops/surface/`** — move `Prober` (and future metrics surface) when grpc/metrics consumers appear; mechanical refactor from `probe/interface.go`
- **Probe actuator `Config` (optional)** — extend `probe/` with ecfg-friendly config (pattern like `transport/http.Config`) to tune heavy probe aggregation without replacing `Actuator`:
  - default: sequential fail-fast (current v1 behavior);
  - opt-in parallel per probe kind (at least `ProbeReady` / `ProbeHealth`; `ProbeLive` stays sequential);
  - context cancel on first error or global deadline;
  - document error policy (first error vs aggregate). Design after first real `ProbeReady`/`ProbeHealth` implementor in app or org-lib — until then latency trade-offs are speculative.
- **SDI Many policy review** — раньше обсуждали `([]T)(nil)` строго `>= 1` implementor (0 → error). Для `probe.Actuator` сейчас разрешён пустой slice (0 → `[]T{}`). Пересмотреть: опционально **warn** при 0 (misconfig / забытый `res.Add`) вместо silent OK или hard error. См. `sdi` backlog.
- **SDI pool-wide Replaceable dedup** — today dedup runs only for types from `Deps()` stubs. Runner-only resources (e.g. replaceable `*systemServer` + explicit `Config.Build`) need dedup by concrete type across the whole pool — no fake dep anchors. Track in sdi.
- **srv-http defer listen** — перенести `net.Listen` из `Build()` в `Start()` so bind failures align with runner lifecycle; ops `systemServer` lazy-build workaround can shrink afterward.
- **Org devtools repo (`taskfiles` + templates)** — отдельный репозиторий (рядом с `lint/`):
  - shared Taskfile fragments: docker `go test` / `go-mutesting`, `pkgsite`, lint wrappers;
  - service/repo templates (Taskfile skeleton, zero-config `use` imports, ecfg layout);
  - подключение через Task `includes` + git URL (`?ref=vX.Y.Z`, опционально `checksum`, `remote.cache-expiry` в `.taskrc`);
  - миграция org repos (`ops`, `sdi`, `res`, `demo`, …) с copy-paste `Taskfile.yml` на `includes` + repo-local vars (`EXTRA_MOUNTS`, `replace` paths).

## References

- [actuator/docs/architecture.md](https://github.com/omcrgnt/actuator/blob/main/docs/architecture.md) — prior design discussion
