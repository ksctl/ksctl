package ksctlhttpserver

import (
	"context"
	"log"

	httpserver "github.com/kubesimplify/ksctl/httpserver/gen/httpserver"
)

// httpserver service example implementation.
// The example methods log the requests and return zero values.
type httpserversrvc struct {
	logger *log.Logger
}

// NewHttpserver returns the httpserver service implementation.
func NewHttpserver(logger *log.Logger) httpserver.Service {
	return &httpserversrvc{logger}
}

// GetHealth implements get health.
func (s *httpserversrvc) GetHealth(ctx context.Context) (res *httpserver.Health, err error) {
	res = &httpserver.Health{}
	s.logger.Print("httpserver.get health")
	return
}
