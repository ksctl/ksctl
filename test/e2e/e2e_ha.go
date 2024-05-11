package main

import (
	ksctlController "github.com/ksctl/ksctl/pkg/resources/controllers"
	"os"
)

func createHACluster(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Started to Create Cluster...")

	err := ksctlClient.CreateHACluster()

	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}

func deleteHACluster(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DeleteHACluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}

func scaleupHACluster(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Started to scaleup Cluster...")

	err := ksctlClient.AddWorkerPlaneNode()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}

func scaleDownHACluster(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DelWorkerPlaneNode()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
