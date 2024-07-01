package main

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"gotest.tools/v3/assert"
)

func TestMain(m *testing.M) {
	log = logger.NewStructuredLogger(-1, os.Stdout)
	agentCtx = context.WithValue(
		context.TODO(),
		consts.KsctlTestFlagKey,
		"true",
	)
	m.Run()
}

func TestVariables(t *testing.T) {
	assert.Check(t, agentCtx != nil)

	assert.Equal(t, helpers.LogVerbosity["DEBUG"], -1)
	assert.Equal(t, helpers.LogVerbosity[""], 0)
	if v, ok := helpers.LogVerbosity["erdrvf"]; ok {
		// it should be absent
		t.Fatalf("This annonomous verbosity convertion is not intended. %v -> %v", "erdrvf", v)
	}
}

func TestApplication(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() {
		lis.Close()
	})
	srv := grpc.NewServer()
	t.Cleanup(func() {
		srv.Stop()
	})

	svc := server{}

	pb.RegisterKsctlAgentServer(srv, &svc)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Error("server serve failed", "Reason", err)
			os.Exit(1)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.DialContext(agentCtx)
	}

	conn, err := grpc.DialContext(agentCtx, "", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))

	t.Cleanup(func() {
		conn.Close()
	})
	if err != nil {
		t.Fatalf("grpc.DialContext %v", err)
	}

	client := pb.NewKsctlAgentClient(conn)

	t.Run("applications are used", func(t *testing.T) {
		res, err := client.Application(agentCtx, &pb.ReqApplication{
			Operation: pb.ApplicationOperation_CREATE,
			Apps: []*pb.Application{
				{
					AppName: "istio",
					AppType: pb.ApplicationType_APP,
					Version: "23e2w",
				},
			},
		})
		if err != nil {
			t.Fatalf("grpc.client.Application %v", err)
		}

		assert.DeepEqual(t, res.FailedApps, []string{"none"})
	})

	t.Run("empty requestApplications", func(t *testing.T) {
		_, err := client.Application(agentCtx, &pb.ReqApplication{
			Operation: pb.ApplicationOperation_CREATE,
		})

		if err == nil {
			t.Fatalf("It should have failed")
		}
		t.Logf("recived from server. Err: %v", err)
	})
}
