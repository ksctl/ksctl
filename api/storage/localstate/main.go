package localstate

import (
	"fmt"
	"io/fs"
	"os"
	"sync"
)

type LocalStorageProvider struct {
	// TODO: implement me
}

var fileMutex sync.Mutex

// Load implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Load(path string) (interface{}, error) {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	fmt.Println("state: local load triggered")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	fmt.Printf("%s", string(data))

	return nil, nil
}

// Save implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Save(path string, data interface{}) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	fmt.Println("state: local save triggered")

	demo := []byte("Hello")
	err := os.WriteFile(path, demo, fs.FileMode(os.O_RDONLY|os.O_CREATE|os.O_WRONLY))
	if err != nil {
		return fmt.Errorf("Unable")
	}

	fmt.Println("state: Save operation")
	return nil
}
