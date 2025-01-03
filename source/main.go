package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Config holds all configuration parameters
type Config struct {
	MQTTHost string
	MQTTPort int
	MQTTUser string
	MQTTPass string
	Topic    string
	TCPHost  string
	TCPPort  int
	USBHost  string
}

// PrintMessage represents the structure of incoming print messages
type PrintMessage struct {
    ID        string    `json:"id"`
    Message   string    `json:"message"`
    Callback  string    `json:"callback"`
    Cut       string    `json:"cut"`
    Timestamp time.Time
}

// CallbackData represents callback information
type CallbackData struct {
	URL  string
	Data string
}

// PrinterService manages the printer communication
type PrinterService struct {
	config        Config
	logger        *zap.Logger
	messageQueue  chan PrintMessage
	callbackQueue chan CallbackData
	mqttClient    mqtt.Client
	httpClient    *http.Client
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewPrinterService creates a new printer service instance
func NewPrinterService(configPath string) (*PrinterService, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PrinterService{
		config:        config,
		logger:        logger,
		messageQueue:  make(chan PrintMessage, 100),
		callbackQueue: make(chan CallbackData, 100),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// loadConfig reads and parses the configuration file
func loadConfig(filePath string) (Config, error) {
	var config Config
	file, err := os.Open(filePath)
	if err != nil {
		return config, fmt.Errorf("error reading configuration file: %w", err)
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

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

		switch key {
		case "MQTT_HOST":
			config.MQTTHost = value
		case "MQTT_PORT":
			config.MQTTPort, _ = strconv.Atoi(value)
		case "MQTT_USER":
			config.MQTTUser = value
		case "MQTT_PASS":
			config.MQTTPass = value
		case "TOPIC":
			config.Topic = value
		case "TCP_HOST":
			config.TCPHost = value
		case "TCP_PORT":
			config.TCPPort, _ = strconv.Atoi(value)
		case "USB_HOST":
			config.USBHost = value
		}
	}

	return config, scanner.Err()
}

// Start initializes and starts all service components
func (s *PrinterService) Start() error {
	if err := s.setupMQTTClient(); err != nil {
		return fmt.Errorf("failed to setup MQTT client: %w", err)
	}

	s.wg.Add(3)
	go s.tcpSender()
	go s.callbackSender()
	go s.healthCheck()

	s.logger.Info("Service started successfully")
	return nil
}

// Stop gracefully shuts down the service
func (s *PrinterService) Stop() {
	s.cancel()
	if s.mqttClient != nil && s.mqttClient.IsConnected() {
		s.mqttClient.Disconnect(250)
	}
	close(s.messageQueue)
	close(s.callbackQueue)
	s.wg.Wait()
	s.logger.Info("Service stopped")
}

// setupMQTTClient configures and connects the MQTT client
func (s *PrinterService) setupMQTTClient() error {
	// First try TLS connection
	if err := s.connectMQTT(true); err != nil {
		s.logger.Warn("TLS connection failed, falling back to TCP",
			zap.Error(err),
			zap.String("host", s.config.MQTTHost))

		// Attempt TCP connection
		if err := s.connectMQTT(false); err != nil {
			return fmt.Errorf("both TLS and TCP connections failed: %w", err)
		}

		s.logger.Warn("Connected using non-secure TCP connection. It is recommended to use TLS in production environments.")
	} else {
		s.logger.Info("Connected successfully using TLS")
	}

	return nil
}

// connectMQTT attempts to connect to the MQTT broker using the specified protocol
func (s *PrinterService) connectMQTT(useTLS bool) error {
    protocol := "tcp"
    if useTLS {
        protocol = "tls"
    }

    brokerURL := fmt.Sprintf("%s://%s:%d", protocol, s.config.MQTTHost, s.config.MQTTPort)

    opts := mqtt.NewClientOptions().
        AddBroker(brokerURL).
        SetClientID(fmt.Sprintf("printer-service-%d", time.Now().Unix())).
        SetUsername(s.config.MQTTUser).
        SetPassword(s.config.MQTTPass).
        SetAutoReconnect(true).
        SetMaxReconnectInterval(1 * time.Minute).
        SetOnConnectHandler(s.onConnect).  // Changed from SetOnConnect to SetOnConnectHandler
        SetConnectTimeout(10 * time.Second)

    if useTLS {
        opts.SetTLSConfig(&tls.Config{})
    }

    client := mqtt.NewClient(opts)
    token := client.Connect()

    if !token.WaitTimeout(10 * time.Second) {
        return fmt.Errorf("connection timeout")
    }

    if err := token.Error(); err != nil {
        return fmt.Errorf("connection failed: %w", err)
    }

    s.mqttClient = client
    return nil
}

// onConnect handles MQTT connection and subscription
func (s *PrinterService) onConnect(client mqtt.Client) {
	const maxRetries = 5
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if token := client.Subscribe(s.config.Topic, 0, s.handleMessage); token.Wait() && token.Error() != nil {
			s.logger.Error("Failed to subscribe",
				zap.String("topic", s.config.Topic),
				zap.Error(token.Error()),
				zap.Int("attempt", attempt))

			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return
		}

		s.logger.Info("Successfully subscribed to topic", zap.String("topic", s.config.Topic))
		return
	}
}

// handleMessage processes incoming MQTT messages
func (s *PrinterService) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	var printMsg PrintMessage
	if err := json.Unmarshal(msg.Payload(), &printMsg); err != nil {
		s.logger.Error("Failed to decode message", zap.Error(err))
		return
	}

	printMsg.Timestamp = time.Now()
	s.logger.Debug("Message received", zap.String("id", printMsg.ID))

	// Handle base64 decoding
	if decodedMsg, err := base64.StdEncoding.DecodeString(printMsg.Message); err == nil {
		printMsg.Message = string(decodedMsg)
		s.logger.Debug("Message decoded from base64", zap.String("id", printMsg.ID))
	}

	// Convert to CP850
	if encodedMsg, err := s.encodeToCP850(printMsg.Message); err == nil {
		printMsg.Message = encodedMsg
		s.logger.Debug("Message encoded to CP850", zap.String("id", printMsg.ID))
	} else {
		s.logger.Error("Failed to encode to CP850", zap.Error(err))
		return
	}

	s.messageQueue <- printMsg
}

// encodeToCP850 converts text to CP850 encoding
func (s *PrinterService) encodeToCP850(input string) (string, error) {
	encoder := charmap.CodePage850.NewEncoder()
	result, _, err := transform.String(encoder, input)
	return result, err
}

// tcpSender manages TCP message sending with retries and backoff
func (s *PrinterService) tcpSender() {
	defer s.wg.Done()

	buffer := make([]PrintMessage, 0, 100)
	backoff := time.Second

	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.messageQueue:
			buffer = append(buffer, msg)
		default:
			if len(buffer) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

//			if err := s.processTCPMessage(&buffer[0]); err != nil {
			if err := s.processMessage(&buffer[0]); err != nil {
				s.logger.Error("Failed to process TCP message",
					zap.Error(err),
					zap.String("id", buffer[0].ID))

				if time.Since(buffer[0].Timestamp) > time.Minute {
					s.sendCallback(buffer[0].ID, buffer[0].Callback, false)
					buffer = buffer[1:]
				} else {
					time.Sleep(backoff)
					backoff = time.Duration(float64(backoff) * 1.5)
					if backoff > 30*time.Second {
						backoff = 30 * time.Second
					}
				}
				continue
			}

			s.sendCallback(buffer[0].ID, buffer[0].Callback, true)
			buffer = buffer[1:]
			backoff = time.Second
		}
	}
}

