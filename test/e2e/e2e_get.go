package main

import (
	ksctlController "github.com/ksctl/ksctl/pkg/types/controllers"
	"os"
)

func getClusters(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Exec ksctl get...")

	err := ksctlClient.GetCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
