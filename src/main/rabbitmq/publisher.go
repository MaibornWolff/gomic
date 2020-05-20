package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func Publish(channel *amqp.Channel, exchange string, routingKey string, data []byte, reliable bool) error {
	if reliable {
		log.Printf("Enabling publishing confirms")

		err := channel.Confirm(false)
		if err != nil {
			return fmt.Errorf("Failed to put channel into confirmation mode: %s", err)
		}

		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
		go handleConfirmations(confirms)
	}

	err := channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
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

func handleConfirmations(confirms <-chan amqp.Confirmation) {
	log.Printf("Waiting for confirms")

	for confirm := range confirms {
		if confirm.Ack {
			log.Printf("Confirmed delivery with delivery tag: %d", confirm.DeliveryTag)
		} else {
			log.Printf("Failed to deliver with delivery tag: %d", confirm.DeliveryTag)
		}
	}
}
