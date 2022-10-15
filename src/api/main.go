//
// ######################################
//		EARLY DEVELOPMENT
//			CREATE & DELETE
//   NOTE: THIS FILE WILL BE REMOVED IN NEXT VERSION
// ######################################
//

package main

import (
	"github.com/kubesimplify/Kubesimpctl/src/api/local"
	"fmt"
)

func main() {
	if err := local.CreateCluster("abcd", "kindest/node:v1.24.0"); err != nil {
		panic(err)
	}

	ch := byte(' ')
	fmt.Println("Do you want to delete the cluster")
	_, err := fmt.Scanf("%c", &ch)
	if err != nil {
		panic(err)
	}
	if ch == 'y' {
		if err := local.DeleteCluster("abcd", "./config.json"); err != nil {
			panic(err)
		}
	}
}