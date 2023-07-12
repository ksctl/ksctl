package civo

import "fmt"

type CivoProvider struct {
	ClusterName string `json:"cluster_name"`
	APIKey      string `json:"api_key"`
	HACluster   bool   `json:"ha_cluster"`
	Region      string `json:"region"`
	//Spec        util.Machine `json:"spec"`
	Application string `json:"application"`
	CNIPlugin   string `json:"cni_plugin"`
}

func (b *CivoProvider) CreateVM() {
	fmt.Println("Civo Create VM")
}

func (b *CivoProvider) DeleteVM() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) GetVM() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) CreateFirewall() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) DeleteFirewall() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) CreateVirtualNetwork() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) DeleteVirtualNetwork() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) CreateManagedKubernetes() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) GetManagedKubernetes() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) DeleteManagedKubernetes() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) InstallApplication() {
	//TODO implement me
	panic("implement me")
}

func (b *CivoProvider) InstallApplications() {
	//TODO implement me
	panic("implement me")
}
