package resources

import (
	"github.com/kubesimplify/ksctl/api/k8s_distro/k3s"
	"github.com/kubesimplify/ksctl/api/k8s_distro/kubeadm"
	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/resources/providers"
)

type ClientHandler interface {
	CloudHandler(provider string) CloudInfrastructure
	DistroHandler(distro string) Distributions
	StateManagementInfrastructure
}

type Builder struct {
	Cloud  CloudInfrastructure
	Distro Distributions
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
		return &k3s.K3sDistro{}
	case "kubeadm":
		return &kubeadm.KubeadmDistro{}
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
}

type LocalStorage struct {
	// will contain data for stornig in the disk
}

type CredentialStorage struct {
}

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

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      Builder
	Version     string
}
