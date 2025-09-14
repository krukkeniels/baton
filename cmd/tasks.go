package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"baton/internal/statemachine"
	"baton/internal/storage"
)

// tasksCmd represents the tasks command
var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Task management commands",
	Long:  `Task management commands for listing, filtering, and manipulating tasks.`,
}

// tasksListCmd represents the tasks list command
var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List tasks with optional filtering by state, priority, owner, and tags.`,
	RunE:  runTasksList,
}

// tasksNextCmd represents the tasks next command
var tasksNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show next task selection",
	Long:  `Show which task would be selected next and the reasoning behind the selection.`,
	RunE:  runTasksNext,
}

// tasksUpdateCmd represents the tasks update command
var tasksUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update task state",
	Long:  `Manually update a task's state with validation and audit logging.`,
	RunE:  runTasksUpdate,
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksNextCmd)
	tasksCmd.AddCommand(tasksUpdateCmd)

	// List command flags
	tasksListCmd.Flags().String("state", "", "filter by state")
	tasksListCmd.Flags().Int("priority", -1, "filter by priority")
	tasksListCmd.Flags().String("owner", "", "filter by owner")
	tasksListCmd.Flags().Bool("json", false, "output in JSON format")

	// Update command flags
	tasksUpdateCmd.Flags().String("id", "", "task ID (required)")
	tasksUpdateCmd.Flags().String("state", "", "new state (required)")
	tasksUpdateCmd.Flags().String("note", "", "optional note")
	tasksUpdateCmd.MarkFlagRequired("id")
	tasksUpdateCmd.MarkFlagRequired("state")
}

func runTasksList(cmd *cobra.Command, args []string) error {
	// Initialize database
	store, err := storage.NewStore(globalConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Build filters
	filters := storage.TaskFilters{}

	if state, _ := cmd.Flags().GetString("state"); state != "" {
		normalizedState := storage.NormalizeState(state)
		filters.State = &normalizedState
	}

	if priority, _ := cmd.Flags().GetInt("priority"); priority >= 0 {
		filters.Priority = &priority
	}

	if owner, _ := cmd.Flags().GetString("owner"); owner != "" {
		filters.Owner = &owner
	}

	// Get tasks
	tasks, err := store.ListTasks(filters)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Check for JSON output
	if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
		data, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Printf("Found %d tasks:\n\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("📝 %s\n", task.ID)
		fmt.Printf("  Title: %s\n", task.Title)
		fmt.Printf("  State: %s\n", task.State)
		fmt.Printf("  Priority: %d\n", task.Priority)
		if task.Owner != "" {
			fmt.Printf("  Owner: %s\n", task.Owner)
		}
		if task.Description != "" {
			fmt.Printf("  Description: %s\n", task.Description)
		}
		fmt.Println()
	}

	return nil
}

func runTasksNext(cmd *cobra.Command, args []string) error {
	// Initialize database
	store, err := storage.NewStore(globalConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Create task selector
	selector := statemachine.NewTaskSelector(store, &globalConfig.Selection)

	// Get next task
	result, err := selector.SelectNext()
	if err != nil {
		return fmt.Errorf("failed to select next task: %w", err)
	}

	// Display result
	fmt.Println("🎯 Next Task Selection")
	fmt.Println("=====================")
	fmt.Printf("Task ID: %s\n", result.Task.ID)
	fmt.Printf("Title: %s\n", result.Task.Title)
	fmt.Printf("State: %s\n", result.Task.State)
	fmt.Printf("Priority: %d\n", result.Task.Priority)
	fmt.Printf("\n🧐 Selection Reasoning:\n%s\n", result.Reason)

	return nil
}

func runTasksUpdate(cmd *cobra.Command, args []string) error {
	taskID, _ := cmd.Flags().GetString("id")
	stateStr, _ := cmd.Flags().GetString("state")
	note, _ := cmd.Flags().GetString("note")

	// Initialize database
	store, err := storage.NewStore(globalConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	// Normalize state
	newState := storage.NormalizeState(stateStr)

	// Create validator
	validator := statemachine.NewTransitionValidator(store)

	// Perform the update
	if err := validator.ValidateAndTransition(taskID, newState, note); err != nil {
		return fmt.Errorf("failed to update task state: %w", err)
	}

	fmt.Printf("✅ Task %s updated to state: %s\n", taskID, newState)
	if note != "" {
		fmt.Printf("Note: %s\n", note)
	}

	return nil
}