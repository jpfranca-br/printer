Here's an improved and visually appealing version of your README without changing the content:

---

# **Printer Service**

This repository contains the implementation and management tools for a custom printer service. It is designed to handle message processing from an MQTT broker, send it to an ESC/POS TCP printer, and call a callback endpoint after printing.

---

## **Installation**

### **Install System Requirements**

Only **Git** is required for cloning this repository:

```bash
sudo apt update -y && sudo apt install git -y
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
