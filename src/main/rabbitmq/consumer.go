package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string, handler func([]byte)) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	log.Printf("Dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	go func() {
		fmt.Printf("Closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("Got connection, getting channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to get channel: %s", err)
	}

	log.Printf("Got Channel, declaring exchange (%q)", exchange)
	err = c.channel.ExchangeDeclare(
		exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare exchange: %s", err)
	}

	log.Printf("Declared Exchange, declaring queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
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

	log.Printf("Declared queue (%q %d messages, %d consumers), binding to exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	err = c.channel.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind queue: %s", err)
	}

	log.Printf("Queue bound to exchange, starting consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name,
		c.tag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to start consuming from queue: %s", err)
	}

	go handle(deliveries, handler, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() {
	err := c.channel.Cancel(c.tag, true)
	if err != nil {
		log.Printf("Failed to cancel deliveries: %s", err)
	}

	err = c.conn.Close()
	if err != nil {
		log.Printf("Failed to close AMQP connection: %s", err)
	}

	defer log.Printf("Shutdown completed")

	// Wait for handle() to exit
	<-c.done
}

func handle(deliveries <-chan amqp.Delivery, handler func([]byte), done chan error) {
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
	done <- nil
}
