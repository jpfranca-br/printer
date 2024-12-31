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

### 1. Clone the repository, make scripts executable, and run the initial setup script:

    ```bash
    cd ~ && rm -rf printer && git clone https://github.com/jpfranca-br/printer.git && cd printer && chmod +x *.sh && ./setup.sh
    ```

### 2. Configure the `printer.config` File

To customize the service for your setup, you need to edit the `printer.config` file. Use the following command to open it in a text editor:

```bash
nano printer.config
```

#### Example Configuration

Below is an example of a typical configuration file:

```bash
MQTT_HOST = broker.hivemq.com
MQTT_PORT = 8883
MQTT_USER = user
MQTT_PASS = password
TOPIC = printerserver
TCP_HOST = 127.0.0.1
TCP_PORT = 9100
```

**Explanation of Fields:**

- **`MQTT_HOST`**: The address of your MQTT broker (e.g., `broker.hivemq.com`).
- **`MQTT_PORT`**: The port used by the MQTT broker (default: `8883` for secure connections).
- **`MQTT_USER`**: Your MQTT broker username (leave empty if not required).
- **`MQTT_PASS`**: Your MQTT broker password (leave empty if not required).
- **`TOPIC`**: The MQTT topic to which the service listens (e.g., `printerserver`).
- **`TCP_HOST`**: The IP address of your printer (usually your printer's network address or `127.0.0.1` for local).
- **`TCP_PORT`**: The port used to communicate with the printer (`9100` is the default for most network printers).

After making changes, save and close the file (`Ctrl + O`, `Enter`, then `Ctrl + X` in Nano).

### 3. Use `manage.sh` for service management:

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

Here’s an improved version with clearer instructions, better formatting, and some added context for ease of understanding:

---

## Testing the Service

After starting the service (using the default `printer.config` file provided), you can test its functionality by publishing a message to the MQTT topic using the following command:

```bash
mosquitto_pub -h broker.hivemq.com -p 1883 -t printerserver -m '{
  "id": "test_id",
  "message": "Hello, world! This is a test message.",
  "callback": "https://apimocha.com/printerserver/callback"
}'
```

### What Happens Next?

- The service should process the message and print it.
- The result of the print operation (success or failure) will be sent to the specified callback URL.

You can check the callback response here:
[https://apimocha.com/printerserver](https://apimocha.com/printerserver)

### Verifying Message Delivery

If you are unsure whether the message was sent correctly to the MQTT topic, you can monitor the topic by subscribing to it in a separate terminal. Use the following command:

```bash
mosquitto_sub -h broker.hivemq.com -p 1883 -t printerserver
```

Each time a message is published to the `printerserver` topic using `mosquitto_pub`, it should appear in the `mosquitto_sub` output.
