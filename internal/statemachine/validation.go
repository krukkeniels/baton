package statemachine

import (
	"encoding/json"
	"fmt"

	"baton/internal/storage"
)

// TransitionValidator handles state transition validation and enforcement
type TransitionValidator struct {
	store *storage.Store
}

// NewTransitionValidator creates a new transition validator
func NewTransitionValidator(store *storage.Store) *TransitionValidator {
	return &TransitionValidator{
		store: store,
	}
}

// ValidateAndTransition validates a transition and updates the task state
func (tv *TransitionValidator) ValidateAndTransition(taskID string, newState storage.State, note string) error {
	// Get current task
	task, err := tv.store.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task %s: %w", taskID, err)
	}

	// Validate the transition
	if err := ValidateTransition(task.State, newState); err != nil {
		return fmt.Errorf("transition validation failed: %w", err)
	}

	// Check dependencies if moving to a new work state
	if err := tv.validateDependencies(task, newState); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// Validate required handover artifacts
	if err := tv.validateRequiredHandovers(task, newState); err != nil {
		return fmt.Errorf("handover validation failed: %w", err)
	}

	// Perform the transition
	return tv.store.UpdateTaskState(taskID, newState, note)
}

// validateDependencies ensures all dependencies are satisfied before transition
func (tv *TransitionValidator) validateDependencies(task *storage.Task, newState storage.State) error {
	// Only check dependencies for certain states
	workStates := []storage.State{
		storage.Planning,
		storage.Implementing,
		storage.Reviewing,
		storage.Committing,
	}

	isWorkState := false
	for _, state := range workStates {
		if newState == state {
			isWorkState = true
			break
		}
	}

	if !isWorkState {
		return nil
	}

	// Parse dependencies
	var dependencies []string
	if len(task.Dependencies) > 0 {
		if err := json.Unmarshal(task.Dependencies, &dependencies); err != nil {
			return fmt.Errorf("failed to parse dependencies: %w", err)
		}
	}

	// Check each dependency
	for _, depID := range dependencies {
		depTask, err := tv.store.GetTask(depID)
		if err != nil {
			return fmt.Errorf("dependency task %s not found: %w", depID, err)
		}

		if depTask.State != storage.Done {
			return fmt.Errorf("dependency task %s (%s) is not complete (current state: %s)", depID, depTask.Title, depTask.State)
		}
	}

	return nil
}

// validateRequiredHandovers checks if required handover artifacts exist
func (tv *TransitionValidator) validateRequiredHandovers(task *storage.Task, newState storage.State) error {
	requiredHandovers := getRequiredHandovers(task.State, newState)

	for _, handover := range requiredHandovers {
		artifact, err := tv.store.GetArtifact(task.ID, handover, 0) // Get latest version
		if err != nil {
			return fmt.Errorf("required handover artifact '%s' not found for transition from %s to %s",
				handover, task.State, newState)
		}

		if artifact.Content == "" {
			return fmt.Errorf("required handover artifact '%s' exists but is empty", handover)
		}
	}

	return nil
}

// getRequiredHandovers returns the required handover artifacts for a state transition
func getRequiredHandovers(from, to storage.State) []string {
	key := fmt.Sprintf("%s->%s", from, to)

	requiredHandovers := map[string][]string{
		"planning->ready_for_implementation":       {"implementation_plan"},
		"implementing->ready_for_code_review":      {"change_summary"},
		"reviewing->ready_for_commit":              {"review_findings"},
		"reviewing->needs_fixes":                   {"review_findings"},
		"fixing->ready_for_code_review":            {"fix_plan"},
		"committing->DONE":                         {"commit_summary"},
	}

	if handovers, exists := requiredHandovers[key]; exists {
		return handovers
	}

	return []string{}
}

// GetTransitionRequirements returns information about what's needed for a transition
type TransitionRequirement struct {
	DependenciesBlocked []string `json:"dependencies_blocked,omitempty"`
	MissingHandovers    []string `json:"missing_handovers,omitempty"`
	IsValid             bool     `json:"is_valid"`
	Reason              string   `json:"reason,omitempty"`
}

// GetTransitionRequirements analyzes what's needed for a specific transition
func (tv *TransitionValidator) GetTransitionRequirements(taskID string, newState storage.State) (*TransitionRequirement, error) {
	req := &TransitionRequirement{
		IsValid: true,
	}

	// Get current task
	task, err := tv.store.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
	}

	// Check basic transition validity
	if err := ValidateTransition(task.State, newState); err != nil {
		req.IsValid = false
		req.Reason = err.Error()
		return req, nil
	}

	// Check dependencies
	var dependencies []string
	if len(task.Dependencies) > 0 {
		if err := json.Unmarshal(task.Dependencies, &dependencies); err == nil {
			for _, depID := range dependencies {
				depTask, err := tv.store.GetTask(depID)
				if err == nil && depTask.State != storage.Done {
					req.DependenciesBlocked = append(req.DependenciesBlocked,
						fmt.Sprintf("%s (%s): %s", depID, depTask.Title, depTask.State))
				}
			}
		}
	}

	// Check handovers
	requiredHandovers := getRequiredHandovers(task.State, newState)
	for _, handover := range requiredHandovers {
		if _, err := tv.store.GetArtifact(task.ID, handover, 0); err != nil {
			req.MissingHandovers = append(req.MissingHandovers, handover)
		}
	}

	// Determine if blocked
	if len(req.DependenciesBlocked) > 0 || len(req.MissingHandovers) > 0 {
		req.IsValid = false
		if len(req.DependenciesBlocked) > 0 {
			req.Reason = fmt.Sprintf("blocked by %d dependencies", len(req.DependenciesBlocked))
		} else {
			req.Reason = fmt.Sprintf("missing %d required handovers", len(req.MissingHandovers))
		}
	}

	return req, nil
}