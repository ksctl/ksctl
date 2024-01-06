package mongodb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/kubesimplify/ksctl/pkg/logger"

	"github.com/kubesimplify/ksctl/internal/storage/types"
	"github.com/kubesimplify/ksctl/pkg/helpers/consts"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	client         *mongo.Client
	mongoURI       string
	context        context.Context
	databaseClient *mongo.Database

	cloudProvider string
	clusterType   string
	clusterName   string
	region        string

	userid any

	mu *sync.Mutex
	wg *sync.WaitGroup
}

var log resources.LoggerFactory

const (
	CredsCollection string = "credentials"
)

func InitStorage(logVerbosity int, logWriter io.Writer) resources.StorageFactory {
	log = logger.NewDefaultLogger(logVerbosity, logWriter)
	log.SetPackageName(string(consts.StoreExtMongo))
	return &Store{mu: &sync.Mutex{}, wg: &sync.WaitGroup{}}
}

func getUserDatabase(userid string) string {
	return fmt.Sprintf("ksctl-%s-db", userid)
}

func getClusterFilters(db *Store) bson.M {
	return bson.M{
		"cluster_type":   db.clusterType,
		"region":         db.region,
		"cluster_name":   db.clusterName,
		"cloud_provider": db.cloudProvider,
	}
}

func getCredentialsFilters(cloud consts.KsctlCloud) bson.M {
	return bson.M{
		"cloud_provider": cloud,
	}
}

func (db *Store) Connect(ctx context.Context) error {
	db.context = context.Background()

	Username := os.Getenv("MONGODB_USER")
	Password := os.Getenv("MONGODB_PASSWORD")
	DNSSeedlist := os.Getenv("MONGODB_DNS")
	if len(Username) == 0 || len(Password) == 0 || len(DNSSeedlist) == 0 {
		return fmt.Errorf("environment vars not set for the storage to work")
	}

	db.mongoURI = fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority", Username, Password, DNSSeedlist)

	opts := mongoOptions.Client().ApplyURI(db.mongoURI)

	var err error
	db.client, err = mongo.Connect(db.context, opts)
	if err != nil {
		return fmt.Errorf("MongoDB failed to connect. Reason: %w", err)
	}

	if err := db.client.Database("admin").RunCommand(db.context, bson.D{{"ping", 1}}).Err(); err != nil {
		return err
	}

	db.userid = ctx.Value("USERID")

	switch o := db.userid.(type) {
	case string:
		db.databaseClient = db.client.Database(getUserDatabase(o))
	default:
		return fmt.Errorf("invalid type for context value `USERID`")
	}
	log.Success("CONN to MongoDB")

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

	log.Debug("storage.external.mongodb.Read", "Store", db)

	if db.isPresent() {
		ret := db.databaseClient.Collection(db.cloudProvider).FindOne(db.context, getClusterFilters(db))
		if ret.Err() != nil {
			return nil, ret.Err()
		}

		var result *types.StorageDocument
		err := ret.Decode(&result)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
	return nil, fmt.Errorf("cluster not present")
}

// ReadCredentials implements resources.StorageFactory.
func (db *Store) ReadCredentials(cloud consts.KsctlCloud) (*types.CredentialsDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	log.Debug("storage.external.mongodb.ReadCredentials", "Store", db)

	if db.isPresentCreds(cloud) {
		ret := db.databaseClient.Collection(CredsCollection).FindOne(db.context, getCredentialsFilters(cloud))
		if ret.Err() != nil {
			return nil, ret.Err()
		}
		var result *types.CredentialsDocument
		err := ret.Decode(&result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, fmt.Errorf("creds not present")
}

// Write implements resources.StorageFactory.
func (db *Store) Write(data *types.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	log.Debug("storage.external.mongodb.Write", "Store", db)

	bsonMap, err := bson.Marshal(data)
	if err != nil {
		return err
	}

	if db.isPresent() {
		res := db.databaseClient.Collection(db.cloudProvider).FindOneAndReplace(db.context, getClusterFilters(db), bsonMap)
		if err := res.Err(); err != nil {
			return err
		}
		return nil
	}

	_, err = db.databaseClient.Collection(db.cloudProvider).InsertOne(db.context, bsonMap)
	return err
}

// WriteCredentials implements resources.StorageFactory.
func (db *Store) WriteCredentials(cloud consts.KsctlCloud, data *types.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.mongodb.WriteCredentials", "Store", db)

	bsonMap, err := bson.Marshal(data)
	if err != nil {
		return err
	}

	if db.isPresentCreds(cloud) {
		res := db.databaseClient.Collection(CredsCollection).FindOneAndReplace(db.context, getCredentialsFilters(cloud), bsonMap)
		if err := res.Err(); err != nil {
			return err
		}
		return nil
	}

	_, err = db.databaseClient.Collection(CredsCollection).InsertOne(db.context, bsonMap)
	return err
}

func (db *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
		db.cloudProvider = string(cloud)
	default:
		return errors.New("invalid cloud")
	}
	if clusterType != consts.ClusterTypeHa && clusterType != consts.ClusterTypeMang {
		return errors.New("invalid cluster type")
	}

	db.clusterName = clusterName
	db.region = region
	db.clusterType = string(clusterType)

	log.Debug("storage.external.mongodb.Setup", "Store", db)
	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug("storage.external.mongodb.Delete", "Store", db)

	if !db.isPresent() {
		return fmt.Errorf("cluster doesn't exist")
	}
	_, err := db.databaseClient.Collection(db.cloudProvider).DeleteOne(db.context, getClusterFilters(db))
	if err != nil {
		return err
	}

	return nil
}

func (db *Store) isPresent() bool {
	c, err := db.databaseClient.Collection(db.cloudProvider).Find(db.context, getClusterFilters(db))
	return !errors.Is(err, mongo.ErrNoDocuments) && c.RemainingBatchLength() == 1
}

func (db *Store) isPresentCreds(cloud consts.KsctlCloud) bool {
	c, err := db.databaseClient.Collection(CredsCollection).Find(db.context, getCredentialsFilters(cloud))
	return !errors.Is(err, mongo.ErrNoDocuments) && c.RemainingBatchLength() == 1
}

func (db *Store) clusterPresent() error {
	c := db.databaseClient.Collection(db.cloudProvider).FindOne(db.context, getClusterFilters(db))
	if c.Err() != nil {
		return c.Err()
	} else {
		var x *types.StorageDocument
		err := c.Decode(&x)
		if err != nil {
			return fmt.Errorf("unable to read data")
		}
	}
	//if c.RemainingBatchLength() == 1 {
	//	return fmt.Errorf("not present")
	//}
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
