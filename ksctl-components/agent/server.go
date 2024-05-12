package main

import (
	"context"
	"encoding/json"
	"net"
	"os"

	"github.com/ksctl/ksctl/pkg/logger"
	"google.golang.org/grpc/health"

	"github.com/ksctl/ksctl/pkg/types"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/application"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/helpers"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/scale"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/storage"
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

	if err := scale.CallManager(log, in); err != nil {
		log.Error(agentCtx, "CallManager", "Reason", err)
		return nil, status.Error(codes.Unimplemented, "failure from calling ksctl manager. Reason:"+err.Error())
	}

	log.Success(agentCtx, "Handled Scale")
	return &pb.ResScale{IsUpdated: true}, nil
}

func (s *server) Storage(ctx context.Context, in *pb.ReqStore) (*pb.ResStore, error) {

	// validate the request
	if in.Operation == pb.StorageOperation_EXPORT {
		log.Error(agentCtx, "Operation is export")
		return nil, status.Error(codes.Unimplemented, "operation is not supported")
	}

	v := in.Data
	exportedData := new(types.StorageStateExportImport)
	if err := json.Unmarshal(v, &exportedData); err != nil {
		log.Error(agentCtx, "Unable to Unmarshal exported data", "Reason", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	client := new(types.KsctlClient)

	if _err := storage.HandleStorageImport(ctx, log, exportedData, client); _err != nil {
		log.Error(agentCtx, "HandleStorageImport", "Reason", _err)
		return nil, status.Error(codes.FailedPrecondition, _err.Error())
	}

	log.Success(agentCtx, "all imports are done")
	return new(pb.ResStore), nil
}

func (s *server) LoadBalancer(ctx context.Context, in *pb.ReqLB) (*pb.ResLB, error) {
	log.Debug(agentCtx, "Request", "ReqLoadBalancer", in)

	log.Success(agentCtx, "Handled LoadBalancer")
	return nil, nil
}

func (s *server) Application(ctx context.Context, in *pb.ReqApplication) (*pb.ResApplication, error) {
	log.Debug(agentCtx, "Request", "ReqApplication", in)

	if err := application.Handler(log, in); err != nil {
		log.Error(agentCtx, "Handler", "Reason", err)
		return &pb.ResApplication{FailedApps: []string{err.Error()}}, nil
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
