package civo

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
)

// NewManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) NewManagedCluster(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] creating managed %s cluster...", obj.Metadata.ResName)
	return nil
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (obj *CivoProvider) DelManagedCluster(state resources.StateManagementInfrastructure) error {
	fmt.Printf("[civo] Del Managed %s cluster....", obj.Metadata.ResName)
	return nil
}

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (obj *CivoProvider) GetManagedKubernetes(state resources.StateManagementInfrastructure) {
	fmt.Printf("[civo] Got Managed %s cluster....", obj.Metadata.ResName)
}
