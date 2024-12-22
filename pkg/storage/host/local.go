// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package host

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/statefile"
	"github.com/ksctl/ksctl/pkg/storage"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
)

const (
	filePerm       = os.FileMode(0640)
	dirPerm        = os.FileMode(0750)
	credentialPerm = os.FileMode(0600)
	subDirState    = "state"
	subDirCreds    = "credentials"
)

var (
	log      logger.Logger
	storeCtx context.Context
)

type Store struct {
	cloudProvider string
	clusterType   string
	clusterName   string
	region        string
	userid        any
	mu            *sync.RWMutex
	wg            *sync.WaitGroup
}

func copyStore(src *Store, dest *Store) {
	dest.cloudProvider = src.cloudProvider
	dest.clusterName = src.clusterName
	dest.clusterType = src.clusterType
	dest.region = src.region
	dest.userid = src.userid
}

func (s *Store) PresentDirectory(_path []string) (loc string, isPresent bool) {
	loc = filepath.Join(_path...)
	_, err := os.ReadDir(loc)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}
	isPresent = true
	return
}

func (s *Store) CreateFileIfNotPresent(_path []string) (loc string, err error) {
	loc = filepath.Join(_path...)

	if _, err = os.ReadFile(loc); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if _, _err := os.Create(loc); _err != nil {
				err = ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					log.NewError(storeCtx, "unable to create file", "Reason", _err),
				)
				return
			}
			err = nil
			return
		}
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(storeCtx, "failed to read the file", "Reason", err),
		)
		return
	}
	return
}

func (s *Store) CreateDirectory(_path []string) error {
	loc := filepath.Join(_path...)

	if err := os.MkdirAll(loc, dirPerm); err != nil {
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to mkdir all folders", "Reason", err),
		)

	}
	return nil
}

func (s *Store) Export(filters map[consts.KsctlSearchFilter]string) (*storage.StateExportImport, error) {

	var cpyS = s
	copyStore(s, cpyS) // for storing the state of the store before import was called!

	dest := new(storage.StateExportImport)

	_cloud := filters[consts.Cloud]
	_clusterType := filters[consts.ClusterType]
	_clusterName := filters[consts.Name]
	_region := filters[consts.Region]

	stateClustersForTypes, err := s.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       _cloud,
		consts.ClusterType: _clusterType,
	})
	if err != nil {
		return nil, err
	}

	for _, states := range stateClustersForTypes {
		// NOTE: make sure both filters are available if not then it will not apply
		if len(_clusterName) == 0 || len(_region) == 0 {
			dest.Clusters = append(dest.Clusters, states...)
		}
		for _, state := range states {
			if _clusterName == state.ClusterName &&
				_region == state.Region {
				dest.Clusters = append(dest.Clusters, state)
			}
		}
	}

	if len(_cloud) == 0 {
		// all the cloud provider credentials
		for _, constsCloud := range []consts.KsctlCloud{
			consts.CloudAws,
			consts.CloudCivo,
			consts.CloudAzure,
		} {
			_v, _err := s.ReadCredentials(constsCloud)

			if _err != nil {
				log.Debug(storeCtx, "error in reading credentials", "Reason", _err)
				if ksctlErrors.IsNoMatchingRecordsFound(_err) {
					continue
				} else {
					return nil, _err
				}
			}
			dest.Credentials = append(dest.Credentials, _v)
		}
	} else {
		_v, _err := s.ReadCredentials(consts.KsctlCloud(_cloud))
		if _cloud != string(consts.CloudLocal) {
			if _err != nil && !ksctlErrors.IsNoMatchingRecordsFound(_err) {
				return nil, _err
			}
		}
		dest.Credentials = append(dest.Credentials, _v)
	}

	copyStore(cpyS, s) // for restoring the state of the store before import was called!
	return dest, nil
}

func (s *Store) Import(src *storage.StateExportImport) error {
	creds := src.Credentials
	states := src.Clusters

	var cpyS = s
	copyStore(s, cpyS) // for storing the state of the store before import was called!

	for _, state := range states {
		cloud := state.InfraProvider
		region := state.Region
		clusterName := state.ClusterName
		clusterType := consts.KsctlClusterType(state.ClusterType)

		if err := s.Setup(cloud, region, clusterName, clusterType); err != nil {
			return err
		}

		if err := s.Write(state); err != nil {
			return err
		}
	}

	for _, cred := range creds {
		cloud := cred.InfraProvider
		if err := s.WriteCredentials(cloud, cred); err != nil {
			return err
		}
	}

	copyStore(cpyS, s) // for restoring the state of the store before import was called!
	return nil
}

