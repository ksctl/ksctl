package main

import "github.com/kubesimplify/ksctl/pkg/resources"

func getClusters(ksctlClient *resources.KsctlClient) {
	l.Println("Exec ksctl get...")

	resp, err := ksctlManager.GetCluster(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}
