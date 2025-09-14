package storage

import (
	"os"
	"testing"
	"time"
)

func TestCreateAndGetTask(t *testing.T) {
	// Create temporary database
	dbFile := "test_tasks.db"
	defer os.Remove(dbFile)

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a test task
	task := &Task{
		Title:       "Test Task",
		Description: "This is a test task",
		State:       ReadyForPlan,
		Priority:    5,
		Owner:       "test-user",
	}

	err = store.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify the task was created with an ID
	if task.ID == "" {
		t.Error("Task ID should not be empty after creation")
	}

	// Retrieve the task
	retrievedTask, err := store.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	// Verify the task data
	if retrievedTask.Title != task.Title {
		t.Errorf("Expected title %s, got %s", task.Title, retrievedTask.Title)
	}

	if retrievedTask.State != task.State {
		t.Errorf("Expected state %s, got %s", task.State, retrievedTask.State)
	}

	if retrievedTask.Priority != task.Priority {
		t.Errorf("Expected priority %d, got %d", task.Priority, retrievedTask.Priority)
	}
}

func TestUpdateTaskState(t *testing.T) {
	// Create temporary database
	dbFile := "test_update.db"
	defer os.Remove(dbFile)

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a test task
	task := &Task{
		Title: "Test Task",
		State: ReadyForPlan,
	}

	err = store.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Update the task state
	newState := Planning
	note := "Started planning phase"
	err = store.UpdateTaskState(task.ID, newState, note)
	if err != nil {
		t.Fatalf("Failed to update task state: %v", err)
	}

	// Retrieve and verify the updated task
	updatedTask, err := store.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if updatedTask.State != newState {
		t.Errorf("Expected state %s, got %s", newState, updatedTask.State)
	}

	// Verify updated_at was changed
	if !updatedTask.UpdatedAt.After(updatedTask.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestListTasks(t *testing.T) {
	// Create temporary database
	dbFile := "test_list.db"
	defer os.Remove(dbFile)

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create multiple test tasks
	tasks := []*Task{
		{Title: "Task 1", State: ReadyForPlan, Priority: 5},
		{Title: "Task 2", State: Planning, Priority: 8},
		{Title: "Task 3", State: ReadyForPlan, Priority: 3},
	}

	for _, task := range tasks {
		err = store.CreateTask(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	// Test listing all tasks
	allTasks, err := store.ListTasks(TaskFilters{})
	if err != nil {
		t.Fatalf("Failed to list all tasks: %v", err)
	}

	if len(allTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(allTasks))
	}

	// Test filtering by state
	state := ReadyForPlan
	filtered, err := store.ListTasks(TaskFilters{State: &state})
	if err != nil {
		t.Fatalf("Failed to list filtered tasks: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 tasks with state ready_for_plan, got %d", len(filtered))
	}

	// Verify tasks are sorted by priority DESC, then updated_at ASC
	if len(allTasks) >= 2 {
		if allTasks[0].Priority < allTasks[1].Priority {
			t.Error("Tasks should be sorted by priority DESC")
		}
	}
}

func TestRequirementOperations(t *testing.T) {
	// Create temporary database
	dbFile := "test_requirements.db"
	defer os.Remove(dbFile)

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a test requirement
	req := &Requirement{
		Key:   "FR-1",
		Title: "Test Requirement",
		Text:  "This is a test requirement",
		Type:  "functional",
	}

	err = store.CreateRequirement(req)
	if err != nil {
		t.Fatalf("Failed to create requirement: %v", err)
	}

	// Retrieve the requirement
	retrievedReq, err := store.GetRequirement(req.Key)
	if err != nil {
		t.Fatalf("Failed to get requirement: %v", err)
	}

	if retrievedReq.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, retrievedReq.Title)
	}

	if retrievedReq.Type != req.Type {
		t.Errorf("Expected type %s, got %s", req.Type, retrievedReq.Type)
	}

	// List requirements
	requirements, err := store.ListRequirements("")
	if err != nil {
		t.Fatalf("Failed to list requirements: %v", err)
	}

	if len(requirements) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(requirements))
	}
}

func TestArtifactOperations(t *testing.T) {
	// Create temporary database
	dbFile := "test_artifacts.db"
	defer os.Remove(dbFile)

	store, err := NewStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a test task first
	task := &Task{
		Title: "Test Task",
		State: Planning,
	}
	err = store.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create an artifact
	artifact := &Artifact{
		TaskID:  task.ID,
		Name:    "implementation_plan",
		Content: "# Implementation Plan\n\nThis is the plan...",
	}

	err = store.UpsertArtifact(artifact)
	if err != nil {
		t.Fatalf("Failed to create artifact: %v", err)
	}

	// Verify version was set
	if artifact.Version != 1 {
		t.Errorf("Expected version 1, got %d", artifact.Version)
	}

	// Retrieve the artifact
	retrievedArtifact, err := store.GetArtifact(task.ID, "implementation_plan", 0) // 0 = latest
	if err != nil {
		t.Fatalf("Failed to get artifact: %v", err)
	}

	if retrievedArtifact.Content != artifact.Content {
		t.Errorf("Expected content %s, got %s", artifact.Content, retrievedArtifact.Content)
	}

	// Create another version
	updatedArtifact := &Artifact{
		TaskID:  task.ID,
		Name:    "implementation_plan",
		Content: "# Updated Implementation Plan\n\nThis is updated...",
	}

	err = store.UpsertArtifact(updatedArtifact)
	if err != nil {
		t.Fatalf("Failed to update artifact: %v", err)
	}

	// Verify new version
	if updatedArtifact.Version != 2 {
		t.Errorf("Expected version 2, got %d", updatedArtifact.Version)
	}

	// List artifacts
	artifacts, err := store.ListArtifacts(task.ID)
	if err != nil {
		t.Fatalf("Failed to list artifacts: %v", err)
	}

	if len(artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(artifacts))
	}
}