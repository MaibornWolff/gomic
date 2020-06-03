package main

var config struct {
	Mongodb struct {
		Host       string
		Database   string
		Collection string
	}
	Rabbitmq struct {
		Host   string
		Source struct {
			Exchange     string
			ExchangeType string `envconfig:"default=direct"`
			Queue        string
			RoutingKey   string
		}
		ConsumerTag string
		Destination struct {
			Exchange     string
			ExchangeType string `envconfig:"default=direct"`
			RoutingKey   string
		}
	}
	HTTPServer struct {
		Port int `envconfig:"default=8080"`
	}
	LogLevel string `envconfig:"default=info"`
}
