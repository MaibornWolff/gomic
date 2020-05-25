package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/etherlabsio/healthcheck"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"maibornwolff.de/gomic/model"
	"maibornwolff.de/gomic/rabbitmq"
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

func handlePersonsRequest(ctx context.Context, mongoClient *mongo.Client, database string, collection string, responseWriter http.ResponseWriter) {
	cursor, err := mongoClient.Database(database).Collection(collection).Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to find documents: %s", err)
		return
	}
	defer cursor.Close(ctx)

	persons := make([]model.Person, 0)
	err = cursor.All(ctx, &persons)
	if err != nil {
		log.Printf("Failed to decode documents: %s", err)
		return
	}
	log.Printf("Found %d Persons", len(persons))

	for _, person := range persons {
		_, err = responseWriter.Write([]byte(person.String() + "\n"))
		if err != nil {
			log.Printf("Failed to write response: %s", err)
			return
		}
	}
}

func handleIncomingMessage(ctx context.Context, personData []byte, mongoClient *mongo.Client, database string, collection string, rabbitChannel *amqp.Channel, exchange string, routingKey string) {
	log.Printf("Trying to insert incoming RabbitMQ message into MongoDB: %s", string(personData))

	var person model.Person
	err := json.Unmarshal(personData, &person)
	if err != nil {
		log.Printf("Failed to unmarshal JSON to Person: %s", err)
		return
	}

	_, err = mongoClient.Database(database).Collection(collection).InsertOne(ctx, person)
	if err != nil {
		log.Printf("Failed to insert Person into MongoDB: %s", err)
		return
	}

	upperCasedPerson := person.WithUpperCase()
	upperCasedPersonData, err := json.Marshal(upperCasedPerson)
	if err != nil {
		log.Printf("Failed to marshal Person in upper case to JSON: %s", err)
	}

	err = rabbitmq.Publish(rabbitChannel, exchange, routingKey, upperCasedPersonData)
	if err != nil {
		log.Printf("Failed to publish RabbitMQ message: %s", err)
	}
}
