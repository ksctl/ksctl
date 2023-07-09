package providers

type KubeadmConfiguration interface {

	// To be Implemented by kubernetes distribution flavour
	// and used by the Kubernetes Controller manager

	ConfigureControlPlane()
	ConfigureWorkerPlane()
	DestroyWorkerPlane()

	ConfigureDataStore()
	DestroyDataStore()

	ConfigureLoadbalancer()
	DestroyLoadbalancer()
}
