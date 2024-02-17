package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/resources"
)

func createManagedCluster(ksctlClient *resources.KsctlClient) {
	l.Print("Started to Create Cluster...")

	err := ksctlManager.CreateManagedCluster(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}

func deleteManagedCluster(ksctlClient *resources.KsctlClient) {
	l.Print("Started to Delete Cluster...")

	err := ksctlManager.DeleteManagedCluster(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}
