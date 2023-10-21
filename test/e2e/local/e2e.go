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
	l            *log.Logger                = log.New(os.Stdout, "[local] ", -1)
	ksctlManager ksctlController.Controller = controllers.GenKsctlController()
)

const (
	clusterNameManaged = "test-e2e-local"
	noOfNodes          = 2
)

// TODO: we can use fuzzy testing
func main() {
	timer := time.Now()

	l.Println("E2E testing starting...")

	ksctlClient := new(resources.KsctlClient)
	ksctlClient.Metadata.ClusterName = clusterNameManaged
	ksctlClient.Metadata.NoMP = noOfNodes
	ksctlClient.Metadata.Provider = consts.CloudLocal
	ksctlClient.Metadata.StateLocation = consts.StoreLocal

	ksctlClient.Metadata.IsHA = false
	ksctlClient.Metadata.K8sVersion = "1.27.1"

	resp, err := controllers.InitializeStorageFactory(ksctlClient, true)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)

	createManagedCluster(ksctlClient)

	getClusters(ksctlClient)

	switchCluster(ksctlClient)

	deleteManagedCluster(ksctlClient)

	l.Printf("E2E testing Completed  ⏰  %v\n", time.Since(timer))
}
