package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// Web UI specific storage methods

// GetTaskCount returns the count of tasks matching the given filters
func (s *Store) GetTaskCount(filters TaskFilters) (int, error) {
	query := "SELECT COUNT(*) FROM tasks WHERE 1=1"
	args := []interface{}{}

	if filters.State != nil {
		query += " AND state = ?"
		args = append(args, *filters.State)
	}

	if filters.Priority != nil {
		query += " AND priority = ?"
		args = append(args, *filters.Priority)
	}

	if filters.Owner != nil {
		query += " AND owner = ?"
		args = append(args, *filters.Owner)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID             string         `json:"id" db:"id"`
	TaskID         string         `json:"task_id" db:"task_id"`
	PrevState      string         `json:"prev_state" db:"prev_state"`
	NextState      string         `json:"next_state" db:"next_state"`
	Actor          string         `json:"actor" db:"actor"`
	SelectionReason string        `json:"selection_reason" db:"selection_reason"`
	Note           string         `json:"note" db:"note"`
	Commands       []byte         `json:"commands" db:"commands"`       // JSON array
	FollowUps      []byte         `json:"follow_ups" db:"follow_ups"`   // JSON array
	InputsSummary  string         `json:"inputs_summary" db:"inputs_summary"`
	OutputsSummary string         `json:"outputs_summary" db:"outputs_summary"`
	Result         string         `json:"result" db:"result"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

// GetRecentAuditEntries returns the most recent audit entries across all tasks
func (s *Store) GetRecentAuditEntries(limit int) ([]AuditEntry, error) {
	query := `
		SELECT a.id, a.task_id, t.title as task_title, a.prev_state, a.next_state,
		       a.actor, a.created_at
		FROM audit_logs a
		LEFT JOIN tasks t ON a.task_id = t.id
		ORDER BY a.created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent audit entries: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var entry AuditEntry
		var taskTitle sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&taskTitle,
			&entry.PrevState,
			&entry.NextState,
			&entry.Actor,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit entry: %w", err)
		}

		// Use task title as a display field (we'll add this to the struct)
		if taskTitle.Valid {
			entry.SelectionReason = taskTitle.String // Temporarily use this field
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// GetAuditHistory returns the complete audit history for a specific task
func (s *Store) GetAuditHistory(taskID string) ([]AuditEntry, error) {
	query := `
		SELECT id, task_id, prev_state, next_state, actor, selection_reason,
		       note, commands, follow_ups, inputs_summary, outputs_summary,
		       result, created_at
		FROM audit_logs
		WHERE task_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit history: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var entry AuditEntry

		err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&entry.PrevState,
			&entry.NextState,
			&entry.Actor,
			&entry.SelectionReason,
			&entry.Note,
			&entry.Commands,
			&entry.FollowUps,
			&entry.InputsSummary,
			&entry.OutputsSummary,
			&entry.Result,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit entry: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// UpdateTask updates an existing task
func (s *Store) UpdateTask(task *Task) error {
	task.UpdatedAt = time.Now()

	query := `
		UPDATE tasks
		SET title = ?, description = ?, state = ?, priority = ?, owner = ?,
		    tags = ?, dependencies = ?, blocked_by = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		task.Title, task.Description, task.State, task.Priority, task.Owner,
		task.Tags, task.Dependencies, task.BlockedBy, task.UpdatedAt, task.ID)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

// Error definitions
var (
	ErrTaskNotFound = fmt.Errorf("task not found")
)