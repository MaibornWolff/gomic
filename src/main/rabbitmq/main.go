package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

func Connect(amqpURI string, enablePublishingConfirms bool) (*amqp.Connection, chan *amqp.Error, *amqp.Channel, error) {
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to dial: %s", err)
	}

	connectionIsClosed := make(chan *amqp.Error)
	connection.NotifyClose(connectionIsClosed)

	channel, err := connection.Channel()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to open channel: %s", err)
	}

	if enablePublishingConfirms {
		err = putIntoConfirmMode(channel)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to enable publishing confirms: %s", err)
		}
	}

	log.Info().Msg("Connected to RabbitMQ")

	return connection, connectionIsClosed, channel, nil
}

func DeclareSimpleExchange(channel *amqp.Channel, exchange string, exchangeType string) error {
	err := channel.ExchangeDeclare(
		exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to declare exchange: %s", err)
	}

	return nil
}
