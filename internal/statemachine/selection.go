package statemachine

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"baton/internal/config"
	"baton/internal/storage"
)

// TaskSelector implements task selection algorithms
type TaskSelector struct {
	store  *storage.Store
	config *config.SelectionConfig
}

// NewTaskSelector creates a new task selector
func NewTaskSelector(store *storage.Store, config *config.SelectionConfig) *TaskSelector {
	return &TaskSelector{
		store:  store,
		config: config,
	}
}

// SelectionResult represents the result of task selection
type SelectionResult struct {
	Task   *storage.Task `json:"task"`
	Reason string        `json:"reason"`
}

// SelectNext selects the next task to work on
func (ts *TaskSelector) SelectNext() (*SelectionResult, error) {
	// Get all non-terminal tasks
	tasks, err := ts.getSelectableTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get selectable tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no selectable tasks available")
	}

	// Apply selection algorithm
	switch ts.config.Algorithm {
	case "priority_dependency":
		return ts.selectByPriorityAndDependency(tasks)
	default:
		return nil, fmt.Errorf("unknown selection algorithm: %s", ts.config.Algorithm)
	}
}

// getSelectableTasks returns tasks that are not in terminal states
func (ts *TaskSelector) getSelectableTasks() ([]*storage.Task, error) {
	allTasks, err := ts.store.ListTasks(storage.TaskFilters{})
	if err != nil {
		return nil, err
	}

	var selectable []*storage.Task
	for _, task := range allTasks {
		if !IsTerminalState(task.State) {
			selectable = append(selectable, task)
		}
	}

	return selectable, nil
}

// selectByPriorityAndDependency implements the priority+dependency selection algorithm
func (ts *TaskSelector) selectByPriorityAndDependency(tasks []*storage.Task) (*SelectionResult, error) {
	// Filter out blocked tasks
	var candidates []*taskCandidate
	for _, task := range tasks {
		candidate := &taskCandidate{
			Task:     task,
			Blocked:  false,
			IsLeaf:   true,
			Priority: task.Priority,
		}

		// Check if blocked by dependencies
		if blocked, reason := ts.isBlockedByDependencies(task); blocked {
			candidate.Blocked = true
			candidate.BlockReason = reason
		}

		// Check if it's a leaf task (no other tasks depend on it)
		if hasDependent, err := ts.hasUnfinishedDependents(task); err != nil {
			return nil, err
		} else {
			candidate.IsLeaf = !hasDependent
		}

		candidates = append(candidates, candidate)
	}

	// Filter out blocked candidates
	var availableCandidates []*taskCandidate
	for _, candidate := range candidates {
		if !candidate.Blocked {
			availableCandidates = append(availableCandidates, candidate)
		}
	}

	if len(availableCandidates) == 0 {
		return nil, fmt.Errorf("no unblocked tasks available")
	}

	// Sort by selection criteria
	ts.sortCandidates(availableCandidates)

	// Select the first candidate
	selected := availableCandidates[0]
	reason := ts.buildSelectionReason(selected, len(candidates), len(availableCandidates))

	return &SelectionResult{
		Task:   selected.Task,
		Reason: reason,
	}, nil
}

// taskCandidate represents a task being considered for selection
type taskCandidate struct {
	Task        *storage.Task
	Blocked     bool
	BlockReason string
	IsLeaf      bool
	Priority    int
}

// isBlockedByDependencies checks if a task is blocked by incomplete dependencies
func (ts *TaskSelector) isBlockedByDependencies(task *storage.Task) (bool, string) {
	if !ts.config.DependencyStrict {
		return false, ""
	}

	var dependencies []string
	if len(task.Dependencies) > 0 {
		if err := json.Unmarshal(task.Dependencies, &dependencies); err != nil {
			return true, fmt.Sprintf("invalid dependencies JSON: %v", err)
		}
	}

	for _, depID := range dependencies {
		depTask, err := ts.store.GetTask(depID)
		if err != nil {
			return true, fmt.Sprintf("dependency %s not found", depID)
		}

		if depTask.State != storage.Done {
			return true, fmt.Sprintf("dependency %s (%s) not complete", depID, depTask.Title)
		}
	}

	return false, ""
}

