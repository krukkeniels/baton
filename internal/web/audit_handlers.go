package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// handleAuditHistory handles GET /api/audit/{task_id}
func (s *Server) handleAuditHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/audit/")
	taskID := strings.Split(path, "/")[0]

	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Get audit history for the task
	entries, err := s.store.GetAuditHistory(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get audit history: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response []AuditHistoryEntry
	for _, entry := range entries {
		historyEntry := AuditHistoryEntry{
			ID:           entry.ID,
			TaskID:       entry.TaskID,
			PrevState:    entry.PrevState,
			NextState:    entry.NextState,
			Actor:        entry.Actor,
			Reason:       entry.SelectionReason,
			Note:         entry.Note,
			CreatedAt:    entry.CreatedAt,
			InputsSummary: entry.InputsSummary,
			OutputsSummary: entry.OutputsSummary,
		}

		// Parse commands if available
		if entry.Commands != nil {
			json.Unmarshal(entry.Commands, &historyEntry.Commands)
		}

		// Parse follow-ups if available
		if entry.FollowUps != nil {
			json.Unmarshal(entry.FollowUps, &historyEntry.FollowUps)
		}

		response = append(response, historyEntry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AuditHistoryEntry represents a single audit entry in the response
type AuditHistoryEntry struct {
	ID             string    `json:"id"`
	TaskID         string    `json:"task_id"`
	PrevState      string    `json:"prev_state"`
	NextState      string    `json:"next_state"`
	Actor          string    `json:"actor"`
	Reason         string    `json:"reason"`
	Note           string    `json:"note"`
	Commands       []string  `json:"commands,omitempty"`
	FollowUps      []string  `json:"follow_ups,omitempty"`
	InputsSummary  string    `json:"inputs_summary"`
	OutputsSummary string    `json:"outputs_summary"`
	CreatedAt      time.Time `json:"created_at"`
}