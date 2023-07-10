package main

import (
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/resources"
)

func NewCli(cmd *resources.CobraCmd) {
	version := os.Getenv("KSCTL_VERSION")

	if len(version) == 0 {
		version = "dummy v11001.2"
	}
}

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cmd := &resources.CobraCmd{ClusterName: "dummy-name", Region: "southindia"}
	NewCli(cmd)
	// HandleError(resources.NewCivoBuilderOrDie(cmd))
	HandleError(resources.NewAzureBuilderOrDie(cmd))
	// HandleError(resources.NewLocalBuilderOrDie(cmd))
	HandleError(resources.NewK3sBuilderOrDie(cmd))
	// HandleError(resources.NewKubeadmBuilderOrDie(cmd))
    HandleError(resources.NewLocalStorageBuilderOrDie(cmd))

	fmt.Println(cmd)
	fmt.Println(cmd.Client.Cloud)
	fmt.Println(cmd.Client.Distro)
	cmd.Client.Cloud.CreateVM() // it will fail if local is present
	// cmd.Client.Cloud.CreateManagedKubernetes()
	cmd.Client.Distro.ConfigureControlPlane()

    cmd.Client.State.Load("$HOME/demo/.ksctl/cred/civo.json")
}
