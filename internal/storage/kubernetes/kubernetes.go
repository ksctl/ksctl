package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ksctl/ksctl/pkg/helpers"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	v1 "k8s.io/api/core/v1"

	"sync"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
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
	log      types.LoggerFactory
	storeCtx context.Context
)

var ksctlNamespace string = "ksctl"

const (
	ksctlStateName      string = "ksctl-state"       // configmap name
	ksctlCredentialName string = "ksctl-credentials" // secret name
)

func copyStore(src *Store, dest *Store) {
	dest.cloudProvider = src.cloudProvider
	dest.clusterName = src.clusterName
	dest.clusterType = src.clusterType
	dest.region = src.region
}

func (s *Store) Export(filters map[consts.KsctlSearchFilter]string) (*types.StorageStateExportImport, error) {

	var cpyS *Store = s
	copyStore(s, cpyS)       // for storing the state of the store before import was called!
	defer copyStore(cpyS, s) // for restoring the state of the store before import was called!

	dest := new(types.StorageStateExportImport)

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

	if len(_cloud) == 0 {
		// all the cloud provider credentials
		for _, constsCloud := range []consts.KsctlCloud{
			consts.CloudAws,
			consts.CloudCivo,
			consts.CloudAzure,
		} {
			_v, _err := s.ReadCredentials(constsCloud)

			if _err != nil {
				log.Debug(storeCtx, "failed to read the credentials for export", "err", _err)
				if ksctlErrors.ErrNoMatchingRecordsFound.Is(_err) {
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
			if _err != nil {
				return nil, _err
			}
		}
		dest.Credentials = append(dest.Credentials, _v)
	}

	return dest, nil
}

func (s *Store) Import(src *types.StorageStateExportImport) error {
	creds := src.Credentials
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

	for _, cred := range creds {
		if cred == nil {
			continue
		}
		cloud := cred.InfraProvider

		log.Debug(storeCtx, "key fields of cred", "cloud", cloud)
		if err := s.WriteCredentials(cloud, cred); err != nil {
			return err
		}
	}

	return nil
}

func NewClient(parentCtx context.Context, _log types.LoggerFactory) *Store {
	storeCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreK8s))
	log = _log
	return &Store{mu: &sync.Mutex{}, wg: &sync.WaitGroup{}}
}

func (db *Store) Connect() error {
	var err error

	if _, ok := helpers.IsContextPresent(storeCtx, consts.KsctlTestFlagKey); ok {
		db.clientSet, err = NewFakeK8sClient(storeCtx)
	} else {
		db.clientSet, err = NewK8sClient(storeCtx)
	}

	if err != nil {
		return err
	}

	log.Success(storeCtx, "CONN to k8s configmap")

	return nil
}

func (db *Store) disconnect() error {
	return nil
}

func (db *Store) Kill() error {
	db.wg.Wait()
	defer log.Success(storeCtx, "K8s Storage Got Killed")

	return db.disconnect()
}

func (db *Store) Read() (*storageTypes.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.kubernetes.Read", "Store", db)

	if c, err := db.isPresent(); err == nil {
		var result *storageTypes.StorageDocument
		if raw, ok := c.BinaryData[helperGenerateKeyForState(db)]; ok {
			if err := json.Unmarshal(raw, &result); err != nil {
				return nil, ksctlErrors.ErrInternal.Wrap(
					log.NewError(storeCtx, "unable to deserialize the state", "Reason", err),
				)
			}
			return result, nil

		} else {
			return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
				log.NewError(storeCtx, "no state as binarydata", "Reason", "c.BinaryData==nil"),
			)
		}

	} else {
		return nil, err
	}
}

func (db *Store) ReadCredentials(cloud consts.KsctlCloud) (*storageTypes.CredentialsDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.kubernetes.ReadCreds", "Store", db)

	if c, err := db.isPresentCreds(string(cloud)); err == nil {
		var result *storageTypes.CredentialsDocument
		if raw, ok := c.Data[string(cloud)]; ok {
			if _err := json.Unmarshal(raw, &result); _err != nil {
				return nil, ksctlErrors.ErrInternal.Wrap(
					log.NewError(storeCtx, "unable to deserialize the creds", "Reason", _err),
				)
			}

			return result, nil
		} else {
			return nil, ksctlErrors.ErrNilCredentials.Wrap(
				log.NewError(storeCtx, "no credentials", "Reason", "c.Data==nil"),
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

func (db *Store) Write(data *storageTypes.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.kubernetes.Write", "Store", db)

	raw, err := json.Marshal(data)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "unable to serialize state", "Reason", err),
		)
	}

	var c *v1.ConfigMap

	c, err = db.isPresent()
	if err != nil {
		if ksctlErrors.ErrNoMatchingRecordsFound.Is(err) {
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

	c.BinaryData[helperGenerateKeyForState(db)] = raw
	if _, err := db.clientSet.WriteConfigMap(ksctlNamespace, c, metav1.UpdateOptions{}); err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to write to the configmap", "Reason", err),
		)
	}
	return nil
}

func (db *Store) WriteCredentials(cloud consts.KsctlCloud, data *storageTypes.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.kubernetes.WriteCreds", "Store", db)

	raw, err := json.Marshal(data)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "unable to serialize state", "Reason", err),
		)
	}

	var s *v1.Secret

	s, err = db.isPresentCreds(string(cloud))
	if err != nil {
		if ksctlErrors.ErrNoMatchingRecordsFound.Is(err) {
			log.Debug(storeCtx, "secret for write was not found")
			s = generateSecret(ksctlCredentialName, ksctlNamespace)

		} else {
			return err
		}
	} else {
		log.Debug(storeCtx, "secret for write was found")
		if s.Data == nil {
			s.Data = make(map[string][]byte)
		}
	}

	s.Data[string(cloud)] = raw

	if _, err := db.clientSet.WriteSecret(ksctlNamespace, s, metav1.UpdateOptions{}); err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to write to the secret", "Reason", err),
		)
	}
	return nil
}

