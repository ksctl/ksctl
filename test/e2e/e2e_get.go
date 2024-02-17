package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/resources"
)

func getClusters(ksctlClient *resources.KsctlClient) {
	l.Print("Exec ksctl get...")

	err := ksctlManager.GetCluster(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}
