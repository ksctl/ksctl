package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoServer struct {
	client          *mongo.Client
	mongoURI        string
	context         context.Context
	mongodbDatabase string
}

type Options struct {
	Hostname string
	Username string
	Password string
}

type StorageConfiguration struct{}

type ConfigurationStore interface {
	ListDatabases() ([]string, error)
	Disconnect() error

	Write(ctx context.Context, cloud string, data StorageConfiguration) error

	Read(ctx context.Context, cloud, region, clustername string) (StorageConfiguration, error)

	Ping() error
}
