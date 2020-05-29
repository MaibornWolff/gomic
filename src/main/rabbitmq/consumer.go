package rabbitmq

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

var (
	incomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomic_incoming_rabbitmq_messages_total",
		Help: "The total number of incoming RabbitMQ messages",
	})
)

func Consume(channel *amqp.Channel, exchange string, queueName string, bindingKey string, consumerTag string, handler func([]byte)) (func() error, error) {
	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare queue: %s", err)
	}

	log.Info().
		Str("queue", queue.Name).
		Int("messageCount", queue.Messages).
		Int("consumerCount", queue.Consumers).
		Str("exchange", exchange).
		Str("bindingKey", bindingKey).
		Msg("Declared queue, binding it to exchange")

	err = channel.QueueBind(
		queue.Name,
		bindingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind queue: %s", err)
	}

	log.Info().Str("consumerTag", consumerTag).Msg("Queue bound to exchange, starting to consume from queue")

	deliveries, err := channel.Consume(
		queue.Name,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to start to consume from queue: %s", err)
	}

	handlerIsDone := make(chan error)
	go handle(deliveries, handler, handlerIsDone)

	cancelConsumer := func() error {
		log.Info().Msg("Cancelling consumer")

		err := channel.Cancel(consumerTag, true)
		if err != nil {
			return fmt.Errorf("Failed to cancel consumer: %s", err)
		}

		return <-handlerIsDone
	}

	return cancelConsumer, nil
}

func handle(deliveries <-chan amqp.Delivery, handler func([]byte), handlerIsDone chan error) {
	for delivery := range deliveries {
		log.Info().
			Uint64("deliveryTag", delivery.DeliveryTag).
			Bytes("deliveryBody", delivery.Body).
			Msgf("Got %dB delivery", len(delivery.Body))

		handler(delivery.Body)
		delivery.Ack(false)

		incomingMessages.Inc()
	}

	log.Info().Msg("Deliveries channel closed")
	handlerIsDone <- nil
}
