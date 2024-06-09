package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
)

func createManagedCluster(ksctlClient *controllers.ManagerClusterManaged) {
	l.Print(ctx, "Started to Create Cluster...")

	err := ksctlClient.CreateCluster()
	if err != nil {
		if ksctlErrors.ErrInvalidCloudProvider.Is(err) {
			l.Error("problem is invalid cloud provider")
		}
		if ksctlErrors.ErrInvalidResourceName.Is(err) {
			l.Error("problem from resource name")
		}
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}

func deleteManagedCluster(ksctlClient *controllers.ManagerClusterManaged) {
	l.Print(ctx, "Started to Delete Cluster...")

	err := ksctlClient.DeleteCluster()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}
