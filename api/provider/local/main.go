package local

import "github.com/kubesimplify/ksctl/api/resources"

type StateConfiguration struct {
	ClusterName string `json:"cluster_name"`
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	// Spec        Machine `json:"spec"`
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
func (*LocalProvider) GetStateForHACluster(state resources.StateManagementInfrastructure) (any, error) {
	panic("unimplemented")
}

// InitState implements resources.CloudInfrastructure.
func (*LocalProvider) InitState() error {
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

func ReturnLocalStruct() *LocalProvider {
	return &LocalProvider{}
}
