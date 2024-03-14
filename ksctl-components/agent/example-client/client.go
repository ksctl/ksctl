package example_client

import (
	// "crypto/tls"

	"context"
	"github.com/ksctl/ksctl/ksctl-components/agent/pb"
	"log/slog"
	"time"

	"google.golang.org/grpc"

	// "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})

	opts := []grpc.DialOption{
		// grpc.WithTransportCredentials(creds),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	url := "127.0.0.1:8080"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, url, opts...)
	if err != nil {
		slog.Error("fail to dial", "reason", err)
	}
	defer conn.Close()

	client := pb.NewKsctlAgentClient(conn)

	res, err := client.Scale(ctx, &pb.ReqScale{Operation: pb.ScaleOperation_UP})
	if err != nil {
		slog.Error("failed to do Scale Operation", "reason", err)
	}
	slog.Info("Result", "NewWorkerNodes", res.UpdatedWP)
}
