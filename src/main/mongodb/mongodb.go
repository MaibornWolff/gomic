package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func Connect(ctx context.Context, mongodbHost string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongodbHost))
	if err != nil {
		return nil, fmt.Errorf("Failed to create client: %s", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize client: %s", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %s", err)
	}

	log.Printf("Connected to MongoDB")

	return client, nil
}
