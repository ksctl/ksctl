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
		ClusterName: "demo",
		HACluster:   false,
		Region:      "eastus",
		Spec: util.Machine{
			Disk:                "Standard_B2s",
			ManagedNodes:        2,
			HAControlPlaneNodes: 3,
			HAWorkerNodes:       1,
			// Disk:                "Standard_F2s", for ha cluster
			// Disk:                "Standard_B1s", for ha cluster
			// Disk: "Standard_DS2_v2", for managed instances
			// Disk: "Standard_B2s", for managed instances
		},
	}
	fmt.Println("Enter [1] to create [0] to delete [2] for adding more worker nodes [3] to delete worker nodes")
	choice := -1
	fmt.Scanf("%d", &choice)
	switch choice {
	case 0:
		fmt.Println(payload.DeleteCluster())
	case 1:
		fmt.Println(payload.CreateCluster())
	case 3:
		fmt.Println(payload.DeleteSomeWorkerNodes())
	case 2:
		fmt.Println(payload.AddMoreWorkerNodes())
	}
}
