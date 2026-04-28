package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Message represents a WebSocket message
type Message struct {
	Type      string `json:"type"`
	Data      any    `json:"data"`
	Timestamp string `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	conn   *websocket.Conn
	send   chan Message
	mutex  sync.Mutex
	closed bool
}

// Hub manages WebSocket clients and broadcasts
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the WebSocket hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(messageType string, data any) {
	message := Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// readPump handles messages from the WebSocket connection
func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		// Handle incoming messages if needed
		log.Printf("Received message: %s", message.Type)
	}
}

// writePump handles sending messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close safely closes the client connection
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		c.closed = true
		close(c.send)
		c.conn.Close()
	}
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan Message, 256),
	}
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
			return
		}

		client := NewClient(conn)
		hub.register <- client

		// Start goroutines for reading and writing
		go client.readPump(hub)
		go client.writePump()
	})
}

// Global hub instance
var globalHub *Hub

// InitializeWebSocket initializes the WebSocket hub
func InitializeWebSocket() {
	globalHub = NewHub()
	go globalHub.Run()
}

// GetHub returns the global WebSocket hub
func GetHub() *Hub {
	return globalHub
}

// BroadcastServiceUpdate broadcasts a service update to all clients
func BroadcastServiceUpdate(serviceName, status string, replicas, readyReplicas int) {
	if globalHub != nil {
		data := map[string]any{
			"service":       serviceName,
			"status":        status,
			"replicas":      replicas,
			"readyReplicas": readyReplicas,
		}
		globalHub.Broadcast("service_update", data)
	}
}

// BroadcastSystemMetrics broadcasts system metrics to all clients
func BroadcastSystemMetrics(cpu, memory, disk, network float64) {
	if globalHub != nil {
		data := map[string]any{
			"cpu":     cpu,
			"memory":  memory,
			"disk":    disk,
			"network": network,
		}
		globalHub.Broadcast("system_metrics", data)
	}
}

// BroadcastAlertUpdate broadcasts an alert update to all clients
func BroadcastAlertUpdate(id, severity, message, status string) {
	if globalHub != nil {
		data := map[string]any{
			"id":       id,
			"severity": severity,
			"message":  message,
			"status":   status,
		}
		globalHub.Broadcast("alert_update", data)
	}
}
