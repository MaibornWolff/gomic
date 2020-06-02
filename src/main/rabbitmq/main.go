package rabbitmq

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type Client struct {
	Connection         *amqp.Connection
	ConnectionIsClosed chan *amqp.Error
	Channel            *amqp.Channel
	ChannelIsClosed    chan *amqp.Error
}

func Connect(amqpURI string, publisherConfirmHandler func(amqp.Confirmation)) (*Client, error) {
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	connectionIsClosed := connection.NotifyClose(make(chan *amqp.Error, 1))

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open channel: %s", err)
	}

	channelIsClosed := channel.NotifyClose(make(chan *amqp.Error, 1))

	if publisherConfirmHandler != nil {
		err = enablePublisherConfirms(channel, publisherConfirmHandler)
		if err != nil {
			return nil, fmt.Errorf("Failed to enable publisher confirms: %s", err)
		}
	}

	log.Info().Msg("Connected to RabbitMQ")

	go func() {
		log.Warn().
			Err(<-connection.NotifyClose(make(chan *amqp.Error))).
			Msg("Connection to RabbitMQ is closed")
	}()

	go func() {
		log.Warn().
			Err(<-channel.NotifyClose(make(chan *amqp.Error))).
			Msg("Channel to RabbitMQ is closed")
	}()

	return &Client{
		connection,
		connectionIsClosed,
		channel,
		channelIsClosed,
	}, nil
}

func (client *Client) Close() {
	client.Channel.Close()
	client.Connection.Close()
}

func (client *Client) DeclareSimpleExchange(exchange string, exchangeType string) error {
	err := client.Channel.ExchangeDeclare(
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
