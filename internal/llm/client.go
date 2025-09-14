package llm

import (
	"context"
	"time"
)

// Client represents an LLM client interface
type Client interface {
	Execute(ctx context.Context, prompt string, agentID string) (*Response, error)
	GetName() string
	IsAvailable() bool
}

// Response represents an LLM response
type Response struct {
	Success    bool            `json:"success"`
	Content    string          `json:"content"`
	Cost       float64         `json:"total_cost_usd"`
	Duration   time.Duration   `json:"duration"`
	SessionID  string          `json:"session_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	Error      error           `json:"error,omitempty"`
}

// ClientFactory creates LLM clients
type ClientFactory struct {
	clients map[string]Client
}

// NewClientFactory creates a new client factory
func NewClientFactory() *ClientFactory {
	return &ClientFactory{
		clients: make(map[string]Client),
	}
}

// Register registers an LLM client
func (f *ClientFactory) Register(name string, client Client) {
	f.clients[name] = client
}

// Get gets an LLM client by name
func (f *ClientFactory) Get(name string) (Client, bool) {
	client, exists := f.clients[name]
	return client, exists
}

// GetAvailable returns all available clients
func (f *ClientFactory) GetAvailable() []Client {
	var available []Client
	for _, client := range f.clients {
		if client.IsAvailable() {
			available = append(available, client)
		}
	}
	return available
}