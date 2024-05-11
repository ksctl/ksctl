package main

import (
	ksctlController "github.com/ksctl/ksctl/pkg/types/controllers"
	"os"
)

func createManagedCluster(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Started to Create Cluster...")

	err := ksctlClient.CreateManagedCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}

func deleteManagedCluster(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DeleteManagedCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
