package main

import (
	"context"
	"github.com/ksctl/ksctl/ksctl-components/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/scale"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
)

type server struct {
	pb.UnimplementedKsctlAgentServer
}

func (s *server) Scale(ctx context.Context, in *pb.ReqScale) (*pb.ResScale, error) {
	slog.DebugContext(ctx, "WorkerNodes", "count", "DUMMY")
	slog.DebugContext(ctx, "Request", "reqScale", in)

	slog.InfoContext(ctx, "Processing Scale Request", "operation", in.Operation, "desired", in.ScaleTo)

	// figure out the how the data will be written to the logs
	if err := scale.CallManager(string(in.Operation), in); err != nil {
		return nil, status.Error(codes.Unimplemented, "failure from calling ksctl manager. Reason:"+err.Error())
	}

	return &pb.ResScale{UpdatedWP: 999}, nil
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
