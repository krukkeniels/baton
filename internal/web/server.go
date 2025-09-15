package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"

	"baton/internal/config"
	"baton/internal/llm"
	"baton/internal/storage"
	"baton/internal/statemachine"
)

// Server represents the web UI server
type Server struct {
	store         *storage.Store
	config        *config.Config
	llmClient     llm.Client
	server        *http.Server
	wsUpgrader    websocket.Upgrader
	wsClients     map[*websocket.Conn]bool
	wsClientsMux  sync.RWMutex
	running       bool
	runningMux    sync.RWMutex
}

// NewServer creates a new web server
func NewServer(store *storage.Store, config *config.Config, llmClient llm.Client) *Server {
	return &Server{
		store:     store,
		config:    config,
		llmClient: llmClient,
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		wsClients: make(map[*websocket.Conn]bool),
	}
}

// Start starts the web server
func (s *Server) Start(port int) error {
	s.runningMux.Lock()
	defer s.runningMux.Unlock()

	if s.running {
		return fmt.Errorf("web server is already running")
	}

	// Create CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskByID)
	mux.HandleFunc("/api/tasks/create", s.handleCreateTask)
	mux.HandleFunc("/api/tasks/update", s.handleUpdateTask)
	mux.HandleFunc("/api/audit/", s.handleAuditHistory)
	mux.HandleFunc("/api/ws", s.handleWebSocket)
	mux.HandleFunc("/api/status", s.handleStatus)

	// Static file serving for the Next.js app
	fs := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", fs)

	handler := c.Handler(mux)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	s.running = true

	log.Printf("Web server starting on port %d", port)
	return s.server.ListenAndServe()
}

// Stop stops the web server
func (s *Server) Stop() error {
	s.runningMux.Lock()
	defer s.runningMux.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	// Close all WebSocket connections
	s.wsClientsMux.Lock()
	for client := range s.wsClients {
		client.Close()
	}
	s.wsClients = make(map[*websocket.Conn]bool)
	s.wsClientsMux.Unlock()

	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}

	return nil
}

// Task response structure for web UI
type TaskResponse struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	State        string                 `json:"state"`
	Priority     int                    `json:"priority"`
	Owner        string                 `json:"owner"`
	Tags         []string               `json:"tags"`
	Dependencies []string               `json:"dependencies"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Artifacts    []storage.Artifact     `json:"artifacts,omitempty"`
}

// handleTasks handles GET /api/tasks
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getTasks(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getTasks returns all tasks
func (s *Server) getTasks(w http.ResponseWriter, r *http.Request) {
	filters := storage.TaskFilters{}

	// Parse query parameters
	if state := r.URL.Query().Get("state"); state != "" {
		filters.State = (*storage.State)(&state)
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil {
			filters.Priority = &p
		}
	}

	tasks, err := s.store.ListTasks(filters)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tasks: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response []TaskResponse
	for _, task := range tasks {
		taskResp := TaskResponse{
			ID:           task.ID,
			Title:        task.Title,
			Description:  task.Description,
			State:        string(task.State),
			Priority:     task.Priority,
			Owner:        task.Owner,
			CreatedAt:    task.CreatedAt,
			UpdatedAt:    task.UpdatedAt,
		}

		// Parse JSON fields
		if task.Tags != nil {
			json.Unmarshal(task.Tags, &taskResp.Tags)
		}
		if task.Dependencies != nil {
			json.Unmarshal(task.Dependencies, &taskResp.Dependencies)
		}

		response = append(response, taskResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTaskByID handles GET/PUT/DELETE /api/tasks/{id}
func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID := strings.Split(path, "/")[0]

	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.getTask(w, taskID)
	case "PUT":
		s.updateTaskState(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getTask returns a single task with artifacts
func (s *Server) getTask(w http.ResponseWriter, taskID string) {
	task, err := s.store.GetTask(taskID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Get artifacts
	artifacts, err := s.store.ListArtifacts(taskID)
	if err != nil {
		log.Printf("Failed to get artifacts for task %s: %v", taskID, err)
		artifacts = []storage.Artifact{} // Continue without artifacts
	}

	taskResp := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		State:       string(task.State),
		Priority:    task.Priority,
		Owner:       task.Owner,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Artifacts:   artifacts,
	}

	// Parse JSON fields
	if task.Tags != nil {
		json.Unmarshal(task.Tags, &taskResp.Tags)
	}
	if task.Dependencies != nil {
		json.Unmarshal(task.Dependencies, &taskResp.Dependencies)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskResp)
}

// CreateTaskRequest represents a request to create a new task via LLM prompt
type CreateTaskRequest struct {
	Prompt string `json:"prompt"`
	Owner  string `json:"owner,omitempty"`
}

// handleCreateTask handles POST /api/tasks/create
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	// Use LLM to analyze the prompt and create task details
	task, err := s.createTaskFromPrompt(req.Prompt, req.Owner)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the task
	if err := s.store.CreateTask(task); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save task: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast update via WebSocket
	s.broadcastTaskUpdate("created", task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTaskRequest represents a request to update a task via LLM prompt
type UpdateTaskRequest struct {
	TaskID string `json:"task_id"`
	Prompt string `json:"prompt"`
}

// handleUpdateTask handles POST /api/tasks/update
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TaskID == "" || req.Prompt == "" {
		http.Error(w, "Task ID and prompt are required", http.StatusBadRequest)
		return
	}

	// Get current task
	task, err := s.store.GetTask(req.TaskID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Use LLM to analyze the prompt and update task
	updatedTask, err := s.updateTaskFromPrompt(task, req.Prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the updated task
	if err := s.store.UpdateTask(updatedTask); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save task: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast update via WebSocket
	s.broadcastTaskUpdate("updated", updatedTask)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// StatusResponse represents the current system status
type StatusResponse struct {
	TasksByState   map[string]int `json:"tasks_by_state"`
	TotalTasks     int            `json:"total_tasks"`
	RecentActivity []AuditEntry   `json:"recent_activity"`
}

type AuditEntry struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	TaskTitle string    `json:"task_title"`
	PrevState string    `json:"prev_state"`
	NextState string    `json:"next_state"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

// handleStatus handles GET /api/status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get task counts by state
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

	// Get recent audit entries (last 10)
	recentActivity, err := s.store.GetRecentAuditEntries(10)
	if err != nil {
		log.Printf("Failed to get recent audit entries: %v", err)
		recentActivity = []AuditEntry{}
	}

	response := StatusResponse{
		TasksByState:   tasksByState,
		TotalTasks:     totalTasks,
		RecentActivity: recentActivity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.runningMux.RLock()
	defer s.runningMux.RUnlock()
	return s.running
}