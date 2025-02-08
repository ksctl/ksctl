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

package mongodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/storage"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	ctx      context.Context
	l        logger.Logger
	client   *mongo.Client
	mongoURI string
	//context        context.Context
	databaseClient *mongo.Database

	cloudProvider string
	clusterType   string
	clusterName   string
	region        string

	userid any

	mu *sync.Mutex
	wg *sync.WaitGroup
}

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

func (db *Store) Export(filters map[consts.KsctlSearchFilter]string) (*storage.StateExportImport, error) {

	var cpyS = db
	copyStore(db, cpyS) // for storing the state of the store before import was called!

	dest := new(storage.StateExportImport)

	_cloud := filters[consts.Cloud]
	_clusterType := filters[consts.ClusterType]
	_clusterName := filters[consts.Name]
	_region := filters[consts.Region]

	stateClustersForTypes, err := db.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
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

	copyStore(cpyS, db) // for restoring the state of the store before import was called!
	return dest, nil
}

func (db *Store) Import(src *storage.StateExportImport) error {
	states := src.Clusters

	var cpyS = db
	copyStore(db, cpyS) // for storing the state of the store before import was called!

	for _, state := range states {
		cloud := state.InfraProvider
		region := state.Region
		clusterName := state.ClusterName
		clusterType := consts.KsctlClusterType(state.ClusterType)

		if err := db.Setup(cloud, region, clusterName, clusterType); err != nil {
			return err
		}

		if err := db.Write(state); err != nil {
			return err
		}
	}

	copyStore(cpyS, db) // for restoring the state of the store before import was called!
	return nil
}

func NewClient(parentCtx context.Context, _log logger.Logger) *Store {
	return &Store{
		ctx: context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreExtMongo)),
		l:   _log,
		mu:  &sync.Mutex{},
		wg:  &sync.WaitGroup{},
	}
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

func URIAssembler(creds statefile.CredentialsMongodb) string {

	schema := "mongodb"
	if creds.SRV {
		schema = "mongodb+srv"
	}

	u := url.URL{
		Scheme: schema,
		User:   url.UserPassword(creds.Username, creds.Password),
		Host: func() string {
			d := creds.Domain
			if creds.Port != nil {
				d = fmt.Sprintf("%s:%d", creds.Domain, *creds.Port)
			}
			return d
		}(),
	}
	u.Query().Add("retryWrites", "true")
	u.Query().Add("w", "majority")

	return u.String()
}

func (db *Store) Connect() error {
	mongoCreds, ok := config.IsContextPresent(db.ctx, consts.KsctlMongodbCredentials)
	if !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			db.l.NewError(db.ctx, "missing mongodb credentials"),
		)
	}
	extractedCreds := statefile.CredentialsMongodb{}
	if err := json.Unmarshal([]byte(mongoCreds), &extractedCreds); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			db.l.NewError(db.ctx, "failed to get the creds", "Reason", err),
		)
	}

	db.mongoURI = URIAssembler(extractedCreds)

	opts := mongoOptions.Client().ApplyURI(db.mongoURI)
	var err error

	db.client, err = mongo.Connect(db.ctx, opts)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			db.l.NewError(db.ctx, "MongoDB failed to connect", "Reason", err),
		)
	}

	if err := db.client.Database("admin").RunCommand(db.ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			db.l.NewError(db.ctx, "MongoDB failed to ping pong the database", "Reason", err),
		)
	}

	if v, ok := config.IsContextPresent(db.ctx, consts.KsctlContextUserID); ok {
		db.userid = v
	} else {
		db.userid = "default"
	}

	switch o := db.userid.(type) {
	case string:
		db.databaseClient = db.client.Database(getUserDatabase(o))
	default:
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			db.l.NewError(db.ctx, "invalid type for context value `USERID`"),
		)
	}

	db.l.Debug(db.ctx, "CONN to MongoDB")

	return nil
}

func (db *Store) disconnect() error {
	return nil
}

func (db *Store) Kill() error {
	db.wg.Wait()
	defer db.l.Debug(db.ctx, "Mongodb Storage Got Killed")

	return db.disconnect()
}

func (db *Store) Read() (*statefile.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	if ret, err := db.isPresent(); err == nil {
		var result *statefile.StorageDocument
		err := ret.Decode(&result)
		if err != nil {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				db.l.NewError(db.ctx, "failed to deserialize the state", "Reason", err),
			)
		}

		return result, nil
	} else {
		return nil, err
	}

}

