package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func Connect(amqpURI string, enablePublishingConfirms bool) (*amqp.Connection, *amqp.Channel, chan *amqp.Error, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to dial: %s", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to open channel: %s", err)
	}

	log.Printf("Connected to RabbitMQ")

	if enablePublishingConfirms {
		err = putIntoConfirmMode(channel)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to enable publishing confirms: %s", err)
		}
	}

	errorChannel := make(chan *amqp.Error)
	conn.NotifyClose(errorChannel)

	return conn, channel, errorChannel, nil
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
