package main

import (
	"log"
	"os"
	"time"

	"github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	ksctlController "github.com/kubesimplify/ksctl/pkg/resources/controllers"
	"github.com/kubesimplify/ksctl/pkg/utils/consts"
)

var (
	l            *log.Logger
	ksctlManager ksctlController.Controller = controllers.GenKsctlController()
)

func main() {
	timer := time.Now()

	operation, meta := GetReqPayload(l)

	loggerPrefix := ""
	switch meta.Provider {
	case consts.CloudAws:
		loggerPrefix = "[aws-e2e]"
	case consts.CloudCivo:
		loggerPrefix = "[civo-e2e]"
	case consts.CloudAzure:
		loggerPrefix = "[azure-e2e]"
	case consts.CloudLocal:
		loggerPrefix = "[local-e2e]"
	case consts.CloudAll:
		loggerPrefix = "[all-e2e]"
	}

	l = log.New(os.Stdout, loggerPrefix, -1)

	l.Println("Testing starting...")
	ksctlClient := new(resources.KsctlClient)

	ksctlClient.Metadata = meta

	resp, err := controllers.InitializeStorageFactory(ksctlClient, true)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)

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
		l.Fatal("This operation is not supported")
	}

	l.Printf("Testing Completed  ⏰  %v\n", time.Since(timer))
}
