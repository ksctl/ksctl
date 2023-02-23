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
	// cp = 2-cp,1wp name demo-cp
	payload = &azure.AzureProvider{
		ClusterName: "demo-dipankar",
		HACluster:   true,
		Region:      "eastus",
		Spec: util.Machine{
			ManagedNodes:        2,
			HAControlPlaneNodes: 2,
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
