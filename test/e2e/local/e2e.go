package main

import (
	"log"
	"os"
	"time"

	"github.com/kubesimplify/ksctl/pkg/controllers"
	"github.com/kubesimplify/ksctl/pkg/resources"
	ksctlController "github.com/kubesimplify/ksctl/pkg/resources/controllers"
	"github.com/kubesimplify/ksctl/test/e2e"
)

var (
	l            *log.Logger                = log.New(os.Stdout, "[local-e2e] ", -1)
	ksctlManager ksctlController.Controller = controllers.GenKsctlController()
)

func main() {
	// Define flags
	// Print the provided arguments
	l.Println("E2E testing starting...")
	timer := time.Now()
	operation, meta := e2e.GetReqPayload(l)

	ksctlClient := new(resources.KsctlClient)

	ksctlClient.Metadata = meta

	resp, err := controllers.InitializeStorageFactory(ksctlClient, true)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)

	switch operation {
	case e2e.OpCreate:
		createManagedCluster(ksctlClient)
	case e2e.OpDelete:
		deleteManagedCluster(ksctlClient)
	case e2e.OpGet:
		getClusters(ksctlClient)
	case e2e.OpSwitch:
		switchCluster(ksctlClient)
	default:
		l.Fatal("This operation is not supported")
	}

	l.Printf("E2E testing Completed  ⏰  %v\n", time.Since(timer))
}
