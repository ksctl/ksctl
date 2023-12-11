package main

import (
	"context"
	"fmt"
	"log"

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
	return db.client.Disconnect(db.context)
}

func (db *MongoServer) IsPresent(cloud, region, clustername, clusterType string) bool {

	c, err := db.client.Database(db.mongodbDatabase).Collection(cloud).Find(db.context, bson.M{
		"clustertype": clusterType,
		"region":      region,
		"clustername": clustername,
	})
	fmt.Printf("%+v\n", err)
	fmt.Printf("%+v\n", c)
	return err != mongo.ErrNoDocuments && c.RemainingBatchLength() == 1
}

func (db *MongoServer) GetAllClusters(cloud string, filters bson.M) ([]StorageConfiguration, error) {

	c, err := db.client.Database(db.mongodbDatabase).Collection(cloud).Find(db.context, filters)
	if err != nil {
		return nil, err
	}
	defer c.Close(context.Background())

	var clusters []StorageConfiguration
	for c.Next(context.Background()) {
		var result *StorageConfiguration
		if err := c.Decode(&result); err != nil {
			log.Fatal(err)
		}
		clusters = append(clusters, *result)
	}
	return clusters, nil
}

func (db *MongoServer) Write(cloud string, data StorageConfiguration) error {

	res, err := db.client.Database(db.mongodbDatabase).Collection(cloud).InsertOne(db.context, data)

	fmt.Println(res)

	return err
}

func (db *MongoServer) ReadOne(cloud, region, clustername, clusterType string) (*StorageConfiguration, error) {
	ret := db.client.Database(db.mongodbDatabase).Collection(cloud).FindOne(db.context, bson.M{
		"clustertype": clusterType,
		"region":      region,
		"clustername": clustername,
	})
	var result *StorageConfiguration
	err := ret.Decode(&result)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", result)
	return result, nil
}

func (db *MongoServer) DeleteOne(cloud, region, clustername, clusterType string) error {
	ret, err := db.client.Database(db.mongodbDatabase).Collection(cloud).DeleteOne(db.context, bson.M{
		"clustertype": clusterType,
		"region":      region,
		"clustername": clustername,
	})
	if err != nil {
		return err
	}

	fmt.Println("Deleted no of records:", ret.DeletedCount)

	return nil
}

func (db *MongoServer) DeleteAllInCloud(cloud string) error {
	ret, err := db.client.Database(db.mongodbDatabase).Collection(cloud).DeleteMany(db.context, bson.D{})
	if err != nil {
		return err
	}

	fmt.Println("Deleted no of records:", ret.DeletedCount)

	return nil
}

func NewClient(options Options) (ConfigurationStore, error) {
	client := &MongoServer{context: context.Background(), mongodbDatabase: "ksctl"}

	client.mongoURI = fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority", options.Username, options.Password, options.Hostname)

	opts := mongoOptions.Client().ApplyURI(client.mongoURI)

	var err error
	client.client, err = mongo.Connect(client.context, opts)
	if err != nil {
		return nil, fmt.Errorf("MongoDB failed to connect. Reason: %w", err)
	}

	return client, nil
}

func (db *MongoServer) Ping() error {

	if err := db.client.Database("admin").RunCommand(db.context, bson.D{{"ping", 1}}).Err(); err != nil {
		return err
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return nil
}
