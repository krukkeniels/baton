package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// Store represents the SQLite database storage
type Store struct {
	db *sql.DB
}

// NewStore creates a new SQLite store
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	store := &Store{db: db}

	// Run migrations
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// migrate runs the database migrations
func (s *Store) migrate() error {
	_, err := s.db.Exec(CreateTablesSQL)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Task operations
func (s *Store) CreateTask(task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	query := `
		INSERT INTO tasks (id, title, description, state, priority, owner, tags, dependencies, blocked_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, task.ID, task.Title, task.Description, task.State, task.Priority,
		task.Owner, task.Tags, task.Dependencies, task.BlockedBy, task.CreatedAt, task.UpdatedAt)

	return err
}

func (s *Store) GetTask(id string) (*Task, error) {
	query := `
		SELECT id, title, description, state, priority, owner, tags, dependencies, blocked_by, created_at, updated_at
		FROM tasks WHERE id = ?
	`

	task := &Task{}
	err := s.db.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.State, &task.Priority,
		&task.Owner, &task.Tags, &task.Dependencies, &task.BlockedBy,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Store) UpdateTaskState(id string, state State, note string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update task state
	_, err = tx.Exec("UPDATE tasks SET state = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", state, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) ListTasks(filters TaskFilters) ([]*Task, error) {
	query := "SELECT id, title, description, state, priority, owner, tags, dependencies, blocked_by, created_at, updated_at FROM tasks WHERE 1=1"
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

	query += " ORDER BY priority DESC, updated_at ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.State, &task.Priority,
			&task.Owner, &task.Tags, &task.Dependencies, &task.BlockedBy,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// Requirement operations
func (s *Store) CreateRequirement(req *Requirement) error {
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	query := `
		INSERT INTO requirements (id, key, title, text, type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, req.ID, req.Key, req.Title, req.Text, req.Type, req.CreatedAt, req.UpdatedAt)
	return err
}

func (s *Store) GetRequirement(key string) (*Requirement, error) {
	query := `
		SELECT id, key, title, text, type, created_at, updated_at
		FROM requirements WHERE key = ?
	`

	req := &Requirement{}
	err := s.db.QueryRow(query, key).Scan(
		&req.ID, &req.Key, &req.Title, &req.Text, &req.Type, &req.CreatedAt, &req.UpdatedAt,
	)

	return req, err
}

func (s *Store) ListRequirements(reqType string) ([]*Requirement, error) {
	query := "SELECT id, key, title, text, type, created_at, updated_at FROM requirements"
	args := []interface{}{}

	if reqType != "" {
		query += " WHERE type = ?"
		args = append(args, reqType)
	}

	query += " ORDER BY key"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requirements []*Requirement
	for rows.Next() {
		req := &Requirement{}
		err := rows.Scan(&req.ID, &req.Key, &req.Title, &req.Text, &req.Type, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, req)
	}

	return requirements, rows.Err()
}

func (s *Store) UpdateRequirement(req *Requirement) error {
	query := `
		UPDATE requirements
		SET title = ?, text = ?, type = ?, updated_at = CURRENT_TIMESTAMP
		WHERE key = ?
	`

	_, err := s.db.Exec(query, req.Title, req.Text, req.Type, req.Key)
	return err
}

// Artifact operations
func (s *Store) UpsertArtifact(artifact *Artifact) error {
	if artifact.ID == "" {
		artifact.ID = uuid.New().String()
	}
	artifact.CreatedAt = time.Now()

	// Get the next version number for this task/name combination
	var maxVersion int
	err := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM artifacts WHERE task_id = ? AND name = ?",
		artifact.TaskID, artifact.Name).Scan(&maxVersion)
	if err != nil {
		return err
	}

	artifact.Version = maxVersion + 1

	query := `
		INSERT INTO artifacts (id, task_id, name, version, content, meta, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, artifact.ID, artifact.TaskID, artifact.Name, artifact.Version,
		artifact.Content, artifact.Meta, artifact.CreatedAt)

	return err
}

func (s *Store) GetArtifact(taskID, name string, version int) (*Artifact, error) {
	query := `
		SELECT id, task_id, name, version, content, meta, created_at
		FROM artifacts WHERE task_id = ? AND name = ? AND version = ?
	`

	if version == 0 {
		// Get latest version
		query = `
			SELECT id, task_id, name, version, content, meta, created_at
			FROM artifacts WHERE task_id = ? AND name = ?
			ORDER BY version DESC LIMIT 1
		`
	}

	artifact := &Artifact{}
	var err error

	if version == 0 {
		err = s.db.QueryRow(query, taskID, name).Scan(
			&artifact.ID, &artifact.TaskID, &artifact.Name, &artifact.Version,
			&artifact.Content, &artifact.Meta, &artifact.CreatedAt,
		)
	} else {
		err = s.db.QueryRow(query, taskID, name, version).Scan(
			&artifact.ID, &artifact.TaskID, &artifact.Name, &artifact.Version,
			&artifact.Content, &artifact.Meta, &artifact.CreatedAt,
		)
	}

	return artifact, err
}

func (s *Store) ListArtifacts(taskID string) ([]*Artifact, error) {
	query := `
		SELECT id, task_id, name, version, content, meta, created_at
		FROM artifacts WHERE task_id = ? ORDER BY name, version DESC
	`

	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []*Artifact
	for rows.Next() {
		artifact := &Artifact{}
		err := rows.Scan(&artifact.ID, &artifact.TaskID, &artifact.Name, &artifact.Version,
			&artifact.Content, &artifact.Meta, &artifact.CreatedAt)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, rows.Err()
}

// Audit operations
func (s *Store) CreateAuditLog(log *AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO audit_logs (id, task_id, cycle_id, prev_state, next_state, actor,
			selection_reason, inputs_summary, outputs_summary, commands, result, note, follow_ups, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, log.ID, log.TaskID, log.CycleID, log.PrevState, log.NextState,
		log.Actor, log.SelectionReason, log.InputsSummary, log.OutputsSummary, log.Commands,
		log.Result, log.Note, log.FollowUps, log.CreatedAt)

	return err
}

func (s *Store) GetAuditLogs(taskID string) ([]*AuditLog, error) {
	query := `
		SELECT id, task_id, cycle_id, prev_state, next_state, actor, selection_reason,
			inputs_summary, outputs_summary, commands, result, note, follow_ups, created_at
		FROM audit_logs WHERE task_id = ? ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		err := rows.Scan(&log.ID, &log.TaskID, &log.CycleID, &log.PrevState, &log.NextState,
			&log.Actor, &log.SelectionReason, &log.InputsSummary, &log.OutputsSummary, &log.Commands,
			&log.Result, &log.Note, &log.FollowUps, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}