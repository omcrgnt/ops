/*
Package http provides the ops HTTP transport: OpenAPI probe routes and srv-http server config.

# Bootstrap

Blank-import the use subpackage at the app composition root:

	import _ "github.com/omcrgnt/ops/transport/http/use"

transport/http/use registers [Handler] (TagFixed), [DefaultServer] (TagReplaceable), and probe actuator via probe/use.

[Config.Build] override (port/host) relies on sdi pool-wide Replaceable dedup by concrete type after materialize.

srv-http depends on [metrics.Recorder] from slok/go-http-metrics. Ops does not register a recorder — the app provides one shared instance for all HTTP servers and metric emitters.

# User override

Optional AppResources field with ecfg tags on [Config] (after builder.Build, sdi removes the replaceable default server).

Pipeline at app root:

	import _ "github.com/omcrgnt/ops/transport/http/use"
	app.Run(&appResources, app.Pipeline{Registry: res.Global(), ...})

[Config.Build] returns a user server for explicit registration;
[DefaultServer] is for transport/http/use system registration.

See https://github.com/omcrgnt/logger — same Default* vs Config.Build split for server port override only (logger remains separately configurable).
*/
package http
