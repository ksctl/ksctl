package main

import (
	"os"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

func createHACluster(ksctlClient *resources.KsctlClient) {
	l.Print("Started to Create Cluster...")

	err := ksctlManager.CreateHACluster(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}

func deleteHACluster(ksctlClient *resources.KsctlClient) {
	l.Print("Started to Delete Cluster...")

	err := ksctlManager.DeleteHACluster(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}

func scaleupHACluster(ksctlClient *resources.KsctlClient) {
	l.Print("Started to scaleup Cluster...")

	err := ksctlManager.AddWorkerPlaneNode(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}

func scaleDownHACluster(ksctlClient *resources.KsctlClient) {
	l.Print("Started to Delete Cluster...")

	err := ksctlManager.DelWorkerPlaneNode(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}
