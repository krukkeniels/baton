package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"baton/internal/storage"
)

// WebSocket message types
const (
	WSMessageTypeTaskCreated = "task_created"
	WSMessageTypeTaskUpdated = "task_updated"
	WSMessageTypeTaskDeleted = "task_deleted"
	WSMessageTypeStatusUpdate = "status_update"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add client to the list
	s.wsClientsMux.Lock()
	s.wsClients[conn] = true
	s.wsClientsMux.Unlock()

	// Remove client when done
	defer func() {
		s.wsClientsMux.Lock()
		delete(s.wsClients, conn)
		s.wsClientsMux.Unlock()
	}()

	// Send initial status
	s.sendStatusUpdate(conn)

	// Handle incoming messages (ping/pong, etc.)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// Echo back or handle specific messages if needed
	}
}

// broadcastTaskUpdate broadcasts a task update to all connected WebSocket clients
func (s *Server) broadcastTaskUpdate(action string, task *storage.Task) {
	messageType := ""
	switch action {
	case "created":
		messageType = WSMessageTypeTaskCreated
	case "updated":
		messageType = WSMessageTypeTaskUpdated
	case "deleted":
		messageType = WSMessageTypeTaskDeleted
	default:
		return
	}

	// Convert task to response format
	taskResp := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		State:       string(task.State),
		Priority:    task.Priority,
		Owner:       task.Owner,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	// Parse JSON fields
	if task.Tags != nil {
		json.Unmarshal(task.Tags, &taskResp.Tags)
	}
	if task.Dependencies != nil {
		json.Unmarshal(task.Dependencies, &taskResp.Dependencies)
	}

	message := WSMessage{
		Type:      messageType,
		Timestamp: task.UpdatedAt.Unix(),
		Data:      taskResp,
	}

	s.broadcastMessage(message)
}

// broadcastStatusUpdate broadcasts a status update to all connected clients
func (s *Server) broadcastStatusUpdate() {
	// Get current status
	tasksByState := make(map[string]int)
	totalTasks := 0

	for _, state := range []storage.State{
		storage.ReadyForPlan,
		storage.Planning,
		storage.ReadyForImplementation,
		storage.Implementing,
		storage.ReadyForCodeReview,
		storage.Reviewing,
		storage.ReadyForCommit,
		storage.Committing,
		storage.NeedsFixes,
		storage.Fixing,
		storage.Done,
	} {
		count, err := s.store.GetTaskCount(storage.TaskFilters{State: &state})
		if err != nil {
			log.Printf("Failed to get count for state %s: %v", state, err)
			continue
		}
		tasksByState[string(state)] = count
		totalTasks += count
	}

	status := map[string]interface{}{
		"tasks_by_state": tasksByState,
		"total_tasks":    totalTasks,
	}

	message := WSMessage{
		Type:      WSMessageTypeStatusUpdate,
		Timestamp: time.Now().Unix(),
		Data:      status,
	}

	s.broadcastMessage(message)
}

// sendStatusUpdate sends status update to a specific client
func (s *Server) sendStatusUpdate(conn *websocket.Conn) {
	// Get current status
	tasksByState := make(map[string]int)
	totalTasks := 0

	for _, state := range []storage.State{
		storage.ReadyForPlan,
		storage.Planning,
		storage.ReadyForImplementation,
		storage.Implementing,
		storage.ReadyForCodeReview,
		storage.Reviewing,
		storage.ReadyForCommit,
		storage.Committing,
		storage.NeedsFixes,
		storage.Fixing,
		storage.Done,
	} {
		count, err := s.store.GetTaskCount(storage.TaskFilters{State: &state})
		if err != nil {
			log.Printf("Failed to get count for state %s: %v", state, err)
			continue
		}
		tasksByState[string(state)] = count
		totalTasks += count
	}

	status := map[string]interface{}{
		"tasks_by_state": tasksByState,
		"total_tasks":    totalTasks,
	}

	message := WSMessage{
		Type:      WSMessageTypeStatusUpdate,
		Timestamp: time.Now().Unix(),
		Data:      status,
	}

	s.sendMessageToClient(conn, message)
}

// broadcastMessage sends a message to all connected WebSocket clients
func (s *Server) broadcastMessage(message WSMessage) {
	s.wsClientsMux.RLock()
	defer s.wsClientsMux.RUnlock()

	for client := range s.wsClients {
		s.sendMessageToClient(client, message)
	}
}

// sendMessageToClient sends a message to a specific WebSocket client
func (s *Server) sendMessageToClient(conn *websocket.Conn, message WSMessage) {
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
		// Connection might be dead, remove it
		s.wsClientsMux.Lock()
		delete(s.wsClients, conn)
		s.wsClientsMux.Unlock()
		conn.Close()
	}
}