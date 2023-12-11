package main

import (
	"fmt"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

var storage ConfigurationStore

func TestMain(m *testing.M) {

	_ = os.Setenv("MONGODB_HOSTNAME", "cluster0.fufupwy.mongodb.net")
	_ = os.Setenv("MONGODB_USER", "dipankar")
	_ = os.Setenv("MONGODB_PASSWORD", "1234")

	hostname := os.Getenv("MONGODB_HOSTNAME")
	user := os.Getenv("MONGODB_USER")
	password := os.Getenv("MONGODB_PASSWORD")

	var err error
	storage, err = NewClient(Options{
		Hostname: hostname,
		Username: user,
		Password: password})
	if err != nil {
		panic(err)
	}

	exitVal := m.Run()

	defer func() {
		if err := storage.Disconnect(); err != nil {
			panic(err)
		}
	}()

	os.Exit(exitVal)
}

func TestPing(t *testing.T) {
	if err := storage.Ping(); err != nil {
		panic(err)
	}
}

func TestListDatabases(t *testing.T) {
	if databases, err := storage.ListDatabases(); err != nil {
		return
	} else {
		fmt.Println(databases)
	}
}

func TestWrite(t *testing.T) {

	if err := storage.Write("civo", StorageConfiguration{ClusterType: "managed", Region: "eastus", ClusterName: "demo", OtherInfo: "nice"}); err != nil {
		t.Fatalf("Unable to write: %v", err)
	}

	if err := storage.Write("civo", StorageConfiguration{ClusterType: "ha", Region: "eastus1", ClusterName: "demo", OtherInfo: "nice"}); err != nil {
		t.Fatalf("Unable to write: %v", err)
	}

	if err := storage.Write("civo", StorageConfiguration{ClusterType: "managed", Region: "westus", ClusterName: "demo1", OtherInfo: "nice"}); err != nil {
		t.Fatalf("Unable to write: %v", err)
	}
}

func TestGetAll(t *testing.T) {
	data, err := storage.GetAllClusters("civo", bson.M{"clustertype": "ha"})

	if err != nil {
		t.Fatalf("unable to get all the clusters %v", err)
	}
	fmt.Printf("[[ CIVO HA ]]%+v\n", data)

	data, err = storage.GetAllClusters("civo", bson.M{"clustertype": "managed"})

	if err != nil {
		t.Fatalf("unable to get all the clusters %v", err)
	}
	fmt.Printf("[[ CIVO Managed ]]%+v\n", data)
}

func TestRead(t *testing.T) {

	if data, err := storage.ReadOne("civo", "eastus", "demo", "managed"); err != nil {
		t.Fatalf("Unable to Read: %v", err)
	} else {
		fmt.Printf("%+v\n", data)
	}
}

func TestPresent(t *testing.T) {
	if ok := storage.IsPresent("civo", "eastus", "demo", "managed"); !ok {
		t.Fatalf("expected that civo, eastus, demo, managed is present but got its false")
	}

	if ok := storage.IsPresent("azure", "eastus", "demo", "managed"); ok {
		t.Fatalf("expected that azure, eastus, demo, managed is absent but got its true")
	}
}

func TestDelete(t *testing.T) {

	if err := storage.DeleteOne("civo", "eastus", "demo", "managed"); err != nil {
		t.Fatalf("Unable to Delete: %v", err)
	}
}

func TestDeleteAll(t *testing.T) {

	if err := storage.DeleteAllInCloud("civo"); err != nil {
		t.Fatalf("Unable to Delete: %v", err)
	}
}
