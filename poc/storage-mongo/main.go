package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func (db *MongoServer) ListDatabases() ([]string, error) {
	out, err := db.client.ListDatabases(db.context, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("Mongodb unable to get databases: %w", err)
	}
	var databases []string
	for _, i := range out.Databases {
		databases = append(databases, i.Name)
	}
	return databases, nil
}

func (db *MongoServer) Disconnect() error {
	fmt.Println("Disconnecting mongodb")
	return db.client.Disconnect(db.context)
}

func (db *MongoServer) IsPresent(cloud, region, clustername, clusterType string) bool {

	c, err := db.databaseClient.Collection(cloud).Find(db.context, bson.M{
		"cluster_type": clusterType,
		"region":       region,
		"cluster_name": clustername,
	})
	return err != mongo.ErrNoDocuments && c.RemainingBatchLength() == 1
}

func (db *MongoServer) GetAllClusters(cloud string, filters bson.M) ([]StorageDocument, error) {

	c, err := db.databaseClient.Collection(cloud).Find(db.context, filters)
	if err != nil {
		return nil, err
	}
	defer c.Close(context.Background())

	var clusters []StorageDocument
	for c.Next(context.Background()) {
		var result StorageDocument
		if err := c.Decode(&result); err != nil {
			return nil, err
		}
		clusters = append(clusters, result)
	}
	return clusters, nil
}

func (db *MongoServer) Write(cloud string, data StorageDocument) error {
	bsonMap, err := bson.Marshal(data)
	if err != nil {
		return err
	}

	if db.IsPresent(cloud, data.Region, data.ClusterName, data.ClusterType) {
		res := db.databaseClient.Collection(cloud).FindOneAndReplace(db.context, bson.M{
			"cluster_type": data.ClusterType,
			"region":       data.Region,
			"cluster_name": data.ClusterName,
		}, bsonMap)
		if err := res.Err(); err != nil {
			return err
		}
		return nil
	}

	_, err = db.databaseClient.Collection(cloud).InsertOne(db.context, bsonMap)
	return err
}

func (db *MongoServer) ReadOne(cloud, region, clustername, clusterType string) (StorageDocument, error) {
	ret := db.databaseClient.Collection(cloud).FindOne(db.context, bson.M{
		"cluster_type": clusterType,
		"region":       region,
		"cluster_name": clustername,
	})
	var result StorageDocument
	err := ret.Decode(&result)
	if err != nil {
		return StorageDocument{}, err
	}

	return result, nil
}

func (db *MongoServer) DeleteOne(cloud, region, clustername, clusterType string) error {
	ret, err := db.databaseClient.Collection(cloud).DeleteOne(db.context, bson.M{
		"cluster_type": clusterType,
		"region":       region,
		"cluster_name": clustername,
	})
	if err != nil {
		return err
	}

	fmt.Println("Deleted no of records:", ret.DeletedCount)

	return nil
}

func (db *MongoServer) DeleteAllInCloud(cloud string) error {
	ret, err := db.databaseClient.Collection(cloud).DeleteMany(db.context, bson.D{})
	if err != nil {
		return err
	}

	fmt.Println("Deleted no of records:", ret.DeletedCount)

	return nil
}

func NewClient(options Options) (ConfigurationStore, error) {
	client := &MongoServer{context: context.Background()}

	client.mongoURI = fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority", options.Username, options.Password, options.Hostname)

	opts := mongoOptions.Client().ApplyURI(client.mongoURI)

	mongoClient, err := mongo.Connect(client.context, opts)
	if err != nil {
		return nil, fmt.Errorf("MongoDB failed to connect. Reason: %w", err)
	}

	if err := mongoClient.Database("admin").RunCommand(client.context, bson.D{{"ping", 1}}).Err(); err != nil {
		return nil, err
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	client.databaseClient = mongoClient.Database("ksctl")
	client.client = mongoClient

	return client, nil
}
