package main

import (
	"context"
	"log"
	"maibornwolff.de/gomic/mongodb"
	"maibornwolff.de/gomic/rabbitmq"
	"net/http"
)

var (
	mongodbHost       = getEnv("MONGODB_HOST", "")
	mongodbDatabase   = getEnv("MONGODB_DATABASE", "db")
	mongodbCollection = getEnv("MONGODB_COLLECTION", "foo")

	rabbitmqHost         = getEnv("RABBITMQ_HOST", "")
	rabbitmqExchange     = getEnv("RABBITMQ_EXCHANGE", "")
	rabbitmqExchangeType = getEnv("RABBITMQ_EXCHANGE_TYPE", "direct")
	rabbitmqQueue        = getEnv("RABBITMQ_QUEUE", "")
	rabbitmqBindingKey   = getEnv("RABBITMQ_BINDING_KEY", "")
	rabbitmqConsumerTag  = getEnv("RABBITMQ_CONSUMER_TAG", "simple-consumer")
)

func main() {
	mongo := mongodb.Connect(mongodbHost)
	defer mongo.Disconnect(context.Background())

	consumer, err := rabbitmq.NewConsumer(
		rabbitmqHost, rabbitmqExchange, rabbitmqExchangeType, rabbitmqQueue, rabbitmqBindingKey, rabbitmqConsumerTag,
		func(data []byte) {
			handleIncomingMessage(data, mongo, mongodbDatabase, mongodbCollection)
		})
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %s", err)
	}
	defer consumer.Shutdown()

	http.HandleFunc("/data", func(writer http.ResponseWriter, request *http.Request) {
		handleFindData(mongo, mongodbDatabase, mongodbCollection, writer)
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to listen and serve: %s", err)
	}
}
