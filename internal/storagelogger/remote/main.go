package remote

import "fmt"

type RemoteStorageProvider struct {
	// TODO: implement me
}

// Load implements resources.StateManagementInfrastructure.
func (*RemoteStorageProvider) Load() ([]byte, error) {
	fmt.Println("remote load")
	return nil, nil
}

// Save implements resources.StateManagementInfrastructure.
func (*RemoteStorageProvider) Save(data []byte) error {
	fmt.Println("remote save")
	return nil
}

func (*RemoteStorageProvider) Destroy() error {
	return nil
}
