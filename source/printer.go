package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"github.com/eclipse/paho.mqtt.golang"
)

// Global configuration variables
var (
	MQTT_HOST   string
	MQTT_PORT   int
	MQTT_USER   string
	MQTT_PASS   string
	TOPIC       string
	TCP_HOST    string
	TCP_PORT    int
	messageQueue = make(chan map[string]interface{}, 100)
	callbackQueue = make(chan map[string]string, 100)
	wg sync.WaitGroup
)

// Function to encode message to CP850
func encodeToCP850(input string) (string, error) {
	// Create a transformer for CP850 encoding
	encoder := charmap.CodePage850.NewEncoder()

	// Transform the string to CP850 encoding
	cp850Bytes, _, err := transform.Bytes(encoder, []byte(input))
	if err != nil {
		return "", err
	}

	// Convert the encoded bytes back to a string
	return string(cp850Bytes), nil
}




// Load configuration
func loadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("[ERROR] reading configuration file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "MQTT_HOST":
			MQTT_HOST = value
		case "MQTT_PORT":
			MQTT_PORT, _ = strconv.Atoi(value)
		case "MQTT_USER":
			MQTT_USER = value
		case "MQTT_PASS":
			MQTT_PASS = value
		case "TOPIC":
			TOPIC = value
		case "TCP_HOST":
			TCP_HOST = value
		case "TCP_PORT":
			TCP_PORT, _ = strconv.Atoi(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("[ERROR] Error scanning configuration file: %v", err)
	}

	return nil
}

func tcpSender() {
    buffer := []map[string]interface{}{}

    for {
        // Add new messages to the buffer
        select {
        case msg := <-messageQueue:
            // Add a timestamp to track when the message was queued
            msg["timestamp"] = time.Now()
            buffer = append(buffer, msg)
        default:
            // No new messages, continue processing the buffer
        }

        if len(buffer) == 0 {
            // Sleep briefly to avoid busy waiting
            time.Sleep(100 * time.Millisecond)
            continue
        }

        for len(buffer) > 0 {
            // Remove the first message from the buffer
            messageData := buffer[0]
            buffer = buffer[1:]

            // Check if the message has timed out
            if time.Since(messageData["timestamp"].(time.Time)) > 60*time.Second {
                fmt.Printf("[WARNING] Message timed out: %v\n", messageData["id"])
                callbackJSON, _ := json.Marshal(map[string]string{
                    "success": "false",
                    "id":      messageData["id"].(string),
                })
                callbackQueue <- map[string]string{
                    "url":  messageData["callback"].(string),
                    "data": string(callbackJSON),
                }
                continue
            }

            // Attempt to establish a TCP connection
            conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", TCP_HOST, TCP_PORT), 10*time.Second)
            if err != nil {
                fmt.Printf("[ERROR] Failed to establish TCP connection: %v\n", err)
                buffer = append([]map[string]interface{}{messageData}, buffer...)
                time.Sleep(5 * time.Second)
                break
            }

            // Send the message
            message := messageData["message"].(string)
            messageID := messageData["id"].(string)
            fmt.Printf("[DEBUG] %s - Processing message from TCP queue.\n", messageID)

            _, err = fmt.Fprintf(conn, message+"\n")
            conn.Close() // Close the connection after sending the message
            if err != nil {
                fmt.Printf("[ERROR] Error sending message via TCP: %v\n", err)
                buffer = append([]map[string]interface{}{messageData}, buffer...)
                break
            }
            fmt.Printf("[DEBUG] %s - Message sent via TCP.\n", messageID)

            // Prepare a success callback
            callbackJSON, _ := json.Marshal(map[string]string{
                "success": "true",
                "id":      messageID,
            })

            if callbackURL := messageData["callback"].(string); callbackURL != "" {
                callbackQueue <- map[string]string{
                    "url":  callbackURL,
                    "data": string(callbackJSON),
                }
            }

            // Remove the successfully sent message from the buffer (it has already been dequeued)
            break
        }
    }
}


// Callback sender thread
func callbackSender() {
	for callbackData := range callbackQueue {
		callbackURL := callbackData["url"]
		callbackJSON := callbackData["data"]

		resp, err := http.Post(callbackURL, "application/json", strings.NewReader(callbackJSON))
		if err != nil {
			fmt.Printf("[ERROR] Failed to call the callback URL: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("[DEBUG] POST successful: %d, Response: %s\n", resp.StatusCode, resp.Status)
	}
	wg.Done()
}

// MQTT message handler
func onMessage(client mqtt.Client, msg mqtt.Message) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		fmt.Printf("[ERROR] Failed to decode JSON received message: %v\n", err)
		return
	}
	messageID := data["id"].(string)
	message := data["message"].(string)
	fmt.Printf("[DEBUG] %s - Message received on MQTT topic.\n", messageID)

	decodedMessage, err := base64.StdEncoding.DecodeString(message)
	if err == nil {
		fmt.Printf("[DEBUG] %s - Base64 Message Decoded\n", messageID)
		data["message"] = string(decodedMessage)
        } else {
		fmt.Printf("[DEBUG] %s - Plain Text Message. No decoding needed\n", messageID)
	}


        // Convert the string to CP850
        message2 := data["message"].(string)
        cp850Encoded, err := encodeToCP850(message2)
        if err != nil {
                fmt.Printf("[ERROR] Error encoding to CP850: %v", err)
        } else {
                fmt.Printf("[DEBUG] %s - Reencoded to CP850.\n", messageID)
                // Replace the "message" value in the map with the CP850-encoded string
                data["message"] = string(cp850Encoded)
        }

	messageQueue <- data

	fmt.Printf("[DEBUG] %s - Message added to TCP queue.\n", messageID)
}

func main() {
	if err := loadConfig("printer.config"); err != nil {
		fmt.Println(err)
		return
	}

	// Start TCP sender
	wg.Add(1)
	go tcpSender()

	// Start callback sender
	wg.Add(1)
	go callbackSender()

	// Configure MQTT client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", MQTT_HOST, MQTT_PORT))
	opts.SetClientID("go-mqtt-client")
	opts.SetUsername(MQTT_USER)
	opts.SetPassword(MQTT_PASS)
	opts.SetTLSConfig(&tls.Config{})
	opts.OnConnect = func(c mqtt.Client) {
		if token := c.Subscribe(TOPIC, 0, onMessage); token.Wait() && token.Error() != nil {
			fmt.Printf("[ERROR] Error subscribing to topic: %v\n", token.Error())
		}
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("[ERROR] Failed to connect to MQTT broker: %v\n", token.Error())
		return
	}

	// Keep the main thread alive
	fmt.Println("Service running... Press Ctrl+C to exit.")
	select {}
}
