package main

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/azure"
	util "github.com/kubesimplify/ksctl/api/utils"
)

func main() {
	var payload azure.AzureOperations
	// what all things users can config
	// also add the node type

	// add option to select size of vm
	payload = &azure.AzureProvider{
		ClusterName: "demo-123",
		HACluster:   true,
		Region:      "eastus",
		Spec: util.Machine{
			ManagedNodes:        2,
			HAControlPlaneNodes: 1,
			HAWorkerNodes:       2,
		},
	}
	fmt.Println("Enter [1] to create [0] to delete")
	choice := -1
	fmt.Scanf("%d", &choice)
	switch choice {
	case 0:
		fmt.Println(payload.DeleteCluster())
	case 1:
		fmt.Println(payload.CreateCluster())
	}
}
