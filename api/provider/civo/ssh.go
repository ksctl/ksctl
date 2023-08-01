package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// DelSSHKeyPair implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelSSHKeyPair(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] delete %s SSHKeypair....\n", obj.Metadata.ResName)
	return nil
}

// CreateUploadSSHKeyPair implements resources.CloudInfrastructure.
func (obj *CivoProvider) CreateUploadSSHKeyPair(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] delete %s SSHKeypair....\n", obj.Metadata.ResName)
	return nil
}
