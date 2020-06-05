#!/bin/bash

set -e

echo "gomic's end-to-end test:"
echo "  * publish a message to the source exchange"
echo "  * let it process by gomic"
echo "  * assert that the message is delivered to the destination queue in upper case"
echo "  * assert that the message is available via the HTTP endpoint"
echo "==="

RETRIES=10
until curl http://gomic:$HTTP_SERVER_PORT/health --silent --fail -o /dev/null; do
  echo "gomic is not ready yet..."
  ((RETRIES--))
  if [[ $RETRIES -eq 0 ]]; then
    echo "Test aborted"
    exit 1
  fi
  sleep 5
done

curl --silent --show-error --fail -O http://rabbitmq:15672/cli/rabbitmqadmin

function rabbitmq() {
  python3 rabbitmqadmin --host rabbitmq "$@"
}

RABBITMQ_DESTINATION_QUEUE=e2e-test-destination-queue
RABBITMQ_MESSAGE_TO_PUBLISH='{"firstName":"John","lastName":"Doe"}'

rabbitmq declare queue \
  name="$RABBITMQ_DESTINATION_QUEUE" durable=false

rabbitmq declare binding \
  source="$RABBITMQ_DESTINATION_EXCHANGE" destination_type=queue destination="$RABBITMQ_DESTINATION_QUEUE" routing_key="$RABBITMQ_DESTINATION_ROUTING_KEY"

rabbitmq publish \
  exchange="$RABBITMQ_SOURCE_EXCHANGE" routing_key="$RABBITMQ_SOURCE_ROUTING_KEY" payload="$RABBITMQ_MESSAGE_TO_PUBLISH"

RABBITMQ_MESSAGE=$(rabbitmq get \
  queue="$RABBITMQ_DESTINATION_QUEUE" ackmode=ack_requeue_false)

HTTP_RESPONSE=$(curl --silent --show-error --fail \
  http://gomic:$HTTP_SERVER_PORT/persons)

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
