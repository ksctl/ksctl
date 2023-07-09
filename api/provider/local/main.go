package local

import "fmt"

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	HACluster   bool   `json:"ha_cluster"`
	// Spec        Machine `json:"spec"`
}

// CreateFirewall implements resources.CloudInfrastructure
func (*LocalProvider) CreateFirewall() {
	panic("unimplemented")
}

// CreateManagedKubernetes implements resources.CloudInfrastructure
func (*LocalProvider) CreateManagedKubernetes() {
	fmt.Println("Local create kind clusters")
}

// CreateVM implements resources.CloudInfrastructure
func (*LocalProvider) CreateVM() {
	panic("unimplemented")
}

// CreateVirtualNetwork implements resources.CloudInfrastructure
func (*LocalProvider) CreateVirtualNetwork() {
	panic("unimplemented")
}

// DeleteFirewall implements resources.CloudInfrastructure
func (*LocalProvider) DeleteFirewall() {
	panic("unimplemented")
}

// DeleteManagedKubernetes implements resources.CloudInfrastructure
func (*LocalProvider) DeleteManagedKubernetes() {
	panic("unimplemented")
}

// DeleteVM implements resources.CloudInfrastructure
func (*LocalProvider) DeleteVM() {
	panic("unimplemented")
}

// DeleteVirtualNetwork implements resources.CloudInfrastructure
func (*LocalProvider) DeleteVirtualNetwork() {
	panic("unimplemented")
}

// GetManagedKubernetes implements resources.CloudInfrastructure
func (*LocalProvider) GetManagedKubernetes() {
	panic("unimplemented")
}

// GetVM implements resources.CloudInfrastructure
func (*LocalProvider) GetVM() {
	panic("unimplemented")
}
