package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

func Publish(channel *amqp.Channel, exchange string, routingKey string, data []byte, contentType string) error {
	err := channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     contentType,
			ContentEncoding: "",
			Body:            data,
			DeliveryMode:    amqp.Transient,
			Priority:        0,
		},
	)
	if err != nil {
		return fmt.Errorf("Failed to publish: %w", err)
	}

	return nil
}

func enablePublisherConfirms(channel *amqp.Channel, confirmHandler func(amqp.Confirmation)) error {
	log.Info().Msg("Enabling publisher confirms")

	err := channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("Failed to enable publisher confirms: %w", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	go processConfirms(confirms, confirmHandler)

	return nil
}

func processConfirms(confirms <-chan amqp.Confirmation, confirmHandler func(amqp.Confirmation)) {
	log.Info().Msg("Ready to process publisher confirms")

	for confirm := range confirms {
		confirmHandler(confirm)
	}
}

func SimplePublisherConfirmHandler(confirm amqp.Confirmation) {
	if confirm.Ack {
		log.Info().Uint64("deliveryTag", confirm.DeliveryTag).Msg("Confirmed delivery")
	} else {
		log.Error().Uint64("deliveryTag", confirm.DeliveryTag).Msg("Failed to deliver")
	}
}
