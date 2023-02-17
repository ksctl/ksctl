package main

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/azure"
	"github.com/kubesimplify/ksctl/api/utils"
)

func main() {
	var payload azure.AzureOperations
	payload = &azure.AzureProvider{
		ClusterName: "dipankar-demo",
		HACluster:   false,
		Region:      "westus",
		Spec: utils.Machine{
			ManagedNodes: 1,
		},
	}
	fmt.Println("Enter [1] to create [0] to delete")
	choice := -1
	fmt.Scanf("%d", &choice)
	switch choice {
	case 0:
		payload.DeleteCluster()
	case 1:
		payload.CreateCluster()
	}
}
