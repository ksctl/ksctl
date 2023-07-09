package providers

type LocalInfrastructure interface {
	CreateManagedKubernetes()
	GetManagedKubernetes()
	DeleteManagedKubernetes()
}
