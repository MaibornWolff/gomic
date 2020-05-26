package main

var config struct {
	Mongodb struct {
		Host       string
		Database   string
		Collection string
	}
	Rabbitmq struct {
		Host                 string
		IncomingExchange     string
		IncomingExchangeType string `envconfig:"default=direct"`
		Queue                string
		BindingKey           string
		ConsumerTag          string
		OutgoingExchange     string
		OutgoingExchangeType string `envconfig:"default=direct"`
		RoutingKey           string
	}
	HTTPServer struct {
		Port int `envconfig:"default=8080"`
	}
}
