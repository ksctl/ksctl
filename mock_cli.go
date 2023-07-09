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
	err := resources.NewCivoBuilderOrDie(cmd)
	HandleError(err)
	fmt.Println(cmd)
}
