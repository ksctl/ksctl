package civo

import "fmt"

func (client *CloudController) CreateManagedCluster() {
	fmt.Println("Implement me[civo managed create]")

	currCloudState = nil
	currCloudState = &StateConfiguration{
		ClusterName: client.ClusterName,
	}
	client.Cloud.CreateManagedKubernetes()

	_, err := client.State.Load("civo.txt")
	fmt.Println(err)
}

func (client *CloudController) DestroyManagedCluster() {

	fmt.Println("Implement me[civo managed delete]")
}
