package local

type StateConfiguration struct {
	ClusterName string `json:"cluster_name"`
}

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	// Spec        Machine `json:"spec"`
}

// CreateUploadSSHKeyPair implements resources.CloudInfrastructure.
func (*LocalProvider) CreateUploadSSHKeyPair() error {
	panic("unimplemented")
}

// DelFirewall implements resources.CloudInfrastructure.
func (*LocalProvider) DelFirewall() error {
	panic("unimplemented")
}

// DelManagedCluster implements resources.CloudInfrastructure.
func (*LocalProvider) DelManagedCluster() error {
	panic("unimplemented")
}

// DelNetwork implements resources.CloudInfrastructure.
func (*LocalProvider) DelNetwork() error {
	panic("unimplemented")
}

// DelSSHKeyPair implements resources.CloudInfrastructure.
func (*LocalProvider) DelSSHKeyPair() error {
	panic("unimplemented")
}

// DelVM implements resources.CloudInfrastructure.
func (*LocalProvider) DelVM() error {
	panic("unimplemented")
}

// GetManagedKubernetes implements resources.CloudInfrastructure.
func (*LocalProvider) GetManagedKubernetes() {
	panic("unimplemented")
}

// GetStateForHACluster implements resources.CloudInfrastructure.
func (*LocalProvider) GetStateForHACluster() (any, error) {
	panic("unimplemented")
}

// InitState implements resources.CloudInfrastructure.
func (*LocalProvider) InitState() error {
	panic("unimplemented")
}

// NewFirewall implements resources.CloudInfrastructure.
func (*LocalProvider) NewFirewall() error {
	panic("unimplemented")
}

// NewManagedCluster implements resources.CloudInfrastructure.
func (*LocalProvider) NewManagedCluster() error {
	panic("unimplemented")
}

// NewNetwork implements resources.CloudInfrastructure.
func (*LocalProvider) NewNetwork() error {
	panic("unimplemented")
}

// NewVM implements resources.CloudInfrastructure.
func (*LocalProvider) NewVM() error {
	panic("unimplemented")
}

func ReturnLocalStruct() *LocalProvider {
	return &LocalProvider{}
}
