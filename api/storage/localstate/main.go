package localstate

import (
	"os"
	"sync"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources"
)

type Metadata struct {
	Path string
	Perm os.FileMode
}

type LocalStorageProvider struct {
	Log logger.LogFactory
	Metadata
}

var fileMutex sync.Mutex

func InitStorage() *LocalStorageProvider {
	return &LocalStorageProvider{
		Log: &logger.Logger{},
	}
}

func (s *LocalStorageProvider) Path(path string) resources.StateManagementInfrastructure {
	s.Metadata.Path = path
	return s
}

func (s *LocalStorageProvider) Permission(perm os.FileMode) resources.StateManagementInfrastructure {
	s.Metadata.Perm = perm
	return s
}

func (s *LocalStorageProvider) CreateDir() error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	if err := os.Mkdir(s.Metadata.Path, s.Metadata.Perm); err != nil {
		return err
	}
	return nil
}

func (s *LocalStorageProvider) DeleteDir() error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	// FIXME: check that create and delete cannot happen in same time
	if err := os.RemoveAll(s.Metadata.Path); err != nil {
		return err
	}
	return nil
}

// Load implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Load() ([]byte, error) {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.ReadFile(storage.Metadata.Path)
}

// Save implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Save(data []byte) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// find the best way to use the paths to store
	// or use a switch to use or else directly call as there is entry point GetPath()
	// try to improve the functionality and simplify it
	return os.WriteFile(
		storage.Metadata.Path,
		data, 0755)
}

func (storage *LocalStorageProvider) Destroy() error {
	return nil
}
