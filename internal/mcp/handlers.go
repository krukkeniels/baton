package mcp

import (
	"encoding/json"
	"io"
	"os"

	"baton/internal/statemachine"
	"baton/internal/storage"
)

// TaskHandler handles task-related MCP operations
type TaskHandler struct {
	store     *storage.Store
	selector  *statemachine.TaskSelector
	validator *statemachine.TransitionValidator
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(store *storage.Store, selector *statemachine.TaskSelector, validator *statemachine.TransitionValidator) *TaskHandler {
	return &TaskHandler{
		store:     store,
		selector:  selector,
		validator: validator,
	}
}

// GetNext handles baton.tasks.get_next
func (h *TaskHandler) GetNext(req *JSONRPCRequest) *JSONRPCResponse {
	result, err := h.selector.SelectNext()
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to select next task", err.Error())
	}

	// Include artifacts in the response
	artifacts, err := h.store.ListArtifacts(result.Task.ID)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to get task artifacts", err.Error())
	}

	response := map[string]interface{}{
		"task": map[string]interface{}{
			"id":           result.Task.ID,
			"title":        result.Task.Title,
			"description":  result.Task.Description,
			"state":        result.Task.State,
			"priority":     result.Task.Priority,
			"owner":        result.Task.Owner,
			"tags":         result.Task.Tags,
			"dependencies": result.Task.Dependencies,
			"blocked_by":   result.Task.BlockedBy,
			"created_at":   result.Task.CreatedAt,
			"updated_at":   result.Task.UpdatedAt,
			"artifacts":    artifacts,
		},
		"selection_reason": result.Reason,
	}

	return NewJSONRPCResponse(req.ID, response)
}

// Get handles baton.tasks.get
func (h *TaskHandler) Get(req *JSONRPCRequest) *JSONRPCResponse {
	taskID, err := req.GetStringParam("task_id")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing task_id parameter", nil)
	}

	task, err := h.store.GetTask(taskID)
	if err != nil {
		return NewJSONRPCError(req.ID, ResourceNotFound, "Task not found", map[string]interface{}{"task_id": taskID})
	}

	// Include artifacts
	artifacts, err := h.store.ListArtifacts(task.ID)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to get task artifacts", err.Error())
	}

	response := map[string]interface{}{
		"id":           task.ID,
		"title":        task.Title,
		"description":  task.Description,
		"state":        task.State,
		"priority":     task.Priority,
		"owner":        task.Owner,
		"tags":         task.Tags,
		"dependencies": task.Dependencies,
		"blocked_by":   task.BlockedBy,
		"created_at":   task.CreatedAt,
		"updated_at":   task.UpdatedAt,
		"artifacts":    artifacts,
	}

	return NewJSONRPCResponse(req.ID, response)
}

// UpdateState handles baton.tasks.update_state
func (h *TaskHandler) UpdateState(req *JSONRPCRequest) *JSONRPCResponse {
	taskID, err := req.GetStringParam("task_id")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing task_id parameter", nil)
	}

	stateStr, err := req.GetStringParam("state")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing state parameter", nil)
	}

	note, _ := req.GetOptionalStringParam("note")

	// Normalize and validate state
	newState := storage.NormalizeState(stateStr)

	// Perform the transition
	if err := h.validator.ValidateAndTransition(taskID, newState, note); err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "State transition failed", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"success": true,
		"task_id": taskID,
		"state":   newState,
	})
}

// AppendNote handles baton.tasks.append_note
func (h *TaskHandler) AppendNote(req *JSONRPCRequest) *JSONRPCResponse {
	taskID, err := req.GetStringParam("task_id")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing task_id parameter", nil)
	}

	note, err := req.GetStringParam("note")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing note parameter", nil)
	}

	// Get current task to maintain state
	task, err := h.store.GetTask(taskID)
	if err != nil {
		return NewJSONRPCError(req.ID, ResourceNotFound, "Task not found", map[string]interface{}{"task_id": taskID})
	}

	// Update with note (keeps same state)
	if err := h.store.UpdateTaskState(taskID, task.State, note); err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to append note", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"success": true,
		"task_id": taskID,
	})
}

// List handles baton.tasks.list
func (h *TaskHandler) List(req *JSONRPCRequest) *JSONRPCResponse {
	params, err := req.GetParams()
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Invalid parameters", nil)
	}

	filters := storage.TaskFilters{}

	// Parse filters
	if stateStr, ok := params["state"].(string); ok {
		state := storage.NormalizeState(stateStr)
		filters.State = &state
	}

	if priority, ok := params["priority"].(float64); ok {
		p := int(priority)
		filters.Priority = &p
	}

	if owner, ok := params["owner"].(string); ok {
		filters.Owner = &owner
	}

	tasks, err := h.store.ListTasks(filters)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to list tasks", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// ArtifactHandler handles artifact-related MCP operations
