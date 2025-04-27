package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"vivalchemy/temperature_sensor/websocket/db/sqlc"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ALERT_URL = "http://localhost:8001/alert"
)

type Temperature struct {
	Temperature float64 `json:"temperature"`
}

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

var (
	queries  *sqlc.Queries
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all connections for simplicity
		},
	}
	clients    = make(map[*websocket.Conn]bool) // Connected clients
	writeMutex sync.Mutex                       // Mutex to protect writes to WebSocket connections
)

func initDB() error {
	// Open SQLite database
	dbConn, err := sql.Open("sqlite3", "../temperature.db")
	if err != nil {
		return err
	}

	// Test connection
	err = dbConn.Ping()
	if err != nil {
		return err
	}

	queries = sqlc.New(dbConn)

	// Create table if it does not exist
	err = queries.CreateTableIfNotExistsTemperature(context.Background())
	if err != nil {
		return err
	}

	// Initialize the queries
	return nil
}

func sendAlert(temp float64) {
	alert := map[string]float64{"temperature": temp}
	jsonData, err := json.Marshal(alert)
	if err != nil {
		log.Printf("Failed to marshal alert data: %v", err)
		return
	}

	resp, err := http.Post(ALERT_URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send alert to FastAPI server: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Alert sent to FastAPI server. Status code: %d", resp.StatusCode)
}

// Broadcast the temperature data to all connected WebSocket clients
// Broadcast the temperature data to all connected WebSocket clients
func broadcastTemperature(temp float64) {
	// Generate the timestamp for the current time
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Create the WebSocket message to be sent
	message := WSMessage{
		Type: "temperature",
		Payload: map[string]any{
			"temperature": temp,
			"timestamp":   timestamp, // Add timestamp to the payload
		},
	}

	// Marshal the message to JSON
	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	// Send the message to all connected clients
	writeMutex.Lock()
	defer writeMutex.Unlock()
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, messageData); err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// WebSocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Register new client
	clients[conn] = true
	defer delete(clients, conn)

	// Listen for messages from the WebSocket client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Handle incoming temperature data from the WebSocket client (e.g., ESP32)
		var temp Temperature
		if err := json.Unmarshal(message, &temp); err == nil {
			// Only process if it looks like a temperature message
			if temp.Temperature != 0 {
				// Log the received temperature before storing
				log.Printf("Received temperature data via WebSocket: %.2f°C", temp.Temperature)

				// Save temperature to database
				_, err = queries.CreateTemperature(context.Background(), temp.Temperature)
				if err != nil {
					log.Printf("Database error: %v", err)
					continue
				}

				// Send alert if temperature is abnormal
				if temp.Temperature > 40 || temp.Temperature < 15 {
					go sendAlert(temp.Temperature)
				}

				// Broadcast the temperature to all connected clients
				go broadcastTemperature(temp.Temperature)
			}
		}
	}
}

// Handle HTTP POST route to receive temperature data as a fallback
func temperaturePostHandler(w http.ResponseWriter, r *http.Request) {
	var temp Temperature
	// Decode the incoming JSON body into the Temperature struct
	if err := json.NewDecoder(r.Body).Decode(&temp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the received temperature
	log.Printf("Received temperature data via HTTP POST: %.2f°C", temp.Temperature)

	// Save temperature to the database
	_, err := queries.CreateTemperature(context.Background(), temp.Temperature)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Failed to save temperature data", http.StatusInternalServerError)
		return
	}

	// Send alert if temperature is abnormal
	if temp.Temperature > 40 || temp.Temperature < 15 {
		go sendAlert(temp.Temperature)
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Temperature data saved successfully"))
}

func main() {
	// Initialize DB
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup WebSocket handler
	http.HandleFunc("/ws", wsHandler)
	// Setup HTTP POST handler for fallback
	http.HandleFunc("/temperature", temperaturePostHandler)

	// Start HTTP server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
