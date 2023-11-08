package main

import (
	"os"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

func switchCluster(ksctlClient *resources.KsctlClient) {

	l.Print("Exec ksctl switch...")

	//ksctlClient.Metadata.Provider = consts.CloudAll

	err := ksctlManager.SwitchCluster(ksctlClient)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
}
