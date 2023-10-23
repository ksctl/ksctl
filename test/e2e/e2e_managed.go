package main

import "github.com/kubesimplify/ksctl/pkg/resources"

func createManagedCluster(ksctlClient *resources.KsctlClient) {
	l.Println("Started to Create Cluster...")

	resp, err := ksctlManager.CreateManagedCluster(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}

func deleteManagedCluster(ksctlClient *resources.KsctlClient) {
	l.Println("Started to Delete Cluster...")

	resp, err := ksctlManager.DeleteManagedCluster(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}
