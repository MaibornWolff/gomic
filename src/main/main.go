package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"maibornwolff.de/gomic/mongodb"
	"maibornwolff.de/gomic/rabbitmq"
	"net/http"
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

	mongo, err := mongodb.Connect(ctx, mongodbHost)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	defer mongo.Disconnect(ctx)

	rabbit, channel, err := rabbitmq.Connect(rabbitmqHost, true)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rabbit.Close()

	err = rabbitmq.DeclareSimpleExchange(channel, rabbitmqIncomingExchange, rabbitmqIncomingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare incoming exchange: %s", err)
	}

	err = rabbitmq.DeclareSimpleExchange(channel, rabbitmqOutgoingExchange, rabbitmqOutgoingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare outgoing exchange: %s", err)
	}

	cancelConsumer, err := rabbitmq.Consume(
		channel, rabbitmqIncomingExchange, rabbitmqQueue, rabbitmqBindingKey, rabbitmqConsumerTag,
		func(data []byte) {
			handleIncomingMessage(ctx, mongo, data, mongodbDatabase, mongodbCollection, channel, rabbitmqOutgoingExchange, rabbitmqRoutingKey)
		})
	if err != nil {
		log.Fatalf("Failed to consume: %s", err)
	}
	defer cancelConsumer()

	http.Handle("/health", handleHealth(mongo, rabbit))

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/persons", func(writer http.ResponseWriter, request *http.Request) {
		handlePersonsRequest(ctx, mongo, mongodbDatabase, mongodbCollection, writer)
	})

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", httpServerPort), nil)
		if err != nil {
			log.Fatalf("Failed to listen and serve: %s", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	log.Printf("Received signal to shutdown")
}
