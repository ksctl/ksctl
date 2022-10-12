//
// ######################################
//		EARLY DEVELOPMENT
//			CREATE & DELETE
//   NOTE: THIS FILE WILL BE REMOVED IN NEXT VERSION
// ######################################
//

package main

import (
	"fmt"
	"github.com/civo/civogo"
	civoHandler "github.com/kubesimplify/Kubesimpctl/src/api/civo"
	"strings"
	"time"
)

func main() {
	clusterID := civoHandler.CreateCluster(civoHandler.RegionLON, "demo")
	client, err := civogo.NewClient(civoHandler.FetchAPIKey(), civoHandler.RegionLON)
	if err != nil {
		panic(err)
	}
	if strings.Compare(clusterID, civoHandler.ERRORCODE) == 0 {
		panic(fmt.Errorf("[FAILED] create-cluster").Error())
	}

	for true {
		clusterDS, _ := client.GetKubernetesCluster(clusterID)
		if clusterDS.Ready {
			//print the new KUBECONFIG
			fmt.Println(clusterDS.KubeConfig)
			break
		}
		fmt.Printf("Waiting.. Status: %v\n", clusterDS.Status)
		time.Sleep(15 * time.Second)
	}

	choice := byte(' ')
	fmt.Println("Do you want to remove the cluster Y/N..")
	_, err = fmt.Scanf("%c", &choice)
	if err != nil {
		return
	}
	if choice == 'Y' || choice == 'y' {
		fmt.Println(civoHandler.DeleteCluster(civoHandler.RegionLON, clusterID))
	}
}
