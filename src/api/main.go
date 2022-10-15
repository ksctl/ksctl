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
	"github.com/kubesimplify/Kubesimpctl/src/api/local"
)

func main() {
	ch := int8(0)
	name := ""

	fmt.Println("Want to create or delete [1/2]:")
	_, err := fmt.Scanf("%d", &ch)
	if err != nil {
		return
	}

	fmt.Println("Enter the cluster name to create / delete")
	_, err = fmt.Scanf("%s", &name)
	if err != nil {
		return
	}
	switch ch {
	case 1:
		if err := local.CreateCluster(name); err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mCREATED!\033[0m\n")
	case 2:
		if err := local.DeleteCluster(name); err != nil {
			fmt.Printf("\033[31;40m%v\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32;40mDELETED!\033[0m\n")
	}
}
