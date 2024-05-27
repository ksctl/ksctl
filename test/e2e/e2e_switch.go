package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
)

func switchCluster(ksctlClient *controllers.ManagerClusterKsctl) {

	l.Print(ctx, "Exec ksctl switch...")

	_, err := ksctlClient.SwitchCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
