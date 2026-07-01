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
| `probe` | Ops facade for readiness/liveness/health |
| `metrics` | Shared Prometheus registry + contributor aggregation + Metricer |
| `transport/http` | HTTP transport (handlers only; server via Config.Build) |
| `srv-http` | TCP server, otel, request metrics via HTTPMetrics |
| `res/unique` + `sdi` | registry and wiring |

## Metrics (v1)

```text
srv-http/use         → HTTPMetrics singleton (MetricsContributor + slok Recorder)
metrics/use          → *prometheus.Registry (Fixed) + metrics.Actuator (Replaceable)

Actuator.Inject:
  registry One + []MetricsContributor Many
  HTTPMetrics.RegisterMetrics(reg) → slok NewRecorder once

transport/http Handler:
  Deps Prober + Metricer
  GET /metrics → Metricer.MetricsMetrics(ctx)
```

**Org rule:** HTTP metrics live in **srv-http** (`HTTPMetrics`), not ops. Ops metrics actuator only aggregates `RegisterMetrics` and exposes scrape. Domain services may implement `MetricsContributor` for custom counters.

## Naming (probe)

SDI matches Many/One deps by `Implements`. Probe ports use `{Actuator}{Action}` pattern:

| Port (inbound) | Method | Surface (outbound) |
|----------------|--------|-------------------|
| `ProbeReadiness` | `ProbeReady(ctx)` | `Prober.ProbeReady(ctx)` |
| `ProbeLiveness` | `ProbeLive(ctx)` | `Prober.ProbeLive(ctx)` |
| `ProbeHealth` | `ProbeHealth(ctx)` | `Prober.ProbeHealth(ctx)` |

Transport `Handler` depends on `(*probe.Prober)(nil)` and `(*metrics.Metricer)(nil)`.

## Registration (`use` packages)

```text
probe/use              → unique.MustAddReplaceable(&Actuator{})
metrics/use            → unique.MustAddFixed(Registry) + MustAddReplaceable(&Actuator{})
transport/http/use     → unique.MustAddFixed(&Handler{}) + MustAddReplaceable(DefaultServer())
```

Apps also import `_ "github.com/omcrgnt/srv-http/use"` for `HTTPMetrics`.

Pipeline: `app.Run(&appResources, app.Pipeline{Registry: unique.Global(), ...})`.

Override ops host/port/label: optional `transport/http.Config` in AppResources with `ecfg` tags; materialize dedup removes replaceable `DefaultServer`.

Default ops port **8080** in `DefaultConfig()` — apps should override (e.g. `:9090`) when domain API uses `:8080`.

## OpenAPI

- Spec: `transport/http/openapi/openapi.yaml` — `/livez`, `/readyz`, `/healthz`, `/metrics`
- Regenerate: `task gen`

## Backlog

Org-wide items: [github.com/omcrgnt/backlog](https://github.com/omcrgnt/backlog)

| Theme | Item |
|-------|------|
| ops follow-ups (grpc, Grafana, surface/) | [ops-probe-v1-followups](https://github.com/omcrgnt/backlog/blob/main/items/ops-probe-v1-followups.md) |
| SDI Many warn | [sdi-v21-followups](https://github.com/omcrgnt/backlog/blob/main/items/sdi-v21-followups.md) |
| srv-http defer listen | [srv-http-defer-listen](https://github.com/omcrgnt/backlog/blob/main/items/srv-http-defer-listen.md) |
