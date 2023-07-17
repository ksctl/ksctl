package main

import (
	"os"

	controller "github.com/kubesimplify/ksctl/api"
	"github.com/kubesimplify/ksctl/api/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/cli"
)

func NewCli(cmd *resources.CobraCmd) {
	cmd.Version = os.Getenv("KSCTL_VERSION")

	if len(cmd.Version) == 0 {
		cmd.Version = "dummy v11001.2"
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
	HandleError(cli.NewCivoBuilderOrDie(cmd))
	HandleError(cli.NewK3sBuilderOrDie(cmd))
	HandleError(cli.NewLocalStorageBuilderOrDie(cmd))

	client := cmd.Client
	api := cloud.ClientBuilder(client)
	controller.NewController(&api)

	// fmt.Println(cmd)
	// fmt.Println(cmd.Client.Cloud)
	// fmt.Println(cmd.Client.Distro)
	// cmd.Client.Cloud.CreateVM() // it will fail if local is present
	// // cmd.Client.Cloud.CreateManagedKubernetes()
	// cmd.Client.Distro.ConfigureControlPlane()
	//
	// cmd.Client.State.Load("$HOME/demo/.ksctl/cred/civo.json")
}
