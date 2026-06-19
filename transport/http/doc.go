/*
Package http provides the ops HTTP transport: OpenAPI probe routes and srv-http server config.

# Bootstrap

Blank-import the use subpackage at the app composition root:

	import _ "github.com/omcrgnt/ops/transport/http/use"

transport/http/use registers [Handler] (TagFixed), [DefaultServer] (TagReplaceable), and probe actuator via probe/use.

[Config.Build] override (port/host) expects sdi to dedupe Replaceable defaults in the pool by concrete type — not only types appearing in Deps stubs (see sdi backlog). No ServerAnchor.

srv-http depends on [metrics.Recorder] from slok/go-http-metrics. Ops does not register a recorder — the app (or future res/core/use meta pack) provides one shared instance for all HTTP servers and metric emitters.

# User override

	type AppConfig struct {
	    OpsServer ophttp.Config `ecfg:"OPS_SERVER"`
	}

Pipeline: builder.Build(cfg, res.Default) → sdi.Resolve(res.Default).
Dedup removes the replaceable default server when an explicit server is built from [Config].

[Config.Build] returns a user server for res.Add;
[DefaultServer] is for transport/http/use system registration.

See https://github.com/omcrgnt/logger — same Default* vs Config.Build split for server port override only (logger remains separately configurable).
*/
package http
