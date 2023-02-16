package main

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/azure"
)

func main() {
	var payload azure.AzureOperations
	payload = &azure.AzureProvider{
		ClusterName: "demo",
		HACluster:   false,
		Region:      "abcd",
	}
	fmt.Println("Enter [1] to create [0] to delete")
	choice := -1
	switch choice {
	case 0:
		payload.DeleteCluster()
	case 1:
		payload.CreateCluster()
	}
}
