package providers

type AzureInfrastructure interface {

	// implemented by the different cloud provider
	// managed by the Cloud Controller manager

	CreateVM()
	DeleteVM()

	CreateFirewall()
	DeleteFirewall()

	CreateManagedKubernetes()
	GetManagedKubernetes()
	DeleteManagedKubernetes()
}
