// WebSocket Progress - Real-time backup progress tracking
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
	}

	wsClients = make(map[*websocket.Conn]bool)
	wsMutex   sync.RWMutex
)

// ProgressMessage represents real-time progress update
type ProgressMessage struct {
	Type           string                 `json:"type"` // backup, restore, verification
	JobID          string                 `json:"job_id"`
	SessionID      string                 `json:"session_id"`
	Percent        float64                `json:"percent"`
	CurrentFile    string                 `json:"current_file,omitempty"`
	FilesTotal     int                    `json:"files_total"`
	FilesProcessed int                    `json:"files_processed"`
	BytesTotal     int64                  `json:"bytes_total"`
	BytesProcessed int64                  `json:"bytes_processed"`
	Speed          string                 `json:"speed,omitempty"`
	ETA            string                 `json:"eta,omitempty"`
	Status         string                 `json:"status"` // running, success, failed, warning
	Message        string                 `json:"message,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

// WSProgressHandler handles WebSocket connections for progress updates
func WSProgressHandler(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	wsMutex.Lock()
	wsClients[conn] = true
	wsMutex.Unlock()

	// Send welcome message
	SendProgress(conn, &ProgressMessage{
		Type:      "system",
		Message:   "Підключено до моніторингу прогресу",
		Timestamp: time.Now(),
	})

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// Unregister client
	wsMutex.Lock()
	delete(wsClients, conn)
	wsMutex.Unlock()
}

// SendProgress sends progress update to all connected clients
func SendProgress(conn *websocket.Conn, msg *ProgressMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling progress: %v", err)
		return
	}

	wsMutex.RLock()
	defer wsMutex.RUnlock()

	for client := range wsClients {
		if conn != nil && client != conn {
			continue // Send to specific client only if conn provided
		}
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Error sending progress: %v", err)
			delete(wsClients, client)
		}
	}
}

// BroadcastProgress broadcasts progress to all clients
func BroadcastProgress(msg *ProgressMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	wsMutex.RLock()
	defer wsMutex.RUnlock()

	for client := range wsClients {
		client.WriteMessage(websocket.TextMessage, data)
	}
}

// GetProgressStats returns WebSocket statistics
func GetProgressStats(c *gin.Context) {
	wsMutex.RLock()
	clientCount := len(wsClients)
	wsMutex.RUnlock()

	c.JSON(200, gin.H{
		"connected_clients": clientCount,
		"status":            "running",
	})
}
