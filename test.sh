#!/bin/bash
# Publish a message to the MQTT topic with a dynamically generated ID and message
mosquitto_pub -h broker.hivemq.com -p 1883 -t printerserver -m '{
  "id": "'$(date +%s)'",
  "message": "Hello, world! Test message ID '$(date +%s)'.",
  "callback": "https://apimocha.com/printerserver/callback"
}'
