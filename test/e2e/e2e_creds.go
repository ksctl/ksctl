package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/controllers"
)

func creds(ksctlClient *controllers.ManagerClusterKsctl) {
	l.Print(ctx, "Exec ksctl creds...")

	err := ksctlClient.Credentials()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
