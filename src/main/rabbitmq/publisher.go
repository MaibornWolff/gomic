package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
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
		return fmt.Errorf("Failed to publish: %s", err)
	}

	return nil
}

func putIntoConfirmMode(channel *amqp.Channel) error {
	log.Printf("Putting channel into confirm mode")

	err := channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("Failed to put channel into confirm mode: %s", err)
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	go handleConfirmations(confirms)

	return nil
}

func handleConfirmations(confirms <-chan amqp.Confirmation) {
	log.Printf("Waiting for confirms")

	for confirm := range confirms {
		if confirm.Ack {
			log.Printf("Confirmed delivery (delivery tag %d)", confirm.DeliveryTag)
		} else {
			log.Printf("Failed to deliver (delivery tag %d)", confirm.DeliveryTag)
		}
	}
}
