package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"maibornwolff.de/gomic/application"
	"maibornwolff.de/gomic/mongodb"
	"maibornwolff.de/gomic/rabbitmq"
	"os"
	"os/signal"
	"syscall"
)

var (
	mongodbHost       = getEnv("MONGODB_HOST", "")
	mongodbDatabase   = getEnv("MONGODB_DATABASE", "")
	mongodbCollection = getEnv("MONGODB_COLLECTION", "")

	rabbitmqHost                 = getEnv("RABBITMQ_HOST", "")
	rabbitmqIncomingExchange     = getEnv("RABBITMQ_INCOMING_EXCHANGE", "")
	rabbitmqIncomingExchangeType = getEnv("RABBITMQ_INCOMING_EXCHANGE_TYPE", "direct")
	rabbitmqQueue                = getEnv("RABBITMQ_QUEUE", "")
	rabbitmqBindingKey           = getEnv("RABBITMQ_BINDING_KEY", "")
	rabbitmqConsumerTag          = getEnv("RABBITMQ_CONSUMER_TAG", "")
	rabbitmqOutgoingExchange     = getEnv("RABBITMQ_OUTGOING_EXCHANGE", "")
	rabbitmqOutgoingExchangeType = getEnv("RABBITMQ_OUTGOING_EXCHANGE_TYPE", "direct")
	rabbitmqRoutingKey           = getEnv("RABBITMQ_ROUTING_KEY", "")

	httpServerPort = getEnv("HTTP_SERVER_PORT", "8080")
)

func main() {
	ctx := context.Background()

	mongoClient, err := mongodb.Connect(ctx, mongodbHost)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	defer mongoClient.Disconnect(ctx)

	rabbitConnection, rabbitConnectionIsClosed, rabbitChannel, err := rabbitmq.Connect(rabbitmqHost, true)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rabbitConnection.Close()

	err = rabbitmq.DeclareSimpleExchange(rabbitChannel, rabbitmqIncomingExchange, rabbitmqIncomingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare incoming exchange: %s", err)
	}

	err = rabbitmq.DeclareSimpleExchange(rabbitChannel, rabbitmqOutgoingExchange, rabbitmqOutgoingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare outgoing exchange: %s", err)
	}

	cancelRabbitConsumer, err := rabbitmq.Consume(
		rabbitChannel, rabbitmqIncomingExchange, rabbitmqQueue, rabbitmqBindingKey, rabbitmqConsumerTag,
		func(data []byte) {
			application.HandleIncomingMessage(ctx, data, mongoClient, mongodbDatabase, mongodbCollection, rabbitChannel, rabbitmqOutgoingExchange, rabbitmqRoutingKey)
		})
	if err != nil {
		log.Fatalf("Failed to consume: %s", err)
	}
	defer cancelRabbitConsumer()

	router := gin.Default()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", gin.WrapH(handleHealthRequest(mongoClient, rabbitConnectionIsClosed)))

	router.GET("/persons", func(ctx *gin.Context) {
		application.HandlePersonsRequest(ctx, mongoClient, mongodbDatabase, mongodbCollection)
	})

	go func() {
		err := router.Run(fmt.Sprintf(":%s", httpServerPort))
		if err != nil {
			log.Fatalf("Failed to listen and serve: %s", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	log.Printf("Received signal to shutdown")
}
