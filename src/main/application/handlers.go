package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"maibornwolff.de/gomic/model"
	"maibornwolff.de/gomic/rabbitmq"
	"net/http"
	"sync"
	"time"
)

const (
	handlerTimeout = 5 * time.Second
)

func HandlePersonsRequest(ginContext *gin.Context, mongoClient *mongo.Client, database string, collection string) {
	timedContext, cancelTimedContext := context.WithTimeout(ginContext, handlerTimeout)
	defer cancelTimedContext()

	cursor, err := mongoClient.Database(database).Collection(collection).Find(timedContext, bson.M{})
	if err != nil {
		abortWithError(ginContext, http.StatusInternalServerError, "Failed to find documents", err)
		return
	}
	defer cursor.Close(timedContext)

	persons := make([]model.Person, 0)
	err = cursor.All(ginContext, &persons)
	if err != nil {
		abortWithError(ginContext, http.StatusInternalServerError, "Failed to decode documents", err)
		return
	}

	log.Info().Msgf("Found %d Persons", len(persons))

	ginContext.JSON(http.StatusOK, persons)
}

func abortWithError(ginContext *gin.Context, httpStatusCode int, publicErrorMessage string, internalError error) {
	log.Error().Err(internalError).Msg(publicErrorMessage)
	ginContext.AbortWithStatusJSON(httpStatusCode, gin.H{
		"code":  httpStatusCode,
		"error": publicErrorMessage,
	})
}

func HandleIncomingMessage(ctx context.Context, wg *sync.WaitGroup, personData []byte, mongoClient *mongo.Client, database string, collection string, rabbitChannel *amqp.Channel, exchange string, routingKey string) error {
	log.Info().Bytes("personData", personData).Msg("Handle incoming RabbitMQ message")

	defer wg.Done()

	timedContext, cancelTimedContext := context.WithTimeout(ctx, handlerTimeout)
	defer cancelTimedContext()

	var person model.Person
	err := json.Unmarshal(personData, &person)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal JSON to Person: %w", err)
	}

	_, err = mongoClient.Database(database).Collection(collection).InsertOne(timedContext, person)
	if err != nil {
		return fmt.Errorf("Failed to insert Person into MongoDB: %w", err)
	}

	upperCasedPerson := person.WithUpperCase()
	upperCasedPersonData, err := json.Marshal(upperCasedPerson)
	if err != nil {
		return fmt.Errorf("Failed to marshal Person in upper case to JSON: %w", err)
	}

	err = rabbitmq.Publish(rabbitChannel, exchange, routingKey, upperCasedPersonData, "application/json")
	if err != nil {
		return fmt.Errorf("Failed to publish RabbitMQ message: %w", err)
	}

	return nil
}
