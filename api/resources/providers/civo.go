package providers

type CivoInfrastructure interface {

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
