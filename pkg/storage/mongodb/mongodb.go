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
	"errors"
	"fmt"
	"sync"

	"github.com/ksctl/ksctl/v2/pkg/config"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"github.com/ksctl/ksctl/v2/pkg/statefile"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConn struct {
	ctx    context.Context
	client *mongo.Client
	mu     *sync.Mutex
}

func NewDBClient(parentCtx context.Context, creds statefile.CredentialsMongodb) (*MongoConn, error) {
	db := &MongoConn{
		ctx: context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreExtMongo)),
		mu:  &sync.Mutex{},
	}

	uri := creds.URI

	opts := mongoOptions.Client().ApplyURI(uri)
	var err error

	db.client, err = mongo.Connect(db.ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("MongoDB failed to connect, Reason: %v", err)
	}

	if err := db.client.Database("admin").RunCommand(db.ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return nil, fmt.Errorf("MongoDB failed to ping pong the database, Reason: %v", err)
	}

	return db, nil
}

type Store struct {
	ctx            context.Context
	l              logger.Logger
	databaseClient *mongo.Database

	cloudProvider string
	clusterType   string
	clusterName   string
	region        string

	mu *sync.Mutex
	wg *sync.WaitGroup
}

func (conn *MongoConn) NewDatabaseClient(ksctlConfig context.Context, l logger.Logger) (*Store, error) {

	db := &Store{
		ctx: conn.ctx,
		l:   l,
		mu:  conn.mu,
		wg:  new(sync.WaitGroup),
	}

	userId := "cli"

	if v, ok := config.IsContextPresent(ksctlConfig, consts.KsctlContextUser); ok {
		userId = v
	}

	db.databaseClient = conn.client.Database(getUserDatabase(userId))

	return db, nil
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

func (db *Store) Connect(_ context.Context) error {

	db.l.Debug(db.ctx, "CONN to MongoDB")

	return nil
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
