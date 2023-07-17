package main

import (
	"os"

	"github.com/kubesimplify/ksctl/api/controllers"
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
	HandleError(cli.NewLocalBuilderOrDie(cmd))
	HandleError(cli.NewKubeadmBuilderOrDie(cmd))
	HandleError(cli.NewLocalStorageBuilderOrDie(cmd))
	cmd.Client.IsHA = true // set by CMD

	controllers.NewController(&cmd.Client)
}