func (db *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
		db.cloudProvider = string(cloud)
	default:
		return ksctlErrors.ErrInvalidCloudProvider
	}
	if clusterType != consts.ClusterTypeHa && clusterType != consts.ClusterTypeMang {
		return ksctlErrors.ErrInvalidClusterType
	}

	db.clusterName = clusterName
	db.region = region
	db.clusterType = string(clusterType)

	log.Debug(storeCtx, "storage.kubernetes.Setup", "Store", db)
	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.kubernetes.Delete", "Store", db)

	if c, err := db.isPresent(); err != nil {
		return err

	} else {
		delete(c.BinaryData, helperGenerateKeyForState(db))
		_, err := db.clientSet.WriteConfigMap(ksctlNamespace, c, metav1.UpdateOptions{})
		if err != nil {
			return ksctlErrors.ErrInternal.Wrap(
				log.NewError(storeCtx, "unable to write the configmap", "Reason", err),
			)
		}
		return nil
	}
}

func helperGenerateKeyForState(db *Store) string {
	return fmt.Sprintf("%s.%s.%s.%s", db.cloudProvider, db.clusterType, db.clusterName, db.region)
}

func (db *Store) isPresent() (*v1.ConfigMap, error) {
	c, err := db.clientSet.ReadConfigMap(ksctlNamespace, ksctlStateName, metav1.GetOptions{})
	if err != nil {
		log.Debug(storeCtx, "storage.kubernetes.isPresent", "err", err)
		if errors.IsNotFound(err) {
			return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
				log.NewError(storeCtx, "no credential is present", "Reason", err),
			)
		}
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to read the secret", "Reason", err),
		)
	}
	return c, nil
}

func (db *Store) isPresentCreds(cloud string) (*v1.Secret, error) {
	c, err := db.clientSet.ReadSecret(ksctlNamespace, ksctlCredentialName, metav1.GetOptions{})
	if err != nil {
		log.Debug(storeCtx, "storage.kubernetes.isPresentCreds", "err", err)
		if errors.IsNotFound(err) {
			return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
				log.NewError(storeCtx, "no credential is present", "Reason", err),
			)
		}
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to read the secret", "Reason", err),
		)
	}
	if _, ok := c.Data[cloud]; !ok {
		return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
			log.NewError(storeCtx, "no credential is present", "Reason", "no entry found for given cloud provider"),
		)
	}
	return c, nil
}

func (db *Store) clusterPresent() error {
	if _, err := db.isPresent(); err != nil {
		return err
	}
	return nil
}

func (db *Store) AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	err := db.Setup(cloud, region, clusterName, clusterType)
	if err != nil {
		return err
	}

	return db.clusterPresent()
}

func (db *Store) GetOneOrMoreClusters(filters map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*storageTypes.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	clusterType := filters[consts.ClusterType]
	cloud := filters[consts.Cloud]

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
	log.Debug(storeCtx, "storage.kubernetes.GetOneOrMoreClusters", "filter", filters, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*storageTypes.StorageDocument)

	c, err := db.clientSet.ReadConfigMap(ksctlNamespace, ksctlStateName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
				log.NewError(storeCtx, "no configmap is present", "Reason", err),
			)
		}
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to read the configmap", "Reason", err),
		)
	}

	data := c.BinaryData
	// storageIdx will index all the required cloud providers and
	storageIdx := make(map[string][]*storageTypes.StorageDocument)
	for k, v := range data {
		_data := strings.Split(k, ".")
		_cloud := _data[0]
		_type := _data[1]

		var result *storageTypes.StorageDocument
		err := json.Unmarshal(v, &result)
		if err != nil {
			return nil, ksctlErrors.ErrInternal.Wrap(
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
