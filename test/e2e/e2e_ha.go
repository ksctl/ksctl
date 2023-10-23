package main

import "github.com/kubesimplify/ksctl/pkg/resources"

func createHACluster(ksctlClient *resources.KsctlClient) {
	l.Println("Started to Create Cluster...")

	resp, err := ksctlManager.CreateHACluster(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}

func deleteHACluster(ksctlClient *resources.KsctlClient) {
	l.Println("Started to Delete Cluster...")

	resp, err := ksctlManager.DeleteHACluster(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}

func scaleupHACluster(ksctlClient *resources.KsctlClient) {
	l.Println("Started to scaleup Cluster...")

	resp, err := ksctlManager.AddWorkerPlaneNode(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}

func scaleDownHACluster(ksctlClient *resources.KsctlClient) {
	l.Println("Started to Delete Cluster...")

	resp, err := ksctlManager.DelWorkerPlaneNode(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}
