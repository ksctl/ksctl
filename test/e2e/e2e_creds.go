package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/resources"
)

func creds(ksctlClient *resources.KsctlClient) {
	l.Print("Exec ksctl creds...")

	err := ksctlManager.Credentials(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}
