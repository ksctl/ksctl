package main

import (
	"fmt"
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
	cmd := &resources.CobraCmd{}
	NewCli(cmd)

	cmd.Client.Metadata.Provider = "civo"
	cmd.Client.Metadata.K8sDistro = "k3s"
	cmd.Client.Metadata.StateLocation = "local"
	cmd.Client.Metadata.ClusterName = "dummy-name"
	cmd.Client.Metadata.ManagedNodeType = "g4s.kube.medium"

	cmd.Client.Metadata.Region = "LON1"
	cmd.Client.Metadata.NoCP = 5
	cmd.Client.Metadata.NoWP = 5
	cmd.Client.Metadata.NoDS = 3

	var controller controllers.Controller = control_pkg.GenKsctlController()
	choice := -1
	fmt.Println(`
[0] enter credential
[1] create HA
[2] Delete HA
[3] Create Managed
[4] Delete Managed

Your Choice`)
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return
	}
	switch choice {
	case 0:
		controller.Credentials(&cmd.Client)
	case 1:
		cmd.Client.Metadata.IsHA = true
		controller.CreateHACluster(&cmd.Client)
	case 2:
		cmd.Client.Metadata.IsHA = true

		controller.DeleteHACluster(&cmd.Client)
	case 3:
		cmd.Client.Metadata.NoWP = 1
		controller.CreateManagedCluster(&cmd.Client)
	case 4:
		controller.DeleteManagedCluster(&cmd.Client)
	}
}
