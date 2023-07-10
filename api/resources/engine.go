package resources

import (
	"github.com/kubesimplify/ksctl/api/k8s_distro/k3s"
	"github.com/kubesimplify/ksctl/api/storage/localstate"
	"github.com/kubesimplify/ksctl/api/storage/remotestate"
	"github.com/kubesimplify/ksctl/api/k8s_distro/kubeadm"
	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources/providers"
)

type ClientHandler interface {
	CloudHandler(provider string) CloudInfrastructure
	DistroHandler(distro string) Distributions
	StateHandler(place string) StateManagementInfrastructure
}

type Builder struct {
	Cloud  CloudInfrastructure
	Distro Distributions
    State StateManagementInfrastructure
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

// Hydrate Builder with the cloudProviders

func NewCivoBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.Cloud = set.CloudHandler("civo")
	return nil
}
func NewAzureBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.Cloud = set.CloudHandler("azure")
	return nil
}
func NewLocalBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.Cloud = set.CloudHandler("local")
	return nil
}

// Hydrate Builder with the Distro

func NewK3sBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.Distro = set.DistroHandler("k3s")
	return nil
}
func NewKubeadmBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.Distro = set.DistroHandler("kubeadm")
	return nil
}


// Hydrate Builder with the state

func NewLocalStorageBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.State = set.StateHandler("local")
	return nil
}
func NewRemoteStorageBuilderOrDie(b *CobraCmd) error {
	set := &ClientSet{}
	b.Client.State = set.StateHandler("remote")
	return nil
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      Builder
	Version     string
}
