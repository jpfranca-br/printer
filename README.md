# **Printer Service**

This repository contains the implementation and management tools for a custom printer service. It is designed to handle message processing from an MQTT broker, send it to an ESC/POS TCP printer, and call a callback endpoint after printing.

---

## **Installation**

### **Install System Requirements**

Only **Git** is required for cloning this repository:

```bash
sudo apt update -y && sudo apt install git -y
```

If you also want to test (publish to MQTT), install Mosquitto

```bash
sudo apt update -y && sudo apt install mosquitto-clients -y
```

---

### **Installation and Setup**

1. Clone the repository, make scripts executable, and run the initial setup script:

    ```bash
    cd ~ && rm -rf printer && git clone https://github.com/jpfranca-br/printer.git && cd printer && chmod +x *.sh && ./setup.sh
    ```

2. Edit the `printer.config` file to match your setup:

    ```bash
    nano printer.config
    ```

   Example configuration:

    ```bash
    MQTT_HOST = io.adafruit.com
    MQTT_PORT = 8883
    MQTT_USER = <user>
    MQTT_PASS = <password like aio_****************************>
    TOPIC = <user>/feeds/<topic_name>
    TCP_HOST = localhost
    TCP_PORT = 9100
    ```

3. Use `manage.sh` for service management:

    ```bash
    ./manage.sh
    ```

---

## **Features**

- **Message Handling**: Processes and encodes messages (e.g., CP850 encoding).
- **MQTT Integration**: Receives and processes messages via MQTT.
- **TCP Communication**: Sends processed messages to a TCP server.
- **Management Scripts**: Includes comprehensive tools for service configuration and management.

---

## **Core Components**

### 1. **Functionalities of `printer.go`**

The `printer.go` application is the core of the printer service, providing the following:

- **Message Processing**:
  - Handles messages from MQTT topics.
  - Encodes messages in CP850.
  - Sends messages via TCP.

- **MQTT Integration**:
  - Connects to an MQTT broker.
  - Subscribes to configurable topics.
  - Decodes and processes MQTT messages.

- **TCP Communication**:
  - Sends messages to a specified TCP server.
  - Manages connection retries and timeouts.

- **Callback Mechanism**:
  - Sends success or failure callbacks to specified URLs.

- **Configuration Loading**:
  - Reads configurations from the `printer.config` file.

---

### 2. **Functionalities of `manage.sh`**

The `manage.sh` script provides an interactive interface for managing the printer service.

#### **Menu Options**:

1. View Log (Real-Time): Displays real-time service logs.
2. View Service Status: Shows the current service status.
3. Enable Service: Enables the service to start at boot.
4. (Re)start Service: Restarts the printer service.
5. Stop Service: Stops the running service.
6. Disable Service: Disables the service from starting at boot.
7. Exit: Closes the management script.

---

## **MQTT Payload**

### **Example Payload**:

```json
{
  "id": "12345",
  "message": "Hello, world!",
  "callback": "http://example.com/callback-endpoint"
}
```

### **How It Works**:

1. **`id`**: A unique identifier for your message. It will appear in logs and callback responses.
2. **`message`**: The plain text or base64-encoded message to print.
3. **`callback`**: A URL where the service sends the result of the message processing.

   - **Success**: Sent if the message is delivered successfully.
   - **Failure**: Sent if the message times out or fails.
   - If no `callback` is provided, the service processes the message without sending a response.

---

### **Callback Payload Example**

#### **Success Example**:

```json
{
  "success": "true",
  "id": "12345"
}
```

#### **Failure Example**:

```json
{
  "success": "false",
  "id": "12345"
}
```

## Testing

Once the service is running (with the default printer.config file provided) you can test it by simply sending a message to the MQTT topic:

```bash
mosquitto_pub -h broker.hivemq.com -p 1883 -t printerserver -m '{
  "id": "test_id",
  "message": "Hello, world! This is a test message.",
  "callback": "https://apimocha.com/printerserver/callback"
}'
```

This should print a message.

If you are in doubt that the message is correctly sent to the MQTT topic, open another terminal window and subscribe to it. Every time you send a topic with mosquitto_pub, a new message should be shown in mosquitto_sub:

```bash
mosquitto_sub -h broker.hivemq.com -p 1883 -t printerserver
```

The result of printing (success or failure) can be checked on the callback endpoint:

[https://apimocha.com/printerserver](https://apimocha.com/printerserver)
