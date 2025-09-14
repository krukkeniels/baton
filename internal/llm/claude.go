package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"baton/internal/config"
)

// ClaudeClient implements the Claude Code LLM client
type ClaudeClient struct {
	config  *config.ClaudeConfig
	mcpPort int
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(config *config.ClaudeConfig, mcpPort int) *ClaudeClient {
	return &ClaudeClient{
		config:  config,
		mcpPort: mcpPort,
	}
}

// Execute executes a prompt using Claude Code
func (c *ClaudeClient) Execute(ctx context.Context, prompt string, agentID string) (*Response, error) {
	start := time.Now()

	// Build command arguments
	args := make([]string, len(c.config.HeadlessArgs))
	copy(args, c.config.HeadlessArgs)

	// Add prompt
	args = append(args, prompt)

	// Add output format
	if c.config.OutputFormat != "" {
		args = append(args, "--output-format", c.config.OutputFormat)
	}

	// Add MCP connection if enabled
	if c.config.MCPConnect && c.mcpPort > 0 {
		args = append(args, "--mcp", fmt.Sprintf("http://localhost:%d", c.mcpPort))
	}

	// Create command
	cmd := exec.CommandContext(ctx, c.config.Command, args...)
	cmd.Env = os.Environ()

	// Get pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude command: %w", err)
	}

	// Read output based on format
	var response *Response
	if c.config.OutputFormat == "stream-json" {
		response, err = c.parseStreamingJSON(stdout, stderr)
	} else {
		response, err = c.parseStandardOutput(stdout, stderr)
	}

	if err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		if response == nil {
			return nil, fmt.Errorf("claude command failed: %w", err)
		}
		// Command failed but we got some output
		response.Success = false
		response.Error = err
	}

	response.Duration = time.Since(start)
	return response, nil
}

// parseStreamingJSON parses streaming JSON output from Claude Code
func (c *ClaudeClient) parseStreamingJSON(stdout, stderr io.Reader) (*Response, error) {
	response := &Response{
		Success:  true,
		Metadata: make(map[string]interface{}),
	}

	// Read stderr in background for errors
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// Log stderr output (debugging info)
			// Don't treat stderr as error for Claude Code
		}
	}()

	// Parse streaming JSON from stdout
	scanner := bufio.NewScanner(stdout)
	var contentParts []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse JSON line
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not JSON, treat as plain text
			contentParts = append(contentParts, line)
			continue
		}

		// Handle different message types
		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "content":
			if content, ok := msg["content"].(string); ok {
				contentParts = append(contentParts, content)
			}
		case "result":
			// Final result message
			if cost, ok := msg["total_cost_usd"].(float64); ok {
				response.Cost = cost
			}
			if sessionID, ok := msg["session_id"].(string); ok {
				response.SessionID = sessionID
			}
			if metadata, ok := msg["metadata"].(map[string]interface{}); ok {
				response.Metadata = metadata
			}
		case "error":
			response.Success = false
			if errMsg, ok := msg["message"].(string); ok {
				response.Error = fmt.Errorf("claude error: %s", errMsg)
			}
		}
	}

	response.Content = strings.Join(contentParts, "\n")
	return response, scanner.Err()
}

// parseStandardOutput parses standard text output
func (c *ClaudeClient) parseStandardOutput(stdout, stderr io.Reader) (*Response, error) {
	// Read stdout
	content, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	// Read stderr
	errorOutput, err := io.ReadAll(stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to read stderr: %w", err)
	}

	response := &Response{
		Success:  true,
		Content:  string(content),
		Metadata: make(map[string]interface{}),
	}

	// If stderr has content, it might be an error
	if len(errorOutput) > 0 {
		errorStr := string(errorOutput)
		// Claude Code uses stderr for logging, not all stderr is errors
		if strings.Contains(strings.ToLower(errorStr), "error") {
			response.Success = false
			response.Error = fmt.Errorf("claude error: %s", errorStr)
		}
	}

	return response, nil
}

// GetName returns the client name
func (c *ClaudeClient) GetName() string {
	return "claude"
}

// IsAvailable checks if Claude Code is available
func (c *ClaudeClient) IsAvailable() bool {
	_, err := exec.LookPath(c.config.Command)
	return err == nil
}