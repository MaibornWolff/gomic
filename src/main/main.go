package main

import (
	"context"
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
	mongodbDatabase   = getEnv("MONGODB_DATABASE", "db")
	mongodbCollection = getEnv("MONGODB_COLLECTION", "foo")

	rabbitmqHost                 = getEnv("RABBITMQ_HOST", "")
	rabbitmqIncomingExchange     = getEnv("RABBITMQ_INCOMING_EXCHANGE", "")
	rabbitmqIncomingExchangeType = getEnv("RABBITMQ_INCOMING_EXCHANGE_TYPE", "direct")
	rabbitmqQueue                = getEnv("RABBITMQ_QUEUE", "")
	rabbitmqBindingKey           = getEnv("RABBITMQ_BINDING_KEY", "")
	rabbitmqConsumerTag          = getEnv("RABBITMQ_CONSUMER_TAG", "simple-consumer")
	rabbitmqOutgoingExchange     = getEnv("RABBITMQ_OUTGOING_EXCHANGE", "")
	rabbitmqOutgoingExchangeType = getEnv("RABBITMQ_OUTGOING_EXCHANGE_TYPE", "direct")
	rabbitmqRoutingKey           = getEnv("RABBITMQ_ROUTING_KEY", "")
)

func main() {
	mongo := mongodb.Connect(mongodbHost)
	defer mongo.Disconnect(context.Background())

	connection, channel, err := rabbitmq.Connect(rabbitmqHost)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer connection.Close()

	err = rabbitmq.DeclareSimpleExchange(channel, rabbitmqIncomingExchange, rabbitmqIncomingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare incoming exchange: %s", err)
	}

	err = rabbitmq.DeclareSimpleExchange(channel, rabbitmqOutgoingExchange, rabbitmqOutgoingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare outgoing exchange: %s", err)
	}

	cancel, err := rabbitmq.Consume(
		channel, rabbitmqIncomingExchange, rabbitmqQueue, rabbitmqBindingKey, rabbitmqConsumerTag,
		func(data []byte) {
			handleIncomingMessage(data, mongo, mongodbDatabase, mongodbCollection, channel, rabbitmqOutgoingExchange, rabbitmqRoutingKey)
		})
	if err != nil {
		log.Fatalf("Failed to consume: %s", err)
	}
	defer cancel()

	http.HandleFunc("/data", func(writer http.ResponseWriter, request *http.Request) {
		handleFindData(mongo, mongodbDatabase, mongodbCollection, writer)
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("Failed to listen and serve: %s", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	log.Printf("Received signal to shutdown")
}