func (db *Store) Write(data *statefile.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	bsonMap, err := bson.Marshal(data)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			db.l.NewError(db.ctx, "failed to serialize the state", "Reason", err),
		)
	}

	if _, err := db.isPresent(); err == nil {
		res := db.databaseClient.Collection(db.cloudProvider).FindOneAndReplace(db.ctx, getClusterFilters(db), bsonMap)
		if _err := res.Err(); _err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				db.l.NewError(db.ctx, "failed to update the state", "Reason", _err),
			)
		}
	} else {
		_, _err := db.databaseClient.Collection(db.cloudProvider).InsertOne(db.ctx, bsonMap)
		if _err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				db.l.NewError(db.ctx, "failed to write state", "Reason", _err),
			)
		}
	}
	return nil
}

func (db *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudLocal:
		db.cloudProvider = string(cloud)
	default:
		return ksctlErrors.NewError(ksctlErrors.ErrInvalidCloudProvider)
	}
	if clusterType != consts.ClusterTypeSelfMang && clusterType != consts.ClusterTypeMang {
		return ksctlErrors.NewError(ksctlErrors.ErrInvalidClusterType)
	}

	db.clusterName = clusterName
	db.region = region
	db.clusterType = string(clusterType)

	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	if _, err := db.isPresent(); err != nil {
		return err
	}
	_, err := db.databaseClient.Collection(db.cloudProvider).DeleteOne(db.ctx, getClusterFilters(db))
	if err != nil {
		return err
	}

	return nil
}

func (db *Store) isPresent() (*mongo.SingleResult, error) {
	c := db.databaseClient.Collection(db.cloudProvider).FindOne(db.ctx, getClusterFilters(db))
	if c.Err() != nil {
		if errors.Is(c.Err(), mongo.ErrNoDocuments) {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				db.l.NewError(db.ctx, "no matching cluster present"),
			)
		}
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			db.l.NewError(db.ctx, "failed to get cluster", "Reason", c.Err()),
		)
	}
	return c, nil
}

func (db *Store) isPresentCreds(cloud consts.KsctlCloud) (*mongo.SingleResult, error) {
	c := db.databaseClient.Collection(CredsCollection).FindOne(db.ctx, getCredentialsFilters(cloud))
	if c.Err() != nil {
		if errors.Is(c.Err(), mongo.ErrNoDocuments) {
			return nil, ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				db.l.NewError(db.ctx, "no matching credentials present"),
			)
		}
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrNilCredentials,
			db.l.NewError(db.ctx, "failed to get credentials", "Reason", c.Err()),
		)
	}
	return c, nil
}

func (db *Store) clusterPresent() error {
	c := db.databaseClient.Collection(db.cloudProvider).FindOne(db.ctx, getClusterFilters(db))
	if c.Err() != nil {
		if errors.Is(c.Err(), mongo.ErrNoDocuments) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrNoMatchingRecordsFound,
				db.l.NewError(db.ctx, "no matching credentials present"),
			)
		}
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInternal,
			db.l.NewError(db.ctx, "failed to get credentials", "Reason", c.Err()),
		)
	} else {
		var x *statefile.StorageDocument
		err := c.Decode(&x)
		if err != nil {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInternal,
				db.l.NewError(db.ctx, "failed to deserialize the state", "Reason", err),
			)
		}
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

func (db *Store) GetOneOrMoreClusters(filters map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*statefile.StorageDocument, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()
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
	db.l.Debug(db.ctx, "storage.external.mongodb.GetOneOrMoreClusters", "filter", filters, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*statefile.StorageDocument)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {

			c, err := db.databaseClient.Collection(cloud).Find(db.ctx, bson.M{
				"cloud_provider": cloud,
				"cluster_type":   clusterType,
			})
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, ksctlErrors.WrapError(
					ksctlErrors.ErrInternal,
					db.l.NewError(db.ctx, "failed to find the state", "Reason", err),
				)
			}

			var clusters []*statefile.StorageDocument
			for c.Next(context.Background()) {
				var result *statefile.StorageDocument
				if err := c.Decode(&result); err != nil {
					_ = c.Close(context.Background())
					return nil, ksctlErrors.WrapError(
						ksctlErrors.ErrInternal,
						db.l.NewError(db.ctx, "failed to deserialize the state", "Reason", err),
					)
				}
				clusters = append(clusters, result)
			}
			_ = c.Close(context.Background())

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], clusters...)
		}
	}

	return clustersInfo, nil
}
