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
	agentCtx context.Context = context.WithValue(
		context.Background(),
		consts.ContextModuleNameKey,
		"ksctl-agent")
)

func (s *server) Scale(ctx context.Context, in *pb.ReqScale) (*pb.ResScale, error) {
	log.Debug(agentCtx, "Request", "ReqScale", in)

	if err := scale.CallManager(agentCtx, log, in); err != nil {
		log.Error(agentCtx, "CallManager", "Reason", err)
		return nil, status.Error(codes.Unimplemented, "failure from calling ksctl manager. Reason:"+err.Error())
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

	if err := application.Handler(agentCtx, log, in); err != nil {
		log.Error(agentCtx, "Handler", "Reason", err)
		return &pb.ResApplication{FailedApps: []string{err.Error()}}, status.Error(codes.Canceled, "invalid returned from manager")
	}
	// TODO: make a function passing for what should be the client this will help
	//  or something different
	if len(os.Getenv("UNIT_TEST_GRPC_KSCTL_AGENT")) != 0 {
		return &pb.ResApplication{FailedApps: []string{"none"}}, nil
	}
	log.Success(agentCtx, "Handled Application")
	return new(pb.ResApplication), nil
}

func main() {

	log = logger.NewStructuredLogger(
		helpers.LogVerbosity[os.Getenv("LOG_LEVEL")],
		helpers.LogWriter)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Error(agentCtx, "unable to do http listener", "err", err)
	}

	s := grpc.NewServer()
	defer s.Stop()
	hs := health.NewServer()                   // will default to respond with SERVING
	grpc_health_v1.RegisterHealthServer(s, hs) // registration

	reflection.Register(s) // for debugging purposes

	pb.RegisterKsctlAgentServer(s, &server{}) // Register the server with the gRPC server

	log.Print(agentCtx, "Server started", "port", "8080")

	if err := s.Serve(listener); err != nil {
		log.Error(agentCtx, "failed to serve", "err", err)
	}
}
