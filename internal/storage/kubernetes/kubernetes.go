package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"

	"io"
	"sync"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
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

var log resources.LoggerFactory

var K8S_NAMESPACE string = "ksctl"

const (
	K8S_STATE_NAME      string = "ksctl-state"       // configmap name
	K8S_CREDENTIAL_NAME string = "ksctl-credentials" // secret name
)

func InitStorage(logVerbosity int, logWriter io.Writer) resources.StorageFactory {
	log = logger.NewDefaultLogger(logVerbosity, logWriter)
	log.SetPackageName(string(consts.StoreExtMongo))
	return &Store{mu: &sync.Mutex{}, wg: &sync.WaitGroup{}}
}

func (db *Store) Connect(ctx context.Context) error {
	var err error

	fakeClient := false
	if str := os.Getenv(string(consts.KsctlFakeFlag)); len(str) != 0 {
		fakeClient = true
	}

	if !fakeClient {
		db.clientSet, err = NewK8sClient(ctx)
	} else {
		db.clientSet, err = NewFakeK8sClient(ctx)
	}
	if err != nil {
		return err
	}

	log.Success("CONN to k8s configmap")

	return nil
}

func (db *Store) disconnect() error {
	return nil
}

func (db *Store) Kill() error {
	db.wg.Wait()
	defer log.Success("Mongodb Storage Got Killed")

	return db.disconnect()
}

func (db *Store) Read() (*types.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.kubernetes.Read", "Store", db)

	if c, ok := db.isPresent(); ok {
		var result *types.StorageDocument
		raw := c.BinaryData[helperGenerateKeyForState(db)]

		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, err
		}

		return result, nil
	}
	return nil, log.NewError("cluster not present")
}

func (db *Store) ReadCredentials(cloud consts.KsctlCloud) (*types.CredentialsDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.kubernetes.ReadCreds", "Store", db)

	if c, ok := db.isPresentCreds(string(cloud)); ok {
		var result *types.CredentialsDocument
		raw := c.Data[string(cloud)]

		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, err
		}

		return result, nil
	}
	return nil, log.NewError("cluster not present")
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

func (db *Store) Write(data *types.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.kubernetes.Write", "Store", db)

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if c, ok := db.isPresent(); ok {
		c.BinaryData[helperGenerateKeyForState(db)] = raw

		if _, err := db.clientSet.WriteConfigMap(K8S_NAMESPACE, c, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}
	c := generateConfigMap(K8S_STATE_NAME, K8S_NAMESPACE)
	c.BinaryData[helperGenerateKeyForState(db)] = raw

	if _, err := db.clientSet.WriteConfigMap(K8S_NAMESPACE, c, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (db *Store) WriteCredentials(cloud consts.KsctlCloud, data *types.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.kubernetes.WriteCreds", "Store", db)

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if c, ok := db.isPresentCreds(string(cloud)); ok {
		c.Data[string(cloud)] = raw

		if _, err := db.clientSet.WriteSecret(K8S_NAMESPACE, c, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}

	c := generateSecret(K8S_CREDENTIAL_NAME, K8S_NAMESPACE)

	c.Data[string(cloud)] = raw

	if _, err := db.clientSet.WriteSecret(K8S_NAMESPACE, c, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (db *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
		db.cloudProvider = string(cloud)
	default:
		return log.NewError("invalid cloud")
	}
	if clusterType != consts.ClusterTypeHa && clusterType != consts.ClusterTypeMang {
		return log.NewError("invalid cluster type")
	}

	db.clusterName = clusterName
	db.region = region
	db.clusterType = string(clusterType)

	log.Debug("storage.kubernetes.Setup", "Store", db)
	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.kubernetes.Delete", "Store", db)

	if c, ok := db.isPresent(); !ok {
		return log.NewError("cluster doesn't exist")
	} else {
		delete(c.BinaryData, helperGenerateKeyForState(db))
		_, err := db.clientSet.WriteConfigMap(K8S_NAMESPACE, c, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	}
}

func helperGenerateKeyForState(db *Store) string {
	return fmt.Sprintf("%s/%s/%s %s", db.cloudProvider, db.clusterType, db.clusterName, db.region)
}

func (db *Store) isPresent() (*v1.ConfigMap, bool) {
	c, err := db.clientSet.ReadConfigMap(K8S_NAMESPACE, K8S_STATE_NAME, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil, false
	}
	if _, ok := c.BinaryData[helperGenerateKeyForState(db)]; !ok {
		return nil, false
	}
	return c, true
}

func (db *Store) isPresentCreds(cloud string) (*v1.Secret, bool) {
	c, err := db.clientSet.ReadSecret(K8S_NAMESPACE, K8S_CREDENTIAL_NAME, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil, false
	}
	if _, ok := c.Data[cloud]; !ok {
		return nil, false
	}
	return c, true
}

func (db *Store) clusterPresent() error {
	if _, ok := db.isPresent(); !ok {
		return log.NewError("cluster not present")
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

func (db *Store) GetOneOrMoreClusters(filters map[string]string) (map[consts.KsctlClusterType][]*types.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	clusterType := filters["clusterType"]
	cloud := filters["cloud"]

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
	log.Debug("storage.kubernetes.GetOneOrMoreClusters", "filter", filters, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*types.StorageDocument)

	c, err := db.clientSet.ReadConfigMap(K8S_NAMESPACE, K8S_STATE_NAME, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := c.BinaryData
	// storageIdx will index all the required cloud providers and
	storageIdx := make(map[string][]*types.StorageDocument)
	for k, v := range data {
		_data := strings.Split(k, "/")
		_cloud := _data[0]
		_type := _data[1]

		var result *types.StorageDocument
		err := json.Unmarshal(v, &result)
		if err != nil {
			return nil, err
		}
		storageIdx[_cloud+" "+_type] = append(storageIdx[_cloud+" "+_type], result)
	}
	clear(data)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {
			clusters := storageIdx[cloud+" "+clusterType]

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], clusters...)
			log.Debug("storage.kubernetes.GetOneOrMoreClusters", "clusterInfo", clustersInfo)
		}
	}

	return clustersInfo, nil
}
