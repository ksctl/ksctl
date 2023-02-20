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
	payload = &azure.AzureProvider{
		ClusterName: "demo",
		HACluster:   true,
		Region:      "westus",
		Spec: util.Machine{
			ManagedNodes:        2,
			HAControlPlaneNodes: 1,
			HAWorkerNodes:       1,
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
