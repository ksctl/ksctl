package localstate

import (
	"sync"

	"github.com/kubesimplify/ksctl/api/logger"
	"github.com/kubesimplify/ksctl/api/resources"
)

type Metadata struct {
	Provider   string
	HACluster  bool
	ClusterDir string
	FileName   string
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

func (s *LocalStorageProvider) Provider(provider string) resources.StateManagementInfrastructure {
	s.Metadata.Provider = provider
	return s
}

func (s *LocalStorageProvider) HA(isHA bool) resources.StateManagementInfrastructure {
	s.Metadata.HACluster = isHA
	return s
}

func (s *LocalStorageProvider) ClusterDir(dir string) resources.StateManagementInfrastructure {
	s.Metadata.ClusterDir = dir
	return s
}

func (s *LocalStorageProvider) File(file string) resources.StateManagementInfrastructure {
	s.Metadata.FileName = file
	return s
}

// Load implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Load() (any, error) {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	return nil, nil
}

// Save implements resources.StateManagementInfrastructure.
func (storage *LocalStorageProvider) Save(data any) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	return nil
}

func (storage *LocalStorageProvider) Destroy() error {
	return nil
}
