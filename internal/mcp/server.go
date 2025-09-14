package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"baton/internal/config"
	"baton/internal/statemachine"
	"baton/internal/storage"
)

// Server represents the MCP server
type Server struct {
	store     *storage.Store
	config    *config.Config
	port      int
	server    *http.Server
	handlers  map[string]HandlerFunc
	mu        sync.RWMutex
	running   bool
}

// HandlerFunc represents a method handler
type HandlerFunc func(*JSONRPCRequest) *JSONRPCResponse

// NewServer creates a new MCP server
func NewServer(store *storage.Store, config *config.Config) *Server {
	server := &Server{
		store:    store,
		config:   config,
		port:     config.MCPPort,
		handlers: make(map[string]HandlerFunc),
	}

	// Register handlers
	server.registerHandlers()

	return server
}

// registerHandlers registers all MCP method handlers
func (s *Server) registerHandlers() {
	// Create handler instances
	selector := statemachine.NewTaskSelector(s.store, &s.config.Selection)
	validator := statemachine.NewTransitionValidator(s.store)

	taskHandler := NewTaskHandler(s.store, selector, validator)
	artifactHandler := NewArtifactHandler(s.store)
	requirementHandler := NewRequirementHandler(s.store)
	planHandler := NewPlanHandler(s.config.PlanFile)

	// Register task methods
	s.handlers["baton.tasks.get_next"] = taskHandler.GetNext
	s.handlers["baton.tasks.get"] = taskHandler.Get
	s.handlers["baton.tasks.update_state"] = taskHandler.UpdateState
	s.handlers["baton.tasks.append_note"] = taskHandler.AppendNote
	s.handlers["baton.tasks.list"] = taskHandler.List

	// Register artifact methods
	s.handlers["baton.artifacts.upsert"] = artifactHandler.Upsert
	s.handlers["baton.artifacts.get"] = artifactHandler.Get
	s.handlers["baton.artifacts.list"] = artifactHandler.List

	// Register requirement methods
	s.handlers["baton.requirements.list"] = requirementHandler.List

	// Register plan methods
	s.handlers["baton.plan.read"] = planHandler.Read

	// Register standard MCP methods
	s.handlers["initialize"] = s.handleInitialize
	s.handlers["ping"] = s.handlePing
}

// Start starts the MCP server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Check if running in STDIO mode (for Claude Code integration)
	if s.isSTDIOMode() {
		return s.runSTDIOMode()
	}

	// HTTP server mode
	return s.runHTTPMode()
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}

	return nil
}

// isSTDIOMode checks if server should run in STDIO mode
func (s *Server) isSTDIOMode() bool {
	// Check if stdin/stdout are connected to pipes (Claude Code integration)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// runSTDIOMode runs the server in STDIO mode for Claude Code integration
func (s *Server) runSTDIOMode() error {
	s.running = true

	scanner := bufio.NewScanner(os.Stdin)
	writer := json.NewEncoder(os.Stdout)

	for scanner.Scan() && s.running {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse JSON-RPC request
		req, err := ParseJSONRPCRequest([]byte(line))
		if err != nil {
			response := NewJSONRPCError(nil, ParseError, "Invalid JSON-RPC request", err.Error())
			if err := writer.Encode(response); err != nil {
				log.Printf("Failed to write error response: %v", err)
			}
			continue
		}

		// Handle request
		response := s.handleRequest(req)

		// Send response (only if not a notification)
		if !req.IsNotification() && response != nil {
			if err := writer.Encode(response); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
		}
	}

	return scanner.Err()
}

// runHTTPMode runs the server in HTTP mode
func (s *Server) runHTTPMode() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHTTP)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.running = true

	log.Printf("MCP server starting on port %d", s.port)
	return s.server.Serve(listener)
}

// handleHTTP handles HTTP requests
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	req, err := ParseJSONRPCRequest(body)
	if err != nil {
		response := NewJSONRPCError(nil, ParseError, "Invalid JSON-RPC request", err.Error())
		s.writeJSONResponse(w, response)
		return
	}

	response := s.handleRequest(req)
	if response != nil {
		s.writeJSONResponse(w, response)
	}
}

// writeJSONResponse writes a JSON response
func (s *Server) writeJSONResponse(w http.ResponseWriter, response *JSONRPCResponse) {
	w.Header().Set("Content-Type", "application/json")

	data, err := response.Marshal()
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(data); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// handleRequest processes a JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	handler, exists := s.handlers[req.Method]
	s.mu.RUnlock()

	if !exists {
		return NewJSONRPCError(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}

	// Call the handler
	return handler(req)
}

// handleInitialize handles the MCP initialize method
func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	params, err := req.GetParams()
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Invalid initialize parameters", nil)
	}

	// Extract client info
	clientInfo := map[string]interface{}{}
	if info, ok := params["clientInfo"].(map[string]interface{}); ok {
		clientInfo = info
	}

	// Build server capabilities
	capabilities := map[string]interface{}{
		"tools": map[string]interface{}{
			"listChanged": false,
		},
		"resources": map[string]interface{}{
			"subscribe":   false,
			"listChanged": false,
		},
	}

	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    capabilities,
		"serverInfo": map[string]interface{}{
			"name":    "baton",
			"title":   "Baton CLI Orchestrator",
			"version": "1.0.0",
		},
		"instructions": "Baton MCP server provides task orchestration capabilities. Use baton.* methods to interact with tasks, artifacts, requirements, and plans.",
	}

	log.Printf("MCP initialized for client: %v", clientInfo)
	return NewJSONRPCResponse(req.ID, result)
}

// handlePing handles the ping method
func (s *Server) handlePing(req *JSONRPCRequest) *JSONRPCResponse {
	return NewJSONRPCResponse(req.ID, map[string]interface{}{})
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetPort returns the server port
func (s *Server) GetPort() int {
	return s.port
}