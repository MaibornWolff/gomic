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

func handleHealthRequest(mongo *mongo.Client, rabbitErrorChannel chan *amqp.Error) http.Handler {
	return healthcheck.Handler(
		healthcheck.WithTimeout(5*time.Second),
		healthcheck.WithChecker(
			"mongodb", healthcheck.CheckerFunc(
				func(ctx context.Context) error {
					return mongo.Ping(ctx, nil)
				},
			),
		),
		healthcheck.WithChecker(
			"rabbitmq", healthcheck.CheckerFunc(
				func(ctx context.Context) error {
					select {
					case errorPointer := <-rabbitErrorChannel:
						if errorPointer != nil {
							return *errorPointer
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

func handlePersonsRequest(ctx context.Context, mongo *mongo.Client, database string, collection string, writer http.ResponseWriter) {
	cursor, err := mongo.Database(database).Collection(collection).Find(ctx, bson.M{})
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
		_, err = writer.Write([]byte(person.String() + "\n"))
		if err != nil {
			log.Printf("Failed to write response: %s", err)
			return
		}
	}
}

func handleIncomingMessage(ctx context.Context, mongo *mongo.Client, personData []byte, database string, collection string, channel *amqp.Channel, exchange string, routingKey string) {
	log.Printf("Trying to insert incoming RabbitMQ message into MongoDB: %s", string(personData))

	var person model.Person
	err := json.Unmarshal(personData, &person)
	if err != nil {
		log.Printf("Failed to unmarshal JSON to Person: %s", err)
		return
	}

	_, err = mongo.Database(database).Collection(collection).InsertOne(ctx, person)
	if err != nil {
		log.Printf("Failed to insert Person into MongoDB: %s", err)
		return
	}

	upperCasedPerson := person.WithUpperCase()
	upperCasedPersonData, err := json.Marshal(upperCasedPerson)
	if err != nil {
		log.Printf("Failed to marshal Person in upper case to JSON: %s", err)
	}

	err = rabbitmq.Publish(channel, exchange, routingKey, upperCasedPersonData)
	if err != nil {
		log.Printf("Failed to publish RabbitMQ message: %s", err)
	}
}
