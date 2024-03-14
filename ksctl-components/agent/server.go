package main

import (
	"context"
	"github.com/ksctl/ksctl/ksctl-components/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/scale"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
)

type server struct {
	pb.UnimplementedKsctlAgentServer
}

func (s *server) Scale(ctx context.Context, in *pb.ReqScale) (*pb.ResScale, error) {
	slog.Debug("WorkerNodes", "count", "DUMMY")
	slog.Debug("Request", "reqScale", in)

	slog.Info("Processing Scale Request", "operation", in.Operation, "desired", in.ScaleTo)

	if err := scale.CallManager(string(in.Operation)); err != nil {
	}

	return &pb.ResScale{UpdatedWP: 999}, nil
}

func (s *server) LoadBalancer(ctx context.Context, in *pb.ReqLB) (*pb.ResLB, error) {
	return nil, nil
}

func (s *server) Application(ctx context.Context, in *pb.ReqApplication) (*pb.ResApplication, error) {
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
