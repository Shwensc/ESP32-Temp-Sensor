#include <OneWire.h>
#include <DallasTemperature.h>
#include <WiFi.h>
#include <ArduinoWebsockets.h>
#include <ArduinoJson.h>

#define LOOP_DELAY 3000
#define BAUD_RATE 115200

// GPIO where the DS18B20 is connected to
const int oneWireBus = 4;

// Setup a OneWire instance to communicate with any OneWire devices
OneWire oneWire(oneWireBus);

// Pass our OneWire reference to Dallas Temperature sensor
DallasTemperature sensors(&oneWire);

// Wi-Fi credentials
const char* ssid = "My Nest";
const char* password = "virajrods3945";

// Server details
const char* websocketServer = "ws://192.168.2.105:8080/ws"; // WebSocket endpoint
const char* httpServer = "http://192.168.2.105:8080/temperature"; // Fallback HTTP endpoint

using namespace websockets;
WebsocketsClient webSocket;

// HTTP Client for fallback
HTTPClient http;

void setup() {
  pinMode(oneWireBus, INPUT_PULLUP);
  // Start the Serial Monitor
  Serial.begin(BAUD_RATE);
  // Start the DS18B20 sensor
  sensors.begin();

  // Connect to Wi-Fi
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.println("Connecting to WiFi...");
  }
  Serial.println("Connected to WiFi");
  
  // Connect to WebSocket server
  connectWebSocket();
}

void connectWebSocket() {
  bool connected = webSocket.connect(websocketServer);
  if (connected) {
    Serial.println("Connected to WebSocket server");
    // Set up callback
    webSocket.onMessage([&](WebsocketsMessage message) {
      Serial.print("Got Message: ");
      Serial.println(message.data());
    });
  } else {
    Serial.println("Failed to connect to WebSocket server, will use HTTP");
  }
}

void loop() {
  // Keep the WebSocket connection alive
  if (webSocket.available()) {
    webSocket.poll();
  } else {
    // Try to reconnect if disconnected
    if (WiFi.status() == WL_CONNECTED) {
      Serial.println("WebSocket disconnected, trying to reconnect...");
      connectWebSocket();
    }
  }

  // Request temperature
  sensors.requestTemperatures();
  float temperatureC = sensors.getTempCByIndex(0);

  // Check if reading is valid (not -127 which indicates an error)
  // if (temperatureC == DEVICE_DISCONNECTED_C) {
    // Serial.println("Error: Could not read temperature data");
    delay(LOOP_DELAY);
    // return;
  // }

  // Print temperature to Serial Monitor
  Serial.print("Temperature: ");
  Serial.print(temperatureC);
  Serial.println("ºC");

  // Create JSON document for temperature data
  StaticJsonDocument<200> doc;
  doc["temperature"] = temperatureC;
  
  // Serialize JSON to string
  String jsonData;
  serializeJson(doc, jsonData);

  // Send data via WebSocket if connected, otherwise use HTTP
  if (webSocket.available()) {
    // Send via WebSocket
    bool sent = webSocket.send(jsonData);
    if (sent) {
      Serial.println("Temperature sent via WebSocket");
    } else {
      Serial.println("Failed to send via WebSocket, falling back to HTTP");
      sendViaHTTP(jsonData);
    }
  } else {
    // Send via HTTP as fallback
    sendViaHTTP(jsonData);
  }

  // Wait before next reading
  delay(LOOP_DELAY);
}

void sendViaHTTP(String jsonData) {
  if (WiFi.status() == WL_CONNECTED) {
    http.begin(httpServer);
    http.addHeader("Content-Type", "application/json");
    
    int httpResponseCode = http.POST(jsonData);
    
    if (httpResponseCode > 0) {
      Serial.print("HTTP Response code: ");
      Serial.println(httpResponseCode);
    } else {
      Serial.print("Error on sending POST request: ");
      Serial.println(httpResponseCode);
    }
    http.end();
  }
}
