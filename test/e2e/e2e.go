package main

import (
	"context"
	"os"
	"time"

	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

var (
	l   resources.LoggerFactory
	ctx = context.WithValue(context.Background(), "USERID", "e2e")
)

func main() {
	timer := time.Now()
	l = logger.NewStructuredLogger(-1, os.Stdout)

	operation, meta := GetReqPayload(l)

	l.Print(ctx, "Testing starting...")
	ksctlClient, err := controllers.GenKsctlController(
		ctx, l, &resources.KsctlClient{
			Metadata: meta,
		})
	if err != nil {
		l.Error(ctx, "unable to initialize the ksctl manager", "Reason", err)
		os.Exit(1)
	}

	switch operation {
	case OpCreate:
		if meta.IsHA {
			createHACluster(ksctlClient)
		} else {
			createManagedCluster(ksctlClient)
		}
	case OpDelete:
		if meta.IsHA {
			deleteHACluster(ksctlClient)
		} else {
			deleteManagedCluster(ksctlClient)
		}
	case OpCreds:
		creds(ksctlClient)
	case OpGet:
		getClusters(ksctlClient)
	case OpSwitch:
		switchCluster(ksctlClient)
	case OpScaleUp:
		scaleupHACluster(ksctlClient)
	case OpScaleDown:
		scaleDownHACluster(ksctlClient)
	default:
		l.Error(ctx, "This operation is not supported")
		os.Exit(1)
	}

	l.Print(ctx, "Testing Completed", "TimeTaken", time.Since(timer).String())
}
