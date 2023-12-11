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
	return db.client.Disconnect(db.context)
}

func (db *MongoServer) Write(ctx context.Context, cloud string, data StorageConfiguration) error {
	return nil
}

func (db *MongoServer) Read(ctx context.Context, cloud, region, clustername string) (StorageConfiguration, error) {
	return StorageConfiguration{}, nil
}

func NewClient(options Options) (ConfigurationStore, error) {
	client := &MongoServer{context: context.Background(), mongodbDatabase: "ksctl"}

	serverAPI := mongoOptions.ServerAPI(mongoOptions.ServerAPIVersion1)

	client.mongoURI = fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority", options.Username, options.Password, options.Hostname)

	opts := mongoOptions.Client().ApplyURI(client.mongoURI).SetServerAPIOptions(serverAPI)

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
