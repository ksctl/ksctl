package local

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type StateConfiguration struct {
	ClusterName string `json:"cluster_name"`
}

type CloudController cloud.ClientBuilder

// FetchState implements cloud.ControllerInterface.
func (client *CloudController) FetchState() cloud.CloudResourceState {
	return cloud.CloudResourceState{
		Metadata: cloud.Metadata{
			ClusterName: client.ClusterName,
			Region:      client.Region,
			Provider:    "local",
		}}
}

func WrapCloudControllerBuilder(b *cloud.ClientBuilder) *CloudController {
	local := (*CloudController)(b)
	return local
}

func (client *CloudController) CreateHACluster() {
	panic("NO SUPPORT")
	// fmt.Println("Implement me[local ha create]")
	// err := client.State.Save("local.txt", nil)
	// fmt.Println(err)
	// client.Distro.ConfigureControlPlane() // no support for custom kubernetes distros
}

func (client *CloudController) CreateManagedCluster() {
	fmt.Println("Implement me[local managed create]")

	client.Cloud.CreateManagedKubernetes()

	err := client.State.Save("local.txt", nil)
	fmt.Println(err)
}

func (client *CloudController) DestroyHACluster() {
	panic("NO SUPPORT")
}

func (client *CloudController) DestroyManagedCluster() {

	fmt.Println("Implement me[local managed delete]")
}
