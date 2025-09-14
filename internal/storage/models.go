package storage

import (
	"encoding/json"
	"time"
)

// State represents the valid task states
type State string

const (
	ReadyForPlan           State = "ready_for_plan"
	Planning               State = "planning"
	ReadyForImplementation State = "ready_for_implementation"
	Implementing           State = "implementing"
	ReadyForCodeReview     State = "ready_for_code_review"
	Reviewing              State = "reviewing"
	ReadyForCommit         State = "ready_for_commit"
	NeedsFixes             State = "needs_fixes"
	Committing             State = "committing"
	Fixing                 State = "fixing"
	Done                   State = "DONE"
)

// StateAliases maps common typos to correct states
var StateAliases = map[string]State{
	"ready_for_implmentation": ReadyForImplementation,
	"ready_for_code_revie":    ReadyForCodeReview,
	"need_fixes":              NeedsFixes,
	"commiting":               Committing,
}

// NormalizeState normalizes input state aliases to canonical states
func NormalizeState(input string) State {
	if alias, exists := StateAliases[input]; exists {
		return alias
	}
	return State(input)
}

// Task represents a unit of work
type Task struct {
	ID           string          `json:"id" db:"id"`
	Title        string          `json:"title" db:"title"`
	Description  string          `json:"description" db:"description"`
	State        State           `json:"state" db:"state"`
	Priority     int             `json:"priority" db:"priority"`
	Owner        string          `json:"owner" db:"owner"`
	Tags         json.RawMessage `json:"tags" db:"tags"`         // JSON array
	Dependencies json.RawMessage `json:"dependencies" db:"dependencies"` // JSON array of task IDs
	BlockedBy    json.RawMessage `json:"blocked_by" db:"blocked_by"`    // JSON array of task IDs
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// Requirement represents a functional or non-functional requirement
type Requirement struct {
	ID        string    `json:"id" db:"id"`
	Key       string    `json:"key" db:"key"` // e.g., "FR-P1"
	Title     string    `json:"title" db:"title"`
	Text      string    `json:"text" db:"text"`
	Type      string    `json:"type" db:"type"` // functional|nonfunctional|constraint|risk
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Artifact represents task-scoped documents (implementation plans, etc.)
type Artifact struct {
	ID        string          `json:"id" db:"id"`
	TaskID    string          `json:"task_id" db:"task_id"`
	Name      string          `json:"name" db:"name"` // implementation_plan, change_summary, etc.
	Version   int             `json:"version" db:"version"`
	Content   string          `json:"content" db:"content"`
	Meta      json.RawMessage `json:"meta" db:"meta"` // JSON metadata
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// Agent represents a role configuration
type Agent struct {
	ID            string          `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Role          string          `json:"role" db:"role"`
	Description   string          `json:"description" db:"description"`
	RoutingPolicy json.RawMessage `json:"routing_policy" db:"routing_policy"` // JSON configuration
	Permissions   json.RawMessage `json:"permissions" db:"permissions"`       // JSON permissions
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// AuditLog represents a cycle execution audit entry
type AuditLog struct {
	ID              string          `json:"id" db:"id"`
	TaskID          string          `json:"task_id" db:"task_id"`
	CycleID         string          `json:"cycle_id" db:"cycle_id"`
	PrevState       string          `json:"prev_state" db:"prev_state"`
	NextState       string          `json:"next_state" db:"next_state"`
	Actor           string          `json:"actor" db:"actor"` // agent ID
	SelectionReason string          `json:"selection_reason" db:"selection_reason"`
	InputsSummary   string          `json:"inputs_summary" db:"inputs_summary"`   // JSON: plan hash, requirements, artifacts used
	OutputsSummary  string          `json:"outputs_summary" db:"outputs_summary"` // JSON: handovers created/updated
	Commands        json.RawMessage `json:"commands" db:"commands"`               // JSON array of external commands executed
	Result          string          `json:"result" db:"result"`
	Note            string          `json:"note" db:"note"`
	FollowUps       json.RawMessage `json:"follow_ups" db:"follow_ups"` // JSON array of follow-up interactions
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// TaskFilters represents filters for task queries
type TaskFilters struct {
	State    *State  `json:"state,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Owner    *string `json:"owner,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// CycleResult represents the outcome of a cycle execution
type CycleResult struct {
	Success         bool          `json:"success"`
	TaskID          string        `json:"task_id"`
	PrevState       State         `json:"prev_state"`
	NextState       State         `json:"next_state"`
	ArtifactsCreated []string      `json:"artifacts_created"`
	Duration        time.Duration `json:"duration"`
	Error           error         `json:"error,omitempty"`
}