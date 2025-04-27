package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func main() {
	// Connect to WebSocket server
	url := "ws://localhost:8080/ws" // Replace with your WebSocket server URL
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()

	// Listen for messages from the WebSocket server
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Print received message
		var wsMessage WSMessage
		err = json.Unmarshal(message, &wsMessage)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch wsMessage.Type {
		case "temperature":
			payload, ok := wsMessage.Payload.(map[string]interface{})
			if ok {
				fmt.Printf("Received temperature: %.2f°C at %v\n", payload["temperature"], payload["timestamp"])
			}
		case "stats":
			payload, ok := wsMessage.Payload.(map[string]interface{})
			if ok {
				fmt.Println("Received stats:", payload)
			}
		default:
			fmt.Println("Received unknown message type:", wsMessage.Type)
		}
	}
}
