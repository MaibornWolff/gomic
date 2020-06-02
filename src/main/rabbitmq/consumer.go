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

func Consume(channel *amqp.Channel, exchange string, queueName string, bindingKey string, consumerTag string, deliveryHandler func(amqp.Delivery) error) (func() error, error) {
	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare queue: %w", err)
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
		return nil, fmt.Errorf("Failed to bind queue: %w", err)
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
		return nil, fmt.Errorf("Failed to start to consume from queue: %w", err)
	}

	processingIsDone := make(chan error)
	go processDeliveries(deliveries, deliveryHandler, processingIsDone)

	cancelConsumer := func() error {
		log.Info().Msg("Cancelling consumer")

		err := channel.Cancel(consumerTag, true)
		if err != nil {
			return fmt.Errorf("Failed to cancel consumer: %w", err)
		}

		return <-processingIsDone
	}

	return cancelConsumer, nil
}

func processDeliveries(deliveries <-chan amqp.Delivery, deliveryHandler func(amqp.Delivery) error, processingIsDone chan error) {
	for delivery := range deliveries {
		log.Info().
			Uint64("deliveryTag", delivery.DeliveryTag).
			Bytes("deliveryBody", delivery.Body).
			Msgf("Got %dB delivery", len(delivery.Body))

		incomingMessages.Inc()

		err := deliveryHandler(delivery)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process delivery")
		} else {
			err = delivery.Ack(false)
			if err != nil {
				log.Error().Err(err).Msg("Failed to acknowledge delivery")
			}
		}
	}

	log.Info().Msg("Deliveries channel closed")
	processingIsDone <- nil
}
