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
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
)

const (
	filePerm       = os.FileMode(0640)
	dirPerm        = os.FileMode(0750)
	credentialPerm = os.FileMode(0600)
	subDirState    = "state"
	subDirCreds    = "credentials"
)

type Store struct {
	ctx           context.Context
	l             logger.Logger
	cloudProvider string
	clusterType   string
	clusterName   string
	region        string
	mu            *sync.RWMutex
	wg            *sync.WaitGroup
}

func NewClient(parentCtx context.Context, _log logger.Logger) *Store {
	return &Store{
		ctx: context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreLocal)),
		l:   _log,
		mu:  &sync.RWMutex{},
		wg:  &sync.WaitGroup{},
	}
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
					s.l.NewError(s.ctx, "unable to create file", "Reason", _err),
				)
				return
			}
			err = nil
			return
		}
		err = ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			s.l.NewError(s.ctx, "failed to read the file", "Reason", err),
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
			s.l.NewError(s.ctx, "failed to mkdir all folders", "Reason", err),
		)

	}
	return nil
}

func (s *Store) genOsClusterPath(subDir ...string) (string, error) {

	var userLoc string
	if v, ok := config.IsContextPresent(s.ctx, consts.KsctlCustomDirLoc); ok {
		userLoc = filepath.Join(strings.Split(strings.TrimSpace(v), " ")...)
	} else {
		v, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		userLoc = v
	}
	subKsctlLoc := subDirState
	pathArr := []string{userLoc, ".ksctl", subKsctlLoc}
	pathArr = append(pathArr, subDir...)
	s.l.Debug(s.ctx, "storage.local.genOsClusterPath", "userLoc", userLoc, "subKsctlLoc", subKsctlLoc, "pathArr", pathArr)

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

	if e := s.clusterPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				s.l.NewError(s.ctx, "cluster not present", "Reason", err),
			)
		}
		return nil
	}); e != nil {
		return nil, e
	}
	dirPath, err := s.genOsClusterPath(s.cloudProvider, s.clusterType, s.clusterName+" "+s.region, "state.json")
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to gen clusterpath in host", "Reason", err),
		)
	}
	s.l.Debug(s.ctx, "storage.local.Read", "dirPath", dirPath)
	if v, e := reader(dirPath); e != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to read in host", "Reason", err),
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

	dirPath, err := s.genOsClusterPath(s.cloudProvider, s.clusterType, s.clusterName+" "+s.region)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to gen clusterpath in host", "Reason", err),
		)
	}
	FileLoc := ""
	s.l.Debug(s.ctx, "storage.local.Write", "dirPath", dirPath)

	if err := s.clusterPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dirPath, dirPerm); err != nil {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					s.l.NewError(s.ctx, "failure in creating directories", "Reason", err),
				)
			}
		}
		return nil
	}); err != nil {
		if !ksctlErrors.IsInternal(err) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrDuplicateRecords,
				s.l.NewError(s.ctx, "cluster already present", "Reason", err),
			)
		} else {
			return err
		}
	}

	FileLoc = filepath.Join(dirPath, "state.json")
	s.l.Debug(s.ctx, "storage.local.Write", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to serialize state", "Reason", err),
		)
	}
	if err := os.WriteFile(FileLoc, data, filePerm); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to write in host", "Reason", err),
		)
	}
	return nil
}

func (s *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudLocal:
		s.cloudProvider = string(cloud)
	default:
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			s.l.NewError(s.ctx, "invalid", "cloud", cloud),
		)
	}
	if clusterType != consts.ClusterTypeSelfMang && clusterType != consts.ClusterTypeMang {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidClusterType,
			s.l.NewError(s.ctx, "invalid", "clusterType", clusterType),
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
			s.l.NewError(s.ctx, "cluster not present", "Reason", e),
		)
	}
	dirPath, err := s.genOsClusterPath(s.cloudProvider, s.clusterType, s.clusterName+" "+s.region)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to perform complete clenup some directories are left behind", "Reason", err),
		)
	}
	return nil
}

func (s *Store) clusterPresent(handleErrFunc func(error) error) error {
	dirPath, _ := s.genOsClusterPath(s.cloudProvider, s.clusterType, s.clusterName+" "+s.region)
	_, err := os.Stat(dirPath)
	if err != nil {
		s.l.Debug(s.ctx, "storage.local.clusterPresent", "err", err)
		if handleErrFunc != nil {
			return handleErrFunc(err)
		}
		return err
	}

	return nil
}

func (s *Store) credentialsPresent(handleErrFunc func(error) error) error {
	dirPath, _ := s.genOsClusterPath()
	_, err := os.Stat(dirPath)
	if err != nil {
		s.l.Debug(s.ctx, "storage.local.credentialsPresent", "err", err)
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
			s.l.NewError(s.ctx, "cluster not present", "Reason", e),
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
		filterCloudPath = append(filterCloudPath, string(consts.CloudAws), string(consts.CloudAzure), string(consts.CloudLocal))

	case string(consts.CloudAzure):
		filterCloudPath = append(filterCloudPath, string(consts.CloudAzure))

	case string(consts.CloudAws):
		filterCloudPath = append(filterCloudPath, string(consts.CloudAws))

	case string(consts.CloudLocal):
		filterCloudPath = append(filterCloudPath, string(consts.CloudLocal))
	}

	switch clusterType {
	case string(consts.ClusterTypeSelfMang):
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeSelfMang))

	case string(consts.ClusterTypeMang):
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeMang))

	case "":
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeMang), string(consts.ClusterTypeSelfMang))
	}
	s.l.Debug(s.ctx, "storage.local.GetOneOrMoreClusters", "filter", filter, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*statefile.StorageDocument)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {
			files, err := s.fetchFilePaths(cloud, clusterType)
			if err != nil {
				return nil, err
			}
			v, err := s.getClustersInfo(files)
			if err != nil {
				return nil, err
			}

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], v...)
		}
	}

	return clustersInfo, nil
}

func (s *Store) getClustersInfo(locs []string) ([]*statefile.StorageDocument, error) {
	var data []*statefile.StorageDocument

	for _, loc := range locs {
		v, err := reader(loc)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				s.l.NewError(s.ctx, "failed to read in host", "Reason", err),
			)
		}
		data = append(data, v)
	}

	return data, nil
}

func (s *Store) fetchFilePaths(cloud string, clusterType string) ([]string, error) {
	dirPath, err := s.genOsClusterPath(cloud, clusterType)
	if err != nil {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			s.l.NewError(s.ctx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	folders, err := os.ReadDir(dirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err // as the other function uses the err code getClustersInfo
	}

	s.l.Debug(s.ctx, "storage.local.fetchFilePaths", "folders", folders)
	var info []string
	for _, file := range folders {
		if file.IsDir() {
			info = append(info, filepath.Join(dirPath, file.Name(), "state.json"))
		}
	}

	return info, nil
}
