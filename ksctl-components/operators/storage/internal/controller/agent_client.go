package controller

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
)

func NewClient(ctx context.Context) (pb.KsctlAgentClient, *grpc.ClientConn, error) {
	ksctlAgentUrl := os.Getenv("KSCTL_AGENT_URL")
	opts := []grpc.DialOption{
		// grpc.WithTransportCredentials(creds),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, ksctlAgentUrl, opts...)
	if err != nil {
		//slog.Error("fail to dial", "reason", err)
		return nil, conn, err
	}

	return pb.NewKsctlAgentClient(conn), conn, nil
}

func ImportData(ctx context.Context, client pb.KsctlAgentClient, data []byte) error {
	_, err := client.Storage(ctx, &pb.ReqStore{Operation: pb.StorageOperation_IMPORT, Data: data})
	fmt.Println("Errors from grpc storage Import: ", err)
	if err != nil {
		return err
	}
	return nil
}
