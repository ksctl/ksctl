package providers

type K3sConfiguration interface {

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
