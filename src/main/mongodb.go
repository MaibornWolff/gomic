package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Foo struct {
	Bar string
}

const (
	DATABASE_NAME   = "ACME"
	COLLECTION_NAME = "foo"
)

var (
	mongoClient    *mongo.Client
	defaultContext context.Context
)

func init() {
	defaultContext, _ = context.WithTimeout(context.Background(), 10*time.Second)
}

func connectToMongo() {
	mongodbHost := os.Getenv("MONGODB_HOST")
	if mongodbHost == "" {
		log.Fatalf("MongoDB host name not supplied")
	}
	var err error
	mongoClient, err = mongo.NewClient(options.Client().ApplyURI(mongodbHost))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %s", err)
	}
	err = mongoClient.Connect(defaultContext)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	err = mongoClient.Ping(defaultContext, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %s", err)
	}
	log.Printf("Connected to MongoDB")
}

func insertData() {
	result, err := mongoClient.Database(DATABASE_NAME).Collection(COLLECTION_NAME).InsertOne(defaultContext, Foo{"Bar"})
	if err != nil {
		log.Printf("Failed to insert data: %s", err)
	} else {
		log.Printf("Inserted data: %s", result)
	}
}

func findData() (result Foo) {
	err := mongoClient.Database(DATABASE_NAME).Collection(COLLECTION_NAME).FindOne(context.TODO(), bson.D{{}}).Decode(&result)
	if err != nil {
		log.Printf("Failed to find data: %s", err)
	}
	return
}

func findDataHandler(w http.ResponseWriter, r *http.Request) {
	data := findData()
	log.Printf("Found data: %v", data)
	w.Write([]byte(data.Bar))
	return
}
