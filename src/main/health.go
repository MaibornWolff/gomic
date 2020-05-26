package main

import (
	"context"
	"errors"
	"github.com/etherlabsio/healthcheck"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func handleHealthRequest(mongoClient *mongo.Client, rabbitConnectionIsClosed chan *amqp.Error) http.Handler {
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
					case err := <-rabbitConnectionIsClosed:
						if err != nil {
							return err
						}
						return errors.New("Connection to RabbitMQ is closed")
					default:
					}
					return nil
				},
			),
		),
	)
}
