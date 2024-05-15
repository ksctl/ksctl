package mongodb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
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

var (
	log      types.LoggerFactory
	storeCtx context.Context
)

const (
	CredsCollection string = "credentials"
)

func copyStore(src *Store, dest *Store) {
	dest.cloudProvider = src.cloudProvider
	dest.clusterName = src.clusterName
	dest.clusterType = src.clusterType
	dest.region = src.region
	dest.userid = src.userid
}

func (s *Store) Export(filters map[consts.KsctlSearchFilter]string) (*types.StorageStateExportImport, error) {

	var cpyS *Store = s
	copyStore(s, cpyS) // for storing the state of the store before import was called!

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
			consts.CloudLocal,
			consts.CloudAzure,
		} {
			_v, _err := s.ReadCredentials(constsCloud)

			if _err != nil {
				if errors.Is(_err, mongo.ErrNoDocuments) {
					continue
				} else {
					return nil, _err
				}
			}
			dest.Credentials = append(dest.Credentials, _v)
		}
	} else {
		_v, _err := s.ReadCredentials(consts.KsctlCloud(_cloud))
		if _err != nil {
			return nil, _err
		}
		dest.Credentials = append(dest.Credentials, _v)
	}

	copyStore(cpyS, s) // for restoring the state of the store before import was called!
	return dest, nil
}

func (s *Store) Import(src *types.StorageStateExportImport) error {
	creds := src.Credentials
	states := src.Clusters

	var cpyS *Store = s
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

func NewClient(parentCtx context.Context, _log types.LoggerFactory) *Store {
	storeCtx = context.WithValue(parentCtx, consts.ContextModuleNameKey, string(consts.StoreExtMongo))
	log = _log
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

func fetchCreds() (string, error) {
	connURI := os.Getenv("MONGODB_URI")

	if len(connURI) == 0 {
		return "", log.NewError(storeCtx, "environment vars not set for the storage to work.", "Hint", "mongodb://${username}:${password}@${domain}:${port} or mongo+atlas mongodb+srv://${username}:${password}@${domain}")
	}

	return fmt.Sprintf("%s/?retryWrites=true&w=majority", connURI), nil
}

func ExportEndpoint() (map[string][]byte, error) {
	v, err := fetchCreds()
	if err != nil {
		return nil, err
	}

	return map[string][]byte{
		"MONGODB_URI": []byte(v),
	}, nil
}

func (db *Store) Connect(ctx context.Context) error {
	db.context = context.Background()

	v, err := fetchCreds()
	if err != nil {
		return err
	}

	db.mongoURI = v

	opts := mongoOptions.Client().ApplyURI(db.mongoURI)

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
	log.Success(storeCtx, "CONN to MongoDB")

	return nil
}

func (db *Store) disconnect() error {
	return nil
}

func (db *Store) Kill() error {
	db.wg.Wait()
	defer log.Success(storeCtx, "Mongodb Storage Got Killed")

	return db.disconnect()
}

// Read implements types.StorageFactory.
func (db *Store) Read() (*storageTypes.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.external.mongodb.Read", "Store", db)

	if db.isPresent() {
		ret := db.databaseClient.Collection(db.cloudProvider).FindOne(db.context, getClusterFilters(db))
		if ret.Err() != nil {
			return nil, ret.Err()
		}

		var result *storageTypes.StorageDocument
		err := ret.Decode(&result)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
	return nil, mongo.ErrNoDocuments
}

// ReadCredentials implements types.StorageFactory.
func (db *Store) ReadCredentials(cloud consts.KsctlCloud) (*storageTypes.CredentialsDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	log.Debug(storeCtx, "storage.external.mongodb.ReadCredentials", "Store", db)

	if db.isPresentCreds(cloud) {
		ret := db.databaseClient.Collection(CredsCollection).FindOne(db.context, getCredentialsFilters(cloud))
		if ret.Err() != nil {
			return nil, ret.Err()
		}
		var result *storageTypes.CredentialsDocument
		err := ret.Decode(&result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, mongo.ErrNoDocuments
}

// Write implements types.StorageFactory.
func (db *Store) Write(data *storageTypes.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
	log.Debug(storeCtx, "storage.external.mongodb.Write", "Store", db)

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

// WriteCredentials implements types.StorageFactory.
func (db *Store) WriteCredentials(cloud consts.KsctlCloud, data *storageTypes.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.external.mongodb.WriteCredentials", "Store", db)

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

	log.Debug(storeCtx, "storage.external.mongodb.Setup", "Store", db)
	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	log.Debug(storeCtx, "storage.external.mongodb.Delete", "Store", db)

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
		var x *storageTypes.StorageDocument
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
	log.Debug(storeCtx, "storage.external.mongodb.GetOneOrMoreClusters", "filter", filters, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*storageTypes.StorageDocument)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {

			c, err := db.databaseClient.Collection(cloud).Find(db.context, bson.M{
				"cloud_provider": cloud,
				"cluster_type":   clusterType,
			})
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}

			var clusters []*storageTypes.StorageDocument
			for c.Next(context.Background()) {
				var result *storageTypes.StorageDocument
				if err := c.Decode(&result); err != nil {
					c.Close(context.Background())
					return nil, err
				}
				clusters = append(clusters, result)
			}
			c.Close(context.Background())

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], clusters...)
			log.Debug(storeCtx, "storage.external.mongodb.GetOneOrMoreClusters", "clusterInfo", clustersInfo)
		}
	}

	return clustersInfo, nil
}
