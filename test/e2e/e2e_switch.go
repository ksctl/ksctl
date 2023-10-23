package main

import "github.com/kubesimplify/ksctl/pkg/resources"

func switchCluster(ksctlClient *resources.KsctlClient) {

	l.Println("Exec ksctl switch...")

	//ksctlClient.Metadata.Provider = consts.CloudAll

	resp, err := ksctlManager.SwitchCluster(ksctlClient)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(resp)
}
