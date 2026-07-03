/*
Package http provides the ops HTTP transport: OpenAPI probe routes and srv-http server config.

# Bootstrap

Blank-import the use subpackage at the app composition root:

	import _ "github.com/omcrgnt/ops/transport/http/use"

transport/http/use registers [Handler] (TagFixed), [DefaultServer] (TagReplaceable), metrics platform, and probe actuator.

Catalog override: [Server] (resource, Configurable) → [Config] (spec) → materialized [*Server]; dedup removes replaceable [DefaultServer].

srv-http depends on [metrics.Recorder] from slok/go-http-metrics. Ops does not register a recorder — the app provides one shared instance for all HTTP servers and metric emitters.

# User override

Optional AppResources field:

	ServerHTTPOps *ophttp.Server `ecfg:"OPS_HTTP"`

Pipeline at app root:

	import _ "github.com/omcrgnt/ops/transport/http/use"
	app.Run(&appResources, app.Pipeline{Registry: unique.Global(), ...})

[DefaultServer] is for transport/http/use system registration with org defaults (:8080).
*/
package http
