package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/vrischmann/envconfig"
	"maibornwolff.de/gomic/application"
	"maibornwolff.de/gomic/mongodb"
	"maibornwolff.de/gomic/rabbitmq"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	err := envconfig.Init(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
	}

	logLevel, err := zerolog.ParseLevel(strings.ToLower(config.LogLevel))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	ctx := context.Background()

	mongoClient, err := mongodb.Connect(ctx, config.Mongodb.Host)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(ctx)

	rabbitConnection, rabbitConnectionIsClosed, rabbitChannel, err := rabbitmq.Connect(config.Rabbitmq.Host, rabbitmq.SimplePublisherConfirmHandler)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbitConnection.Close()

	err = rabbitmq.DeclareSimpleExchange(rabbitChannel, config.Rabbitmq.IncomingExchange, config.Rabbitmq.IncomingExchangeType)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to declare incoming exchange")
	}

	err = rabbitmq.DeclareSimpleExchange(rabbitChannel, config.Rabbitmq.OutgoingExchange, config.Rabbitmq.OutgoingExchangeType)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to declare outgoing exchange")
	}

	cancelRabbitConsumer, err := rabbitmq.Consume(
		rabbitChannel, config.Rabbitmq.IncomingExchange, config.Rabbitmq.Queue, config.Rabbitmq.BindingKey, config.Rabbitmq.ConsumerTag,
		func(delivery amqp.Delivery) {
			application.HandleIncomingMessage(ctx, delivery.Body, mongoClient, config.Mongodb.Database, config.Mongodb.Collection, rabbitChannel, config.Rabbitmq.OutgoingExchange, config.Rabbitmq.RoutingKey)
		})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to consume")
	}
	defer cancelRabbitConsumer()

	router := createHTTPRouter(logLevel)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", gin.WrapH(handleHealthRequest(mongoClient, rabbitConnectionIsClosed)))

	router.GET("/persons", func(ctx *gin.Context) {
		application.HandlePersonsRequest(ctx, mongoClient, config.Mongodb.Database, config.Mongodb.Collection)
	})

	go func() {
		err := router.Run(fmt.Sprintf(":%d", config.HTTPServer.Port))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to listen and serve")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	log.Info().Msg("Received signal to shutdown")
}
