# **Printer Service**

This repository contains the implementation and management tools for a custom printer service. It is designed to handle message processing from an MQTT broker, send it to an ESC/POS TCP or USB printer, and call a callback endpoint after printing.

## **Installation**

### **Step 1: Install System Requirements**

The only mandatory requirement for cloning this repository is **Git**. To install Git, run the following command:

```bash
sudo apt update -y && sudo apt install git -y
```

### **Step 2: (Optional) Install Mosquitto for Testing**

If you plan to test by publishing messages to an MQTT topic, you’ll need the Mosquitto client tools. Install them with the following command:

```bash
sudo apt update -y && sudo apt install mosquitto-clients -y
```

### **Step 3: Clone the repository, make scripts executable, and run the initial setup script**

```bash
cd ~ && rm -rf printer && git clone https://github.com/jpfranca-br/printer.git && cd printer && chmod +x *.sh && ./setup.sh
```

### **Step 4: Configure the `printer.config` File**

To customize the service for your setup, you need to edit the `printer.config` file. 

**The provided `printer.config` is a working version using free services, which can be kept for basic testing purposes.**

Use the following command to open it in a text editor:

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
#USB_PORT = /dev/usb/lp0
```

**Explanation of Fields:**

- **`MQTT_HOST`**: The address of your MQTT broker (e.g., `broker.hivemq.com`).
- **`MQTT_PORT`**: The port used by the MQTT broker (default: `8883` for secure connections).
- **`MQTT_USER`**: Your MQTT broker username (leave empty if not required).
- **`MQTT_PASS`**: Your MQTT broker password (leave empty if not required).
- **`TOPIC`**: The MQTT topic to which the service listens (e.g., `printerserver`).
- **`TCP_HOST`**: The IP address of your printer (usually your printer's network address or `127.0.0.1` for local).
- **`TCP_PORT`**: The port used to communicate with the printer (`9100` is the default for most network printers).
- **`USB_PORT`**: The USB host  used to communicate with the printer.

You should either comment TCP_HOST + TCP_PORT *OR* USB_HOST

After making changes, save and close the file (`Ctrl + O`, `Enter`, then `Ctrl + X` in Nano) and restart the service with

```bash
sudo systemctl restart printer
```

### **Step 5. Use `manage.sh` for service management**

```bash
./manage.sh
```

---

## **Testing the Service**

After starting the service (using the default `printer.config` file provided), you can test its functionality by publishing a message to the MQTT topic using the following command:

```bash
./test.sh
```

or

```bash
mosquitto_pub -h broker.hivemq.com -p 1883 -t printerserver -m '{
  "id": "'$(date +%s)'",
  "message": "Hello, world! Test message ID '$(date +%s)'.",
  "callback": "https://apimocha.com/printerserver/callback",
  "cut": "true"
}'
```

The service should process the message and print it and the result of the print operation (success or failure) will be sent to the specified callback URL.

You can check the callback response here: [https://apimocha.com/printerserver](https://apimocha.com/printerserver)

If you are unsure whether the message was sent correctly to the MQTT topic, you can monitor the topic by subscribing to it in a separate terminal. Use the following command:

```bash
mosquitto_sub -h broker.hivemq.com -p 1883 -t printerserver
```

---

## **Features**

- **Message Handling**: Processes and encodes messages (e.g., CP850 encoding).
- **MQTT Integration**: Receives and processes messages via MQTT.
- **Communication**: Sends processed messages to a TCP or USB printer.
- **Management Scripts**: Includes comprehensive tools for service configuration and management.

---

## **Core Components**

### 1. **Functionalities of `printer.go`**

The `printer.go` application is the core of the printer service, providing the following:

- **Message Processing**:
  - Handles messages from MQTT topics.
  - Encodes messages in CP850.
  - Sends messages via TCP or USB.

- **MQTT Integration**:
  - Connects to an MQTT broker.
  - Subscribes to configurable topics.
  - Decodes and processes MQTT messages.

- **Communication**:
  - Sends messages to a specified TCP or USB printer.
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
  "callback": "http://example.com/callback-endpoint",
  "cut": "true"
}
```

### **How It Works**:

1. **`id`**: A unique identifier for your message. It will appear in logs and callback responses.
2. **`message`**: The plain text or base64-encoded message to print.
3. **`callback`**: optional - URL where the service sends the result of the message processing. If key is not provided, the service processes the message without sending a response.
4. **`cut`**: optional - cut paper after each message. Can be `true` or `false`. If key is not provided, does not cut the paper.

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
