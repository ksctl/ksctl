package localstate

import "fmt"

type LocalStorageProvider struct {
	// TODO: implement me
}

type payload struct {
    data string
}

// Load implements resources.StateManagementInfrastructure.
func (*LocalStorageProvider) Load(path string) (interface{}, error) {
	fmt.Println("local load")
    return payload{}, nil
}

// Save implements resources.StateManagementInfrastructure.
func (*LocalStorageProvider) Save(path string, data interface{}) error {
	fmt.Println("local save")
    return nil
}

