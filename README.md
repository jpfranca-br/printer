# Printer Service

This repository contains the implementation and management tools for a custom printer service. It is designed to handle message processing messages from a MQTT broker and sending it to a ESC/POS TCP printer, and then calling a callback endpoint after printing.

---

# **Installation**

## **Install System Requirements**
Only **Git** is required for cloning this repository.
    ```bash
    sudo apt update -y && sudo apt install git -y
    ```  
## **Installation and Setup**

1. Clone the repository, make scripts executable, run the initial setup script

    ```bash
    cd ~ && rm -rf printer && git clone https://github.com/jpfranca-br/printer.git && cd printer && chmod +x *.sh && ./setup.sh
    ```

2. edit `printer.config` to match your setup:

    ```bash
    nano printer.config
    ```

Config file should be something like:

    ```bash
    MQTT_HOST = io.adafruit.com
    MQTT_PORT = 8883
    MQTT_USER = <user>
    MQTT_PASS = <password like aio_****************************>
    TOPIC = <user>/feeds/<topic_name>
    TCP_HOST = localhost
    TCP_PORT = 9100
    ```

3. You can then use `manage.sh` for service management:

    ```bash
    ./manage.sh
    ```

---

# Features

- Message handling and encoding (including CP850).
- MQTT integration for receiving and processing messages.
- TCP communication for sending messages.
- Comprehensive management scripts for configuring and managing the service.

---

## 1. **Functionalities of `printer.go`**

`printer.go` is the core application of the printer service. It provides the following features:

1. **Message Processing**:
   - Handles messages received from MQTT topics.
   - Encodes messages to CP850.
   - Sends processed messages via TCP.

2. **MQTT Integration**:
   - Connects to an MQTT broker.
   - Subscribes to a configurable topic.
   - Decodes and processes MQTT messages.

3. **TCP Communication**:
   - Sends messages to a specified TCP server.
   - Handles connection retries and message timeouts.

4. **Callback Mechanism**:
   - Sends success or failure callbacks to specified URLs.

5. **Configuration Loading**:
   - Reads and loads configurations from the `printer.config` file.

## 2. **Functionalities of `manage.sh`**

`manage.sh` is an interactive script for managing the printer service.

### Menu Options:
1. **View Log (Real-Time)**: Displays the real-time service log.
2. **View Service Status**: Shows the status of the printer service.
3. **Enable Service**: Enables the service at startup.
4. **(Re)start Service**: Restarts the service.
5. **Stop Service**: Stops the running service.
6. **Disable Service**: Disables the service at startup.
7. **Exit**: Exits the script.

# MQTT Payload

## Example

    ```json
    {
      "id": "12345",
      "message": "Hello, world!",
      "callback": "http://example.com/callback-endpoint"
    }
    ```

## How It Works
1. Id is an id of your choice for your message. This will be logged to the log files and sent back to the callback url after printing.
2. Message is a plain text or base64-encoded message to be printed.
3. The `callback` field in the JSON payload is read when the message is processed.
4. After the message is sent via TCP:
   - A **success callback** is sent to the specified URL if the message is successfully delivered.
   - A **failure callback** is sent if the message times out or fails to be delivered.
5. The data sent to the `callback` URL is also in JSON format, with the following structure:
   - **`success`**: A boolean value (`"true"` or `"false"`).
   - **`id`**: The unique identifier of the processed message.

### Example Callback Payload
**Success Example:**

    ```json
    {
      "success": "true",
      "id": "12345"
    }
    ```

**Failure Example:**

    ```json
    {
      "success": "false",
      "id": "12345"
    }
    ```

If no `callback` is provided, the service will process the message but will not send any callback.
