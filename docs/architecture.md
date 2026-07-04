# ops architecture

## Layers

```text
domain resources     implement probe.Probe* / metrics.MetricsContributor → res.Add
probe.Actuator       SDI many → aggregate → Prober API
metrics.Actuator     SDI registry One + contributors Many → Metricer scrape
transport/http       SDI Prober + Metricer → OpenAPI routes → http.Handler
srv-http (lazy bind) listen on SDI Resolve, Start/Close via runner
runner (in app)      process lifecycle
```

| Package | Role |
|---------|------|
| `probe` | Ops facade for readiness and liveness |
| `metrics` | Shared Prometheus registry + contributor aggregation + Metricer |
| `transport/http` | HTTP transport (handlers; [Server] resource → [Config] spec) |
| `srv-http` | TCP server, otel, request metrics via HTTPMetrics |
| `res/unique` + `sdi` | registry and wiring |

## Metrics (v1)

```text
srv-http (init)      → HTTPMetrics singleton Fixed in unique.Global
metrics/use          → *prometheus.Registry (Fixed) + metrics.Actuator (Replaceable)
transport/http/use   → probe/use + metrics/use + Handler + DefaultServer

Actuator.Inject:
  registry One + []MetricsContributor Many
  HTTPMetrics.RegisterMetrics(reg) → slok NewRecorder once

transport/http Handler:
  Deps Prober + Metricer
  GET /metrics → Metricer.MetricsMetrics(ctx)
```

**Org rule:** HTTP metrics live in **srv-http** (`HTTPMetrics`), not ops. Ops metrics actuator only aggregates `RegisterMetrics` and exposes scrape. Domain services may implement `MetricsContributor` for custom counters.

## Import modes

| Mode | Imports | Result |
|------|---------|--------|
| Full ops | `srv-http` + `transport/http/use` | HTTP metrics + scrape + probes |
| Metrics only | `srv-http` + `metrics/use` | registry + actuator, no ops REST |
| HTTP only | `srv-http` only | Server runs; recorder no-op until actuator |

## Naming (probe)

SDI matches Many/One deps by `Implements`. Probe ports use `{Actuator}{Action}` pattern:

| Port (inbound) | Method | Surface (outbound) |
|----------------|--------|-------------------|
| `ProbeReadiness` | `ProbeReady(ctx)` | `Prober.ProbeReady(ctx)` |
| `ProbeLiveness` | `ProbeLive(ctx)` | `Prober.ProbeLive(ctx)` |

Readiness implementors: domain `srv-http.Server[T]` (transport serve error). Ops HTTP wrapper does not implement `ProbeReadiness` — SDI cycle if it also depends on `Handler` → `Actuator`.

Transport `Handler` depends on `(*probe.Prober)(nil)` and `(*metrics.Metricer)(nil)`.

## Registration (`use` packages)

```text
probe/use              → unique.MustAddReplaceable(&Actuator{})  [via transport/http/use]
metrics/use            → unique.MustAddFixed(Registry) + MustAddReplaceable(&Actuator{})  [via transport/http/use]
transport/http/use     → probe + metrics + Handler + DefaultServer()
srv-http               → unique.MustAddFixed(&HTTPMetrics{}) in package init
```

Pipeline: `app.Run(&appResources, app.Pipeline{Registry: unique.Global(), ...})`.

Override ops host/port/label: optional `transport/http.Server` in AppResources with `ecfg` tags (resource → `Config` spec → `*Server`); materialize dedup removes replaceable `DefaultServer`.

Default ops port **8080** in `DefaultConfig()` — apps should override (e.g. `:9090`) when domain API uses `:8080`.

## OpenAPI

- Spec: `transport/http/openapi/openapi.yaml` — `/livez`, `/readyz`, `/metrics`
- Regenerate: `task gen`

## Backlog

Org-wide items: [github.com/omcrgnt/backlog](https://github.com/omcrgnt/backlog)

| Theme | Item |
|-------|------|
| ops follow-ups (grpc, Grafana, surface/) | [ops-probe-v1-followups](https://github.com/omcrgnt/backlog/blob/main/items/ops-probe-v1-followups.md) |
| SDI Many warn | [sdi-v21-followups](https://github.com/omcrgnt/backlog/blob/main/items/sdi-v21-followups.md) |
| srv-http defer listen | [srv-http-defer-listen](https://github.com/omcrgnt/backlog/blob/main/items/srv-http-defer-listen.md) |
