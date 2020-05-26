package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vrischmann/envconfig"
	"log"
	"maibornwolff.de/gomic/application"
	"maibornwolff.de/gomic/mongodb"
	"maibornwolff.de/gomic/rabbitmq"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := envconfig.Init(&config)
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	ctx := context.Background()

	mongoClient, err := mongodb.Connect(ctx, config.Mongodb.Host)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	defer mongoClient.Disconnect(ctx)

	rabbitConnection, rabbitConnectionIsClosed, rabbitChannel, err := rabbitmq.Connect(config.Rabbitmq.Host, true)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rabbitConnection.Close()

	err = rabbitmq.DeclareSimpleExchange(rabbitChannel, config.Rabbitmq.IncomingExchange, config.Rabbitmq.IncomingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare incoming exchange: %s", err)
	}

	err = rabbitmq.DeclareSimpleExchange(rabbitChannel, config.Rabbitmq.OutgoingExchange, config.Rabbitmq.OutgoingExchangeType)
	if err != nil {
		log.Fatalf("Failed to declare outgoing exchange: %s", err)
	}

	cancelRabbitConsumer, err := rabbitmq.Consume(
		rabbitChannel, config.Rabbitmq.IncomingExchange, config.Rabbitmq.Queue, config.Rabbitmq.BindingKey, config.Rabbitmq.ConsumerTag,
		func(data []byte) {
			application.HandleIncomingMessage(ctx, data, mongoClient, config.Mongodb.Database, config.Mongodb.Collection, rabbitChannel, config.Rabbitmq.OutgoingExchange, config.Rabbitmq.RoutingKey)
		})
	if err != nil {
		log.Fatalf("Failed to consume: %s", err)
	}
	defer cancelRabbitConsumer()

	router := gin.Default()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", gin.WrapH(handleHealthRequest(mongoClient, rabbitConnectionIsClosed)))

	router.GET("/persons", func(ctx *gin.Context) {
		application.HandlePersonsRequest(ctx, mongoClient, config.Mongodb.Database, config.Mongodb.Collection)
	})

	go func() {
		err := router.Run(fmt.Sprintf(":%d", config.HTTPServer.Port))
		if err != nil {
			log.Fatalf("Failed to listen and serve: %s", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	log.Printf("Received signal to shutdown")
}
