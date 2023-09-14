// Code generated by goa v3.12.4, DO NOT EDIT.
//
// httpserver endpoints
//
// Command:
// $ goa gen github.com/kubesimplify/ksctl/httpserver/design -o httpserver

package httpserver

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Endpoints wraps the "httpserver" service endpoints.
type Endpoints struct {
	CreateHa    goa.Endpoint
	DeleteHa    goa.Endpoint
	Scaledown   goa.Endpoint
	Scaleup     goa.Endpoint
	GetHealth   goa.Endpoint
	GetClusters goa.Endpoint
}

// NewEndpoints wraps the methods of the "httpserver" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		CreateHa:    NewCreateHaEndpoint(s),
		DeleteHa:    NewDeleteHaEndpoint(s),
		Scaledown:   NewScaledownEndpoint(s),
		Scaleup:     NewScaleupEndpoint(s),
		GetHealth:   NewGetHealthEndpoint(s),
		GetClusters: NewGetClustersEndpoint(s),
	}
}

// Use applies the given middleware to all the "httpserver" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.CreateHa = m(e.CreateHa)
	e.DeleteHa = m(e.DeleteHa)
	e.Scaledown = m(e.Scaledown)
	e.Scaleup = m(e.Scaleup)
	e.GetHealth = m(e.GetHealth)
	e.GetClusters = m(e.GetClusters)
}

// NewCreateHaEndpoint returns an endpoint function that calls the method
// "create ha" of service "httpserver".
func NewCreateHaEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*Metadata)
		return s.CreateHa(ctx, p)
	}
}

// NewDeleteHaEndpoint returns an endpoint function that calls the method
// "delete ha" of service "httpserver".
func NewDeleteHaEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*Metadata)
		return s.DeleteHa(ctx, p)
	}
}

// NewScaledownEndpoint returns an endpoint function that calls the method
// "scaledown" of service "httpserver".
func NewScaledownEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*Metadata)
		return s.Scaledown(ctx, p)
	}
}

// NewScaleupEndpoint returns an endpoint function that calls the method
// "scaleup" of service "httpserver".
func NewScaleupEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*Metadata)
		return s.Scaleup(ctx, p)
	}
}

// NewGetHealthEndpoint returns an endpoint function that calls the method "get
// health" of service "httpserver".
func NewGetHealthEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		return s.GetHealth(ctx)
	}
}

// NewGetClustersEndpoint returns an endpoint function that calls the method
// "get clusters" of service "httpserver".
func NewGetClustersEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		return s.GetClusters(ctx)
	}
}
