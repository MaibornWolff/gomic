package application

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"maibornwolff.de/gomic/model"
	"maibornwolff.de/gomic/rabbitmq"
)

func HandlePersonsRequest(ctx *gin.Context, mongoClient *mongo.Client, database string, collection string) {
	cursor, err := mongoClient.Database(database).Collection(collection).Find(ctx, bson.M{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to find documents")
		return
	}
	defer cursor.Close(ctx)

	persons := make([]model.Person, 0)
	err = cursor.All(ctx, &persons)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode documents")
		return
	}
	log.Info().Msgf("Found %d Persons", len(persons))

	ctx.JSON(200, persons)
}

func HandleIncomingMessage(ctx context.Context, personData []byte, mongoClient *mongo.Client, database string, collection string, rabbitChannel *amqp.Channel, exchange string, routingKey string) {
	log.Info().Bytes("personData", personData).Msg("Handle incoming RabbitMQ message")

	var person model.Person
	err := json.Unmarshal(personData, &person)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal JSON to Person")
		return
	}

	_, err = mongoClient.Database(database).Collection(collection).InsertOne(ctx, person)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert Person into MongoDB")
		return
	}

	upperCasedPerson := person.WithUpperCase()
	upperCasedPersonData, err := json.Marshal(upperCasedPerson)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal Person in upper case to JSON")
		return
	}

	err = rabbitmq.Publish(rabbitChannel, exchange, routingKey, upperCasedPersonData, "application/json")
	if err != nil {
		log.Error().Err(err).Msg("Failed to publish RabbitMQ message")
	}
}
