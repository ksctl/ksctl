package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

func (client *CloudController) CreateHACluster() {

	fmt.Println("Implement me[civo ha create]")
	currCloudState = &StateConfiguration{
		ClusterName: client.ClusterName,
		Region:      client.Region,
		K8s: cloud.CloudResourceState{
			SSHState: cloud.SSHPayload{UserName: "root"},
			Metadata: cloud.Metadata{
				ClusterName: client.ClusterName,
				Region:      client.Region,
				Provider:    "civo",
			},
		},
	}
	err := client.State.Save("civo.txt", nil)
	fmt.Println(err)
	client.Distro.ConfigureControlPlane()
}

func (client *CloudController) DestroyHACluster() {
	fmt.Println("Implement me[civo ha delete]")
}
