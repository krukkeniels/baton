package web

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"baton/internal/storage"
	"baton/internal/statemachine"
)

// LLM prompts for task creation and updates
const (
	taskCreationPrompt = `You are an expert task analyst for a software development project.
Based on the user's request, analyze and extract the following information to create a well-structured task:

User Request: "%s"

Please provide a JSON response with the following structure:
{
  "title": "Clear, concise task title (max 80 chars)",
  "description": "Detailed description of what needs to be done",
  "priority": 5,
  "state": "ready_for_plan",
  "owner": "%s",
  "tags": ["tag1", "tag2"],
  "dependencies": [],
  "estimated_complexity": "low|medium|high",
  "acceptance_criteria": [
    "Specific, testable criteria"
  ]
}

Guidelines:
- Priority scale: 1-10 (1=lowest, 10=highest, 5=normal)
- State should be "ready_for_plan" for new tasks
- Tags should be relevant technology or domain keywords
- Dependencies should reference existing task IDs if mentioned
- Acceptance criteria should be specific and testable

Respond with ONLY the JSON object, no additional text.`

	taskUpdatePrompt = `You are an expert task analyst. Based on the user's update request, analyze the current task and provide the necessary changes.

Current Task:
Title: %s
Description: %s
State: %s
Priority: %d
Tags: %s
Dependencies: %s

User Update Request: "%s"

Please provide a JSON response with ONLY the fields that should be updated:
{
  "title": "New title if changed",
  "description": "Updated description if changed",
  "priority": 7,
  "state": "new_state_if_changed",
  "tags": ["updated", "tags", "if", "changed"],
  "dependencies": ["updated_deps_if_changed"],
  "update_reason": "Brief explanation of what was changed and why"
}

State transition rules:
- ready_for_plan → planning
- planning → ready_for_implementation
- ready_for_implementation → implementing
- implementing → ready_for_code_review
- ready_for_code_review → reviewing
- reviewing → ready_for_commit | needs_fixes
- ready_for_commit → committing
- committing → DONE
- needs_fixes → fixing
- fixing → ready_for_code_review

Only include fields that actually need to be updated. If no changes are needed, return {"update_reason": "No changes needed"}.

Respond with ONLY the JSON object, no additional text.`
)

// TaskCreationResponse represents the LLM response for task creation
type TaskCreationResponse struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           int      `json:"priority"`
	State              string   `json:"state"`
	Owner              string   `json:"owner"`
	Tags               []string `json:"tags"`
	Dependencies       []string `json:"dependencies"`
	EstimatedComplexity string   `json:"estimated_complexity"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
}

// TaskUpdateResponse represents the LLM response for task updates
type TaskUpdateResponse struct {
	Title        *string  `json:"title,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	State        *string  `json:"state,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	UpdateReason string   `json:"update_reason"`
}

// createTaskFromPrompt uses LLM to create a task from a natural language prompt
func (s *Server) createTaskFromPrompt(prompt string, owner string) (*storage.Task, error) {
	if owner == "" {
		owner = "system"
	}

	// Format the prompt for the LLM
	llmPrompt := fmt.Sprintf(taskCreationPrompt, prompt, owner)

	// Call the LLM
	response, err := s.llmClient.GenerateText(llmPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse the JSON response
	var taskResp TaskCreationResponse
	if err := json.Unmarshal([]byte(response), &taskResp); err != nil {
		// Try to extract JSON from the response if it's wrapped in other text
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}") + 1
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart:jsonEnd]
			if err := json.Unmarshal([]byte(jsonStr), &taskResp); err != nil {
				return nil, fmt.Errorf("failed to parse LLM response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse LLM response: %w", err)
		}
	}

	// Validate and normalize the response
	if taskResp.Title == "" {
		return nil, fmt.Errorf("LLM did not provide a task title")
	}

	if taskResp.Priority < 1 || taskResp.Priority > 10 {
		taskResp.Priority = 5 // Default to normal priority
	}

	// Normalize state
	state := storage.NormalizeState(taskResp.State)
	if state == "" {
		state = storage.ReadyForPlan
	}

	// Convert to storage format
	tags, _ := json.Marshal(taskResp.Tags)
	deps, _ := json.Marshal(taskResp.Dependencies)

	task := &storage.Task{
		ID:           uuid.New().String(),
		Title:        taskResp.Title,
		Description:  taskResp.Description,
		State:        state,
		Priority:     taskResp.Priority,
		Owner:        taskResp.Owner,
		Tags:         tags,
		Dependencies: deps,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create initial artifacts if we have acceptance criteria
	if len(taskResp.AcceptanceCriteria) > 0 {
		criteriaContent := "## Acceptance Criteria\n\n"
		for i, criteria := range taskResp.AcceptanceCriteria {
			criteriaContent += fmt.Sprintf("%d. %s\n", i+1, criteria)
		}

		if taskResp.EstimatedComplexity != "" {
			criteriaContent += fmt.Sprintf("\n## Estimated Complexity\n%s\n", taskResp.EstimatedComplexity)
		}

		// This artifact will be created after the task is saved
		// We'll add it as metadata for now
		task.Description += "\n\n" + criteriaContent
	}

	return task, nil
}

