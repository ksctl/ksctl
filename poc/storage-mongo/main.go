package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListDatabases(client *mongo.Client) {
	out, err := client.ListDatabases(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	fmt.Println("@@ List of databases @@")
	for _, i := range out.Databases {
		fmt.Println(i.Name)
	}
}

func main() {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://dipankar:1234@cluster0.zako1lr.mongodb.net/?retryWrites=true&w=majority").SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	ListDatabases(client)

	res, err := client.Database("ksctl").Collection("civo").InsertOne(context.TODO(), bson.M{"key12312": "value2342wq"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", res)

	cursor, err := client.Database("ksctl").Collection("civo").Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	defer cursor.Close(context.Background())

	// Iterate through the cursor and print each document
	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}

	ListDatabases(client)

	if _, err := client.Database("ksctl").Collection("civo").DeleteMany(context.TODO(), bson.D{}); err != nil {
		panic(err)
	}

	fmt.Println("Deleted all the documents/records")
}
