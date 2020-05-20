package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"maibornwolff.de/gomic/model"
	"maibornwolff.de/gomic/rabbitmq"
	"net/http"
)

func handleFindData(mongo *mongo.Client, database string, collection string, writer http.ResponseWriter) {
	cursor, err := mongo.Database(database).Collection(collection).Find(context.Background(), bson.M{"firstName": bson.M{"$exists": true}})
	if err != nil {
		log.Printf("Failed to find data: %s", err)
		return
	}
	defer cursor.Close(context.Background())

	persons := make([]model.Person, 1)
	for cursor.Next(context.Background()) {
		var person model.Person
		err = cursor.Decode(&person)
		if err != nil {
			log.Printf("Failed to decode Person: %s", err)
		}
		persons = append(persons, person)
	}
	log.Printf("Found %d Persons", len(persons))

	for _, person := range persons {
		_, err = writer.Write([]byte(fmt.Sprintf("%s %s\n", person.FirstName, person.LastName)))
		if err != nil {
			log.Printf("Failed to write response: %s", err)
			return
		}
	}
}

func handleIncomingMessage(data []byte, mongo *mongo.Client, database string, collection string, channel *amqp.Channel, exchange string, routingKey string) {
	log.Printf("Trying to insert incoming message into MongoDB: %s", string(data))

	var person model.Person
	err := json.Unmarshal(data, &person)
	if err != nil {
		log.Printf("Failed to unmarshal JSON to Person: %s", err)
	}

	_, err = mongo.Database(database).Collection(collection).InsertOne(context.Background(), person)
	if err != nil {
		log.Printf("Failed to insert Person into MongoDB: %s", err)
	}

	err = rabbitmq.Publish(channel, exchange, routingKey, data, true)
	if err != nil {
		log.Printf("Failed to publish message: %s", err)
	}
}
