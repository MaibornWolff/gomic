package rabbitmq

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
	"log"
)

var (
	incomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gomic_incoming_rabbitmq_messages_total",
		Help: "The total number of incoming RabbitMQ messages",
	})
)

func Consume(channel *amqp.Channel, exchange string, queueName string, key string, consumerTag string, handler func([]byte)) (func() error, error) {
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

	log.Printf("Queue bound to exchange, starting to consume from queue (consumer tag %q)", consumerTag)

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
		log.Printf("Cancelling consumer")

		err := channel.Cancel(consumerTag, true)
		if err != nil {
			return fmt.Errorf("Failed to cancel consumer: %s", err)
		}

		return <-handlerIsDone
	}

	return cancelConsumer, nil
}

func handle(deliveries <-chan amqp.Delivery, handler func([]byte), handlerIsDone chan error) {
	for d := range deliveries {
		log.Printf("Got %dB delivery: [%v] %q", len(d.Body), d.DeliveryTag, d.Body)

		handler(d.Body)
		d.Ack(false)

		incomingMessages.Inc()
	}

	log.Printf("Deliveries channel closed")
	handlerIsDone <- nil
}