type ArtifactHandler struct {
	store *storage.Store
}

// NewArtifactHandler creates a new artifact handler
func NewArtifactHandler(store *storage.Store) *ArtifactHandler {
	return &ArtifactHandler{store: store}
}

// Upsert handles baton.artifacts.upsert
func (h *ArtifactHandler) Upsert(req *JSONRPCRequest) *JSONRPCResponse {
	taskID, err := req.GetStringParam("task_id")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing task_id parameter", nil)
	}

	name, err := req.GetStringParam("name")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing name parameter", nil)
	}

	content, err := req.GetStringParam("content")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing content parameter", nil)
	}

	params, _ := req.GetParams()
	var meta json.RawMessage
	if metaData, ok := params["meta"]; ok {
		if metaBytes, err := json.Marshal(metaData); err == nil {
			meta = metaBytes
		}
	}

	artifact := &storage.Artifact{
		TaskID:  taskID,
		Name:    name,
		Content: content,
		Meta:    meta,
	}

	if err := h.store.UpsertArtifact(artifact); err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to upsert artifact", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"success": true,
		"artifact": map[string]interface{}{
			"id":      artifact.ID,
			"task_id": artifact.TaskID,
			"name":    artifact.Name,
			"version": artifact.Version,
		},
	})
}

// Get handles baton.artifacts.get
func (h *ArtifactHandler) Get(req *JSONRPCRequest) *JSONRPCResponse {
	taskID, err := req.GetStringParam("task_id")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing task_id parameter", nil)
	}

	name, err := req.GetStringParam("name")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing name parameter", nil)
	}

	version := 0 // Default to latest
	if v, err := req.GetIntParam("version"); err == nil {
		version = v
	}

	artifact, err := h.store.GetArtifact(taskID, name, version)
	if err != nil {
		return NewJSONRPCError(req.ID, ResourceNotFound, "Artifact not found", map[string]interface{}{
			"task_id": taskID,
			"name":    name,
			"version": version,
		})
	}

	return NewJSONRPCResponse(req.ID, artifact)
}

// List handles baton.artifacts.list
func (h *ArtifactHandler) List(req *JSONRPCRequest) *JSONRPCResponse {
	taskID, err := req.GetStringParam("task_id")
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Missing task_id parameter", nil)
	}

	artifacts, err := h.store.ListArtifacts(taskID)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to list artifacts", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"artifacts": artifacts,
		"count":     len(artifacts),
	})
}

// RequirementHandler handles requirement-related MCP operations
type RequirementHandler struct {
	store *storage.Store
}

// NewRequirementHandler creates a new requirement handler
func NewRequirementHandler(store *storage.Store) *RequirementHandler {
	return &RequirementHandler{store: store}
}

// List handles baton.requirements.list
func (h *RequirementHandler) List(req *JSONRPCRequest) *JSONRPCResponse {
	params, err := req.GetParams()
	if err != nil {
		return NewJSONRPCError(req.ID, InvalidParams, "Invalid parameters", nil)
	}

	reqType := ""
	if t, ok := params["type"].(string); ok {
		reqType = t
	}

	requirements, err := h.store.ListRequirements(reqType)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to list requirements", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"requirements": requirements,
		"count":        len(requirements),
	})
}

// PlanHandler handles plan-related MCP operations
type PlanHandler struct {
	planFile string
}

// NewPlanHandler creates a new plan handler
func NewPlanHandler(planFile string) *PlanHandler {
	return &PlanHandler{planFile: planFile}
}

// Read handles baton.plan.read
func (h *PlanHandler) Read(req *JSONRPCRequest) *JSONRPCResponse {
	// Check if plan file exists
	if _, err := os.Stat(h.planFile); os.IsNotExist(err) {
		return NewJSONRPCError(req.ID, ResourceNotFound, "Plan file not found", map[string]interface{}{
			"path": h.planFile,
		})
	}

	// Read the file
	file, err := os.Open(h.planFile)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to open plan file", err.Error())
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to read plan file", err.Error())
	}

	// Get file info for metadata
	info, err := file.Stat()
	if err != nil {
		return NewJSONRPCError(req.ID, InternalError, "Failed to get plan file info", err.Error())
	}

	return NewJSONRPCResponse(req.ID, map[string]interface{}{
		"content":     string(content),
		"path":        h.planFile,
		"size":        info.Size(),
		"modified_at": info.ModTime(),
	})
}