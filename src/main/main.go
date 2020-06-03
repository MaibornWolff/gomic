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
	"sync"
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

	rabbitClient, err := rabbitmq.Connect(config.Rabbitmq.Host, rabbitmq.SimplePublisherConfirmHandler)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbitClient.Close()

	err = rabbitClient.DeclareSimpleExchange(config.Rabbitmq.IncomingExchange, config.Rabbitmq.IncomingExchangeType)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to declare incoming exchange")
	}

	err = rabbitClient.DeclareSimpleExchange(config.Rabbitmq.OutgoingExchange, config.Rabbitmq.OutgoingExchangeType)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to declare outgoing exchange")
	}

	var rabbitHandlerWaitGroup sync.WaitGroup

	cancelRabbitConsumer, err := rabbitmq.Consume(
		rabbitClient.Channel, config.Rabbitmq.IncomingExchange, config.Rabbitmq.Queue, config.Rabbitmq.BindingKey, config.Rabbitmq.ConsumerTag,
		func(delivery amqp.Delivery) error {
			rabbitHandlerWaitGroup.Add(1)
			return application.HandleIncomingMessage(ctx, &rabbitHandlerWaitGroup, delivery.Body, mongoClient, config.Mongodb.Database, config.Mongodb.Collection, rabbitClient.Channel, config.Rabbitmq.OutgoingExchange, config.Rabbitmq.RoutingKey)
		})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to consume")
	}
	defer cancelRabbitConsumer()

	router := createHTTPRouter(logLevel)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", gin.WrapH(handleHealthRequest(mongoClient, rabbitClient)))

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
	log.Info().Msg("Waiting for handlers to finish")
	rabbitHandlerWaitGroup.Wait()
}
