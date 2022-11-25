//
// ######################################
//		EARLY DEVELOPMENT [STUB AND DRIVER CONFIG]
//   NOTE: THIS FILE WILL BE REMOVED IN NEXT VERSION (remane it to main.go_template
// ######################################
//

package main

import (
	"fmt"
	"log"

	"github.com/kubesimplify/ksctl/src/api/ha_civo"
)

func main() {

	fmt.Println("Enter 1 to create and 0 to delete: ")
	choice := -1
	fmt.Scanf("%d", &choice)
	var err error
	switch choice {
	case 0:
		err = ha_civo.DeleteCluster("dipankar", "FRA1")
	case 1:
		// controlplane and workernode nodeSize
		err = ha_civo.CreateCluster("dipankar", "FRA1", "g3.medium", 3, 2)
	}
	if err != nil {
		log.Panicln(err)
	}
}
