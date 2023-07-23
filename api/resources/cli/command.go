package cli

import (
	"github.com/kubesimplify/ksctl/api/resources"
)

// Hydrate Builder with the cloudProviders

func NewCivoBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Provider = "civo"
	b.Client.Cloud = set.CloudHandler("civo")
	b.Client.ClusterName = b.ClusterName
	b.Client.Region = b.Region
	return nil
}
func NewAzureBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Provider = "azure"
	b.Client.Cloud = set.CloudHandler("azure")
	b.Client.ClusterName = b.ClusterName
	b.Client.Region = b.Region
	return nil
}
func NewLocalBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Provider = "local"
	b.Client.Cloud = set.CloudHandler("local")
	b.Client.ClusterName = b.ClusterName
	b.Client.Region = b.Region
	return nil
}

// Hydrate Builder with the Distro

func NewK3sBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.K8sDistro = "k3s"
	b.Client.Distro = set.DistroHandler("k3s")
	return nil
}
func NewKubeadmBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.K8sDistro = "kubeadm"
	b.Client.Distro = set.DistroHandler("kubeadm")
	return nil
}

// Hydrate Builder with the state

func NewLocalStorageBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.StateLocation = "local"
	b.Client.State = set.StateHandler("local")
	return nil
}
func NewRemoteStorageBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.StateLocation = "remote"
	b.Client.State = set.StateHandler("remote")
	return nil
}
