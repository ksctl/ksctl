package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("ksctl-http-server", func() {
	Title("ksctl http server")
	Description("ksctl server for http based usage")
	Server("server", func() {
		Host("production", func() {
			URI("http://0.0.0.0:8080")
		})
	})
})

var HealthAuth = Type("Health", func() {
	Attribute("msg", String, "message")
})

var _ = Service("httpserver", func() {
	Description("server handlers")

	Method("get health", func() {
		Result(HealthAuth)
		HTTP(func() {
			GET("/healthz")
		})
	})

	Files("/openapi3.json", "./gen/http/openapi3.json")
	Files("/swaggerui/{*path}", "./swaggerui")
})
