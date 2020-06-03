package mongodb

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	connectionTimeout = 10 * time.Second
)

func Connect(ctx context.Context, mongodbHost string) (*mongo.Client, error) {
	timedContext, cancelTimedContext := context.WithTimeout(ctx, connectionTimeout)
	defer cancelTimedContext()

	client, err := mongo.NewClient(options.Client().ApplyURI(mongodbHost))
	if err != nil {
		return nil, fmt.Errorf("Failed to create client: %w", err)
	}

	err = client.Connect(timedContext)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize client: %w", err)
	}

	err = client.Ping(timedContext, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %w", err)
	}

	log.Info().Msg("Connected to MongoDB")

	return client, nil
}
