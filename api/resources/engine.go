package resources

import (
	k3s "github.com/kubesimplify/ksctl/api/k8s_distro/k3s/interfaces"
	kubeadm "github.com/kubesimplify/ksctl/api/k8s_distro/kubeadm/interfaces"
	azure "github.com/kubesimplify/ksctl/api/provider/azure/interfaces"
	civo "github.com/kubesimplify/ksctl/api/provider/civo/interfaces"
	local "github.com/kubesimplify/ksctl/api/provider/local/interfaces"
	"github.com/kubesimplify/ksctl/api/resources/providers"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/storage/remotestate"
)

type ClientHandler interface {
	CloudHandler(provider string) CloudInfrastructure
	DistroHandler(distro string) Distributions
	StateHandler(place string) StateManagementInfrastructure
}

type Builder struct {
	Cloud  CloudInfrastructure
	Distro Distributions
	State  StateManagementInfrastructure

	Provider      string
	K8sDistro     string
	K8sVersion    string
	StateLocation string
	IsHA          bool
}

type ClientSet struct {
}

func (h *ClientSet) CloudHandler(provider string) CloudInfrastructure {
	switch provider {
	case "civo":
		return &civo.CivoProvider{}
	case "azure":
		return &azure.AzureProvider{}
	case "local":
		return &local.LocalProvider{}
	}
	return nil
}

func (h *ClientSet) DistroHandler(distro string) Distributions {
	switch distro {
	case "k3s":
		return &k3s.K3sDistro{}
	case "kubeadm":
		return &kubeadm.KubeadmDistro{}
	}
	return nil
}

// StateHandler place local, remote
func (h *ClientSet) StateHandler(place string) StateManagementInfrastructure {
	switch place {
	case "local":
		return &localstate.LocalStorageProvider{}
	case "remote":
		return &remotestate.RemoteStorageProvider{}
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
	// NonKubernetesInfrastructure
}

type StateManagementInfrastructure interface {
	providers.LocalStorage
	providers.RemoteStorage
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      Builder
	Version     string
}
