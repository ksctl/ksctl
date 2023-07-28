package localstate

import (
	"fmt"
	"os"
	"sync"
)

type LocalStorageProvider struct {
}

var fileMutex sync.Mutex

// Load implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Load(path string) (any, error) {
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
func (storage *LocalStorageProvider) Save(path string, data any) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	fmt.Println("state: local save triggered")

	demo := []byte("Hello from ksctl [stores configs]")
	err := os.WriteFile(path, demo, 0644)
	if err != nil {
		return fmt.Errorf("Unable")
	}

	fmt.Println("state: Save operation")
	return nil
}
