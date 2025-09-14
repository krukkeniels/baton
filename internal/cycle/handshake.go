package cycle

import (
	"context"
	"fmt"
	"time"

	"baton/internal/config"
	"baton/internal/llm"
	"baton/internal/storage"
)

// CompletionHandshake enforces completion handshake after cycle execution
type CompletionHandshake struct {
	store  *storage.Store
	config *config.CompletionConfig
}

// HandshakeResult represents the result of a completion handshake
type HandshakeResult struct {
	Success          bool     `json:"success"`
	FinalState       storage.State `json:"final_state"`
	ArtifactsCreated []string `json:"artifacts_created"`
	FollowUps        []string `json:"follow_ups"`
	Note             string   `json:"note"`
}

// NewCompletionHandshake creates a new completion handshake enforcer
func NewCompletionHandshake(store *storage.Store, config *config.CompletionConfig) *CompletionHandshake {
	return &CompletionHandshake{
		store:  store,
		config: config,
	}
}

// Enforce enforces the completion handshake
func (ch *CompletionHandshake) Enforce(ctx context.Context, taskID string, llmResponse *llm.Response) (*HandshakeResult, error) {
	result := &HandshakeResult{
		Success: false,
	}

	// Get the task to check its current state
	task, err := ch.store.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	initialState := task.State

	// Check if the task state was updated (the primary success condition)
	updatedTask, err := ch.store.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated task: %w", err)
	}

	// If state changed, handshake is successful
	if updatedTask.State != initialState {
		result.Success = true
		result.FinalState = updatedTask.State
		result.Note = "Task state successfully updated"

		// Check for artifacts created during this cycle
		artifacts, err := ch.store.ListArtifacts(taskID)
		if err == nil {
			for _, artifact := range artifacts {
				// Consider artifacts created in the last few seconds as "new"
				if time.Since(artifact.CreatedAt) < 30*time.Second {
					result.ArtifactsCreated = append(result.ArtifactsCreated, artifact.Name)
				}
			}
		}

		return result, nil
	}

	// State not updated - need to enforce completion handshake
	return ch.enforceHandshake(ctx, taskID, initialState, llmResponse)
}

// enforceHandshake performs the completion handshake enforcement
func (ch *CompletionHandshake) enforceHandshake(ctx context.Context, taskID string, initialState storage.State, llmResponse *llm.Response) (*HandshakeResult, error) {
	result := &HandshakeResult{
		Success:    false,
		FinalState: initialState,
	}

	// Attempt follow-up prompts with bounded retries
	for retry := 0; retry < ch.config.MaxRetries; retry++ {
		// Add delay between retries
		if retry > 0 {
			select {
			case <-time.After(time.Duration(ch.config.RetryDelaySeconds) * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Check if state was updated in the meantime
		task, err := ch.store.GetTask(taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task during handshake: %w", err)
		}

		if task.State != initialState {
			// Success - state was updated
			result.Success = true
			result.FinalState = task.State
			result.Note = fmt.Sprintf("Task state updated after follow-up %d", retry+1)
			return result, nil
		}

		// Add follow-up to record
		followUpMsg := ch.config.FollowUpTemplate
		result.FollowUps = append(result.FollowUps, followUpMsg)

		// In a real implementation, we would send the follow-up to the LLM
		// For now, we'll simulate waiting and checking again
		// TODO: Implement actual follow-up mechanism with LLM
	}

	// All retries exhausted - set task to needs_fixes
	if ch.config.RequireExplicitStateUpdate {
		note := fmt.Sprintf("Completion handshake failed after %d retries. Agent did not update task state.", ch.config.MaxRetries)
		
		if err := ch.store.UpdateTaskState(taskID, storage.NeedsFixes, note); err != nil {
			return nil, fmt.Errorf("failed to set task to needs_fixes: %w", err)
		}

			result.FinalState = storage.NeedsFixes
		result.Note = note
	}

	return result, nil
}

// ValidateCompletion validates that completion requirements are met
func (ch *CompletionHandshake) ValidateCompletion(taskID string, fromState, toState storage.State) error {
	// Check required handover artifacts
	requiredArtifacts := getRequiredHandovers(fromState, toState)

	for _, artifactName := range requiredArtifacts {
		artifact, err := ch.store.GetArtifact(taskID, artifactName, 0) // Get latest version
		if err != nil {
			return fmt.Errorf("required handover artifact '%s' not found for transition %s->%s", artifactName, fromState, toState)
		}

		if artifact.Content == "" {
			return fmt.Errorf("required handover artifact '%s' exists but is empty", artifactName)
		}
	}

	return nil
}

// getRequiredHandovers returns required handover artifacts for a transition
func getRequiredHandovers(from, to storage.State) []string {
	key := fmt.Sprintf("%s->%s", from, to)

	requiredHandovers := map[string][]string{
		"planning->ready_for_implementation":  {"implementation_plan"},
		"implementing->ready_for_code_review": {"change_summary"},
		"reviewing->ready_for_commit":         {"review_findings"},
		"reviewing->needs_fixes":              {"review_findings"},
		"fixing->ready_for_code_review":       {"fix_plan"},
		"committing->DONE":                    {"commit_summary"},
	}

	if handovers, exists := requiredHandovers[key]; exists {
		return handovers
	}

	return []string{}
}