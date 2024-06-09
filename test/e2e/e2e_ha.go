package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
)

func createHACluster(ksctlClient *controllers.ManagerClusterSelfManaged) {
	l.Print(ctx, "Started to Create Cluster...")

	err := ksctlClient.CreateCluster()

	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}

func deleteHACluster(ksctlClient *controllers.ManagerClusterSelfManaged) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DeleteCluster()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}

func scaleupHACluster(ksctlClient *controllers.ManagerClusterSelfManaged) {
	l.Print(ctx, "Started to scaleup Cluster...")

	err := ksctlClient.AddWorkerPlaneNodes()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}

func scaleDownHACluster(ksctlClient *controllers.ManagerClusterSelfManaged) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DelWorkerPlaneNodes()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}