func NewClient(parentCtx context.Context, _log logger.Logger) *Store {
	storeCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreLocal))
	log = _log
	return &Store{mu: &sync.RWMutex{}, wg: &sync.WaitGroup{}}
}

func (s *Store) disconnect() error {
	return nil
}

func (s *Store) Kill() error {
	s.wg.Wait()
	defer log.Success(storeCtx, "Local Storage Got Killed")

	return s.disconnect()
}

func (s *Store) Connect() error {
	if v, ok := config.IsContextPresent(storeCtx, consts.KsctlContextUserID); ok {
		s.userid = v
	} else {
		s.userid = "default"
	}

	log.Success(storeCtx, "CONN to HostOS")
	return nil
}

func genOsClusterPath(creds bool, subDir ...string) (string, error) {

	var userLoc string
	if v, ok := config.IsContextPresent(storeCtx, consts.KsctlCustomDirLoc); ok {
		userLoc = filepath.Join(strings.Split(strings.TrimSpace(v), " ")...)
	} else {
		v, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		userLoc = v
	}
	subKsctlLoc := subDirState
	if creds {
		subKsctlLoc = subDirCreds
	}
	pathArr := []string{userLoc, ".ksctl", subKsctlLoc}

	if !creds {
		pathArr = append(pathArr, subDir...)
	}
	log.Debug(storeCtx, "storage.local.genOsClusterPath", "userLoc", userLoc, "subKsctlLoc", subKsctlLoc, "pathArr", pathArr)

	return filepath.Join(pathArr...), nil
}

func reader(loc string) (*statefile.StorageDocument, error) {
	data, err := os.ReadFile(loc)
	if err != nil {
		return nil, err
	}

	var v *statefile.StorageDocument
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Store) Read() (*statefile.StorageDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.wg.Add(1)
	defer s.wg.Done()

	if e := s.clusterPresent(nil); e != nil {
		return nil, e
	}
	dirPath, err := genOsClusterPath(false, s.cloudProvider, s.clusterType, s.clusterName+" "+s.region, "state.json")
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}
	log.Debug(storeCtx, "storage.local.Read", "dirPath", dirPath)
	if v, e := reader(dirPath); e != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to read in host", "Reason", err),
		)
	} else {
		return v, nil
	}
}

func (s *Store) Write(v *statefile.StorageDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	dirPath, err := genOsClusterPath(false, s.cloudProvider, s.clusterType, s.clusterName+" "+s.region)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}
	FileLoc := ""
	log.Debug(storeCtx, "storage.local.Write", "dirPath", dirPath)

	if err := s.clusterPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dirPath, dirPerm); err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					log.NewError(storeCtx, "failure in creating directories", "Reason", err),
				)
			}
		}
		return nil
	}); err != nil {
		if !ksctlErrors.IsInternal(err) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				log.NewError(storeCtx, "cluster already present", "Reason", err),
			)
		} else {
			return err
		}
	}

	FileLoc = filepath.Join(dirPath, "state.json")
	log.Debug(storeCtx, "storage.local.Write", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to serialize state", "Reason", err),
		)
	}
	if err := os.WriteFile(FileLoc, data, filePerm); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to write in host", "Reason", err),
		)
	}
	return nil
}

func (s *Store) ReadCredentials(cloud consts.KsctlCloud) (*statefile.CredentialsDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.wg.Add(1)
	defer s.wg.Done()

	// for now create multiple files civo.json, azure.json, etc.
	if e := s.credentialsPresent(nil); e != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			log.NewError(storeCtx, "credentials are absent", "Reason", e),
		)
	}
	dirPath, err := genOsClusterPath(true)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	data, err := os.ReadFile(filepath.Join(dirPath, string(cloud)+".json"))
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			log.NewError(storeCtx, "failed to read a host file", "Reason", err),
		)
	}

	var v *statefile.CredentialsDocument
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrNilCredentials,
			log.NewError(storeCtx, "failed to deserialize the credentials", "Reason", err),
		)
	}

	return v, nil
}

func (s *Store) WriteCredentials(cloud consts.KsctlCloud, v *statefile.CredentialsDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	dirPath, err := genOsClusterPath(true)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	if _err := s.credentialsPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if __err := os.MkdirAll(dirPath, dirPerm); __err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					log.NewError(storeCtx, "failure in creating directories", "Reason", __err),
				)
			}
		}
		return nil
	}); _err != nil {
		if !ksctlErrors.IsInternal(_err) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				log.NewError(storeCtx, "credentials already present", "Reason", _err),
			)
		} else {
			return _err
		}
	}

	FileLoc := ""

	FileLoc = filepath.Join(dirPath, string(cloud)+".json")
	log.Debug(storeCtx, "storage.local.WriteCredentials", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to serialize the credentials", "Reason", err),
		)
	}
	if e := os.WriteFile(FileLoc, data, credentialPerm); e != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to write file in host", "Reason", err),
		)
	}
	return nil
}

