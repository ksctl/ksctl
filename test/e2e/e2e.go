package main

import (
	"os"
	"time"

	"github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/logger"
	"github.com/kubesimplify/ksctl/pkg/resources"
	ksctlController "github.com/kubesimplify/ksctl/pkg/resources/controllers"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

var (
	l            resources.LoggerFactory
	ksctlManager ksctlController.Controller = controllers.GenKsctlController()
)

func main() {
	timer := time.Now()
	l = logger.NewDefaultLogger(-1, os.Stdout)

	operation, meta := GetReqPayload(l)

	loggerPrefix := ""
	switch meta.Provider {
	case consts.CloudAws:
		loggerPrefix = "aws-e2e"
	case consts.CloudCivo:
		loggerPrefix = "civo-e2e"
	case consts.CloudAzure:
		loggerPrefix = "azure-e2e"
	case consts.CloudLocal:
		loggerPrefix = "local-e2e"
	case consts.CloudAll:
		loggerPrefix = "all-e2e"
	}

	l.SetPackageName(loggerPrefix)

	l.Print("Testing starting...")
	ksctlClient := new(resources.KsctlClient)

	ksctlClient.Metadata = meta

	err := controllers.InitializeStorageFactory(ksctlClient, true)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	switch operation {
	case OpCreate:
		if ksctlClient.Metadata.IsHA {
			createHACluster(ksctlClient)
		} else {
			createManagedCluster(ksctlClient)
		}
	case OpDelete:
		if ksctlClient.Metadata.IsHA {
			deleteHACluster(ksctlClient)
		} else {
			deleteManagedCluster(ksctlClient)
		}
	case OpGet:
		getClusters(ksctlClient)
	case OpSwitch:
		switchCluster(ksctlClient)
	case OpScaleUp:
		scaleupHACluster(ksctlClient)
	case OpScaleDown:
		scaleDownHACluster(ksctlClient)
	default:
		l.Error("This operation is not supported")
		os.Exit(1)
	}

	l.Print("Testing Completed", " ⏰ ", time.Since(timer))
}
