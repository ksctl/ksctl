package cli

import (
	"github.com/kubesimplify/ksctl/api/resources"
)

// Hydrate Builder with the cloudProviders

func NewCivoBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Cloud = set.CloudHandler("civo")
	return nil
}
func NewAzureBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Cloud = set.CloudHandler("azure")
	return nil
}
func NewLocalBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Cloud = set.CloudHandler("local")
	return nil
}

// Hydrate Builder with the Distro

func NewK3sBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Distro = set.DistroHandler("k3s")
	return nil
}
func NewKubeadmBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.Distro = set.DistroHandler("kubeadm")
	return nil
}

// Hydrate Builder with the state

func NewLocalStorageBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.State = set.StateHandler("local")
	return nil
}
func NewRemoteStorageBuilderOrDie(b *resources.CobraCmd) error {
	set := &resources.ClientSet{}
	b.Client.State = set.StateHandler("remote")
	return nil
}
