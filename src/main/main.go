package main

import (
	"github.com/micro/go-micro/v2/web"
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
	defer mongo.Shutdown()

	consumer, err := rabbitmq.NewConsumer(
		rabbitmqHost, rabbitmqExchange, rabbitmqExchangeType, rabbitmqQueue, rabbitmqBindingKey, rabbitmqConsumerTag,
		func(data []byte) {
			handleAmqpMessage(data, mongo, mongodbDatabase, mongodbCollection)
		})
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %s", err)
	}
	defer consumer.Shutdown()

	service := web.NewService(
		web.Name("foobar"),
		web.Address(":8080"),
	)

	service.HandleFunc("/data", func(writer http.ResponseWriter, request *http.Request) {
		handleFindData(mongo, mongodbDatabase, mongodbCollection, writer)
	})

	err = service.Init()
	if err != nil {
		log.Fatalf("Failed to init service: %s", err)
	}

	err = service.Run()
	if err != nil {
		log.Fatalf("Failed to run service: %s", err)
	}
}
