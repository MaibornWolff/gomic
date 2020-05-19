package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func Connect(mongodbHost string) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongodbHost))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %s", err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %s", err)
	}

	log.Printf("Connected to MongoDB")

	return client
}
