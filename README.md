Gomic is a minimal microservice skeleton in Golang, including support for
* MongoDB
* RabbitMQ
* Prometheus
* health check
* custom HTTP endpoints

The sample business logic consumes a JSON document from a RabbitMQ queue, stores it in a MongoDB database and finally forwards it in upper case to a RabbitMQ exchange.
