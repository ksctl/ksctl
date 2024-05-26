package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
)

func createManagedCluster(ksctlClient *controllers.ManagerClusterManaged) {
	l.Print(ctx, "Started to Create Cluster...")

	err := ksctlClient.CreateCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}

func deleteManagedCluster(ksctlClient *controllers.ManagerClusterManaged) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DeleteCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
