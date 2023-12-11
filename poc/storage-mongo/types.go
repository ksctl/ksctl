package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoServer struct {
	client         *mongo.Client
	mongoURI       string
	context        context.Context
	databaseClient *mongo.Database
}

type Options struct {
	Hostname string
	Username string
	Password string
}

type ConfigurationStore interface {
	ListDatabases() ([]string, error)

	Disconnect() error

	Write(cloud string, data StorageDocument) error

	ReadOne(cloud, region, clustername, clusterType string) (StorageDocument, error)

	DeleteOne(cloud, region, clustername, clusterType string) error

	Ping() error

	IsPresent(cloud, region, clustername, clusterType string) bool

	GetAllClusters(cloud string, filters bson.M) ([]StorageDocument, error)

	DeleteAllInCloud(cloud string) error
}
