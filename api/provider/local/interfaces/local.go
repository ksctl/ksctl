package interfaces

import "fmt"

type LocalProvider struct {
	ClusterName string `json:"cluster_name"`
	// Spec        Machine `json:"spec"`
}

// CreateFirewall implements resources.CloudInfrastructure
func (*LocalProvider) CreateFirewall() {
	panic("NO SUPPORT")
}

// CreateManagedKubernetes implements resources.CloudInfrastructure
func (*LocalProvider) CreateManagedKubernetes() {
	fmt.Println("Local create kind clusters")
}

// CreateVM implements resources.CloudInfrastructure
func (*LocalProvider) CreateVM() {
	panic("NO SUPPORT")
}

// CreateVirtualNetwork implements resources.CloudInfrastructure
func (*LocalProvider) CreateVirtualNetwork() {
	panic("NO SUPPORT")
}

// DeleteFirewall implements resources.CloudInfrastructure
func (*LocalProvider) DeleteFirewall() {
	panic("NO SUPPORT")
}

// DeleteManagedKubernetes implements resources.CloudInfrastructure
func (*LocalProvider) DeleteManagedKubernetes() {
	fmt.Println("delete local managed kubernetes")
}

// DeleteVM implements resources.CloudInfrastructure
func (*LocalProvider) DeleteVM() {
	panic("NO SUPPORT")
}

// DeleteVirtualNetwork implements resources.CloudInfrastructure
func (*LocalProvider) DeleteVirtualNetwork() {
	panic("NO SUPPORT")
}

// GetManagedKubernetes implements resources.CloudInfrastructure
func (*LocalProvider) GetManagedKubernetes() {
	fmt.Println("local managed k8s")
}

// GetVM implements resources.CloudInfrastructure
func (*LocalProvider) GetVM() {
	panic("NO SUPPORT")
}
