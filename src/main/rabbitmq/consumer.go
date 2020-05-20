package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func Consume(channel *amqp.Channel, exchange string, queueName string, key string, ctag string, handler func([]byte)) (func() error, error) {
	log.Printf("Declaring queue %q", queueName)
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

	log.Printf("Declared queue (%q, %d messages, %d consumers), binding to exchange (%q, key %q)",
		queue.Name, queue.Messages, queue.Consumers, exchange, key)

	err = channel.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind queue: %s", err)
	}

	log.Printf("Queue bound to exchange, starting consume (consumer tag %q)", ctag)
	deliveries, err := channel.Consume(
		queue.Name,
		ctag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to start consuming from queue: %s", err)
	}

	handlerIsDone := make(chan error)
	go handle(deliveries, handler, handlerIsDone)

	return func() error {
		log.Printf("Cancelling consumer")
		err := channel.Cancel(ctag, true)
		if err != nil {
			return fmt.Errorf("Failed to cancel consumer: %s", err)
		}
		return <-handlerIsDone
	}, nil
}

func handle(deliveries <-chan amqp.Delivery, handler func([]byte), handlerIsDone chan error) {
	for d := range deliveries {
		log.Printf(
			"Got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		handler(d.Body)
		d.Ack(false)
	}
	log.Printf("Deliveries channel closed")
	handlerIsDone <- nil
}
