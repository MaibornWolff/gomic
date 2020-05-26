# gomic

![Gopher](resources/gomic_gopher_small.png)

## Aim
gomic is a minimal microservice skeleton in Golang, including support for
* MongoDB
* RabbitMQ
* Prometheus
* health checks
* custom HTTP endpoints

## Behavior
The sample business logic consumes a JSON document from a RabbitMQ queue, stores it in a MongoDB database and finally forwards it in upper case to a RabbitMQ exchange.

## Usage
`make build start`