// processMessage processes a single message based on the configuration
func (s *PrinterService) processMessage(msg *PrintMessage) error {
    if s.config.USBHost == "" {
        // USBHost is nil or empty, use TCP
        conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.config.TCPHost, s.config.TCPPort), 10*time.Second)
        if err != nil {
            return fmt.Errorf("TCP connection failed: %w", err)
        }
        defer conn.Close()

        if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
            return fmt.Errorf("failed to set write deadline: %w", err)
        }

        // Send the message
        if _, err := fmt.Fprintf(conn, msg.Message); err != nil {
            return fmt.Errorf("failed to send message: %w", err)
        }
		s.logger.Debug("TCP Printed Message:", zap.String("msg",msg.Message))
        // Handle paper cut
        if strings.ToLower(msg.Cut) == "true" {
			s.logger.Info("TCP Paper Cut")
            cutCommand := []byte{'\n',0x1B, 0x6D, '\n'} // ESC/POS partial cut command
            if _, err := conn.Write(cutCommand); err != nil {
                return fmt.Errorf("failed to send cut command: %w", err)
            }
            s.logger.Debug("Paper cut command sent via TCP", zap.String("id", msg.ID))
        }
    } else {
        // Use USBHost to send the message
        file, err := os.OpenFile(s.config.USBHost, os.O_WRONLY, 0666)
        if err != nil {
            return fmt.Errorf("failed to open USB device: %w", err)
        }
        defer file.Close()

        // Write the message
		time.Sleep(100 * time.Millisecond) // Wait before sending the next command
        if _, err := file.WriteString(msg.Message); err != nil {
            return fmt.Errorf("failed to send message to USB device: %w", err)
        } 
		s.logger.Debug("USB Printed Message:", zap.String("msg",msg.Message))
		time.Sleep(100 * time.Millisecond) // Wait before sending the next command
        // Handle paper cut
        if strings.ToLower(msg.Cut) == "true" {
            cutCommand := []byte{'\n',0x1B, 0x6D, '\n'} // ESC/POS partial cut command
            if _, err := file.Write(cutCommand); err != nil {
                return fmt.Errorf("failed to send cut command to USB device: %w", err)
            }
            s.logger.Debug("Paper cut command sent to USB", zap.String("id", msg.ID))
        }
    }
    return nil
}

