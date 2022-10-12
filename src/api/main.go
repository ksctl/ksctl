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
	civoHandler "github.com/kubesimplify/Kubesimpctl/src/api/civo"
)

func main() {

	fmt.Println("Enter 1 to create and 2 to delete")
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		panic(err)
	}
	switch choice {
	case 1:
		clusterConfig := civoHandler.ClusterInfoInjecter("demo", "FRA1", "g4s.kube.xsmall", 1)
		err := civoHandler.CreateCluster(clusterConfig)
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mCREATED!\033[0m\n")
	case 2:
		err = civoHandler.DeleteCluster("FRA1", "demo")
		if err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mDELETED!\033[0m\n")
	default:
		fmt.Println("INVALID CHOICE")
	}
}
