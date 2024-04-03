package main

import (
	"context"
	"log/slog"
	"net"

	"github.com/ksctl/ksctl/ksctl-components/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/scale"
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

	return &pb.ResScale{ActualNoOfWP: 999}, nil
}

func (s *server) LoadBalancer(ctx context.Context, in *pb.ReqLB) (*pb.ResLB, error) {
	slog.DebugContext(ctx, "Request", "ReqLB", in)
	return nil, nil
}

func (s *server) Application(ctx context.Context, in *pb.ReqApplication) (*pb.ResApplication, error) {
	slog.DebugContext(ctx, "Request", "ReqApplication", in)
	return nil, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("unable to do http listener", "err", err)
	}

	s := grpc.NewServer()
	reflection.Register(s) // for debugging purposes

	pb.RegisterKsctlAgentServer(s, &server{}) // Register the server with the gRPC server

	slog.Info("Server started", "port", "8080")

	if err := s.Serve(listener); err != nil {
		slog.Error("failed to serve", "err", err)
	}
}
