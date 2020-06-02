package main

import (
	"context"
	"fmt"
	"github.com/etherlabsio/healthcheck"
	"go.mongodb.org/mongo-driver/mongo"
	"maibornwolff.de/gomic/rabbitmq"
	"net/http"
	"time"
)

func handleHealthRequest(mongoClient *mongo.Client, rabbitClient *rabbitmq.Client) http.Handler {
	return healthcheck.Handler(
		healthcheck.WithTimeout(5*time.Second),
		healthcheck.WithChecker(
			"mongodb", healthcheck.CheckerFunc(
				func(ctx context.Context) error {
					return mongoClient.Ping(ctx, nil)
				},
			),
		),
		healthcheck.WithChecker(
			"rabbitmq", healthcheck.CheckerFunc(
				func(ctx context.Context) error {
					select {
					case err := <-rabbitClient.ConnectionIsClosed:
						return fmt.Errorf("Connection to RabbitMQ is closed: %s", err)
					case err := <-rabbitClient.ChannelIsClosed:
						return fmt.Errorf("Channel to RabbitMQ is closed: %s", err)
					default:
					}
					return nil
				},
			),
		),
	)
}
