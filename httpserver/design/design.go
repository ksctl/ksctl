package design

// import (
// 	. "goa.design/goa/v3/dsl"
// )
//
// var _ = API("ksctl-http-server", func() {
// 	Title("ksctl http server")
// 	Description("ksctl server for http based usage")
//
// 	Server("server", func() {
// 		Host("production", func() {
// 			URI("http://0.0.0.0:8080")
// 		})
// 	})
// })
//
// var metadata = Type("metadata", func() {
// 	Description("metadata for the ksctl to work")
//
// 	Attribute("no_wp", Int, func() {
// 		Description("desired no of workerplane nodes")
// 		Example(1)
// 		Minimum(0)
// 	})
//
// 	Attribute("no_cp", Int32, func() {
// 		Description("desired no of workerplane nodes")
// 		Minimum(3)
// 		Example(3)
// 		Maximum(10)
// 	})
//
// 	Attribute("no_ds", Int32, func() {
// 		Description("desired no of workerplane nodes")
// 		Minimum(1)
// 		Example(1)
// 		Maximum(3)
// 	})
//
// 	Attribute("no_mp", Int32, func() {
// 		Description("desired no of workerplane nodes")
// 		Minimum(1)
// 		Example(1)
// 		Maximum(3)
// 	})
//
// 	Attribute("vm_size_cp", String, func() {
// 		Description("virtual machine size for the controlplane")
// 		Example("g3.small")
// 	})
//
// 	Attribute("vm_size_ds", String, func() {
// 		Description("virtual machine size for the datastore")
// 		Example("g3.small")
// 	})
//
// 	Attribute("vm_size_wp", String, func() {
// 		Description("virtual machine size for the workerplane")
// 		Example("g3.small")
// 	})
//
// 	Attribute("vm_size_lb", String, func() {
// 		Description("virtual machine size for the loadbalancer")
// 		Example("g3.small")
// 	})
//
// 	Attribute("cluster_name", String, func() {
// 		Description("Cluster name")
// 		Example("demo")
// 	})
//
// 	Attribute("region", String, func() {
// 		Description("Region")
// 		Example("XYZ")
// 	})
//
// 	Attribute("cloud", String, func() {
// 		Description("cloud provider")
// 		Example("azure")
// 	})
//
// 	Attribute("distro", String, func() {
// 		Description("kubernetes distribution")
// 		Example("k3s")
// 	})
//
// 	Required("cluster_name", "region", "cloud", "distro") // Required attributes
// })
//
// var responseScale = Type("response", func() {
// 	Description("response type")
//
// 	Attribute("ok", Boolean, "successful")
// 	Attribute("errors", String, "reason of failure")
// 	Attribute("response", Any, "response")
// })
//
// var HealthAuth = Type("Health", func() {
// 	Attribute("msg", String, "message")
// })
//
// var _ = Service("httpserver", func() {
// 	Description("server handlers")
//
// 	Method("create ha", func() {
// 		Payload(metadata)
// 		Result(responseScale)
//
// 		HTTP(func() {
// 			PUT("/createha")
// 		})
// 	})
//
// 	Method("delete ha", func() {
// 		Payload(metadata)
// 		Result(responseScale)
//
// 		HTTP(func() {
// 			PUT("/deleteha")
// 		})
// 	})
//
// 	Method("scaledown", func() {
// 		Payload(metadata)
// 		Result(responseScale)
//
// 		HTTP(func() {
// 			PUT("/deletenodes")
// 		})
// 	})
// 	Method("scaleup", func() {
// 		Payload(metadata)
// 		Result(responseScale)
//
// 		HTTP(func() {
// 			PUT("/addnodes")
// 		})
// 	})
//
// 	Method("get health", func() {
// 		Result(HealthAuth)
// 		HTTP(func() {
// 			GET("/healthz")
// 		})
// 	})
//
// 	Method("get clusters", func() {
// 		Result(responseScale)
// 		HTTP(func() {
// 			GET("/list")
// 		})
// 	})
//
// 	Files("/openapi3.json", "./gen/http/openapi3.json")
// 	Files("/swaggerui/{*path}", "./swaggerui")
// })
