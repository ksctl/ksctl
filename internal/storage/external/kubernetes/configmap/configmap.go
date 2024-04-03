package configmap

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	v1 "k8s.io/api/core/v1"

	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/resources"
)

type Store struct {
	context context.Context

	cloudProvider string
	clusterType   string
	clusterName   string
	region        string

	userid any

	mu        *sync.Mutex
	wg        *sync.WaitGroup
	clientSet *kubernetes.Clientset
}

var log resources.LoggerFactory

const (
	CredsCollection      string = "credentials"
	K8S_NAMESPACE        string = "ksctl"
	K8S_CONFIGMAP_NAME   string = "ksctl-state"
	K8S_STATE_FILE_NAME  string = "state.json"
	K8S_CREDENTIALS_NAME string = "credentials.json"
)

func InitStorage(logVerbosity int, logWriter io.Writer) resources.StorageFactory {
	log = logger.NewDefaultLogger(logVerbosity, logWriter)
	log.SetPackageName(string(consts.StoreExtMongo))
	return &Store{mu: &sync.Mutex{}, wg: &sync.WaitGroup{}}
}

func WorkOnK8s() error {

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching ConfigMap: %v\n", err)
		return err
	}

	type CM struct {
		ClusterName string `json:"cluster_name"`
		Region      string `json:"region"`
		Count       int8   `json:"count"`
	}
	slog.Info("[original] ksctl-state.ksctl.cm", "entire", cm.Data)

	fmt.Println("ConfigMap Data:")
	if v, ok := cm.Data["state.json"]; ok {
		slog.Info("ksctl-state.ksctl.cm", "state.json", v)
		var updatedData CM

		if err := json.Unmarshal([]byte(v), &updatedData); err != nil {
			return err
		}
		updatedData.Count++
		raw, err := json.Marshal(updatedData)
		if err != nil {
			return err
		}

		cm.Data["state.json"] = string(raw)

	} else {
		return fmt.Errorf("not found the correct key in the configmap")
	}

	updated, err := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	slog.Info("[updated] ksctl-state.ksctl.cm", "entire", updated.Data)

	return nil
}

func (db *Store) Connect(ctx context.Context) error {
	db.context = ctx

	config, err := rest.InClusterConfig()
	if err != nil {
		return log.NewError("Error loading in-cluster config: %v\n", err)
	}

	db.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return log.NewError("Error creating Kubernetes client: %v\n", err)
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

// Read implements resources.StorageFactory.
func (db *Store) Read() (*types.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.kubernetes.configmap.Read", "Store", db)

	if c, ok := db.isPresent(); ok {
		var result *types.StorageDocument
		raw := c.BinaryData[K8S_STATE_FILE_NAME]

		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, err
		}

		return result, nil
	}
	return nil, log.NewError("cluster not present")
}

// ReadCredentials implements resources.StorageFactory.
func (db *Store) ReadCredentials(_ consts.KsctlCloud) (*types.CredentialsDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.kubernetes.configmap.ReadCreds", "Store", db)

	if c, ok := db.isPresent(); ok {
		var result *types.CredentialsDocument
		raw := c.BinaryData[K8S_CREDENTIALS_NAME]

		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, err
		}

		return result, nil
	}
	return nil, log.NewError("cluster not present")
}

func generateConfigMap(name string, namespace string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		BinaryData: map[string][]byte{},
	}
}

// Write implements resources.StorageFactory.
func (db *Store) Write(data *types.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.kubernetes.configmap.Write", "Store", db)

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if c, ok := db.isPresent(); ok {
		c.BinaryData[K8S_STATE_FILE_NAME] = raw

		if _, err := db.clientSet.CoreV1().
			ConfigMaps(K8S_NAMESPACE).
			Update(db.context, c, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}
	c := generateConfigMap(K8S_CONFIGMAP_NAME, K8S_NAMESPACE)
	c.BinaryData[K8S_STATE_FILE_NAME] = raw

	if _, err := db.clientSet.CoreV1().
		ConfigMaps(K8S_NAMESPACE).
		Update(db.context, c, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

// WriteCredentials implements resources.StorageFactory.
func (db *Store) WriteCredentials(_ consts.KsctlCloud, data *types.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.kubernetes.configmap.WriteCreds", "Store", db)

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if c, ok := db.isPresent(); ok {
		c.BinaryData[K8S_CREDENTIALS_NAME] = raw

		if _, err := db.clientSet.CoreV1().
			ConfigMaps(K8S_NAMESPACE).
			Update(db.context, c, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}
	c := generateConfigMap(K8S_CONFIGMAP_NAME, K8S_NAMESPACE)
	c.BinaryData[K8S_CREDENTIALS_NAME] = raw

	if _, err := db.clientSet.CoreV1().
		ConfigMaps(K8S_NAMESPACE).
		Update(db.context, c, metav1.UpdateOptions{}); err != nil {
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

	log.Debug("storage.external.kubernetes.configmap.Setup", "Store", db)
	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.kubernetes.configmap.Delete", "Store", db)

	if c, ok := db.isPresent(); !ok {
		return log.NewError("cluster doesn't exist")
	} else {
		clear(c.BinaryData[K8S_STATE_FILE_NAME])
		_, err := db.clientSet.CoreV1().ConfigMaps(K8S_NAMESPACE).Update(db.context, c, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	}
}

func (db *Store) isPresent() (*v1.ConfigMap, bool) {
	c, err := db.clientSet.CoreV1().ConfigMaps(K8S_NAMESPACE).Get(db.context, K8S_CONFIGMAP_NAME, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return nil, false
	}
	return c, true
}

func (db *Store) isPresentCreds() (*v1.ConfigMap, bool) {
	c, err := db.clientSet.CoreV1().ConfigMaps(K8S_NAMESPACE).Get(db.context, K8S_CONFIGMAP_NAME, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return nil, false
	}
	return c, true
}

// func (db *Store) clusterPresent() error {
// 	//c := db.databaseClient.Collection(db.cloudProvider).FindOne(db.context, getClusterFilters(db))
// 	//if c.Err() != nil {
// 	//	return c.Err()
// 	//} else {
// 	//	var x *types.StorageDocument
// 	//	err := c.Decode(&x)
// 	//	if err != nil {
// 	//		return fmt.Errorf("unable to read data")
// 	//	}
// 	//}
// 	////if c.RemainingBatchLength() == 1 {
// 	////	return fmt.Errorf("not present")
// 	////}
// 	//return nil
// }

func (db *Store) AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	//db.mu.Lock()
	//defer db.mu.Unlock()
	//db.wg.Add(1)
	//defer db.wg.Done()
	//
	//err := db.Setup(cloud, region, clusterName, clusterType)
	//if err != nil {
	//	return err
	//}
	//
	//return db.clusterPresent()
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
	log.Debug("storage.external.mongodb.GetOneOrMoreClusters", "filter", filters, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*types.StorageDocument)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {

			c, err := db.databaseClient.Collection(cloud).Find(db.context, bson.M{
				"cloud_provider": cloud,
				"cluster_type":   clusterType,
			})
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}

			var clusters []*types.StorageDocument
			for c.Next(context.Background()) {
				var result *types.StorageDocument
				if err := c.Decode(&result); err != nil {
					c.Close(context.Background())
					return nil, err
				}
				clusters = append(clusters, result)
			}
			c.Close(context.Background())

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], clusters...)
			log.Debug("storage.external.mongodb.GetOneOrMoreClusters", "clusterInfo", clustersInfo)
		}
	}

	return clustersInfo, nil
}
