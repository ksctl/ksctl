package main

import (
	"context"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
)

var (
	l   types.LoggerFactory
	ctx context.Context
)

func main() {
	ctx = context.WithValue(
		context.Background(),
		"USERID",
		"e2e",
	)
	// ctx = context.WithValue(
	// 	ctx,
	// 	consts.KsctlCustomDirLoc,
	// 	fmt.Sprintf("%s ksctl-e2e", os.TempDir()),
	// )

	timer := time.Now()

	flags := os.Getenv("E2E_FLAGS")
	rFlags := strings.Split(flags, ",")

	verbosityLevel := 0

	if slices.Contains[[]string, string](rFlags, "debug") {
		verbosityLevel = -1
	}

	if slices.Contains[[]string, string](rFlags, "new_logging") {
		l = logger.NewGeneralLogger(verbosityLevel, os.Stdout)
	} else {
		l = logger.NewStructuredLogger(verbosityLevel, os.Stdout)
	}

	operation, meta := GetReqPayload(l)

	l.Print(ctx, "Testing starting...")

	switch operation {
	case OpCreate, OpDelete, OpScaleUp, OpScaleDown:
		switch meta.IsHA {
		case true:

			if meta.Provider == consts.CloudLocal {
				err := l.NewError(ctx, "ha not supported for local")
				l.Error(ctx, "handled error", "catch", err)
				os.Exit(1)
			}
			managerClient, err := controllers.NewManagerClusterSelfManaged(
				ctx,
				l,
				&types.KsctlClient{
					Metadata: meta,
				},
			)
			if err != nil {
				l.Error(ctx, "unable to initialize the ksctl manager", "Reason", err)
				os.Exit(1)
			}
			switch operation {
			case OpCreate:
				_ = managerClient.CreateCluster()
			case OpDelete:
				_ = managerClient.DeleteCluster()
			case OpScaleUp:
				_ = managerClient.AddWorkerPlaneNodes()
			case OpScaleDown:
				_ = managerClient.DelWorkerPlaneNodes()
			}

		case false:
			managerClient, err := controllers.NewManagerClusterManaged(
				ctx,
				l,
				&types.KsctlClient{
					Metadata: meta,
				},
			)
			if err != nil {
				l.Error(ctx, "unable to initialize the ksctl manager", "Reason", err)
				os.Exit(1)
			}
			switch operation {
			case OpCreate:
				_ = managerClient.CreateCluster()
			case OpDelete:
				_ = managerClient.DeleteCluster()
			}
		}

	case OpCreds, OpGet, OpSwitch:
		managerClient, err := controllers.NewManagerClusterKsctl(
			ctx,
			l,
			&types.KsctlClient{
				Metadata: meta,
			},
		)
		if err != nil {
			l.Error(ctx, "unable to initialize the ksctl manager", "Reason", err)
			os.Exit(1)
		}
		switch operation {
		case OpCreds:
			_ = managerClient.Credentials()
		case OpGet:
			_ = managerClient.GetCluster()
		case OpSwitch:
			_, _ = managerClient.SwitchCluster()
		}

	default:
		l.Error(ctx, "This operation is not supported")
		os.Exit(1)
	}

	l.Print(ctx, "Testing Completed", "TimeTaken", time.Since(timer).String())
}
