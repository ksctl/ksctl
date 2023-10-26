package aws

import (
	"fmt"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

func (obj *AwsProvider) CreateUploadSSHKeyPair(storage resources.StorageFactory) error {
	//TODO implement me

	obj.mxName.Unlock()

	fmt.Println("AWS Create Upload SSH Key Pair")
	return nil

}
