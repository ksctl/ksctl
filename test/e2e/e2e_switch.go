package main

import (
	ksctlController "github.com/ksctl/ksctl/pkg/resources/controllers"
	"os"
)

func switchCluster(ksctlClient ksctlController.Controller) {

	l.Print(ctx, "Exec ksctl switch...")

	_, err := ksctlClient.SwitchCluster()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
