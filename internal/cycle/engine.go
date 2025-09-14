package cycle

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"baton/internal/config"
	"baton/internal/llm"
	"baton/internal/mcp"
	"baton/internal/statemachine"
	"baton/internal/storage"
	"baton/internal/audit"
)

// CycleEngine orchestrates the execution of a single cycle
type CycleEngine struct {
	store     *storage.Store
	config    *config.Config
	mcpServer *mcp.Server
	llmClient llm.Client
	selector  *statemachine.TaskSelector
	validator *statemachine.TransitionValidator
	auditor   *audit.Logger
	handshake *CompletionHandshake
}

// NewCycleEngine creates a new cycle engine
func NewCycleEngine(store *storage.Store, config *config.Config, llmClient llm.Client) *CycleEngine {
	selector := statemachine.NewTaskSelector(store, &config.Selection)
	validator := statemachine.NewTransitionValidator(store)
	auditor := audit.NewLogger(store)
	mcpServer := mcp.NewServer(store, config)
	handshake := NewCompletionHandshake(store, &config.Completion)

	return &CycleEngine{
		store:     store,
		config:    config,
		mcpServer: mcpServer,
		llmClient: llmClient,
		selector:  selector,
		validator: validator,
		auditor:   auditor,
		handshake: handshake,
	}
}

// ExecuteCycle executes a complete cycle
func (ce *CycleEngine) ExecuteCycle(ctx context.Context, dryRun bool) (*storage.CycleResult, error) {
	cycleID := uuid.New().String()
	start := time.Now()

	result := &storage.CycleResult{
		Success: false,
	}

	// Add timeout context if configured
	if ce.config.Development.CycleTimeboxSeconds > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(ce.config.Development.CycleTimeboxSeconds)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	// Step 1: Context reset (conceptual - new cycle starts fresh)
	// Step 2: Rehydrate context from stored sources (handled by task selection)

	// Step 3: Select next task
	selectionResult, err := ce.selector.SelectNext()
	if err != nil {
		return nil, fmt.Errorf("task selection failed: %w", err)
	}

	task := selectionResult.Task
	result.TaskID = task.ID
	result.PrevState = task.State

	// Step 4: Start MCP server
	if !dryRun {
		if err := ce.mcpServer.Start(); err != nil {
			return nil, fmt.Errorf("failed to start MCP server: %w", err)
		}
		defer ce.mcpServer.Stop()
	}

	// Step 5: Execute agent logic via LLM
	agent, err := ce.getAgentForTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent for task: %w", err)
	}

	prompt, err := ce.buildPrompt(task, agent)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	var llmResponse *llm.Response
	if !dryRun {
		llmResponse, err = ce.llmClient.Execute(ctx, prompt, agent.Name)
		if err != nil {
			return nil, fmt.Errorf("LLM execution failed: %w", err)
		}
	} else {
		// Dry run - simulate response
		llmResponse = &llm.Response{
			Success: true,
			Content: "[DRY RUN] Simulated agent response",
		}
	}

	// Step 6: Enforce completion handshake
	if !dryRun {
		handshakeResult, err := ce.handshake.Enforce(ctx, task.ID, llmResponse)
		if err != nil {
			return nil, fmt.Errorf("completion handshake failed: %w", err)
		}
		result.NextState = handshakeResult.FinalState
		result.ArtifactsCreated = handshakeResult.ArtifactsCreated
	} else {
		// Dry run - predict next state
		allowedStates, _ := statemachine.GetAllowedTransitions(task.State)
		if len(allowedStates) > 0 {
			result.NextState = allowedStates[0] // Take first allowed state
		} else {
			result.NextState = task.State // Stay in same state
		}
	}

	// Step 7: Create/update handover artifacts (handled by completion handshake)

	// Step 8: Record audit entry
	auditEntry := &storage.AuditLog{
		TaskID:          task.ID,
		CycleID:         cycleID,
		PrevState:       string(result.PrevState),
		NextState:       string(result.NextState),
		Actor:           agent.Name,
		SelectionReason: selectionResult.Reason,
		InputsSummary:   ce.buildInputsSummary(task),
		OutputsSummary:  ce.buildOutputsSummary(result.ArtifactsCreated),
		Result:          "success",
	}

	if llmResponse != nil {
		auditEntry.Note = fmt.Sprintf("LLM Response: %s", llmResponse.Content[:min(len(llmResponse.Content), 200)])
	}

	if !dryRun {
		if err := ce.auditor.LogCycle(auditEntry); err != nil {
			return nil, fmt.Errorf("failed to log audit entry: %w", err)
		}
	}

	// Step 9: Stop MCP server (handled by defer)

	// Step 10: Return cycle result
	result.Success = true
	result.Duration = time.Since(start)

	return result, nil
}

// getAgentForTask determines which agent should handle a task
func (ce *CycleEngine) getAgentForTask(task *storage.Task) (*config.Agent, error) {
	for agentID, agent := range ce.config.Agents {
		// Check if agent can handle this state
		for _, allowedState := range agent.AllowedStates {
			if allowedState == string(task.State) {
				return &agent, nil
			}
		}
		_ = agentID // Use the variable to avoid unused warning
	}

	return nil, fmt.Errorf("no agent configured for state %s", task.State)
}

// buildPrompt constructs the prompt for the LLM
func (ce *CycleEngine) buildPrompt(task *storage.Task, agent *config.Agent) (string, error) {
	// Base prompt structure
	prompt := fmt.Sprintf(`# %s Role

You are the %s for this project. %s

## Current Context
- **Task**: %s
- **Description**: %s
- **State**: %s
- **Priority**: %d

## Your Responsibilities
Handle the current task state (%s) according to your role.

## Important Rules
- Use the MCP tools to update task state and artifacts
- Follow the implementation plan exactly if one exists
- Create required handover artifacts before state transitions
- Update the task state when your work is complete

## Available MCP Methods
- baton.tasks.get - Get task details
- baton.tasks.update_state - Update task state
- baton.tasks.append_note - Add notes to task
- baton.artifacts.upsert - Create/update artifacts
- baton.artifacts.get - Get existing artifacts
- baton.plan.read - Read the project plan
- baton.requirements.list - List requirements

Please proceed with handling this task.`,
		agent.Name,
		agent.Name,
		agent.Role,
		task.Title,
		task.Description,
		task.State,
		task.Priority,
		task.State,
	)

	return prompt, nil
}

// buildInputsSummary creates a summary of cycle inputs
func (ce *CycleEngine) buildInputsSummary(task *storage.Task) string {
	return fmt.Sprintf("Task: %s (State: %s, Priority: %d)", task.Title, task.State, task.Priority)
}

// buildOutputsSummary creates a summary of cycle outputs
func (ce *CycleEngine) buildOutputsSummary(artifactsCreated []string) string {
	if len(artifactsCreated) == 0 {
		return "No artifacts created"
	}
	return fmt.Sprintf("Artifacts created: %v", artifactsCreated)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}