package controller

import (
	"context"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func NewClient(ctx context.Context) (pb.KsctlAgentClient, *grpc.ClientConn, error) {
	ksctlAgentUrl := os.Getenv("KSCTL_AGENT_URL")
	opts := []grpc.DialOption{
		// grpc.WithTransportCredentials(creds),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if strings.Compare(os.Getenv(string(consts.KsctlFakeFlag)), ControllerTestSkip) == 0 { // to ecape test
		return pb.NewKsctlAgentClient(&grpc.ClientConn{}), &grpc.ClientConn{}, nil
	}

	conn, err := grpc.DialContext(ctx, ksctlAgentUrl, opts...)
	if err != nil {
		return nil, conn, err
	}

	return pb.NewKsctlAgentClient(conn), conn, nil
}

func ImportData(ctx context.Context, client pb.KsctlAgentClient, data []byte) error {

	if strings.Compare(os.Getenv(string(consts.KsctlFakeFlag)), ControllerTestSkip) == 0 { // to ecape test
		return nil
	}
	_, err := client.Storage(ctx, &pb.ReqStore{Operation: pb.StorageOperation_IMPORT, Data: data})
	if err != nil {
		return err
	}
	return nil
}
