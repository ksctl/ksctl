package remotestate

import "fmt"

type RemoteStorageProvider struct {
	// TODO: implement me
}

type payload struct {
    data string
}

// Load implements resources.StateManagementInfrastructure.
func (*RemoteStorageProvider) Load(path string) (interface{}, error) {
	fmt.Println("remote load")
    return payload{}, nil
}

// Save implements resources.StateManagementInfrastructure.
func (*RemoteStorageProvider) Save(path string, data interface{}) error {
	fmt.Println("remote save")
    return nil
}

