package local

import (
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type StateConfiguration struct {
	ClusterName string `json:"cluster_name"`
}

type Metadata struct {
	ResName string
	Role    string
	VmType  string
	Public  bool
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	// Spec        Machine `json:"spec"`
	Metadata
}

// CreateUploadSSHKeyPair implements resources.CloudInfrastructure.
func (*LocalProvider) CreateUploadSSHKeyPair(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelFirewall implements resources.CloudInfrastructure.
func (*LocalProvider) DelFirewall(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (*LocalProvider) DelManagedCluster(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelNetwork implements resources.CloudInfrastructure.
func (*LocalProvider) DelNetwork(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudInfrastructure.
func (*LocalProvider) DelSSHKeyPair(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// DelVM implements resources.CloudInfrastructure.
func (*LocalProvider) DelVM(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (*LocalProvider) GetManagedKubernetes(state resources.StateManagementInfrastructure) {
	panic("unimplemented")
}

// GetStateForHACluster implements resources.CloudInfrastructure.
func (*LocalProvider) GetStateForHACluster(state resources.StateManagementInfrastructure) (cloud.CloudResourceState, error) {
	panic("unimplemented")
}

// InitState implements resources.CloudInfrastructure.
func (*LocalProvider) InitState(state resources.StateManagementInfrastructure, operation string) error {
	panic("unimplemented")
}

// NewFirewall implements resources.CloudInfrastructure.
func (*LocalProvider) NewFirewall(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewManagedCluster implements resources.CloudInfrastructure.
func (*LocalProvider) NewManagedCluster(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewNetwork implements resources.CloudInfrastructure.
func (*LocalProvider) NewNetwork(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

// NewVM implements resources.CloudInfrastructure.
func (*LocalProvider) NewVM(state resources.StateManagementInfrastructure) error {
	panic("unimplemented")
}

func ReturnLocalStruct(metadata resources.Metadata) *LocalProvider {
	return &LocalProvider{
		ClusterName: metadata.ClusterName,
	}
}

// it will contain the name of the resource to be created
func (cloud *LocalProvider) Name(resName string) resources.CloudInfrastructure {
	cloud.Metadata.ResName = resName
	return cloud
}

// it will contain whether the resource to be created belongs for controlplane component or loadbalancer...
func (cloud *LocalProvider) Role(resRole string) resources.CloudInfrastructure {
	cloud.Metadata.Role = resRole
	return cloud
}

// it will contain which vmType to create
func (cloud *LocalProvider) VMType(size string) resources.CloudInfrastructure {
	cloud.Metadata.VmType = size
	return cloud
}

// whether to have the resource as public or private (i.e. VMs)
func (cloud *LocalProvider) Visibility(toBePublic bool) resources.CloudInfrastructure {
	cloud.Metadata.Public = toBePublic
	return cloud
}
