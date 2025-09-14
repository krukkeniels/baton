package audit

import (
	"encoding/json"
	"fmt"

	"baton/internal/storage"
)

// Logger handles audit trail logging
type Logger struct {
	store *storage.Store
}

// NewLogger creates a new audit logger
func NewLogger(store *storage.Store) *Logger {
	return &Logger{
		store: store,
	}
}

// LogCycle logs a complete cycle execution
func (al *Logger) LogCycle(entry *storage.AuditLog) error {
	return al.store.CreateAuditLog(entry)
}

// LogStateTransition logs a state transition
func (al *Logger) LogStateTransition(taskID, actor string, prevState, nextState storage.State, reason string) error {
	entry := &storage.AuditLog{
		TaskID:    taskID,
		CycleID:   "manual", // For manual transitions
		PrevState: string(prevState),
		NextState: string(nextState),
		Actor:     actor,
		Note:      reason,
		Result:    "success",
	}

	return al.store.CreateAuditLog(entry)
}

// LogError logs an error during cycle execution
func (al *Logger) LogError(taskID, cycleID, actor string, err error, context map[string]interface{}) error {
	contextJSON, _ := json.Marshal(context)

	entry := &storage.AuditLog{
		TaskID:  taskID,
		CycleID: cycleID,
		Actor:   actor,
		Result:  "error",
		Note:    err.Error(),
		InputsSummary: string(contextJSON),
	}

	return al.store.CreateAuditLog(entry)
}

// GetTaskHistory returns the audit history for a task
func (al *Logger) GetTaskHistory(taskID string) ([]*storage.AuditLog, error) {
	return al.store.GetAuditLogs(taskID)
}

// GetRecentCycles returns recent cycle executions
func (al *Logger) GetRecentCycles(limit int) ([]*storage.AuditLog, error) {
	// This would require a new store method, for now return empty
	return []*storage.AuditLog{}, nil
}

// GenerateReport generates a summary report of recent activity
func (al *Logger) GenerateReport() (map[string]interface{}, error) {
	// Get all tasks to analyze
	tasks, err := al.store.ListTasks(storage.TaskFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for report: %w", err)
	}

	report := map[string]interface{}{
		"total_tasks":      len(tasks),
		"by_state":         make(map[string]int),
		"recent_activity":  []map[string]interface{}{},
		"completion_rate":  0.0,
	}

	// Count by state
	stateCount := make(map[string]int)
	completedCount := 0

	for _, task := range tasks {
		stateCount[string(task.State)]++
		if task.State == storage.Done {
			completedCount++
		}
	}

	report["by_state"] = stateCount
	if len(tasks) > 0 {
		report["completion_rate"] = float64(completedCount) / float64(len(tasks)) * 100.0
	}

	return report, nil
}