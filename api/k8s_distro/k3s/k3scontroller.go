package k3s

type K3sDistro struct {
	IsHA bool
}

func (k3s *K3sDistro) ConfigureControlPlane() {
	//TODO implement me
	panic("K3s Config CP")
}

func (k3s *K3sDistro) DestroyControlPlane() {
	//TODO implement me
}

func (k3s *K3sDistro) ConfigureWorkerPlane() {
	//TODO implement me
}

func (k3s *K3sDistro) DestroyWorkerPlane() {
	//TODO implement me
}

func (k3s *K3sDistro) ConfigureLoadbalancer() {
	//TODO implement me
}

func (k3s *K3sDistro) DestroyLoadbalancer() {
	//TODO implement me
}

func (k3s *K3sDistro) ConfigureDataStore() {
	//TODO implement me
}

func (k3s *K3sDistro) DestroyDataStore() {
	//TODO implement me
}

func (k3s *K3sDistro) InstallApplication() {
	//TODO implement me
}
