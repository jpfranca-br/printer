#!/bin/bash

# Generate a unique ID based on the current epoch time
message_id=$(date +%s)

# Publish a message to the MQTT topic with the generated ID
mosquitto_pub -h broker.hivemq.com -p 1883 -t printerserver -m '{
  "id": "'"$message_id"'",
  "message": "Hello, world! Test message ID '"$message_id"'.",
  "callback": "https://apimocha.com/printerserver/callback",
  "cut": "true"
}'

# Echo the message ID
echo "Message sent with ID: $message_id"
