package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"

	"github.com/ksctl/ksctl/pkg/resources"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
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

// NOTE: assumption going to use
//
//	os.stdout as writer
//	LOG_LEVEL env variable for verbosity

func (s *server) Scale(ctx context.Context, in *pb.ReqScale) (*pb.ResScale, error) {
	slog.DebugContext(ctx, "Request", "reqScale", in)

	slog.InfoContext(ctx, "Processing Scale Request", "operation", in.Operation, "desired", in.DesiredNoOfWP)

	if err := scale.CallManager(in); err != nil {
		return nil, status.Error(codes.Unimplemented, "failure from calling ksctl manager. Reason:"+err.Error())
	}

	return &pb.ResScale{IsUpdated: true}, nil
}

func (s *server) Storage(ctx context.Context, in *pb.ReqStore) (*pb.ResStore, error) {
	slog.DebugContext(ctx, "Request", "reqStore", in)

	// validate the request
	if in.Operation == pb.StorageOperation_EXPORT {
		return nil, status.Error(codes.Unimplemented, "operation is not supported")
	}

	v := in.Data
	exportedData := new(resources.StorageStateExportImport)
	if err := json.Unmarshal(v, &exportedData); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	client := new(resources.KsctlClient)

	if _err := storage.NewStorageClient(ctx, client); _err != nil {
		return nil, status.Error(codes.FailedPrecondition, _err.Error())
	}

	if _err := storage.HandleImport(client, exportedData); _err != nil {
		return nil, status.Error(codes.Internal, _err.Error())
	}

	slog.Info("all imports are done")
	return new(pb.ResStore), nil
}

func (s *server) LoadBalancer(ctx context.Context, in *pb.ReqLB) (*pb.ResLB, error) {
	slog.DebugContext(ctx, "Request", "ReqLB", in)
	return nil, nil
}

func (s *server) Application(ctx context.Context, in *pb.ReqApplication) (*pb.ResApplication, error) {
	slog.DebugContext(ctx, "Request", "ReqApplication", in)
	return nil, nil
}

type Health struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (h Health) Check(ctx context.Context, g *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	slog.InfoContext(ctx, "serving health", "grpc_health_v1", g.String())
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("unable to do http listener", "err", err)
	}

	s := grpc.NewServer()
	defer s.Stop()
	//hs := health.NewServer()                   // will default to respond with SERVING
	//grpc_health_v1.RegisterHealthServer(s, hs) // registration
	grpc_health_v1.RegisterHealthServer(s, &Health{}) // registration

	reflection.Register(s) // for debugging purposes

	pb.RegisterKsctlAgentServer(s, &server{}) // Register the server with the gRPC server

	slog.Info("Server started", "port", "8080")

	if err := s.Serve(listener); err != nil {
		slog.Error("failed to serve", "err", err)
	}
}
