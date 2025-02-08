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

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"

	v1 "k8s.io/api/core/v1"

	"sync"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Store struct {
	cloudProvider string
	clusterType   string
	clusterName   string
	region        string

	mu        *sync.Mutex
	wg        *sync.WaitGroup
	clientSet ClientSet
}

var (
	log      logger.Logger
	storeCtx context.Context
)

var ksctlNamespace string = "ksctl"

const (
	ksctlStateName string = "ksctl-state" // configmap name
)

func copyStore(src *Store, dest *Store) {
	dest.cloudProvider = src.cloudProvider
	dest.clusterName = src.clusterName
	dest.clusterType = src.clusterType
	dest.region = src.region
}

func (s *Store) Export(filters map[consts.KsctlSearchFilter]string) (*storage.StateExportImport, error) {

	var cpyS = s
	copyStore(s, cpyS)       // for storing the state of the store before import was called!
	defer copyStore(cpyS, s) // for restoring the state of the store before import was called!

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
			continue
		}
		for _, state := range states {
			if _clusterName == state.ClusterName &&
				_region == state.Region {
				dest.Clusters = append(dest.Clusters, state)
			}
		}
	}

	return dest, nil
}

func (s *Store) Import(src *storage.StateExportImport) error {
	states := src.Clusters

	var cpyS *Store = s
	copyStore(s, cpyS)       // for storing the state of the store before import was called!
	defer copyStore(cpyS, s) // for restoring the state of the store before import was called!

	for _, state := range states {
		cloud := state.InfraProvider
		region := state.Region
		clusterName := state.ClusterName
		clusterType := consts.KsctlClusterType(state.ClusterType)

		log.Debug(storeCtx, "key fields of state", "cloud", cloud, "region", region, "clusterName", clusterName, "clusterType", clusterType)

		if err := s.Setup(cloud, region, clusterName, clusterType); err != nil {
			return err
		}

		if err := s.Write(state); err != nil {
			return err
		}
	}

	return nil
}

func NewClient(parentCtx context.Context, _log logger.Logger) *Store {
	storeCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreK8s))
	log = _log
	return &Store{mu: &sync.Mutex{}, wg: &sync.WaitGroup{}}
}

func (s *Store) Connect() error {
	var err error

	if _, ok := config.IsContextPresent(storeCtx, consts.KsctlTestFlagKey); ok {
		s.clientSet, err = NewFakeK8sClient(storeCtx)
	} else {
		s.clientSet, err = NewK8sClient(storeCtx)
	}

	if err != nil {
		return err
	}

	log.Debug(storeCtx, "CONN to k8s configmap")

	return nil
}

func (s *Store) disconnect() error {
	return nil
}

func (s *Store) Kill() error {
	s.wg.Wait()
	defer log.Debug(storeCtx, "K8s Storage Got Killed")

	return s.disconnect()
}

func (s *Store) Read() (*statefile.StorageDocument, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	if c, err := s.isPresent(); err == nil {
		var result *statefile.StorageDocument
		if raw, ok := c.BinaryData[helperGenerateKeyForState(s)]; ok {
			if err := json.Unmarshal(raw, &result); err != nil {
				return nil, ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					log.NewError(storeCtx, "unable to deserialize the state", "Reason", err),
				)
			}
			return result, nil

		} else {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				log.NewError(storeCtx, "no state as binarydata", "Reason", "c.BinaryData==nil"),
			)
		}

	} else {
		return nil, err
	}
}

func generateSecret(name string, namespace string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{},
	}
}

func generateConfigMap(name string, namespace string) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		BinaryData: map[string][]byte{},
	}
}

func (s *Store) Write(data *statefile.StorageDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	raw, err := json.Marshal(data)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "unable to serialize state", "Reason", err),
		)
	}

	var c *v1.ConfigMap

	c, err = s.isPresent()
	if err != nil {
		if ksctlErrors.IsNoMatchingRecordsFound(err) {
			log.Debug(storeCtx, "configmap for write was not found")

			c = generateConfigMap(ksctlStateName, ksctlNamespace)
		} else {
			return err
		}
	} else {
		log.Debug(storeCtx, "configmap for write was found")
		if c.BinaryData == nil {
			c.BinaryData = make(map[string][]byte)
		}
	}

	c.BinaryData[helperGenerateKeyForState(s)] = raw
	if _, err := s.clientSet.WriteConfigMap(ksctlNamespace, c, metav1.UpdateOptions{}); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to write to the configmap", "Reason", err),
		)
	}
	return nil
}

func (s *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudLocal:
		s.cloudProvider = string(cloud)
	default:
		return ksctlErrors.NewError(ksctlErrors.ErrInvalidCloudProvider)
	}
	if clusterType != consts.ClusterTypeSelfMang && clusterType != consts.ClusterTypeMang {
		return ksctlErrors.NewError(ksctlErrors.ErrInvalidClusterType)
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

	if c, err := s.isPresent(); err != nil {
		return err

	} else {
		delete(c.BinaryData, helperGenerateKeyForState(s))
		_, err := s.clientSet.WriteConfigMap(ksctlNamespace, c, metav1.UpdateOptions{})
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				log.NewError(storeCtx, "unable to write the configmap", "Reason", err),
			)
		}
		return nil
	}
}

func helperGenerateKeyForState(db *Store) string {
	return fmt.Sprintf("%s.%s.%s.%s", db.cloudProvider, db.clusterType, db.clusterName, db.region)
}

func (s *Store) isPresent() (*v1.ConfigMap, error) {
	c, err := s.clientSet.ReadConfigMap(ksctlNamespace, ksctlStateName, metav1.GetOptions{})
	if err != nil {
		log.Debug(storeCtx, "storage.kubernetes.isPresent", "err", err)
		if errors.IsNotFound(err) {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				log.NewError(storeCtx, "no credential is present", "Reason", err),
			)
		}
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to read the secret", "Reason", err),
		)
	}
	return c, nil
}

func (s *Store) clusterPresent() error {
	if _, err := s.isPresent(); err != nil {
		return err
	}
	return nil
}

func (s *Store) AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	err := s.Setup(cloud, region, clusterName, clusterType)
	if err != nil {
		return err
	}

	return s.clusterPresent()
}

func (s *Store) GetOneOrMoreClusters(filters map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*statefile.StorageDocument, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()
	clusterType := filters[consts.ClusterType]
	cloud := filters[consts.Cloud]

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
	log.Debug(storeCtx, "storage.kubernetes.GetOneOrMoreClusters", "filter", filters, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*statefile.StorageDocument)

	c, err := s.clientSet.ReadConfigMap(ksctlNamespace, ksctlStateName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				log.NewError(storeCtx, "no configmap is present", "Reason", err),
			)
		}
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			log.NewError(storeCtx, "failed to read the configmap", "Reason", err),
		)
	}

	data := c.BinaryData
	// storageIdx will index all the required cloud providers and
	storageIdx := make(map[string][]*statefile.StorageDocument)
	for k, v := range data {
		_data := strings.Split(k, ".")
		_cloud := _data[0]
		_type := _data[1]

		var result *statefile.StorageDocument
		err := json.Unmarshal(v, &result)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				log.NewError(storeCtx, "failed to desearialize the state", "Reason", err),
			)
		}
		storageIdx[_cloud+" "+_type] = append(storageIdx[_cloud+" "+_type], result)
	}
	clear(data)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {
			clusters := storageIdx[cloud+" "+clusterType]

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], clusters...)
		}
	}

	return clustersInfo, nil
}
