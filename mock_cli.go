package main

import (
	"fmt"
	"os"

	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/cli"
)

func NewCli(cmd *cli.CobraCmd) {
	version := os.Getenv("KSCTL_VERSION")

	if len(version) == 0 {
		version = "dummy v11001.2"
	}

	cmd.Client.Client = cli.Builder{}
}

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cmd := &cli.CobraCmd{ClusterName: "dummy-name", Region: "southindia"}
	NewCli(cmd)
	err := resources.NewCivoBuilderOrDie(cmd)
	HandleError(err)
	fmt.Println(cmd)
}
