package main

import (
	"context"
	"net"
	"os"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"google.golang.org/grpc/health"

	"github.com/ksctl/ksctl/pkg/types"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/application"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/helpers"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/scale"
	ksctlHelpers "github.com/ksctl/ksctl/pkg/helpers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedKsctlAgentServer
}

var (
	log      types.LoggerFactory
	agentCtx context.Context
)

func (s *server) Scale(ctx context.Context, in *pb.ReqScale) (*pb.ResScale, error) {
	log.Debug(agentCtx, "Request", "ReqScale", in)

	if err := scale.CallManager(agentCtx, log, in); err != nil {
		log.Error("CallManager", "Reason", err)
		return nil, status.Error(codes.Unimplemented, "failure from calling ksctl manager. Reason: "+err.Error())
	}

	log.Success(agentCtx, "Handled Scale")
	return &pb.ResScale{IsUpdated: true}, nil
}

func (s *server) LoadBalancer(ctx context.Context, in *pb.ReqLB) (*pb.ResLB, error) {
	log.Debug(agentCtx, "Request", "ReqLoadBalancer", in)

	log.Success(agentCtx, "Handled LoadBalancer")
	return nil, nil
}

func (s *server) Application(ctx context.Context, in *pb.ReqApplication) (*pb.ResApplication, error) {
	log.Debug(agentCtx, "Request", "ReqApplication", in)
	if len(in.Apps) == 0 {
		return nil, status.Error(codes.Unimplemented, "invalid argument, cannot contain empty apps")
	}
	// FIXME: important the agentCtx is used instead of ctx
	// Reason: the context from the unit test were not transfarable

	if err := application.Handler(agentCtx, log, in); err != nil {
		log.Error("Handler", "Reason", err)
		return &pb.ResApplication{FailedApps: []string{err.Error()}}, status.Error(codes.Canceled, "invalid returned from manager")
	}
	if _, ok := ksctlHelpers.IsContextPresent(agentCtx, consts.KsctlTestFlagKey); ok {
		return &pb.ResApplication{FailedApps: []string{"none"}}, nil
	}
	log.Success(agentCtx, "Handled Application")
	return new(pb.ResApplication), nil
}

func main() {
	agentCtx = context.WithValue(
		context.Background(),
		consts.KsctlModuleNameKey,
		"ksctl-agent")

	agentCtx = context.WithValue(
		agentCtx,
		consts.KsctlContextUserID,
		"ksctl-agent")

	log = logger.NewStructuredLogger(
		helpers.LogVerbosity[os.Getenv("LOG_LEVEL")],
		helpers.LogWriter)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Error("unable to do http listener", "err", err)
	}

	s := grpc.NewServer()
	defer s.Stop()
	hs := health.NewServer()                   // will default to respond with SERVING
	grpc_health_v1.RegisterHealthServer(s, hs) // registration

	reflection.Register(s) // for debugging purposes

	pb.RegisterKsctlAgentServer(s, &server{}) // Register the server with the gRPC server

	log.Print(agentCtx, "Server started", "port", "8080")

	if err := s.Serve(listener); err != nil {
		log.Error("failed to serve", "err", err)
	}
}
