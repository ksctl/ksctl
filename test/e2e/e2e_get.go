package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
)

func getClusters(ksctlClient *controllers.ManagerClusterKsctl) {
	l.Print(ctx, "Exec ksctl get...")

	err := ksctlClient.GetCluster()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}
