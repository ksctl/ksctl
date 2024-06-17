package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
)

func infoClusters(ksctlClient *controllers.ManagerClusterKsctl) {
	l.Print(ctx, "Exec ksctl get...")

	_, err := ksctlClient.InfoCluster()
	if err != nil {
		l.Error("Failure", "err", err)
		os.Exit(1)
	}
}
