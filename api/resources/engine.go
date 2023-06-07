package resources

import (
	"github.com/kubesimplify/ksctl/api/resources/cli"
	"github.com/kubesimplify/ksctl/api/resources/providers"
)

type CloudInfrastructure interface {

	// implemented by the different cloud provider
	// managed by the Cloud Controller manager

	CreateVM()
	DeleteVM()

	CreateFirewall()
	DeleteFirewall()

	CreateVirtualNetwork()
	DeleteVirtualNetwork()

	GetVM()

	// managed clusters are managed by the Cloud provider so no need for the Kubernetes abstraction layer

	CreateManagedKubernetes()
	GetManagedKubernetes()
	DeleteManagedKubernetes()
}

type KubernetesInfrastructure interface {

	// To be Implemented by kubernetes distribution flavour
	// and used by the Kubernetes Controller manager

	ConfigureControlPlane()
	DestroyControlPlane()

	ConfigureWorkerPlane()
	DestroyWorkerPlane()

	ConfigureLoadbalancer()
	DestroyLoadbalancer()

	ConfigureDataStore()
	DestroyDataStore()

	InstallApplication() // not planned yet
}

type StateManagementInfrastructure interface {
}

type LocalStorage struct {
	// will contain data for stornig in the disk
}

type CredentialStorage struct {
}

func NewCivoBuilderOrDie(b *cli.CobraCmd) error {
	b.Client.Client = providers.CivoProvider{
		Region:      b.Region,
		ClusterName: b.ClusterName,
		HACluster:   false,
		APIKey:      "XYBXSBWJHDSBCHJSBDZCBHJDBHJDSB",
	}
	return nil
}

func NewAzureBuilderOrDie(b *cli.CobraCmd) {}

// func (b *BehaviourCobraCmd) NewAWSBuilderOrDie(b *BehaviourCobraCmd)   {}
// func (b *BehaviourCobraCmd) NewGCPBuilderOrDie()   {}
// func (b *BehaviourCobraCmd) NewLocalBuilderOrDie() {}
