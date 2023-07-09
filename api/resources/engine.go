package resources

import (
	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/resources/providers"
)

type ClientHandler interface {
	CloudHandler(provider string) CloudInfrastructure
	DistroHandler(distro string) Distributions
	StateManagementInfrastructure
}

type ClientSet struct {
}

func (h *ClientSet) CloudHandler(provider string) CloudInfrastructure {
	switch provider {
	case "civo":
		return &civo.CivoProvider{}
	case "azure":
		return &azure.AzureProvider{}
	}
	return nil
}

func (h *ClientSet) DistroHandler(distro string) Distributions {
	switch distro {
	case "k3s":
		return nil
	case "kubeadm":
		return nil
	}
	return nil
}

type CloudInfrastructure interface {
	providers.CivoInfrastructure
	providers.AzureInfrastructure
}

type KubernetesInfrastructure interface {
	providers.K3sConfiguration
	providers.KubeadmConfiguration
}

type NonKubernetesInfrastructure interface {
	InstallApplications()
}

type Distributions interface {
	KubernetesInfrastructure
	NonKubernetesInfrastructure
}

type StateManagementInfrastructure interface {
}

type LocalStorage struct {
	// will contain data for stornig in the disk
}

type CredentialStorage struct {
}
