package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func Connect(amqpURI string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to dial: %s", err)
	}

	go func() {
		fmt.Printf("Closing: %s", <-conn.NotifyClose(make(chan *amqp.Error)))
	}()

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open channel: %s", err)
	}

	log.Printf("Connected to RabbitMQ")

	return conn, channel, nil
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
