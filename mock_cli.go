package main

import (
	"os"

	control_pkg "github.com/kubesimplify/ksctl/api/controllers"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers"
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

	var controller controllers.Controller = control_pkg.GenKsctlController()
	controller.CreateHACluster(cmd.Client)

	// HandleError(cli.NewCivoBuilderOrDie(cmd))
	// HandleError(cli.NewK3sBuilderOrDie(cmd))
	// HandleError(cli.NewLocalStorageBuilderOrDie(cmd))
	// cmd.Client.IsHA = true // set by CMD

	// controllers.NewController(&cmd.Client)
}