// updateTaskFromPrompt uses LLM to update a task based on a natural language prompt
func (s *Server) updateTaskFromPrompt(task *storage.Task, prompt string) (*storage.Task, error) {
	// Parse current tags and dependencies for the prompt
	var tags []string
	var deps []string

	if task.Tags != nil {
		json.Unmarshal(task.Tags, &tags)
	}
	if task.Dependencies != nil {
		json.Unmarshal(task.Dependencies, &deps)
	}

	tagsStr, _ := json.Marshal(tags)
	depsStr, _ := json.Marshal(deps)

	// Format the prompt for the LLM
	llmPrompt := fmt.Sprintf(taskUpdatePrompt,
		task.Title,
		task.Description,
		string(task.State),
		task.Priority,
		string(tagsStr),
		string(depsStr),
		prompt,
	)

	// Call the LLM
	response, err := s.llmClient.GenerateText(llmPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse the JSON response
	var updateResp TaskUpdateResponse
	if err := json.Unmarshal([]byte(response), &updateResp); err != nil {
		// Try to extract JSON from the response if it's wrapped in other text
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}") + 1
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart:jsonEnd]
			if err := json.Unmarshal([]byte(jsonStr), &updateResp); err != nil {
				return nil, fmt.Errorf("failed to parse LLM response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse LLM response: %w", err)
		}
	}

	// Create updated task
	updatedTask := *task
	updatedTask.UpdatedAt = time.Now()

	// Apply updates only for fields that were changed
	if updateResp.Title != nil {
		updatedTask.Title = *updateResp.Title
	}
	if updateResp.Description != nil {
		updatedTask.Description = *updateResp.Description
	}
	if updateResp.Priority != nil {
		if *updateResp.Priority >= 1 && *updateResp.Priority <= 10 {
			updatedTask.Priority = *updateResp.Priority
		}
	}
	if updateResp.State != nil {
		newState := storage.NormalizeState(*updateResp.State)
		if newState != "" {
			// Validate state transition
			validator := statemachine.NewTransitionValidator(s.store)
			if validator.IsValidTransition(task.State, newState) {
				updatedTask.State = newState
			} else {
				return nil, fmt.Errorf("invalid state transition from %s to %s", task.State, newState)
			}
		}
	}
	if updateResp.Tags != nil {
		newTags, _ := json.Marshal(updateResp.Tags)
		updatedTask.Tags = newTags
	}
	if updateResp.Dependencies != nil {
		newDeps, _ := json.Marshal(updateResp.Dependencies)
		updatedTask.Dependencies = newDeps
	}

	// Add update reason as a note if provided
	if updateResp.UpdateReason != "" && updateResp.UpdateReason != "No changes needed" {
		// This could be logged or added as an artifact
		// For now, we'll append it to the description as a timestamped note
		updatedTask.Description += fmt.Sprintf("\n\n---\n**Update (%s):** %s",
			time.Now().Format("2006-01-02 15:04:05"), updateResp.UpdateReason)
	}

	return &updatedTask, nil
}