package main

import (
	ksctlController "github.com/ksctl/ksctl/pkg/types/controllers"
	"os"
)

func creds(ksctlClient ksctlController.Controller) {
	l.Print(ctx, "Exec ksctl creds...")

	err := ksctlClient.Credentials()
	if err != nil {
		l.Error(ctx, "Failure", "err", err)
		os.Exit(1)
	}
}