// callbackSender manages HTTP callbacks
func (s *PrinterService) callbackSender() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case callback := <-s.callbackQueue:
			if err := s.sendHTTPCallback(callback); err != nil {
				s.logger.Error("Callback failed", zap.Error(err))
			}
		}
	}
}

// sendCallback queues a callback message
func (s *PrinterService) sendCallback(messageID, callbackURL string, success bool) {
	if callbackURL == "" {
		return
	}

	callbackJSON, _ := json.Marshal(map[string]string{
		"success": strconv.FormatBool(success),
		"id":      messageID,
	})

	s.callbackQueue <- CallbackData{
		URL:  callbackURL,
		Data: string(callbackJSON),
	}
}

// sendHTTPCallback sends an HTTP callback
func (s *PrinterService) sendHTTPCallback(callback CallbackData) error {
	resp, err := s.httpClient.Post(callback.URL, "application/json", strings.NewReader(callback.Data))
	if err != nil {
		return fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP POST returned status: %d", resp.StatusCode)
	}

	return nil
}

// healthCheck periodically checks service health
func (s *PrinterService) healthCheck() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if !s.mqttClient.IsConnected() {
				s.logger.Error("MQTT client disconnected")
			}
			s.logger.Info("Health check",
				zap.Int("message_queue_len", len(s.messageQueue)),
				zap.Int("callback_queue_len", len(s.callbackQueue)))
		}
	}
}

func main() {
	service, err := NewPrinterService("printer.config")
	if err != nil {
		fmt.Printf("Failed to initialize service: %v\n", err)
		os.Exit(1)
	}

	if err := service.Start(); err != nil {
		fmt.Printf("Failed to start service: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	service.Stop()
}