func (s *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
		s.cloudProvider = string(cloud)
	default:
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			log.NewError(storeCtx, "invalid", "cloud", cloud),
		)
	}
	if clusterType != consts.ClusterTypeHa && clusterType != consts.ClusterTypeMang {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidClusterType,
			log.NewError(storeCtx, "invalid", "clusterType", clusterType),
		)
	}
	s.clusterName = clusterName
	s.region = region
	s.clusterType = string(clusterType)
	return nil
}

func (s *Store) DeleteCluster() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	if e := s.clusterPresent(nil); e != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			log.NewError(storeCtx, "cluster not present", "Reason", e),
		)
	}
	dirPath, err := genOsClusterPath(false, s.cloudProvider, s.clusterType, s.clusterName+" "+s.region)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to perform complete clenup some directories are left behind", "Reason", err),
		)
	}
	return nil
}

func (s *Store) clusterPresent(handleErrFunc func(error) error) error {
	dirPath, _ := genOsClusterPath(false, s.cloudProvider, s.clusterType, s.clusterName+" "+s.region)
	_, err := os.Stat(dirPath)
	if err != nil {
		log.Debug(storeCtx, "storage.local.clusterPresent", "err", err)
		if handleErrFunc != nil {
			return handleErrFunc(err)
		}
		return err
	}

	return nil
}

func (s *Store) credentialsPresent(handleErrFunc func(error) error) error {
	dirPath, _ := genOsClusterPath(true)
	_, err := os.Stat(dirPath)
	if err != nil {
		log.Debug(storeCtx, "storage.local.credentialsPresent", "err", err)
		if handleErrFunc != nil {
			return handleErrFunc(err)
		}
		return err
	}

	return nil
}

func (s *Store) AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.wg.Add(1)
	defer s.wg.Done()

	err := s.Setup(cloud, region, clusterName, clusterType)
	if err != nil {
		return err
	}

	if e := s.clusterPresent(nil); e != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrNoMatchingRecordsFound,
			log.NewError(storeCtx, "cluster not present", "Reason", e),
		)
	}
	return nil
}

func (s *Store) GetOneOrMoreClusters(filter map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*statefile.StorageDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.wg.Add(1)
	defer s.wg.Done()

	clusterType := filter[consts.ClusterType]
	cloud := filter[consts.Cloud]

	var filterCloudPath, filterClusterType []string

	switch cloud {
	case string(consts.CloudAll), "":
		filterCloudPath = append(filterCloudPath, string(consts.CloudCivo), string(consts.CloudAws), string(consts.CloudAzure), string(consts.CloudLocal))

	case string(consts.CloudCivo):
		filterCloudPath = append(filterCloudPath, string(consts.CloudCivo))

	case string(consts.CloudAzure):
		filterCloudPath = append(filterCloudPath, string(consts.CloudAzure))

	case string(consts.CloudAws):
		filterCloudPath = append(filterCloudPath, string(consts.CloudAws))

	case string(consts.CloudLocal):
		filterCloudPath = append(filterCloudPath, string(consts.CloudLocal))
	}

	switch clusterType {
	case string(consts.ClusterTypeHa):
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeHa))

	case string(consts.ClusterTypeMang):
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeMang))

	case "":
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeMang), string(consts.ClusterTypeHa))
	}
	log.Debug(storeCtx, "storage.local.GetOneOrMoreClusters", "filter", filter, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*statefile.StorageDocument)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {
			files, err := fetchFilePaths(cloud, clusterType)
			if err != nil {
				return nil, err
			}
			v, err := getClustersInfo(files)
			if err != nil {
				return nil, err
			}

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], v...)
		}
	}

	return clustersInfo, nil
}

func getClustersInfo(locs []string) ([]*statefile.StorageDocument, error) {
	var data []*statefile.StorageDocument

	for _, loc := range locs {
		v, err := reader(loc)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				log.NewError(storeCtx, "failed to read in host", "Reason", err),
			)
		}
		data = append(data, v)
	}

	return data, nil
}

func fetchFilePaths(cloud string, clusterType string) ([]string, error) {
	dirPath, err := genOsClusterPath(false, cloud, clusterType)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	folders, err := os.ReadDir(dirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err // as the other function uses the err code getClustersInfo
	}

	log.Debug(storeCtx, "storage.local.fetchFilePaths", "folders", folders)
	var info []string
	for _, file := range folders {
		if file.IsDir() {
			info = append(info, filepath.Join(dirPath, file.Name(), "state.json"))
		}
	}

	return info, nil
}