// hasUnfinishedDependents checks if other unfinished tasks depend on this task
func (ts *TaskSelector) hasUnfinishedDependents(task *storage.Task) (bool, error) {
	allTasks, err := ts.store.ListTasks(storage.TaskFilters{})
	if err != nil {
		return false, err
	}

	for _, otherTask := range allTasks {
		if otherTask.ID == task.ID || otherTask.State == storage.Done {
			continue
		}

		var dependencies []string
		if len(otherTask.Dependencies) > 0 {
			if err := json.Unmarshal(otherTask.Dependencies, &dependencies); err != nil {
				continue
			}

			for _, depID := range dependencies {
				if depID == task.ID {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// sortCandidates sorts task candidates according to selection policy
func (ts *TaskSelector) sortCandidates(candidates []*taskCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]

		// 1. Priority (higher priority first)
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}

		// 2. Leaf preference (if enabled)
		if ts.config.PreferLeafTasks {
			if a.IsLeaf != b.IsLeaf {
				return a.IsLeaf // prefer leaf tasks
			}
		}

		// 3. Tie breaker
		switch ts.config.TieBreaker {
		case "oldest_updated":
			return a.Task.UpdatedAt.Before(b.Task.UpdatedAt)
		case "newest_created":
			return a.Task.CreatedAt.After(b.Task.CreatedAt)
		case "alphabetical":
			return a.Task.Title < b.Task.Title
		default:
			return a.Task.UpdatedAt.Before(b.Task.UpdatedAt)
		}
	})
}

// buildSelectionReason builds a human-readable explanation of why a task was selected
func (ts *TaskSelector) buildSelectionReason(selected *taskCandidate, totalTasks, availableTasks int) string {
	reason := fmt.Sprintf("Selected from %d total tasks (%d available)", totalTasks, availableTasks)

	criteria := []string{}

	// Priority
	if selected.Priority > 5 {
		criteria = append(criteria, fmt.Sprintf("high priority (%d)", selected.Priority))
	} else if selected.Priority == 5 {
		criteria = append(criteria, fmt.Sprintf("normal priority (%d)", selected.Priority))
	} else {
		criteria = append(criteria, fmt.Sprintf("low priority (%d)", selected.Priority))
	}

	// Leaf status
	if ts.config.PreferLeafTasks && selected.IsLeaf {
		criteria = append(criteria, "leaf task (no dependents)")
	}

	// Dependencies
	var depCount int
	if len(selected.Task.Dependencies) > 0 {
		var deps []string
		if err := json.Unmarshal(selected.Task.Dependencies, &deps); err == nil {
			depCount = len(deps)
		}
	}

	if depCount > 0 {
		criteria = append(criteria, fmt.Sprintf("%d dependencies satisfied", depCount))
	} else {
		criteria = append(criteria, "no dependencies")
	}

	// Tie breaker
	switch ts.config.TieBreaker {
	case "oldest_updated":
		age := time.Since(selected.Task.UpdatedAt)
		if age > 24*time.Hour {
			criteria = append(criteria, fmt.Sprintf("oldest update (%dd ago)", int(age.Hours()/24)))
		} else {
			criteria = append(criteria, fmt.Sprintf("oldest update (%dh ago)", int(age.Hours())))
		}
	}

	if len(criteria) > 0 {
		reason += ": " + fmt.Sprintf("%s", criteria[0])
		for _, criterion := range criteria[1:] {
			reason += ", " + criterion
		}
	}

	return reason
}

// GetTaskStatus returns detailed status information for all tasks
func (ts *TaskSelector) GetTaskStatus() (map[string]interface{}, error) {
	allTasks, err := ts.store.ListTasks(storage.TaskFilters{})
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"total_tasks":    len(allTasks),
		"by_state":       make(map[string]int),
		"blocked_tasks":  []map[string]interface{}{},
		"ready_tasks":    []map[string]interface{}{},
		"completed_tasks": 0,
	}

	var blockedTasks []map[string]interface{}
	var readyTasks []map[string]interface{}

	for _, task := range allTasks {
		// Count by state
		stateCount := status["by_state"].(map[string]int)
		stateCount[string(task.State)]++

		// Count completed tasks
		if task.State == storage.Done {
			status["completed_tasks"] = status["completed_tasks"].(int) + 1
		}

		// Check if blocked
		if !IsTerminalState(task.State) {
			if blocked, reason := ts.isBlockedByDependencies(task); blocked {
				blockedTasks = append(blockedTasks, map[string]interface{}{
					"id":     task.ID,
					"title":  task.Title,
					"state":  task.State,
					"reason": reason,
				})
			} else {
				readyTasks = append(readyTasks, map[string]interface{}{
					"id":       task.ID,
					"title":    task.Title,
					"state":    task.State,
					"priority": task.Priority,
				})
			}
		}
	}

	status["blocked_tasks"] = blockedTasks
	status["ready_tasks"] = readyTasks

	return status, nil
}