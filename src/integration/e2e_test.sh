#!/bin/bash
set -e

curl --silent --show-error --fail -O http://localhost:15672/cli/rabbitmqadmin

function rabbitmq() {
  python3 rabbitmqadmin "$@"
}

# Set up by gomic
SOURCE_EXCHANGE=source-exchange
SOURCE_ROUTING_KEY=source-routing-key

# Set up by this script
DESTINATION_EXCHANGE=destination-exchange
DESTINATION_QUEUE=e2e-test-destination-queue
DESTINATION_ROUTING_KEY=destination-routing-key

RABBITMQ_MESSAGE_TO_PUBLISH='{"firstName":"John","lastName":"Doe"}'

echo "gomic's end-to-end test:"
echo "  * publish a message to the source exchange"
echo "  * let it process by gomic"
echo "  * assert that the message is delivered to the destination queue in upper case"
echo "  * assert that the message is available via the HTTP endpoint"
echo "==="

rabbitmq declare queue \
  name=$DESTINATION_QUEUE durable=false

rabbitmq declare binding \
  source=$DESTINATION_EXCHANGE destination_type=queue destination=$DESTINATION_QUEUE routing_key=$DESTINATION_ROUTING_KEY

rabbitmq publish \
  exchange=$SOURCE_EXCHANGE routing_key=$SOURCE_ROUTING_KEY payload="$RABBITMQ_MESSAGE_TO_PUBLISH"

RABBITMQ_MESSAGE=$(rabbitmq get \
  queue=$DESTINATION_QUEUE ackmode=ack_requeue_false)

HTTP_RESPONSE=$(curl --silent --show-error --fail \
  http://localhost:8080/persons)

echo "==="

EXPECTED_RABBITMQ_MESSAGE='{"firstName":"JOHN","lastName":"DOE"}'
EXPECTED_HTTP_RESPONSE='{"firstName":"John","lastName":"Doe"}'

if [[ "$RABBITMQ_MESSAGE" == *$EXPECTED_RABBITMQ_MESSAGE* ]]; then
  echo "Received expected RabbitMQ message"
else
  echo "Did not receive expected RabbitMQ message"
  echo "Test failed"
  exit 1
fi

if [[ "$HTTP_RESPONSE" == *$EXPECTED_HTTP_RESPONSE* ]]; then
  echo "Retrieved expected HTTP response"
else
  echo "Did not retrieve expected HTTP response"
  echo "Test failed"
  exit 1
fi

echo "Test successfully completed"
