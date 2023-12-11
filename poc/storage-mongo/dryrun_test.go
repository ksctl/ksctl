package main

import (
	"fmt"
	"os"
	"testing"
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

// res, err := client.Database("ksctl").Collection("civo").InsertOne(context.TODO(), bson.M{"key12312": "value2342wq"})
//
//	if err != nil {
//		panic(err)
//	}
//
// fmt.Printf("%v\n", res)
//
// cursor, err := client.Database("ksctl").Collection("civo").Find(context.TODO(), bson.M{})
//
//	if err != nil {
//		panic(err)
//	}
//
// defer cursor.Close(context.Background())
//
// // Iterate through the cursor and print each document
//
//	for cursor.Next(context.Background()) {
//		var result bson.M
//		if err := cursor.Decode(&result); err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println(result)
//	}
//
// ListDatabases(client)
//
//	if _, err := client.Database("ksctl").Collection("civo").DeleteMany(context.TODO(), bson.D{}); err != nil {
//		panic(err)
//	}
//
// fmt.Println("Deleted all the documents/records")

func TestWrite(t *testing.T) {}
